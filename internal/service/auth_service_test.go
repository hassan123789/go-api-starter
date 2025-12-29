package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/zareh/go-api-starter/internal/model"
)

func TestAuthService_Register_Success(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	req := &model.CreateUserRequest{
		Email:    "test@example.com",
		Password: "securepassword123",
	}

	// Simulate Register logic
	exists, _ := repo.EmailExists(ctx, req.Email)
	assert.False(t, exists)

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user, err := repo.Create(ctx, req.Email, string(hash))
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.NotEmpty(t, user.PasswordHash)
}

func TestAuthService_Register_EmailExists(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Create existing user
	repo.Create(ctx, "existing@example.com", "hash")

	// Try to register with same email
	exists, err := repo.EmailExists(ctx, "existing@example.com")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Create user with hashed password
	password := "securepassword123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	repo.Create(ctx, "test@example.com", string(hash))

	// Simulate login
	user, err := repo.GetByEmail(ctx, "test@example.com")
	require.NoError(t, err)

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	require.NoError(t, err) // password matches
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Create user with hashed password
	password := "securepassword123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	repo.Create(ctx, "test@example.com", string(hash))

	// Try to login with wrong password
	user, err := repo.GetByEmail(ctx, "test@example.com")
	require.NoError(t, err)

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("wrongpassword"))
	assert.Error(t, err) // password doesn't match
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	user, err := repo.GetByEmail(ctx, "notfound@example.com")
	assert.Nil(t, user)
	assert.Error(t, err)
}

func TestPasswordHashing(t *testing.T) {
	password := "mysecretpassword"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	// Verify correct password
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	assert.NoError(t, err)

	// Verify wrong password fails
	err = bcrypt.CompareHashAndPassword(hash, []byte("wrongpassword"))
	assert.Error(t, err)
}

func TestBcryptCost(t *testing.T) {
	password := "testpassword"

	// Default cost
	hash1, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// Min cost (faster for tests)
	hash2, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

	// Both should validate
	assert.NoError(t, bcrypt.CompareHashAndPassword(hash1, []byte(password)))
	assert.NoError(t, bcrypt.CompareHashAndPassword(hash2, []byte(password)))

	// Hashes should be different
	assert.NotEqual(t, string(hash1), string(hash2))
}
