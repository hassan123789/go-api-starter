// Package middleware provides HTTP middleware for the API.
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/zareh/go-api-starter/pkg/apperrors"
)

// RequestIDKey is the context key for request ID
const RequestIDKey = "request_id"

// RequestID middleware adds a unique request ID to each request
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if request ID already exists in header
			requestID := c.Request().Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Set request ID in context and response header
			c.Set(RequestIDKey, requestID)
			c.Response().Header().Set("X-Request-ID", requestID)

			return next(c)
		}
	}
}

// GetRequestID retrieves the request ID from context
func GetRequestID(c echo.Context) string {
	if id, ok := c.Get(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// StructuredLogger middleware provides structured logging using slog
func StructuredLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()

			// Create a logger with request context
			reqLogger := logger.With(
				slog.String("request_id", GetRequestID(c)),
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.String("remote_ip", c.RealIP()),
				slog.String("user_agent", req.UserAgent()),
			)

			// Store logger in context for handlers to use
			c.Set("logger", reqLogger)

			// Process request
			err := next(c)

			// Calculate duration
			duration := time.Since(start)
			status := c.Response().Status

			// Log based on status code
			logAttrs := []any{
				slog.Int("status", status),
				slog.Duration("duration", duration),
				slog.Int64("bytes_out", c.Response().Size),
			}

			if err != nil {
				logAttrs = append(logAttrs, slog.String("error", err.Error()))
				reqLogger.Error("request failed", logAttrs...)
			} else if status >= 500 {
				reqLogger.Error("server error", logAttrs...)
			} else if status >= 400 {
				reqLogger.Warn("client error", logAttrs...)
			} else {
				reqLogger.Info("request completed", logAttrs...)
			}

			return err
		}
	}
}

// GetLogger retrieves the logger from context
func GetLogger(c echo.Context) *slog.Logger {
	if logger, ok := c.Get("logger").(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// Recover middleware recovers from panics and logs the error
func Recover(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					// Get stack trace
					stack := debug.Stack()

					// Log the panic
					logger.Error("panic recovered",
						slog.String("request_id", GetRequestID(c)),
						slog.Any("panic", r),
						slog.String("stack", string(stack)),
					)

					// Return internal server error
					err := apperrors.ErrInternal.WithDetails("an unexpected error occurred")
					c.JSON(err.HTTPStatus, map[string]interface{}{
						"error": map[string]interface{}{
							"code":    err.Code,
							"message": err.Message,
						},
					})
				}
			}()
			return next(c)
		}
	}
}

// RateLimiter implements a simple in-memory rate limiter using token bucket algorithm
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // requests per interval
	interval time.Duration // time interval
}

type visitor struct {
	tokens    int
	lastVisit time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		interval: interval,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.interval * 2)
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastVisit) > rl.interval*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request from the given IP is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{
			tokens:    rl.rate - 1,
			lastVisit: time.Now(),
		}
		return true
	}

	// Refill tokens based on time elapsed
	elapsed := time.Since(v.lastVisit)
	tokensToAdd := int(elapsed/rl.interval) * rl.rate
	v.tokens += tokensToAdd
	if v.tokens > rl.rate {
		v.tokens = rl.rate
	}
	v.lastVisit = time.Now()

	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

// RateLimitMiddleware returns a middleware that limits requests per IP
func RateLimitMiddleware(limiter *RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()

			if !limiter.Allow(ip) {
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error": map[string]interface{}{
						"code":    "RATE_LIMIT_EXCEEDED",
						"message": "too many requests, please try again later",
					},
				})
			}

			return next(c)
		}
	}
}

// Timeout middleware adds a timeout to requests
func Timeout(timeout time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create a channel to signal completion
			done := make(chan error, 1)

			go func() {
				done <- next(c)
			}()

			select {
			case err := <-done:
				return err
			case <-time.After(timeout):
				return c.JSON(http.StatusGatewayTimeout, map[string]interface{}{
					"error": map[string]interface{}{
						"code":    "TIMEOUT",
						"message": "request timed out",
					},
				})
			}
		}
	}
}

// CORS middleware with configurable options
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns sensible CORS defaults
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

// ErrorHandler is a custom error handler that returns structured errors
func ErrorHandler(logger *slog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		code := http.StatusInternalServerError
		message := "internal server error"
		errorCode := "INTERNAL_ERROR"
		var details string

		// Check if it's an Echo HTTP error
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			if m, ok := he.Message.(string); ok {
				message = m
			}
			errorCode = http.StatusText(code)
		}

		// Check if it's our custom AppError
		if appErr, ok := apperrors.GetAppError(err); ok {
			code = appErr.HTTPStatus
			message = appErr.Message
			errorCode = string(appErr.Code)
			details = appErr.Details
		}

		// Log the error
		logger.Error("request error",
			slog.String("request_id", GetRequestID(c)),
			slog.Int("status", code),
			slog.String("error", err.Error()),
		)

		// Build response
		response := map[string]interface{}{
			"error": map[string]interface{}{
				"code":    errorCode,
				"message": message,
			},
		}

		if details != "" {
			response["error"].(map[string]interface{})["details"] = details
		}

		c.JSON(code, response)
	}
}
