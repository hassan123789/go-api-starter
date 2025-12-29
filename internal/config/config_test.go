package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Success(t *testing.T) {
	// Setup environment
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("PORT", "3000")
	os.Setenv("JWT_EXPIRY", "48")
	os.Setenv("GRPC_PORT", "9091")
	os.Setenv("GRPC_ENABLED", "true")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("PORT")
		os.Unsetenv("JWT_EXPIRY")
		os.Unsetenv("GRPC_PORT")
		os.Unsetenv("GRPC_ENABLED")
	}()

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "3000", cfg.Port)
	assert.Equal(t, "postgres://localhost/test", cfg.DatabaseURL)
	assert.Equal(t, "test-secret", cfg.JWTSecret)
	assert.Equal(t, 48, cfg.JWTExpiry)
	assert.Equal(t, "9091", cfg.GRPCPort)
	assert.True(t, cfg.GRPCEnabled)
}

func TestLoad_Defaults(t *testing.T) {
	// Setup required environment only
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Unsetenv("PORT")
	os.Unsetenv("JWT_EXPIRY")
	os.Unsetenv("GRPC_PORT")
	os.Unsetenv("GRPC_ENABLED")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)     // default
	assert.Equal(t, 24, cfg.JWTExpiry)    // default
	assert.Equal(t, "9090", cfg.GRPCPort) // default
	assert.False(t, cfg.GRPCEnabled)      // default (not "true")
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	cfg, err := Load()
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL")
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("DATABASE_URL")

	cfg, err := Load()
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

func TestLoad_InvalidJWTExpiry(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_EXPIRY", "invalid")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_EXPIRY")
	}()

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 24, cfg.JWTExpiry) // falls back to default
}

func TestLoad_GRPCEnabledVariations(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("JWT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("GRPC_ENABLED")
	}()

	testCases := []struct {
		name     string
		value    string
		expected bool
	}{
		{"true", "true", true},
		{"false", "false", false},
		{"empty", "", false},
		{"TRUE", "TRUE", false}, // case sensitive
		{"1", "1", false},       // only "true" works
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.value == "" {
				os.Unsetenv("GRPC_ENABLED")
			} else {
				os.Setenv("GRPC_ENABLED", tc.value)
			}
			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, tc.expected, cfg.GRPCEnabled)
		})
	}
}
