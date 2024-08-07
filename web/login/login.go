package login

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"web/token"
)

// Add a global variable to hold the database connection
var db *sql.DB

// InitDB initializes the database connection
func InitDB(database *sql.DB) {
	db = database
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Check if user is already logged in
		cookie, err := r.Cookie("session_token")
		if err == nil {
			// Verify token from the cookie
			_, err := token.VerifyToken(cookie.Value, "session")
			if err == nil {
				// If token is valid, redirect to index
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		}
		serveTemplate(w, "login.html")
	} else if r.Method == http.MethodPost {
		// Process login form submission
		email := r.FormValue("email")
		password := r.FormValue("password")

		data := struct {
			ErrorMessage string
		}{}
		isValidUser, emailVerified, agreePrivacy, err := validateUser(email, password)
		if err == nil {
			if isValidUser {
				userID := getUserID(email)
				if !emailVerified {
					// Delete the token
					_, err = db.Exec("DELETE FROM tokens WHERE user_id = ? AND purpose = 'resend_email'", userID)
					if err != nil && err != sql.ErrNoRows {
						data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1004"
						serveTemplateWithData(w, "error_page.html", data)
						return
					}

					// Redirect to email verification page
					resendEmailToken, err := token.GenerateToken("resend_email", userID, 1*time.Hour)
					if err != nil {
						data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1003"
						serveTemplateWithData(w, "error_page.html", data)
						return
					}

					// Set the token in a cookie
					http.SetCookie(w, &http.Cookie{
						Name:    "rem",
						Value:   resendEmailToken,
						Expires: time.Now().Add(1 * time.Hour),
					})

					serveTemplateWithData(w, "email_verification.html", nil)
					return
				}
				if !agreePrivacy {
					privacyToken, err := token.GenerateToken("privacy_agreement", getUserID(email), 1*time.Hour)
					if err != nil {
						data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1003"
						serveTemplateWithData(w, "error_page.html", data)
						return
					}
					http.SetCookie(w, &http.Cookie{
						Name:    "privacy_agreement",
						Value:   privacyToken,
						Expires: time.Now().Add(1 * time.Hour),
					})
					http.Redirect(w, r, "/privacy", http.StatusSeeOther)
					return
				}
				// Generate a new token for the user
				sessionToken, err := token.GenerateToken("session", getUserID(email), 24*time.Hour)
				if err != nil {
					data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1003"
					serveTemplateWithData(w, "error_page.html", data)
					return
				}

				// Set the token in a cookie
				http.SetCookie(w, &http.Cookie{
					Name:    "session_token",
					Value:   sessionToken,
					Expires: time.Now().Add(24 * time.Hour),
				})

				// Redirect to index
				http.Redirect(w, r, "/", http.StatusSeeOther)
			} else {
				// Invalid credentials
				data.ErrorMessage = "電子郵件或密碼錯誤！"
				serveTemplateWithData(w, "login.html", data)
				return
			}
		} else {
			if err == sql.ErrNoRows {
				data.ErrorMessage = "電子郵件或密碼錯誤！"
				serveTemplateWithData(w, "login.html", data)
				return
			}
			data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1000"
			serveTemplateWithData(w, "error_page.html", data)
			return
		}
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		// Delete token from the cookie
		token.DeleteToken(cookie.Value, "session")
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return
}

func PrivacyHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("privacy_agreement")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		serveTemplate(w, "privacy.html")
	} else if r.Method == http.MethodPost {
		consent := r.FormValue("consent")
		data := struct {
			ErrorMessage string
		}{}
		if consent != "on" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		tokenValue := cookie.Value
		userID, err := token.VerifyToken(tokenValue, "privacy_agreement")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		_, err = db.Exec("UPDATE users SET agree_privacy = 1 WHERE user_id = ?", userID)
		if err != nil {
			data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1005"
			serveTemplateWithData(w, "error_page.html", data)
		}

		_, err = db.Exec("DELETE FROM tokens WHERE token = ? AND purpose = 'privacy_agreement'", tokenValue)
		if err != nil {
			data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1004"
			serveTemplateWithData(w, "error_page.html", data)
		}

		sessionToken, err := token.GenerateToken("session", userID, 24*time.Hour)
		if err != nil {
			data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1003"
			serveTemplateWithData(w, "error_page.html", data)
			return
		}

		// Set the token in a cookie
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: time.Now().Add(24 * time.Hour),
		})

		// Redirect to index
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// VerifyHandler handles email verification requests
func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the token from the request
	queryParams := r.URL.Query()
	tokenValue := queryParams.Get("token")

	data := struct {
		ErrorMessage string
	}{}

	if tokenValue == "" {
		data.ErrorMessage = "缺少密鑰"
		serveTemplateWithData(w, "error_page.html", data)
		return
	}

	// Verify the token
	userID, err := token.VerifyToken(tokenValue, "email_verification")
	if err != nil {
		data.ErrorMessage = "非法或過期的密鑰"
		serveTemplateWithData(w, "error_page.html", data)
		return
	}

	// Update the user's email_verified status
	_, err = db.Exec("UPDATE users SET email_verified = 1 WHERE user_id = ?", userID)
	if err != nil {
		data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1005"
		serveTemplateWithData(w, "error_page.html", data)
		return
	}

	// Delete the token
	_, err = db.Exec("DELETE FROM tokens WHERE user_id = ? AND purpose = 'email_verification'", userID)
	if err != nil && err != sql.ErrNoRows {
		data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1004"
		serveTemplateWithData(w, "error_page.html", data)
		return
	}

	_, err = db.Exec("DELETE FROM tokens WHERE user_id = ? AND purpose = 'resend_email'", userID)
	if err != nil && err != sql.ErrNoRows {
		data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1004"
		serveTemplateWithData(w, "error_page.html", data)
		return
	}

	// Redirect to login page after successful verification
	serveTemplate(w, "email_verification_success.html")
}

