package main

import (
	"context"
	"log"
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
	"github.com/zareh/go-api-starter/internal/repository"
	"github.com/zareh/go-api-starter/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := repository.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
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

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
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
		log.Printf("Starting server on %s", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}

	log.Println("Server shutdown completed")
}
