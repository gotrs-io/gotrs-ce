package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/gotrs-io/gotrs-ce/internal/services/registry"
)

// PostgresService implements DatabaseService for PostgreSQL
type PostgresService struct {
	config   *registry.ServiceConfig
	db       *sql.DB
	health   *registry.ServiceHealth
	metrics  *registry.ServiceMetrics
}

// NewPostgresService creates a new PostgreSQL service instance
func NewPostgresService(config *registry.ServiceConfig) (DatabaseService, error) {
	return &PostgresService{
		config: config,
		health: &registry.ServiceHealth{
			ServiceID: config.ID,
			Status:    registry.StatusInitializing,
		},
		metrics: &registry.ServiceMetrics{
			ServiceID: config.ID,
		},
	}, nil
}

// Connect establishes connection to PostgreSQL
func (s *PostgresService) Connect(ctx context.Context) error {
	// Build connection string
	connStr := s.buildConnectionString()
	
	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		s.health.Status = registry.StatusUnhealthy
		s.health.Error = err.Error()
		return fmt.Errorf("failed to open database: %w", err)
	}
	
	// Configure connection pool
	if s.config.MaxConns > 0 {
		db.SetMaxOpenConns(s.config.MaxConns)
	}
	if s.config.MinConns > 0 {
		db.SetMaxIdleConns(s.config.MinConns)
	}
	
	// Test connection
	if err := db.PingContext(ctx); err != nil {
		s.health.Status = registry.StatusUnhealthy
		s.health.Error = err.Error()
		return fmt.Errorf("failed to ping database: %w", err)
	}
	
	s.db = db
	s.health.Status = registry.StatusHealthy
	s.health.LastChecked = time.Now()
	
	return nil
}

// Disconnect closes the database connection
func (s *PostgresService) Disconnect(ctx context.Context) error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Ping checks if the database is reachable
func (s *PostgresService) Ping(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("database not connected")
	}
	return s.db.PingContext(ctx)
}

// Health returns the health status of the service
func (s *PostgresService) Health(ctx context.Context) (*registry.ServiceHealth, error) {
	start := time.Now()
	err := s.Ping(ctx)
	s.health.Latency = time.Since(start)
	s.health.LastChecked = time.Now()
	
	if err != nil {
		s.health.Status = registry.StatusUnhealthy
		s.health.Error = err.Error()
		return s.health, err
	}
	
	// Get additional health metrics
	stats := s.db.Stats()
	s.health.Status = registry.StatusHealthy
	s.health.Error = ""
	s.health.Metadata = map[string]interface{}{
		"open_connections":  stats.OpenConnections,
		"in_use":           stats.InUse,
		"idle":             stats.Idle,
		"wait_count":       stats.WaitCount,
		"wait_duration":    stats.WaitDuration.String(),
		"max_idle_closed":  stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
	
	return s.health, nil
}

// Metrics returns performance metrics
func (s *PostgresService) Metrics(ctx context.Context) (*registry.ServiceMetrics, error) {
	stats := s.db.Stats()
	
	s.metrics.Connections = stats.OpenConnections
	s.metrics.CustomMetrics = map[string]interface{}{
		"in_use":            stats.InUse,
		"idle":              stats.Idle,
		"wait_count":        stats.WaitCount,
		"wait_duration_ms":  stats.WaitDuration.Milliseconds(),
		"max_idle_closed":   stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
	
	return s.metrics, nil
}

// GetConfig returns the service configuration
func (s *PostgresService) GetConfig() *registry.ServiceConfig {
	return s.config
}

// UpdateConfig updates the service configuration
func (s *PostgresService) UpdateConfig(config *registry.ServiceConfig) error {
	// For database connections, this typically requires reconnection
	s.config = config
	
	// Reconnect with new configuration
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := s.Disconnect(ctx); err != nil {
		return err
	}
	
	return s.Connect(ctx)
}

// Type returns the service type
func (s *PostgresService) Type() registry.ServiceType {
	return registry.ServiceTypeDatabase
}

// Provider returns the service provider
func (s *PostgresService) Provider() registry.ServiceProvider {
	return registry.ProviderPostgres
}

// ID returns the service ID
func (s *PostgresService) ID() string {
	return s.config.ID
}

// Query executes a query that returns rows
func (s *PostgresService) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	
	s.metrics.Requests++
	start := time.Now()
	rows, err := s.db.QueryContext(ctx, query, args...)
	s.metrics.Latency = time.Since(start)
	
	if err != nil {
		s.metrics.Errors++
	}
	
	return rows, err
}

// QueryRow executes a query that returns a single row
func (s *PostgresService) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if s.db == nil {
		return nil
	}
	
	s.metrics.Requests++
	return s.db.QueryRowContext(ctx, query, args...)
}

// Exec executes a query that doesn't return rows
func (s *PostgresService) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	
	s.metrics.Requests++
	start := time.Now()
	result, err := s.db.ExecContext(ctx, query, args...)
	s.metrics.Latency = time.Since(start)
	
	if err != nil {
		s.metrics.Errors++
	}
	
	return result, err
}

// BeginTx starts a database transaction
func (s *PostgresService) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	
	return s.db.BeginTx(ctx, opts)
}

// GetDB returns the underlying database connection
func (s *PostgresService) GetDB() *sql.DB {
	return s.db
}

// SetMaxOpenConns sets the maximum number of open connections
func (s *PostgresService) SetMaxOpenConns(n int) {
	if s.db != nil {
		s.db.SetMaxOpenConns(n)
	}
}

// SetMaxIdleConns sets the maximum number of idle connections
func (s *PostgresService) SetMaxIdleConns(n int) {
	if s.db != nil {
		s.db.SetMaxIdleConns(n)
	}
}

// SetConnMaxLifetime sets the maximum connection lifetime
func (s *PostgresService) SetConnMaxLifetime(d time.Duration) {
	if s.db != nil {
		s.db.SetConnMaxLifetime(d)
	}
}

// RunMigrations runs database migrations
func (s *PostgresService) RunMigrations(ctx context.Context, migrationsPath string) error {
	// This would integrate with a migration tool like golang-migrate
	// For now, return not implemented
	return fmt.Errorf("migrations not implemented")
}

// GetSchemaVersion returns the current schema version
func (s *PostgresService) GetSchemaVersion() (int, error) {
	var version int
	err := s.db.QueryRow("SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// Backup creates a database backup
func (s *PostgresService) Backup(ctx context.Context, path string) error {
	// This would use pg_dump or similar
	// For now, return not implemented
	return fmt.Errorf("backup not implemented")
}

// Restore restores from a backup
func (s *PostgresService) Restore(ctx context.Context, path string) error {
	// This would use pg_restore or similar
	// For now, return not implemented
	return fmt.Errorf("restore not implemented")
}

// buildConnectionString builds the PostgreSQL connection string
func (s *PostgresService) buildConnectionString() string {
	sslMode := "disable"
	if s.config.TLS {
		sslMode = "require"
	}
	
	// Check for sslmode in options
	if s.config.Options != nil {
		if mode, ok := s.config.Options["sslmode"].(string); ok {
			sslMode = mode
		}
	}
	
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		s.config.Host,
		s.config.Port,
		s.config.Username,
		s.config.Password,
		s.config.Database,
		sslMode,
	)
}