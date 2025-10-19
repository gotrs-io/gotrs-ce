package database

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gotrs-io/gotrs-ce/internal/services/adapter"
	_ "github.com/lib/pq"
)

// GetDB returns the database connection singleton from the service registry.
// Service registry is the single source of truth for database connections.
func GetDB() (*sql.DB, error) {
	if testDBOverride && testDB != nil {
		return testDB, nil
	}

	if testDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		err := testDB.PingContext(ctx)
		cancel()
		if err == nil {
			return testDB, nil
		}
		// Drop stale pointer and fall back to the service-managed connection.
		testDB = nil
	}

	db, err := adapter.GetDB()
	if err != nil {
		return nil, err
	}

	if !testDBOverride {
		testDB = db
	}

	return db, nil
}
