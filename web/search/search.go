package search

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"web/token"

	_ "github.com/go-sql-driver/mysql"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

// Add a global variable to hold the database connection
var db *sql.DB
var current_year_table, last_year_table, two_years_ago_table string

// InitDB initializes the database connection
func InitDB(database *sql.DB, currentYearTable, lastYearTable, TwoYearsAgoTable string) {
	db = database
	current_year_table, last_year_table, two_years_ago_table = currentYearTable, lastYearTable, TwoYearsAgoTable
}

// Result holds the search result with its Levenshtein distance
type Result struct {
	CombinedName    string `json:"combined_name"`
	Code            string `json:"code"`
	Subject         string `json:"subject"`
	SubjectCriteria string `json:"subject_criteria"`
	Distance        int    `json:"distance"`
}

type Data struct {
	Tag                         string   `json:"tag"`
	Code                        string   `json:"code"`
	Subject                     []string `json:"subject"`
	SchoolName                  string   `json:"school_name"`
	SpecifyItems                []string `json:"specify_items"`
	ScreeningDate               string   `json:"screening_date"`
	DepartmentName              string   `json:"department_name"`
	SubjectScoring              []string `json:"subject_scoring"`
	EnrollmentQuota             string   `json:"enrollment_quota"`
	SubjectCriteria             []string `json:"subject_criteria"`
	ExcessiveScreening          []string `json:"excessive_screening"`
	SubjectMagnification        []string `json:"subject_magnification"`
	SpecifyItemsCriteria        []string `json:"specify_items_criteria"`
	SubjectScoreProportion      string   `json:"subject_score_proportion"`
	SpecifyItemsScoreProportion []string `json:"specify_items_score_proportion"`
	Candidates                  string   `json:"candidates"`
}

// ByDistance implements sort.Interface for sorting results by distance
type ByDistance []Result

func (a ByDistance) Len() int           { return len(a) }
func (a ByDistance) Less(i, j int) bool { return a[i].Distance < a[j].Distance }
func (a ByDistance) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// IsNumber checks if the given string is a number
func IsNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// getMyListHandler handles requests to get a list of schools
func GetMyListHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	userID := -1
	if err == nil {
		userID, err = token.VerifyToken(cookie.Value, "session")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Prepare a slice to hold the search results
	var results []Result

	// Fetch the school_list for the user
	var schoolList string
	err = db.QueryRow("SELECT school_list FROM users WHERE user_id = ?", userID).Scan(&schoolList)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
		return
	}

	n, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		n = 10 // default value if page is not provided
	} else {
		n = n * 10
	}

	// Split the school_list into a slice
	schools := strings.Split(schoolList, ",")

	// Limit the number of schools to n
	if len(schools) >= n {
		schools = schools[n-10 : n]
	} else if n-len(schools) < 10 {
		schools = schools[n-10:]
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
		return
	}

	// Search for each school in the list
	query := fmt.Sprintf(`SELECT combined_name, code, subject, subject_criteria FROM %s WHERE combined_name = ?`, current_year_table)
	for _, school := range schools {
		var combinedName, code, subject, subjectCriteria string
		err = db.QueryRow(query, school).Scan(&combinedName, &code, &subject, &subjectCriteria)
		if err != nil {
			continue
		}
		results = append(results, Result{CombinedName: combinedName, Code: code, Subject: subject, SubjectCriteria: subjectCriteria})
	}

	// Return the results as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// LoadReplaceWords loads the replace words from a file and returns a map
func LoadReplaceWords(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	replaceWords := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) == 2 {
			replaceWords[parts[0]] = parts[1]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return replaceWords, nil
}

// ReplaceKeyword replaces keyword based on the replace words map
func ReplaceKeyword(keyword string, replaceWords map[string]string) string {
	for oldWord, newWord := range replaceWords {
		keyword = strings.ReplaceAll(keyword, oldWord, newWord)
	}
	return keyword
}

