package dbconnector

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

const (
	maxRetries      = 2
	retryDelay      = 1 * time.Second
	maxOpenConns    = 100
	maxIdleConns    = 25
	connMaxLifetime = 5 * time.Minute
)

// InitDB initializes the database connection with retry logic
func InitDB(dbUser, dbPassword, dbHost, dbPort, dbName string) {
	var err error

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)

	for i := 0; i <= maxRetries; i++ {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Printf("Failed to connect to database: %v", err)
			if i < maxRetries {
				log.Printf("Retrying in %v...", retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			log.Fatalf("Exceeded maximum retries. Exiting...")
		}

		// Set up connection pool
		db.SetMaxOpenConns(maxOpenConns)
		db.SetMaxIdleConns(maxIdleConns)
		db.SetConnMaxLifetime(connMaxLifetime)

		// Check if the connection is working
		if err = db.Ping(); err != nil {
			log.Printf("Failed to ping database: %v", err)
			if i < maxRetries {
				log.Printf("Retrying in %v...", retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			log.Fatalf("Exceeded maximum retries. Exiting...")
		}

		// Connection is successfully established
		log.Println("Database connection established successfully")
		return
	}
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return db
}
