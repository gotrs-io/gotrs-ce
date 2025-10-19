package api

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/gotrs-io/gotrs-ce/internal/database"
)

func TestMain(m *testing.M) {
	ensureTestEnvironment()

	if err := waitForTestDatabase(45 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "test database unavailable: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	database.CloseTestDB()
	os.Exit(code)
}

func init() {
	ensureTestEnvironment()
}

func ensureTestEnvironment() {
	os.Setenv("APP_ENV", "test")
	os.Setenv("DB_DRIVER", "postgres")
	os.Setenv("DB_HOST", "postgres-test")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "gotrs_test")
	os.Setenv("DB_USER", "gotrs_user")
	os.Setenv("DB_PASSWORD", "gotrs_password")
}

func waitForTestDatabase(timeout time.Duration) error {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if host == "" || port == "" {
		return fmt.Errorf("test database host/port not configured")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)

	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		sqlDB, err := sql.Open("postgres", dsn)
		if err == nil {
			err = sqlDB.Ping()
			sqlDB.Close()
			if err == nil {
				database.ResetDB()
				return database.InitTestDB()
			}
			lastErr = err
		} else {
			lastErr = err
		}

		time.Sleep(500 * time.Millisecond)
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("timeout waiting for database")
	}
	return lastErr
}
