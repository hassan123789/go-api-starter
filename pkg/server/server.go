// Package server provides a configurable HTTP server with functional options pattern.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
)

// Server wraps echo.Echo with additional configuration and lifecycle management.
type Server struct {
	echo            *echo.Echo
	port            int
	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	shutdownTimeout time.Duration
	logger          *slog.Logger
	onShutdown      []func()
}

// Option is a functional option for configuring the server.
type Option func(*Server)

// WithPort sets the server port.
func WithPort(port int) Option {
	return func(s *Server) {
		s.port = port
	}
}

// WithReadTimeout sets the read timeout.
func WithReadTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.readTimeout = timeout
	}
}

// WithWriteTimeout sets the write timeout.
func WithWriteTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.writeTimeout = timeout
	}
}

// WithIdleTimeout sets the idle timeout.
func WithIdleTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.idleTimeout = timeout
	}
}

// WithShutdownTimeout sets the graceful shutdown timeout.
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = timeout
	}
}

// WithLogger sets the logger.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithTimeouts sets all timeouts at once.
func WithTimeouts(read, write, idle, shutdown time.Duration) Option {
	return func(s *Server) {
		s.readTimeout = read
		s.writeTimeout = write
		s.idleTimeout = idle
		s.shutdownTimeout = shutdown
	}
}

// OnShutdown registers a function to be called during graceful shutdown.
func OnShutdown(fn func()) Option {
	return func(s *Server) {
		s.onShutdown = append(s.onShutdown, fn)
	}
}

// New creates a new server with the given options.
func New(opts ...Option) *Server {
	s := &Server{
		echo:            echo.New(),
		port:            8080,
		readTimeout:     15 * time.Second,
		writeTimeout:    15 * time.Second,
		idleTimeout:     60 * time.Second,
		shutdownTimeout: 10 * time.Second,
		logger:          slog.Default(),
		onShutdown:      make([]func(), 0),
	}

	s.echo.HideBanner = true
	s.echo.HidePort = true

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Echo returns the underlying echo instance for route registration.
func (s *Server) Echo() *echo.Echo {
	return s.echo
}

// Start starts the server and blocks until shutdown signal is received.
func (s *Server) Start() error {
	// Configure server
	s.echo.Server.ReadTimeout = s.readTimeout
	s.echo.Server.WriteTimeout = s.writeTimeout
	s.echo.Server.IdleTimeout = s.idleTimeout

	// Start server in goroutine
	addr := fmt.Sprintf(":%d", s.port)
	go func() {
		s.logger.Info("Starting server", "address", addr)
		if err := s.echo.Start(addr); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Server error", "error", err)
		}
	}()

	// Wait for interrupt signal
	return s.waitForShutdown()
}

// StartTLS starts the server with TLS and blocks until shutdown signal is received.
func (s *Server) StartTLS(certFile, keyFile string) error {
	s.echo.Server.ReadTimeout = s.readTimeout
	s.echo.Server.WriteTimeout = s.writeTimeout
	s.echo.Server.IdleTimeout = s.idleTimeout

	addr := fmt.Sprintf(":%d", s.port)
	go func() {
		s.logger.Info("Starting TLS server", "address", addr)
		if err := s.echo.StartTLS(addr, certFile, keyFile); err != nil && err != http.ErrServerClosed {
			s.logger.Error("TLS server error", "error", err)
		}
	}()

	return s.waitForShutdown()
}

// waitForShutdown waits for a shutdown signal and performs graceful shutdown.
func (s *Server) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutting down server...")

	// Run shutdown hooks
	for _, fn := range s.onShutdown {
		fn()
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.echo.Shutdown(ctx); err != nil {
		s.logger.Error("Graceful shutdown failed", "error", err)
		return err
	}

	s.logger.Info("Server shutdown completed")
	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	for _, fn := range s.onShutdown {
		fn()
	}
	return s.echo.Shutdown(ctx)
}

// Builder provides a fluent interface for building a server.
type Builder struct {
	opts []Option
}

// NewBuilder creates a new server builder.
func NewBuilder() *Builder {
	return &Builder{
		opts: make([]Option, 0),
	}
}

// Port sets the server port.
func (b *Builder) Port(port int) *Builder {
	b.opts = append(b.opts, WithPort(port))
	return b
}

// ReadTimeout sets the read timeout.
func (b *Builder) ReadTimeout(timeout time.Duration) *Builder {
	b.opts = append(b.opts, WithReadTimeout(timeout))
	return b
}

// WriteTimeout sets the write timeout.
func (b *Builder) WriteTimeout(timeout time.Duration) *Builder {
	b.opts = append(b.opts, WithWriteTimeout(timeout))
	return b
}

// IdleTimeout sets the idle timeout.
func (b *Builder) IdleTimeout(timeout time.Duration) *Builder {
	b.opts = append(b.opts, WithIdleTimeout(timeout))
	return b
}

// ShutdownTimeout sets the shutdown timeout.
func (b *Builder) ShutdownTimeout(timeout time.Duration) *Builder {
	b.opts = append(b.opts, WithShutdownTimeout(timeout))
	return b
}

// Logger sets the logger.
func (b *Builder) Logger(logger *slog.Logger) *Builder {
	b.opts = append(b.opts, WithLogger(logger))
	return b
}

// OnShutdown adds a shutdown hook.
func (b *Builder) OnShutdown(fn func()) *Builder {
	b.opts = append(b.opts, OnShutdown(fn))
	return b
}

// Build creates the server with all configured options.
func (b *Builder) Build() *Server {
	return New(b.opts...)
}
