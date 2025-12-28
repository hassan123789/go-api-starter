//go:build integration

// Package repository_test provides integration tests for repository layer.
package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// IntegrationTestSuite is the test suite for repository integration tests.
type IntegrationTestSuite struct {
	suite.Suite
	container testcontainers.Container
	db        *sql.DB
	ctx       context.Context
}

// SetupSuite runs once before all tests in the suite.
func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			wait.ForListeningPort("5432/tcp"),
		).WithDeadline(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(s.T(), err)
	s.container = container

	// Get connection details
	host, err := container.Host(s.ctx)
	require.NoError(s.T(), err)

	port, err := container.MappedPort(s.ctx, "5432")
	require.NoError(s.T(), err)

	dsn := "postgres://test:test@" + host + ":" + port.Port() + "/testdb?sslmode=disable"

	// Connect to database
	db, err := sql.Open("postgres", dsn)
	require.NoError(s.T(), err)
	s.db = db

	// Wait for connection
	require.Eventually(s.T(), func() bool {
		return db.Ping() == nil
	}, 30*time.Second, 100*time.Millisecond)

	// Run migrations
	s.runMigrations()
}

// TearDownSuite runs once after all tests in the suite.
func (s *IntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.container != nil {
		s.container.Terminate(s.ctx)
	}
}

// SetupTest runs before each test.
func (s *IntegrationTestSuite) SetupTest() {
	// Clean tables before each test
	s.cleanTables()
}

func (s *IntegrationTestSuite) runMigrations() {
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
		`CREATE INDEX IF NOT EXISTS idx_todos_user_id ON todos(user_id)`,
	}

	for _, m := range migrations {
		_, err := s.db.Exec(m)
		require.NoError(s.T(), err)
	}
}

func (s *IntegrationTestSuite) cleanTables() {
	_, err := s.db.Exec("TRUNCATE TABLE todos, users CASCADE")
	require.NoError(s.T(), err)
}

// ===== User Repository Tests =====

func (s *IntegrationTestSuite) TestUserRepository_Create() {
	// Arrange
	email := "test@example.com"
	passwordHash := "$2a$10$abcdefghijklmnopqrstuvwxyz123456"

	// Act
	result, err := s.db.Exec(
		"INSERT INTO users (email, password_hash) VALUES ($1, $2)",
		email, passwordHash,
	)

	// Assert
	require.NoError(s.T(), err)
	rowsAffected, _ := result.RowsAffected()
	assert.Equal(s.T(), int64(1), rowsAffected)

	// Verify user was created
	var count int
	err = s.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", email).Scan(&count)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, count)
}

func (s *IntegrationTestSuite) TestUserRepository_Create_DuplicateEmail() {
	// Arrange - create first user
	email := "duplicate@example.com"
	_, err := s.db.Exec(
		"INSERT INTO users (email, password_hash) VALUES ($1, $2)",
		email, "hash1",
	)
	require.NoError(s.T(), err)

	// Act - try to create user with same email
	_, err = s.db.Exec(
		"INSERT INTO users (email, password_hash) VALUES ($1, $2)",
		email, "hash2",
	)

	// Assert - should fail with unique constraint
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "duplicate key")
}

func (s *IntegrationTestSuite) TestUserRepository_GetByEmail() {
	// Arrange
	email := "find@example.com"
	_, err := s.db.Exec(
		"INSERT INTO users (email, password_hash) VALUES ($1, $2)",
		email, "hash",
	)
	require.NoError(s.T(), err)

	// Act
	var foundEmail string
	var id int64
	err = s.db.QueryRow(
		"SELECT id, email FROM users WHERE email = $1",
		email,
	).Scan(&id, &foundEmail)

	// Assert
	require.NoError(s.T(), err)
	assert.Equal(s.T(), email, foundEmail)
	assert.Greater(s.T(), id, int64(0))
}

func (s *IntegrationTestSuite) TestUserRepository_GetByEmail_NotFound() {
	// Act
	var email string
	err := s.db.QueryRow(
		"SELECT email FROM users WHERE email = $1",
		"nonexistent@example.com",
	).Scan(&email)

	// Assert
	assert.ErrorIs(s.T(), err, sql.ErrNoRows)
}

// ===== Todo Repository Tests =====

