package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Port        string
	GRPCPort    string
	DatabaseURL string
	JWTSecret   string
	JWTExpiry   int  // hours
	GRPCEnabled bool // Enable gRPC server
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	jwtExpiry := 24 // default 24 hours
	if exp := os.Getenv("JWT_EXPIRY"); exp != "" {
		if parsed, err := strconv.Atoi(exp); err == nil {
			jwtExpiry = parsed
		}
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}

	grpcEnabled := os.Getenv("GRPC_ENABLED") == "true"

	return &Config{
		Port:        port,
		GRPCPort:    grpcPort,
		DatabaseURL: databaseURL,
		JWTSecret:   jwtSecret,
		JWTExpiry:   jwtExpiry,
		GRPCEnabled: grpcEnabled,
	}, nil
}