func validateUser(email, password string) (bool, bool, bool, error) {
	// Check the email, password, email_verified, and agree_privacy against the database
	var storedPassword, salt string
	var emailVerified, agreePrivacy bool
	err := db.QueryRow("SELECT password, salt, email_verified, agree_privacy FROM users WHERE email = ?", email).
		Scan(&storedPassword, &salt, &emailVerified, &agreePrivacy)
	if err != nil {
		return false, false, false, err
	}

	// Hash the provided password with the stored salt
	hashedPassword := hashPassword(password, salt)

	return hashedPassword == storedPassword, emailVerified, agreePrivacy, nil
}

func getUserID(email string) int {
	var userID int
	err := db.QueryRow("SELECT user_id FROM users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		return -1
	}
	return userID
}

func hashPassword(password, salt string) string {
	hash := sha256.New()
	hash.Write([]byte(password + salt))
	return hex.EncodeToString(hash.Sum(nil))
}

func serveTemplate(w http.ResponseWriter, filename string) {
	serveTemplateWithData(w, filename, nil)
}

func serveTemplateWithData(w http.ResponseWriter, filename string, data interface{}) {
	// Ensure the security of the file name to prevent path traversal attacks
	if filepath.Clean(filename) != filename {
		http.Error(w, "內部錯誤，請聯絡團隊！\nerror code: 1006", http.StatusBadRequest)
		return
	}

	// Build the file path
	filePath := filepath.Join("tem", filename)

	// Parse the template
	tmpl, err := template.ParseFiles(filePath)
	if err != nil {
		http.Error(w, "內部錯誤，請聯絡團隊！\nerror code: 1006", http.StatusInternalServerError)
		return
	}

	// Render the template
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "內部錯誤，請聯絡團隊！\nerror code: 1006", http.StatusInternalServerError)
	}
}
