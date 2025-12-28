package token_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zareh/go-api-starter/pkg/token"
)

func TestManager_GenerateTokenPair(t *testing.T) {
	manager := token.NewManager("access-secret", "refresh-secret")

	pair, err := manager.GenerateTokenPair(1, "test@example.com", "user")
	require.NoError(t, err)

	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.Equal(t, "Bearer", pair.TokenType)
	assert.Greater(t, pair.ExpiresIn, int64(0))
	assert.True(t, pair.ExpiresAt.After(time.Now()))
}

func TestManager_ValidateAccessToken(t *testing.T) {
	manager := token.NewManager("access-secret", "refresh-secret")

	// Generate token
	pair, err := manager.GenerateTokenPair(42, "user@test.com", "admin")
	require.NoError(t, err)

	// Validate access token
	claims, err := manager.ValidateAccessToken(pair.AccessToken)
	require.NoError(t, err)

	assert.Equal(t, int64(42), claims.UserID)
	assert.Equal(t, "user@test.com", claims.Email)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, token.TypeAccess, claims.TokenType)
}

func TestManager_ValidateRefreshToken(t *testing.T) {
	manager := token.NewManager("access-secret", "refresh-secret")

	// Generate token
	pair, err := manager.GenerateTokenPair(42, "user@test.com", "user")
	require.NoError(t, err)

	// Validate refresh token
	claims, err := manager.ValidateRefreshToken(pair.RefreshToken)
	require.NoError(t, err)

	assert.Equal(t, int64(42), claims.UserID)
	assert.Equal(t, token.TypeRefresh, claims.TokenType)
}

func TestManager_InvalidTokenType(t *testing.T) {
	// Use same secret for both to test token type validation
	// (in production, secrets would be different)
	manager := token.NewManager("same-secret", "same-secret")

	pair, err := manager.GenerateTokenPair(1, "test@example.com", "user")
	require.NoError(t, err)

	// Try to validate access token as refresh token - should fail due to token type mismatch
	_, err = manager.ValidateRefreshToken(pair.AccessToken)
	assert.ErrorIs(t, err, token.ErrInvalidTokenType)

	// Try to validate refresh token as access token - should fail due to token type mismatch
	_, err = manager.ValidateAccessToken(pair.RefreshToken)
	assert.ErrorIs(t, err, token.ErrInvalidTokenType)
}

func TestManager_ExpiredToken(t *testing.T) {
	// Create manager with very short expiry
	manager := token.NewManager(
		"access-secret",
		"refresh-secret",
		token.WithAccessExpiry(1*time.Millisecond),
	)

	pair, err := manager.GenerateTokenPair(1, "test@example.com", "user")
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Should return expired error
	_, err = manager.ValidateAccessToken(pair.AccessToken)
	assert.ErrorIs(t, err, token.ErrExpiredToken)
}

func TestManager_RefreshAccessToken(t *testing.T) {
	manager := token.NewManager("access-secret", "refresh-secret")

	// Generate initial tokens
	pair1, err := manager.GenerateTokenPair(1, "test@example.com", "user")
	require.NoError(t, err)

	// Small delay to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	// Refresh tokens
	pair2, err := manager.RefreshAccessToken(pair1.RefreshToken)
	require.NoError(t, err)

	// New access token should be different
	assert.NotEqual(t, pair1.AccessToken, pair2.AccessToken)

	// New access token should be valid
	claims, err := manager.ValidateAccessToken(pair2.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, int64(1), claims.UserID)
}

func TestManager_WithOptions(t *testing.T) {
	manager := token.NewManager(
		"access-secret",
		"refresh-secret",
		token.WithAccessExpiry(30*time.Minute),
		token.WithRefreshExpiry(14*24*time.Hour),
		token.WithIssuer("test-issuer"),
		token.WithAudience([]string{"test-audience"}),
	)

	assert.Equal(t, 30*time.Minute, manager.GetAccessExpiry())
	assert.Equal(t, 14*24*time.Hour, manager.GetRefreshExpiry())

	// Generate and validate token
	pair, err := manager.GenerateTokenPair(1, "test@example.com", "user")
	require.NoError(t, err)

	claims, err := manager.ValidateAccessToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "test-issuer", claims.Issuer)
	assert.Contains(t, claims.Audience, "test-audience")
}

func TestHashToken(t *testing.T) {
	token1 := "test-token-1"
	token2 := "test-token-2"

	hash1 := token.HashToken(token1)
	hash2 := token.HashToken(token2)

	// Same input should produce same hash
	assert.Equal(t, hash1, token.HashToken(token1))

	// Different inputs should produce different hashes
	assert.NotEqual(t, hash1, hash2)

	// Hash should not be empty
	assert.NotEmpty(t, hash1)
}

func TestGenerateOpaqueToken(t *testing.T) {
	token1, err := token.GenerateOpaqueToken()
	require.NoError(t, err)

	token2, err := token.GenerateOpaqueToken()
	require.NoError(t, err)

	// Tokens should be unique
	assert.NotEqual(t, token1, token2)

	// Tokens should have reasonable length
	assert.Greater(t, len(token1), 20)
}
