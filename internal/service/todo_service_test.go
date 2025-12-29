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

// MockTodoRepository is a mock implementation of TodoRepository for testing.
type MockTodoRepository struct {
	todos        map[int64]*model.Todo
	nextID       int64
	createErr    error
	getByIDErr   error
	getByUserErr error
	updateErr    error
	deleteErr    error
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
	if todo, ok := m.todos[id]; ok {
		return todo, nil
	}
	return nil, sql.ErrNoRows
}

func (m *MockTodoRepository) GetByUserID(ctx context.Context, userID int64) ([]*model.Todo, error) {
	if m.getByUserErr != nil {
		return nil, m.getByUserErr
	}
	var result []*model.Todo
	for _, todo := range m.todos {
		if todo.UserID == userID {
			result = append(result, todo)
		}
	}
	return result, nil
}

func (m *MockTodoRepository) GetByUserIDWithPagination(ctx context.Context, userID int64, limit, offset int) ([]*model.Todo, int64, error) {
	todos, err := m.GetByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(todos))

	if offset >= len(todos) {
		return []*model.Todo{}, total, nil
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

// MockUserRepository is a mock implementation of UserRepository for testing.
type MockUserRepository struct {
	users     map[int64]*model.User
	byEmail   map[string]*model.User
	nextID    int64
	createErr error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:   make(map[int64]*model.User),
		byEmail: make(map[string]*model.User),
		nextID:  1,
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
	m.byEmail[email] = user
	m.nextID++
	return user, nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, sql.ErrNoRows
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if user, ok := m.byEmail[email]; ok {
		return user, nil
	}
	return nil, sql.ErrNoRows
}

func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	_, exists := m.byEmail[email]
	return exists, nil
}

func (m *MockUserRepository) Update(ctx context.Context, id int64, email, passwordHash *string) (*model.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, sql.ErrNoRows
	}

	if email != nil {
		delete(m.byEmail, user.Email)
		user.Email = *email
		m.byEmail[*email] = user
	}
	if passwordHash != nil {
		user.PasswordHash = *passwordHash
	}
	user.UpdatedAt = time.Now()

	return user, nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	user, ok := m.users[id]
	if !ok {
		return sql.ErrNoRows
	}
	delete(m.byEmail, user.Email)
	delete(m.users, id)
	return nil
}

// Tests for TodoRepository mock
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

	created, err := repo.Create(ctx, 1, "Test Todo")
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Title, retrieved.Title)
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
	_, err := repo.Create(ctx, 1, "Todo 1")
	require.NoError(t, err)
	_, err = repo.Create(ctx, 1, "Todo 2")
	require.NoError(t, err)
	// Create todo for user 2
	_, err = repo.Create(ctx, 2, "Todo 3")
	require.NoError(t, err)

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
	created, err := repo.Create(ctx, 1, "Original Title")
	require.NoError(t, err)

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
	created, err := repo.Create(ctx, 1, "Original Title")
	require.NoError(t, err)

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
	created, err := repo.Create(ctx, 1, "Test Todo")
	require.NoError(t, err)

	// Delete it
	err = repo.Delete(ctx, created.ID)
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
	_, err = repo.Create(ctx, "test@example.com", "hashedpassword")
	require.NoError(t, err)
	exists, err = repo.EmailExists(ctx, "test@example.com")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	_, err := repo.Create(ctx, "test@example.com", "hashedpassword")
	require.NoError(t, err)

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
		_, err := repo.Create(ctx, 1, "Todo")
		require.NoError(t, err)
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
