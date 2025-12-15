package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/zareh/go-api-starter/internal/config"
	"github.com/zareh/go-api-starter/internal/handler"
	custommw "github.com/zareh/go-api-starter/internal/middleware"
	"github.com/zareh/go-api-starter/internal/repository"
	"github.com/zareh/go-api-starter/internal/service"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Connect to database
	db, err := repository.NewDB(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	todoRepo := repository.NewTodoRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiry)
	todoService := service.NewTodoService(todoRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	todoHandler := handler.NewTodoHandler(todoService)

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true

	// Custom error handler
	e.HTTPErrorHandler = custommw.ErrorHandler(logger)

	// Initialize rate limiter: 100 requests per second
	rateLimiter := custommw.NewRateLimiter(100, time.Second)

	// Middleware - Order matters!
	e.Use(custommw.RequestID())                      // Add request ID first
	e.Use(custommw.StructuredLogger(logger))         // Structured logging with slog
	e.Use(custommw.Recover(logger))                  // Panic recovery with stack trace
	e.Use(custommw.RateLimitMiddleware(rateLimiter)) // Rate limiter
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"X-Request-ID",
		},
	}))

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	// API v1 routes
	v1 := e.Group("/api/v1")
	{
		// Public routes (no authentication required)
		v1.POST("/users", authHandler.Register)
		v1.POST("/auth/login", authHandler.Login)

		// Protected routes (authentication required)
		auth := v1.Group("")
		auth.Use(echojwt.WithConfig(echojwt.Config{
			SigningKey: []byte(cfg.JWTSecret),
			Skipper: func(c echo.Context) bool {
				// Explicitly skip auth for public endpoints
				switch c.Path() {
				case "/health", "/api/v1/users", "/api/v1/auth/login":
					return true
				default:
					return false
				}
			},
		}))
		{
			auth.GET("/todos", todoHandler.List)
			auth.POST("/todos", todoHandler.Create)
			auth.GET("/todos/:id", todoHandler.Get)
			auth.PUT("/todos/:id", todoHandler.Update)
			auth.DELETE("/todos/:id", todoHandler.Delete)
		}
	}

	// Start server
	go func() {
		addr := ":" + cfg.Port
		slog.Info("Starting server", "address", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown server", "error", err)
		os.Exit(1)
	}

	slog.Info("Server shutdown completed")
}
