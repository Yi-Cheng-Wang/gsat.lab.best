package token

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Add a global variable to hold the database connection
var db *sql.DB

// Token represents the structure of a token in the database
type Token struct {
	ID        int
	Purpose   string
	UserID    int
	Token     string
	ExpiresAt time.Time
}

// InitDB initializes the database connection
func InitDB(database *sql.DB) {
	db = database
}

// GenerateToken generates a new token and stores it in the database
func GenerateToken(purpose string, userID int, duration time.Duration) (string, error) {
	token := generateRandomToken()

	expiresAt := time.Now().Add(duration)
	_, err := db.Exec("INSERT INTO tokens (purpose, user_id, token, expires_at) VALUES (?, ?, ?, ?)",
		purpose, userID, token, expiresAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

// VerifyToken verifies if the provided token is valid for the given purpose
func VerifyToken(tokenString, purpose string) (int, error) {
	var userID int
	var expiresAtStr string

	err := db.QueryRow("SELECT user_id, expires_at FROM tokens WHERE token = ? AND purpose = ?", tokenString, purpose).Scan(&userID, &expiresAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, fmt.Errorf("token not found")
		}
		return -1, err
	}

	// Parse the expiresAtStr to time.Time
	expiresAt, err := time.Parse("2006-01-02 15:04:05", expiresAtStr)
	if err != nil {
		return -1, fmt.Errorf("failed to parse expires_at: %v", err)
	}

	// Check if the token has expired
	if expiresAt.Before(time.Now()) {
		// If expired, delete the token from the database
		err := DeleteToken(tokenString, purpose)
		if err != nil {
			return -1, fmt.Errorf("failed to delete token: %v", err)
		}
		return -1, fmt.Errorf("token expired")
	}

	return userID, nil
}

// generateRandomToken generates a random token
func generateRandomToken() string {
	token := make([]byte, 256)
	_, err := rand.Read(token)
	if err != nil {
		fmt.Printf("Failed to generate random token: %v\n", err)
		return ""
	}
	return hex.EncodeToString(token)
}

func DeleteToken(tokenString, purpose string) error {
	_, err := db.Exec("DELETE FROM tokens WHERE token = ? AND purpose = ?", tokenString, purpose)
	if err != nil {
		return err
	}
	return nil
}

// CleanExpiredTokens periodically cleans up expired tokens from the database
func CleanExpiredTokens(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		err := deleteExpiredTokens()
		if err != nil {
			fmt.Printf("Failed to clean expired tokens: %v\n", err)
		}
	}
}

func deleteExpiredTokens() error {
	_, err := db.Exec("DELETE FROM tokens WHERE expires_at < ?", time.Now())
	if err != nil {
		return err
	}
	return nil
}
