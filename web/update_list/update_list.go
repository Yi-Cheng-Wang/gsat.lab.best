package update_list

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"web/token"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var current_year_table string

// AddToListHandler handles the request to add a school name to the list
func AddToListHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	UserID := -1
	if err == nil {
		UserID, err = token.VerifyToken(cookie.Value, "session")
		if err != nil {
			http.Error(w, "您必須登入才能使用這項服務！", http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "您必須登入才能使用這項服務！", http.StatusBadRequest)
		return
	}

	// Get school name from URL query parameters
	schoolName := r.URL.Query().Get("school_name")
	if schoolName == "" {
		http.Error(w, "school_name parameter is required", http.StatusBadRequest)
		return
	}

	schoolName = strings.ReplaceAll(schoolName, "<br>", " ")

	query := fmt.Sprintf("SELECT combined_name FROM %s WHERE combined_name = ?", current_year_table)
	err = db.QueryRow(query, schoolName).Scan(&schoolName)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No school_name found", http.StatusInternalServerError)
		} else {
			http.Error(w, fmt.Sprintf("Failed to query database: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Query the current list from the database
	var currentList sql.NullString
	err = db.QueryRow("SELECT school_list FROM users WHERE user_id = ?", UserID).Scan(&currentList)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No record found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to query database: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Handle NULL value
	list := ""
	if currentList.Valid {
		list = currentList.String
	}

	// Add the new school name to the list
	newList := addToCommaSeparatedList(list, schoolName)

	// Update the list in the database
	_, err = db.Exec("UPDATE users SET school_list = ? WHERE user_id = ?", newList, UserID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update database: %v", err), http.StatusInternalServerError)
		return
	}

	// Return a success message
	w.Write([]byte("School name added successfully"))
}

// RemoveFromListHandler handles the request to remove a school name from the list
func RemoveFromListHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	UserID := -1
	if err == nil {
		UserID, err = token.VerifyToken(cookie.Value, "session")
		if err != nil {
			http.Error(w, "您必須登入才能使用這項服務！", http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "您必須登入才能使用這項服務！", http.StatusBadRequest)
		return
	}

	// Get school name from URL query parameters
	schoolName := r.URL.Query().Get("school_name")
	if schoolName == "" {
		http.Error(w, "school_name parameter is required", http.StatusBadRequest)
		return
	}

	schoolName = strings.ReplaceAll(schoolName, "<br>", " ")

	// Query the current list from the database
	var currentList sql.NullString
	err = db.QueryRow("SELECT school_list FROM users WHERE user_id = ?", UserID).Scan(&currentList)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No record found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to query database: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Handle NULL value
	list := ""
	if currentList.Valid {
		list = currentList.String
	}

	// Remove the school name from the list
	newList := removeFromCommaSeparatedList(list, schoolName)

	// Update the list in the database
	_, err = db.Exec("UPDATE users SET school_list = ? WHERE user_id = ?", newList, UserID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update database: %v", err), http.StatusInternalServerError)
		return
	}

	// Return a success message
	w.Write([]byte("School name removed successfully"))
}

// addToCommaSeparatedList adds a new item to a comma-separated list
func addToCommaSeparatedList(list, newItem string) string {
	if list == "" {
		return newItem
	}
	// Avoid adding duplicate entries
	items := strings.Split(list, ",")
	for _, item := range items {
		if item == newItem {
			return list
		}
	}
	return list + "," + newItem
}

// removeFromCommaSeparatedList removes an item from a comma-separated list
func removeFromCommaSeparatedList(list, removeItem string) string {
	if list == "" {
		return list
	}
	items := strings.Split(list, ",")
	newItems := []string{}
	for _, item := range items {
		if item != removeItem {
			newItems = append(newItems, item)
		}
	}
	return strings.Join(newItems, ",")
}

// InitDB initializes the database connection
func InitDB(database *sql.DB, currentYearTable string) {
	db = database
	current_year_table = currentYearTable
}
