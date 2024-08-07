package registration

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"web/mailer"
	"web/token"

	_ "github.com/go-sql-driver/mysql"
)

// Add a global variable to hold the database connection
var db *sql.DB

// InitDB initializes the database connection
func InitDB(database *sql.DB) {
	db = database
}

// RegisterHandler handles the registration request
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		data := struct {
			ErrorMessage string
		}{}

		if password != confirmPassword {
			data.ErrorMessage = "密碼不相符！"
			serveTemplateWithData(w, "register.html", data)
			return
		}

		// Check if email already exists
		exists, err := emailExists(email)
		if err != nil {
			data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1000"
			serveTemplateWithData(w, "error_page.html", data)
			return
		}
		if exists {
			data.ErrorMessage = "Email 已經註冊！"
			serveTemplateWithData(w, "register.html", data)
			return
		}

		// Generate salt
		salt := generateSalt()

		// Encrypt password (with salt)
		hashedPassword := hashPassword(password, salt)

		// Insert user information into the database
		err = insertUser(email, hashedPassword, salt)
		if err != nil {
			data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1001"
			serveTemplateWithData(w, "error_page.html", data)
			return
		}

		var userID int
		err = db.QueryRow("SELECT user_id FROM users WHERE email = ?", email).Scan(&userID)
		if err != nil {
			data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1001"
			serveTemplateWithData(w, "error_page.html", data)
			return
		}

		// Generate verification token
		emailVerificationToken, err := token.GenerateToken("email_verification", userID, 15*time.Minute)
		if err != nil {
			data := struct {
				ErrorMessage string
			}{
				ErrorMessage: "內部錯誤，請聯絡團隊！\nerror code: 1003",
			}
			serveTemplateWithData(w, "error_page.html", data)
			return
		}

		// Send verification email
		emailData := struct {
			Token string
		}{
			Token: emailVerificationToken,
		}

		err = mailer.SendEmail(db, email, "Email Verification", "email_verification_code.html", emailData)
		if err != nil {
			data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1002"
			serveTemplateWithData(w, "error_page.html", data)
			return
		}

		resendEmailToken, err := token.GenerateToken("resend_email", userID, 1*time.Hour)
		if err != nil {
			data := struct {
				ErrorMessage string
			}{
				ErrorMessage: "內部錯誤，請聯絡團隊！\nerror code: 1003",
			}
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
	} else {
		serveTemplateWithData(w, "register.html", nil)
	}
}

func ResendVerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		ErrorMessage string
	}{}

	cookie, err := r.Cookie("rem")
	if err != nil {
		data.ErrorMessage = "缺少密鑰"
		serveTemplateWithData(w, "error_page.html", data)
		return
	}

	// Verify token from the cookie
	userID, err := token.VerifyToken(cookie.Value, "resend_email")
	if err != nil {
		// If token is valid, redirect to index
		data.ErrorMessage = "非法或過期的密鑰"
		serveTemplateWithData(w, "error_page.html", data)
		return
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM tokens WHERE user_id = ? AND purpose = 'email_verification'", userID).Scan(&count)
	if err != nil {
		data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1005"
		serveTemplateWithData(w, "error_page.html", data)
		return
	}

	if count > 5 {
		data.ErrorMessage = "操作太頻繁！\n60分鐘內操作6次以上！"
		serveTemplateWithData(w, "error_page.html", data)
		return
	}

	var email string

	err = db.QueryRow("SELECT email FROM users WHERE user_id = ?", userID).Scan(&email)
	if err != nil {
		data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1001"
		serveTemplateWithData(w, "error_page.html", data)
		return
	}

	// Generate verification token
	emailVerificationToken, err := token.GenerateToken("email_verification", userID, 15*time.Minute)
	if err != nil {
		data := struct {
			ErrorMessage string
		}{
			ErrorMessage: "內部錯誤，請聯絡團隊！\nerror code: 1003",
		}
		serveTemplateWithData(w, "error_page.html", data)
		return
	}

	// Send verification email
	emailData := struct {
		Token string
	}{
		Token: emailVerificationToken,
	}

	err = mailer.SendEmail(db, email, "Email Verification", "email_verification_code.html", emailData)
	if err != nil {
		data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1002"
		serveTemplateWithData(w, "error_page.html", data)
		return
	}

	serveTemplateWithData(w, "email_verification.html", nil)
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		serveTemplateWithData(w, "reset_password.html", nil)
	} else if r.Method == http.MethodPost {
		var userID string

		email := r.FormValue("email")
		err := db.QueryRow("SELECT user_id FROM users WHERE email = ?", email).Scan(&userID)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		var count int
		db.QueryRow("SELECT COUNT(*) FROM tokens WHERE user_id = ? AND purpose = 'reset_password'", userID).Scan(&count)
		if count > 2 {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		num, _ := strconv.Atoi(userID)
		resetPasswordToken, _ := token.GenerateToken("reset_password", num, 15*time.Minute)
		emailData := struct {
			Token string
		}{
			Token: resetPasswordToken,
		}

		go mailer.SendEmail(db, email, "Password reset", "reset_password_code.html", emailData)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func ResetPasswordSetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		queryParams := r.URL.Query()
		tokenValue := queryParams.Get("token")

		if tokenValue == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Verify the token
		_, err := token.VerifyToken(tokenValue, "reset_password")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		serveTemplateWithData(w, "reset_password_setting.html", nil)
	} else if r.Method == http.MethodPost {
		queryParams := r.URL.Query()
		tokenValue := queryParams.Get("token")

		if tokenValue == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Verify the token
		userID, err := token.VerifyToken(tokenValue, "reset_password")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		data := struct {
			ErrorMessage string
		}{}

		if password != confirmPassword {
			data.ErrorMessage = "密碼不相符！"
			serveTemplateWithData(w, "reset_password_setting.html", data)
			return
		}

		// Generate salt
		salt := generateSalt()

		// Encrypt password (with salt)
		hashedPassword := hashPassword(password, salt)

		_, err = db.Exec("UPDATE users SET password = ?, salt = ? WHERE user_id = ?", hashedPassword, salt, userID)
		if err != nil {
			data.ErrorMessage = "內部錯誤，請聯絡團隊！\nerror code: 1001"
			serveTemplateWithData(w, "error_page.html", data)
			return
		}
		token.DeleteToken(tokenValue, "reset_password")
		serveTemplateWithData(w, "reset_password_success.html", nil)
	}
}

func generateSalt() string {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		fmt.Printf("Failed to generate salt: %v\n", err)
		return ""
	}
	return hex.EncodeToString(salt)
}

func hashPassword(password, salt string) string {
	hash := sha256.New()
	hash.Write([]byte(password + salt))
	return hex.EncodeToString(hash.Sum(nil))
}

func insertUser(email, hashedPassword, salt string) error {
	_, err := db.Exec("INSERT INTO users (email, password, salt, agree_privacy, email_verified) VALUES (?, ?, ?, ?, ?)", email, hashedPassword, salt, false, false)
	return err
}

func emailExists(email string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=?)", email).Scan(&exists)
	return exists, err
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
