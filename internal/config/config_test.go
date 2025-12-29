package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	require.NoError(t, os.Setenv(key, value))
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	require.NoError(t, os.Unsetenv(key))
}

func TestLoad_Success(t *testing.T) {
	// Setup environment
	setEnv(t, "DATABASE_URL", "postgres://localhost/test")
	setEnv(t, "JWT_SECRET", "test-secret")
	setEnv(t, "PORT", "3000")
	setEnv(t, "JWT_EXPIRY", "48")
	setEnv(t, "GRPC_PORT", "9091")
	setEnv(t, "GRPC_ENABLED", "true")
	t.Cleanup(func() {
		unsetEnv(t, "DATABASE_URL")
		unsetEnv(t, "JWT_SECRET")
		unsetEnv(t, "PORT")
		unsetEnv(t, "JWT_EXPIRY")
		unsetEnv(t, "GRPC_PORT")
		unsetEnv(t, "GRPC_ENABLED")
	})

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
	setEnv(t, "DATABASE_URL", "postgres://localhost/test")
	setEnv(t, "JWT_SECRET", "test-secret")
	unsetEnv(t, "PORT")
	unsetEnv(t, "JWT_EXPIRY")
	unsetEnv(t, "GRPC_PORT")
	unsetEnv(t, "GRPC_ENABLED")
	t.Cleanup(func() {
		unsetEnv(t, "DATABASE_URL")
		unsetEnv(t, "JWT_SECRET")
	})

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)     // default
	assert.Equal(t, 24, cfg.JWTExpiry)    // default
	assert.Equal(t, "9090", cfg.GRPCPort) // default
	assert.False(t, cfg.GRPCEnabled)      // default (not "true")
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	unsetEnv(t, "DATABASE_URL")
	setEnv(t, "JWT_SECRET", "test-secret")
	t.Cleanup(func() {
		unsetEnv(t, "JWT_SECRET")
	})

	cfg, err := Load()
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL")
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	setEnv(t, "DATABASE_URL", "postgres://localhost/test")
	unsetEnv(t, "JWT_SECRET")
	t.Cleanup(func() {
		unsetEnv(t, "DATABASE_URL")
	})

	cfg, err := Load()
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

func TestLoad_InvalidJWTExpiry(t *testing.T) {
	setEnv(t, "DATABASE_URL", "postgres://localhost/test")
	setEnv(t, "JWT_SECRET", "test-secret")
	setEnv(t, "JWT_EXPIRY", "invalid")
	t.Cleanup(func() {
		unsetEnv(t, "DATABASE_URL")
		unsetEnv(t, "JWT_SECRET")
		unsetEnv(t, "JWT_EXPIRY")
	})

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 24, cfg.JWTExpiry) // falls back to default
}

func TestLoad_GRPCEnabledVariations(t *testing.T) {
	setEnv(t, "DATABASE_URL", "postgres://localhost/test")
	setEnv(t, "JWT_SECRET", "test-secret")
	t.Cleanup(func() {
		unsetEnv(t, "DATABASE_URL")
		unsetEnv(t, "JWT_SECRET")
		unsetEnv(t, "GRPC_ENABLED")
	})

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
				unsetEnv(t, "GRPC_ENABLED")
			} else {
				setEnv(t, "GRPC_ENABLED", tc.value)
			}
			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, tc.expected, cfg.GRPCEnabled)
		})
	}
}
