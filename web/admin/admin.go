package admin

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"web/token"
)

// Add a global variable to hold the database connection
var db *sql.DB

// InitDB initializes the database connection
func InitDB(database *sql.DB) {
	db = database
}

type Announcement struct {
	Content string `json:"announcement"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func AdminPanelHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		userID, err := token.VerifyToken(cookie.Value, "session")
		if err == nil {
			var permit int
			err = db.QueryRow("SELECT permit FROM users WHERE user_id = ?", userID).Scan(&permit)
			if err == nil && permit == 0 {
				data := struct {
					Login               bool
					AnnouncementMessage []string
				}{}
				file, err := os.Open("sys_info/announcement")
				if err != nil {
					fmt.Println("Error opening file:", err)
					return
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					line := scanner.Text()
					data.AnnouncementMessage = append(data.AnnouncementMessage, line)
				}
				data.Login = true
				serveTemplateWithData(w, "admin.html", data)
				return
			}
		}
	}
	permissionDenied(w, r)
}

func SaveAnnouncementHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		permissionDenied(w, r)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err == nil {
		userID, err := token.VerifyToken(cookie.Value, "session")
		if err == nil {
			var permit int
			err = db.QueryRow("SELECT permit FROM users WHERE user_id = ?", userID).Scan(&permit)
			if err == nil && permit == 0 {
				var ann Announcement
				err := json.NewDecoder(r.Body).Decode(&ann)
				if err != nil {
					response := Response{Success: false, Message: "System return: " + err.Error()}
					jsonResponse(w, response)
					return
				}

				err = ioutil.WriteFile("sys_info/announcement", []byte(ann.Content), 0644)
				if err != nil {
					response := Response{Success: false, Message: "System return: " + err.Error()}
					jsonResponse(w, response)
					return
				}

				response := Response{Success: true, Message: "System return: Announcement saved successfully"}
				jsonResponse(w, response)
				return
			}
		}
	}
	permissionDenied(w, r)
}

func jsonResponse(w http.ResponseWriter, response Response) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the template
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func permissionDenied(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	data := struct {
		Login               bool
		AnnouncementMessage []string
	}{}
	if err == nil {
		_, err = token.VerifyToken(cookie.Value, "session")
		if err == nil {
			data.Login = true
		}
	}

	file, err := os.Open("sys_info/announcement")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if index := strings.Index(line, "//"); index != -1 {
			line = strings.TrimSpace(line[:index])
		}

		if line == "" {
			continue
		}

		data.AnnouncementMessage = append(data.AnnouncementMessage, line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}

	serveTemplateWithData(w, "index.html", data)
}