// SearchDatabase searches the database for a given keyword using fuzzy matching
func SearchDatabase(db *sql.DB, tableName, keyword string, n int, replaceWords map[string]string) ([]Result, error) {
	var query string
	var rows *sql.Rows
	var err error

	threshold := len(keyword)
	if threshold < 9 {
		threshold = (threshold / 3) * 32
	} else {
		threshold = ((threshold/3)-3)*16 + 64
	}

	if IsNumber(keyword) {
		query = fmt.Sprintf(`SELECT combined_name, code FROM %s WHERE code LIKE ?`, tableName)
		keyword = "%" + keyword + "%"
		rows, err = db.Query(query, keyword)
		threshold = 1024
	} else {
		keyword = ReplaceKeyword(keyword, replaceWords)
		query = fmt.Sprintf(`SELECT combined_name, code FROM %s`, tableName)
		rows, err = db.Query(query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Result
	DefaultOptions := levenshtein.Options{
		InsCost: 0,
		DelCost: 32,
		SubCost: 64,
		Matches: levenshtein.IdenticalRunes,
	}

	for rows.Next() {
		var combinedName string
		var code string
		if err := rows.Scan(&combinedName, &code); err != nil {
			return nil, err
		}

		distance := levenshtein.DistanceForStrings([]rune(keyword), []rune(combinedName), DefaultOptions)
		if distance < threshold {
			num, _ := strconv.Atoi(code)
			results = append(results, Result{CombinedName: combinedName, Code: code, Distance: distance*1000000 + num})
		}
	}

	sort.Sort(ByDistance(results))
	if len(results) >= n {
		results = results[n-10 : n]
	} else if n-len(results) < 10 {
		results = results[n-10:]
	} else {
		results = nil
	}

	// Fetch additional fields for the selected results
	for i := range results {
		var subject, subjectCriteria string
		query = fmt.Sprintf(`SELECT subject, subject_criteria FROM %s WHERE code = ?`, tableName)
		err := db.QueryRow(query, results[i].Code).Scan(&subject, &subjectCriteria)
		if err != nil {
			return nil, err
		}
		results[i].Subject = subject
		results[i].SubjectCriteria = subjectCriteria
	}

	return results, nil
}

// SearchHandler handles search requests
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	n, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		n = 10
	} else {
		n = n * 10
	}

	replaceWordsFile := "./search/replace_words.txt"
	replaceWords, err := LoadReplaceWords(replaceWordsFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	results, err := SearchDatabase(db, current_year_table, keyword, n, replaceWords)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// DetailHandler handles requests to view details of a specific entry
func DetailHandler(w http.ResponseWriter, r *http.Request) {
	// Get combined_name from URL query parameters
	combinedName := r.URL.Query().Get("combined_name")
	if combinedName == "" {
		http.Error(w, "combined_name parameter is required", http.StatusBadRequest)
		return
	}
	combinedName = strings.TrimSpace(combinedName)
	combinedName = strings.ReplaceAll(combinedName, "<br>", " ")
	combinedNameSearch := "%" + combinedName + "%"

	// List of table names to query
	tables := []string{current_year_table, last_year_table, two_years_ago_table}

	var jsonArray []Data
	for _, table := range tables {
		var jsonData Data
		var fullJSON, Candidates string
		query := fmt.Sprintf("SELECT full_json, candidates FROM %s WHERE combined_name LIKE ?", table)
		err := db.QueryRow(query, combinedNameSearch).Scan(&fullJSON, &Candidates)
		if err != nil {
			if err == sql.ErrNoRows {
				jsonArray = append(jsonArray, jsonData)
				continue
			} else {
				http.Error(w, fmt.Sprintf("Failed to query database: %v", fullJSON), http.StatusInternalServerError)
				return
			}
		}

		if err := json.Unmarshal([]byte(fullJSON), &jsonData); err != nil {
			http.Error(w, fmt.Sprintf("Failed to unmarshal JSON: %v", err), http.StatusInternalServerError)
			return
		}

		jsonData.Candidates = Candidates
		jsonArray = append(jsonArray, jsonData)
	}

	if len(jsonArray) == 0 {
		http.Error(w, "No data found", http.StatusNotFound)
		return
	}

	subjects := findMaxSet(jsonArray, "subject")
	specifyItems := findMaxSet(jsonArray, "specify_items")

	for i := range jsonArray {
		_, jsonArray[i].SubjectScoring = mergeArrays(jsonArray[i].Subject, jsonArray[i].SubjectScoring, subjects)
		_, jsonArray[i].SubjectCriteria = mergeArrays(jsonArray[i].Subject, jsonArray[i].SubjectCriteria, subjects)
		jsonArray[i].Subject, jsonArray[i].SubjectMagnification = mergeArrays(jsonArray[i].Subject, jsonArray[i].SubjectMagnification, subjects)
		_, jsonArray[i].SpecifyItemsCriteria = mergeArrays(jsonArray[i].SpecifyItems, jsonArray[i].SpecifyItemsCriteria, specifyItems)
		jsonArray[i].SpecifyItems, jsonArray[i].SpecifyItemsScoreProportion = mergeArrays(jsonArray[i].SpecifyItems, jsonArray[i].SpecifyItemsScoreProportion, specifyItems)
	}

	result, err := json.MarshalIndent(jsonArray, "", "    ")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal combined JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a data struct containing combined_name and the combined JSON array
	data := struct {
		CombinedName string
		FullJSON     string
		CurrentYear  string
		LastYear     string
		TwoYearsAgo  string
	}{
		CombinedName: combinedName,
		FullJSON:     string(result),
		CurrentYear:  current_year_table,
		LastYear:     last_year_table,
		TwoYearsAgo:  two_years_ago_table,
	}

	// Render the detail.html template with the data
	tmpl, err := template.ParseFiles("./tem/detail.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse template: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
	}
}

func findMaxSet(data []Data, field string) []string {
	itemSet := make(map[string]struct{})
	for _, d := range data {
		var items []string
		if field == "subject" {
			items = d.Subject
		} else if field == "specify_items" {
			items = d.SpecifyItems
		}

		for _, item := range items {
			itemSet[item] = struct{}{}
		}
	}

	var result []string
	for item := range itemSet {
		result = append(result, item)
	}

	// Sort the result to ensure consistent order
	sort.Strings(result)

	return result
}

func mergeArrays(arr []string, arrA []string, fullSet []string) ([]string, []string) {
	itemIndexMap := make(map[string]string)
	for i, item := range arr {
		itemIndexMap[item] = arrA[i]
	}

	var mergedArr []string
	var mergedArrA []string
	for _, item := range fullSet {
		if value, exists := itemIndexMap[item]; exists {
			mergedArr = append(mergedArr, item)
			mergedArrA = append(mergedArrA, value)
		} else {
			mergedArr = append(mergedArr, item)
			mergedArrA = append(mergedArrA, "--")
		}
	}

	return mergedArr, mergedArrA
}
