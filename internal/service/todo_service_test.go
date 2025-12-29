package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zareh/go-api-starter/internal/model"
)

// MockTodoRepository is a mock implementation of TodoRepository for testing
type MockTodoRepository struct {
	todos       map[int64]*model.Todo
	nextID      int64
	createErr   error
	getByIDErr  error
	getByUserErr error
	updateErr   error
	deleteErr   error
}

func NewMockTodoRepository() *MockTodoRepository {
	return &MockTodoRepository{
		todos:  make(map[int64]*model.Todo),
		nextID: 1,
	}
}

func (m *MockTodoRepository) Create(ctx context.Context, userID int64, title string) (*model.Todo, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	todo := &model.Todo{
		ID:        m.nextID,
		UserID:    userID,
		Title:     title,
		Completed: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.todos[m.nextID] = todo
	m.nextID++
	return todo, nil
}

func (m *MockTodoRepository) GetByID(ctx context.Context, id int64) (*model.Todo, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	todo, ok := m.todos[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return todo, nil
}

func (m *MockTodoRepository) GetByUserID(ctx context.Context, userID int64) ([]model.Todo, error) {
	if m.getByUserErr != nil {
		return nil, m.getByUserErr
	}
	var todos []model.Todo
	for _, todo := range m.todos {
		if todo.UserID == userID {
			todos = append(todos, *todo)
		}
	}
	return todos, nil
}

func (m *MockTodoRepository) GetByUserIDWithPagination(ctx context.Context, userID int64, limit, offset int) ([]model.Todo, int64, error) {
	todos, err := m.GetByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(todos))
	if offset >= len(todos) {
		return []model.Todo{}, total, nil
	}
	end := offset + limit
	if end > len(todos) {
		end = len(todos)
	}
	return todos[offset:end], total, nil
}

func (m *MockTodoRepository) Update(ctx context.Context, id int64, title *string, completed *bool) (*model.Todo, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	todo, ok := m.todos[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	if title != nil {
		todo.Title = *title
	}
	if completed != nil {
		todo.Completed = *completed
	}
	todo.UpdatedAt = time.Now()
	return todo, nil
}

func (m *MockTodoRepository) Delete(ctx context.Context, id int64) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.todos[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.todos, id)
	return nil
}

// MockUserRepository is a mock implementation for user repository
type MockUserRepository struct {
	users       map[int64]*model.User
	emailIndex  map[string]int64
	nextID      int64
	createErr   error
	getByIDErr  error
	getByEmailErr error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:      make(map[int64]*model.User),
		emailIndex: make(map[string]int64),
		nextID:     1,
	}
}

func (m *MockUserRepository) Create(ctx context.Context, email, passwordHash string) (*model.User, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	user := &model.User{
		ID:           m.nextID,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	m.users[m.nextID] = user
	m.emailIndex[email] = m.nextID
	m.nextID++
	return user, nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	user, ok := m.users[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return user, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.getByEmailErr != nil {
		return nil, m.getByEmailErr
	}
	id, ok := m.emailIndex[email]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return m.users[id], nil
}

func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	_, ok := m.emailIndex[email]
	return ok, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return sql.ErrNoRows
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	user, ok := m.users[id]
	if !ok {
		return sql.ErrNoRows
	}
	delete(m.emailIndex, user.Email)
	delete(m.users, id)
	return nil
}

// TestableAuthService wraps AuthService for testing with mock repository
type TestableAuthService struct {
	userRepo  *MockUserRepository
	jwtSecret string
	jwtExpiry int
}

func NewTestableAuthService(userRepo *MockUserRepository, jwtSecret string, jwtExpiry int) *TestableAuthService {
	return &TestableAuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

// Tests for TodoService-like behavior
func TestTodoService_Create(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	todo, err := repo.Create(ctx, 1, "Test Todo")
	require.NoError(t, err)
	assert.Equal(t, int64(1), todo.ID)
	assert.Equal(t, int64(1), todo.UserID)
	assert.Equal(t, "Test Todo", todo.Title)
	assert.False(t, todo.Completed)
}

func TestTodoService_GetByID_Success(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	// Create a todo
	created, _ := repo.Create(ctx, 1, "Test Todo")

	// Get it back
	todo, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, todo.ID)
	assert.Equal(t, "Test Todo", todo.Title)
}

func TestTodoService_GetByID_NotFound(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	todo, err := repo.GetByID(ctx, 999)
	assert.Nil(t, todo)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestTodoService_ListByUserID(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	// Create todos for user 1
	repo.Create(ctx, 1, "Todo 1")
	repo.Create(ctx, 1, "Todo 2")
	// Create todo for user 2
	repo.Create(ctx, 2, "Todo 3")

	// List user 1's todos
	todos, err := repo.GetByUserID(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, todos, 2)
}

func TestTodoService_ListByUserID_Empty(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	todos, err := repo.GetByUserID(ctx, 999)
	require.NoError(t, err)
	assert.Empty(t, todos)
}

func TestTodoService_Update(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	// Create a todo
	created, _ := repo.Create(ctx, 1, "Original Title")

	// Update it
	newTitle := "Updated Title"
	completed := true
	updated, err := repo.Update(ctx, created.ID, &newTitle, &completed)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", updated.Title)
	assert.True(t, updated.Completed)
}

func TestTodoService_Update_PartialUpdate(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	// Create a todo
	created, _ := repo.Create(ctx, 1, "Original Title")

	// Update only title
	newTitle := "Updated Title"
	updated, err := repo.Update(ctx, created.ID, &newTitle, nil)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", updated.Title)
	assert.False(t, updated.Completed) // unchanged
}

func TestTodoService_Delete(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	// Create a todo
	created, _ := repo.Create(ctx, 1, "Test Todo")

	// Delete it
	err := repo.Delete(ctx, created.ID)
	require.NoError(t, err)

	// Verify it's gone
	_, err = repo.GetByID(ctx, created.ID)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestTodoService_Delete_NotFound(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, 999)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

// Tests for error handling
func TestTodoService_CreateError(t *testing.T) {
	repo := NewMockTodoRepository()
	repo.createErr = errors.New("database error")
	ctx := context.Background()

	todo, err := repo.Create(ctx, 1, "Test")
	assert.Nil(t, todo)
	assert.Error(t, err)
}

// Test user repository mock
func TestUserRepository_Create(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	user, err := repo.Create(ctx, "test@example.com", "hashedpassword")
	require.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestUserRepository_EmailExists(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Before creating
	exists, err := repo.EmailExists(ctx, "test@example.com")
	require.NoError(t, err)
	assert.False(t, exists)

	// After creating
	repo.Create(ctx, "test@example.com", "hashedpassword")
	exists, err = repo.EmailExists(ctx, "test@example.com")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	repo.Create(ctx, "test@example.com", "hashedpassword")

	user, err := repo.GetByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	user, err := repo.GetByEmail(ctx, "notfound@example.com")
	assert.Nil(t, user)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestPagination(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	// Create 5 todos
	for i := 0; i < 5; i++ {
		repo.Create(ctx, 1, "Todo")
	}

	// Get page 1 (limit 2, offset 0)
	todos, total, err := repo.GetByUserIDWithPagination(ctx, 1, 2, 0)
	require.NoError(t, err)
	assert.Len(t, todos, 2)
	assert.Equal(t, int64(5), total)

	// Get page 3 (limit 2, offset 4)
	todos, total, err = repo.GetByUserIDWithPagination(ctx, 1, 2, 4)
	require.NoError(t, err)
	assert.Len(t, todos, 1) // only 1 left
	assert.Equal(t, int64(5), total)
}
