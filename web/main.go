package main

import (
	"bufio"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"web/admin"
	"web/dbconnector"
	"web/login"
	"web/registration"
	"web/search"
	"web/token"
	"web/update_list"
)

func main() {
	// Define command-line parameters
	dbUser := flag.String("db_user", "", "Database user")
	dbPassword := flag.String("db_password", "", "Database password")
	dbHost := flag.String("db_host", "127.0.0.1", "Database host")
	dbPort := flag.String("db_port", "3306", "Database port")
	dbName := flag.String("db_name", "", "Database name")
	currentYearTable := flag.String("current_year_table", "", "Current year table")
	lastYearTable := flag.String("last_year_table", "", "Last year table")
	twoYearsAgoTable := flag.String("two_years_ago_table", "", "Two years ago table")

	// Parse command-line parameters
	flag.Parse()

	dbconnector.InitDB(*dbUser, *dbPassword, *dbHost, *dbPort, *dbName)

	// Initialize database connection
	admin.InitDB(dbconnector.GetDB())
	login.InitDB(dbconnector.GetDB())
	registration.InitDB(dbconnector.GetDB())
	search.InitDB(dbconnector.GetDB(), *currentYearTable, *lastYearTable, *twoYearsAgoTable)
	token.InitDB(dbconnector.GetDB())
	update_list.InitDB(dbconnector.GetDB(), *currentYearTable)

	// Set routes and handler functions
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/about_us", aboutUsHandler)
	http.HandleFunc("/add_score", addScoreHandler)
	http.HandleFunc("/admin", admin.AdminPanelHandler)
	http.HandleFunc("/admin/save_announcement", admin.SaveAnnouncementHandler)
	http.HandleFunc("/contact_us", contactUsHandler)
	http.HandleFunc("/detail", search.DetailHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/get_my_list", search.GetMyListHandler)
	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/login", login.LoginHandler)
	http.HandleFunc("/logout", login.LogoutHandler)
	http.HandleFunc("/privacy", login.PrivacyHandler)
	http.HandleFunc("/register", registration.RegisterHandler)
	http.HandleFunc("/remove_from_list", update_list.RemoveFromListHandler)
	http.HandleFunc("/reset/password", registration.ResetPasswordHandler)
	http.HandleFunc("/reset_password", registration.ResetPasswordSetHandler)
	http.HandleFunc("/robots.txt", robotsTxtHandler)
	http.HandleFunc("/search", search.SearchHandler)
	http.HandleFunc("/add_to_list", update_list.AddToListHandler)
	http.HandleFunc("/verify_email", login.VerifyHandler)
	http.HandleFunc("/verify_email/resend", registration.ResendVerifyEmailHandler)
	http.HandleFunc("/what_is_this", whatIsThisHandler)

	// Start the token cleaner with a cleanup interval of 30 minute
	go token.CleanExpiredTokens(30 * time.Minute)

	// Set server listening address and port
	serverAddress := "127.0.0.1:8080"

	// Start the server
	fmt.Printf("Server is running at %s\n", serverAddress)
	err := http.ListenAndServe(serverAddress, nil)
	if err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
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

func aboutUsHandler(w http.ResponseWriter, r *http.Request) {
	serveTemplate(w, "about_us.html")
}

func addScoreHandler(w http.ResponseWriter, r *http.Request) {
	serveTemplate(w, "add_score.html")
}

func contactUsHandler(w http.ResponseWriter, r *http.Request) {
	serveTemplate(w, "contact_us.html")
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/favicon.ico")
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	serveTemplate(w, "list.html")
}

func robotsTxtHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/robots.txt")
}

func whatIsThisHandler(w http.ResponseWriter, r *http.Request) {
	serveTemplate(w, "what_is_this.html")
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the template
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