func (s *IntegrationTestSuite) TestTodoRepository_Create() {
	// Arrange - create user first
	var userID int64
	err := s.db.QueryRow(
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		"todo@example.com", "hash",
	).Scan(&userID)
	require.NoError(s.T(), err)

	// Act
	var todoID int64
	err = s.db.QueryRow(
		"INSERT INTO todos (user_id, title) VALUES ($1, $2) RETURNING id",
		userID, "Test Todo",
	).Scan(&todoID)

	// Assert
	require.NoError(s.T(), err)
	assert.Greater(s.T(), todoID, int64(0))
}

func (s *IntegrationTestSuite) TestTodoRepository_GetByUserID() {
	// Arrange
	var userID int64
	err := s.db.QueryRow(
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		"list@example.com", "hash",
	).Scan(&userID)
	require.NoError(s.T(), err)

	// Create multiple todos
	titles := []string{"Todo 1", "Todo 2", "Todo 3"}
	for _, title := range titles {
		_, err := s.db.Exec(
			"INSERT INTO todos (user_id, title) VALUES ($1, $2)",
			userID, title,
		)
		require.NoError(s.T(), err)
	}

	// Act
	rows, err := s.db.Query("SELECT title FROM todos WHERE user_id = $1 ORDER BY id", userID)
	require.NoError(s.T(), err)
	defer rows.Close()

	// Assert
	var foundTitles []string
	for rows.Next() {
		var title string
		err := rows.Scan(&title)
		require.NoError(s.T(), err)
		foundTitles = append(foundTitles, title)
	}
	assert.Equal(s.T(), titles, foundTitles)
}

func (s *IntegrationTestSuite) TestTodoRepository_Update() {
	// Arrange
	var userID int64
	err := s.db.QueryRow(
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		"update@example.com", "hash",
	).Scan(&userID)
	require.NoError(s.T(), err)

	var todoID int64
	err = s.db.QueryRow(
		"INSERT INTO todos (user_id, title, completed) VALUES ($1, $2, $3) RETURNING id",
		userID, "Original Title", false,
	).Scan(&todoID)
	require.NoError(s.T(), err)

	// Act
	_, err = s.db.Exec(
		"UPDATE todos SET title = $1, completed = $2, updated_at = NOW() WHERE id = $3",
		"Updated Title", true, todoID,
	)

	// Assert
	require.NoError(s.T(), err)

	var title string
	var completed bool
	err = s.db.QueryRow("SELECT title, completed FROM todos WHERE id = $1", todoID).Scan(&title, &completed)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Updated Title", title)
	assert.True(s.T(), completed)
}

func (s *IntegrationTestSuite) TestTodoRepository_Delete() {
	// Arrange
	var userID int64
	err := s.db.QueryRow(
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		"delete@example.com", "hash",
	).Scan(&userID)
	require.NoError(s.T(), err)

	var todoID int64
	err = s.db.QueryRow(
		"INSERT INTO todos (user_id, title) VALUES ($1, $2) RETURNING id",
		userID, "To Delete",
	).Scan(&todoID)
	require.NoError(s.T(), err)

	// Act
	result, err := s.db.Exec("DELETE FROM todos WHERE id = $1", todoID)
	require.NoError(s.T(), err)

	// Assert
	rowsAffected, _ := result.RowsAffected()
	assert.Equal(s.T(), int64(1), rowsAffected)

	// Verify deletion
	var count int
	err = s.db.QueryRow("SELECT COUNT(*) FROM todos WHERE id = $1", todoID).Scan(&count)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 0, count)
}

func (s *IntegrationTestSuite) TestTodoRepository_CascadeDelete() {
	// Arrange - create user with todos
	var userID int64
	err := s.db.QueryRow(
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		"cascade@example.com", "hash",
	).Scan(&userID)
	require.NoError(s.T(), err)

	for i := 0; i < 3; i++ {
		_, err := s.db.Exec(
			"INSERT INTO todos (user_id, title) VALUES ($1, $2)",
			userID, "Todo",
		)
		require.NoError(s.T(), err)
	}

	// Act - delete user
	_, err = s.db.Exec("DELETE FROM users WHERE id = $1", userID)
	require.NoError(s.T(), err)

	// Assert - todos should be deleted too
	var count int
	err = s.db.QueryRow("SELECT COUNT(*) FROM todos WHERE user_id = $1", userID).Scan(&count)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 0, count)
}

// TestIntegrationSuite runs the integration test suite.
func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}
	suite.Run(t, new(IntegrationTestSuite))
}
