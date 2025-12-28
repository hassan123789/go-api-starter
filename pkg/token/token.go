// Package token provides JWT and refresh token management.
package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Token types
const (
	TypeAccess  = "access"
	TypeRefresh = "refresh"
)

// Errors
var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidTokenType = errors.New("invalid token type")
	ErrTokenRevoked     = errors.New("token has been revoked")
)

// Claims represents JWT claims with custom fields.
type Claims struct {
	jwt.RegisteredClaims
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
}

// Pair represents an access/refresh token pair.
type Pair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"` // seconds
	ExpiresAt    time.Time `json:"expires_at"`
}

// Manager handles token generation and validation.
type Manager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	issuer        string
	audience      []string
	signingMethod jwt.SigningMethod
}

// ManagerOption is a functional option for Manager.
type ManagerOption func(*Manager)

// WithAccessExpiry sets the access token expiry duration.
func WithAccessExpiry(d time.Duration) ManagerOption {
	return func(m *Manager) {
		m.accessExpiry = d
	}
}

// WithRefreshExpiry sets the refresh token expiry duration.
func WithRefreshExpiry(d time.Duration) ManagerOption {
	return func(m *Manager) {
		m.refreshExpiry = d
	}
}

// WithIssuer sets the token issuer.
func WithIssuer(issuer string) ManagerOption {
	return func(m *Manager) {
		m.issuer = issuer
	}
}

// WithAudience sets the token audience.
func WithAudience(audience []string) ManagerOption {
	return func(m *Manager) {
		m.audience = audience
	}
}

// NewManager creates a new token manager.
func NewManager(accessSecret, refreshSecret string, opts ...ManagerOption) *Manager {
	m := &Manager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessExpiry:  15 * time.Minute,   // Default: 15 minutes
		refreshExpiry: 7 * 24 * time.Hour, // Default: 7 days
		issuer:        "go-api-starter",
		audience:      []string{"go-api-starter"},
		signingMethod: jwt.SigningMethodHS256,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// GenerateTokenPair generates both access and refresh tokens.
func (m *Manager) GenerateTokenPair(userID int64, email, role string) (*Pair, error) {
	now := time.Now()

	// Generate access token
	accessToken, err := m.generateToken(userID, email, role, TypeAccess, m.accessSecret, now, m.accessExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := m.generateToken(userID, email, role, TypeRefresh, m.refreshSecret, now, m.refreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &Pair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(m.accessExpiry.Seconds()),
		ExpiresAt:    now.Add(m.accessExpiry),
	}, nil
}

func (m *Manager) generateToken(userID int64, email, role, tokenType string, secret []byte, now time.Time, expiry time.Duration) (string, error) {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			Audience:  m.audience,
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        generateTokenID(),
		},
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
	}

	token := jwt.NewWithClaims(m.signingMethod, claims)
	return token.SignedString(secret)
}

// ValidateAccessToken validates an access token and returns its claims.
func (m *Manager) ValidateAccessToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, TypeAccess, m.accessSecret)
}

// ValidateRefreshToken validates a refresh token and returns its claims.
func (m *Manager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, TypeRefresh, m.refreshSecret)
}

func (m *Manager) validateToken(tokenString, expectedType string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if token.Method != m.signingMethod {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Validate token type
	if claims.TokenType != expectedType {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token.
func (m *Manager) RefreshAccessToken(refreshToken string) (*Pair, error) {
	claims, err := m.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return m.GenerateTokenPair(claims.UserID, claims.Email, claims.Role)
}

// GetAccessExpiry returns the access token expiry duration.
func (m *Manager) GetAccessExpiry() time.Duration {
	return m.accessExpiry
}

// GetRefreshExpiry returns the refresh token expiry duration.
func (m *Manager) GetRefreshExpiry() time.Duration {
	return m.refreshExpiry
}

// generateTokenID generates a unique token ID.
func generateTokenID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b) // crypto/rand.Read always succeeds on supported platforms
	return base64.RawURLEncoding.EncodeToString(b)
}

// HashToken hashes a token for secure storage.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// GenerateOpaqueToken generates a random opaque token.
func GenerateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
