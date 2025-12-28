// Package testutil provides testing utilities including database containers.
package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer represents a running PostgreSQL container for testing.
type PostgresContainer struct {
	Container testcontainers.Container
	Host      string
	Port      string
	User      string
	Password  string
	Database  string
	DSN       string
}

var (
	pgContainer *PostgresContainer
	pgOnce      sync.Once
	errPG       error
)

// GetPostgresContainer returns a shared PostgreSQL container for tests.
// The container is started once and reused across all tests.
func GetPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	pgOnce.Do(func() {
		pgContainer, errPG = startPostgresContainer(ctx)
	})
	return pgContainer, errPG
}

func startPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	const (
		user     = "testuser"
		password = "testpass"
		dbName   = "testdb"
	)

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     user,
			"POSTGRES_PASSWORD": password,
			"POSTGRES_DB":       dbName,
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			wait.ForListeningPort("5432/tcp"),
		).WithDeadline(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, mappedPort.Port(), dbName,
	)

	return &PostgresContainer{
		Container: container,
		Host:      host,
		Port:      mappedPort.Port(),
		User:      user,
		Password:  password,
		Database:  dbName,
		DSN:       dsn,
	}, nil
}

// GetDB creates a new database connection from the container.
func (c *PostgresContainer) GetDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", c.DSN)
	if err != nil {
		return nil, err
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// Stop terminates the container.
func (c *PostgresContainer) Stop(ctx context.Context) error {
	if c.Container != nil {
		return c.Container.Terminate(ctx)
	}
	return nil
}

// SetupTestDB sets up the database with migrations for testing.
func SetupTestDB(ctx context.Context, db *sql.DB) error {
	// Create tables
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS todos (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			completed BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			revoked BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
			action VARCHAR(100) NOT NULL,
			resource_type VARCHAR(100) NOT NULL,
			resource_id BIGINT,
			details JSONB,
			ip_address INET,
			user_agent TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_todos_user_id ON todos(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at)`,
	}

	for _, migration := range migrations {
		if _, err := db.ExecContext(ctx, migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// CleanupTestDB cleans all test data from the database.
func CleanupTestDB(ctx context.Context, db *sql.DB) error {
	tables := []string{"audit_logs", "refresh_tokens", "todos", "users"}
	for _, table := range tables {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			return fmt.Errorf("failed to truncate %s: %w", table, err)
		}
	}
	return nil
}

// SkipIfShort skips the test if running in short mode.
func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
}

// TestMain sets up and tears down the test container.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start container
	container, err := GetPostgresContainer(ctx)
	if err != nil {
		log.Fatalf("Failed to start postgres container: %v", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if container != nil {
		_ = container.Stop(ctx)
	}

	os.Exit(code)
}
