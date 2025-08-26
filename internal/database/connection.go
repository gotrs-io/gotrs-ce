package database

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	_ "github.com/lib/pq"
	"github.com/gotrs-io/gotrs-ce/internal/services/adapter"
)

var (
	db   *sql.DB
	once sync.Once
)

// GetDB returns the database connection singleton
// Now uses service registry by default
func GetDB() (*sql.DB, error) {
	// Try to get from service registry first (now default)
	db, err := adapter.GetDB()
	if err == nil {
		return db, nil
	}
	
	// If service registry fails, fall back to direct connection
	// This ensures the system still works during transition
	var fallbackErr error
	once.Do(func() {
		// Get database configuration from environment
		host := os.Getenv("DB_HOST")
		if host == "" {
			host = "postgres"
		}
		
		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "5432"
		}
		
		user := os.Getenv("DB_USER")
		if user == "" {
			user = "gotrs"
		}
		
		password := os.Getenv("DB_PASSWORD")
		if password == "" {
			password = "gotrs_password"
		}
		
		dbname := os.Getenv("DB_NAME")
		if dbname == "" {
			dbname = "gotrs"
		}
		
		sslmode := os.Getenv("DB_SSLMODE")
		if sslmode == "" {
			sslmode = "disable"
		}
		
		// Create connection string
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)
		
		// Open database connection
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			return
		}
		
		// Set connection pool settings
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		
		// Test the connection
		fallbackErr = db.Ping()
	})
	
	if fallbackErr != nil {
		return nil, fmt.Errorf("service registry: %v, fallback: %v", err, fallbackErr)
	}
	
	return db, fallbackErr
}