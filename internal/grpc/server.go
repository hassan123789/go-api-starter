package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	todov1 "github.com/zareh/go-api-starter/gen/go/todo/v1"
	"github.com/zareh/go-api-starter/internal/service"
)

// ServerConfig holds the gRPC server configuration.
type ServerConfig struct {
	Port       int
	JWTSecret  string
	Logger     *slog.Logger
	Reflection bool // Enable gRPC reflection for debugging
}

// Server wraps the gRPC server.
type Server struct {
	grpcServer  *grpc.Server
	todoServer  *TodoServer
	healthCheck *health.Server
	config      ServerConfig
	logger      *slog.Logger
}

// NewServer creates a new gRPC server.
func NewServer(cfg ServerConfig, todoService *service.TodoService) *Server {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	// Create interceptor chain
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			RecoveryInterceptor(logger),
			LoggingInterceptor(logger),
			AuthInterceptor(cfg.JWTSecret),
		),
		grpc.ChainStreamInterceptor(
			StreamAuthInterceptor(cfg.JWTSecret),
		),
	}

	grpcServer := grpc.NewServer(opts...)

	// Create Todo server
	todoServer := NewTodoServer(todoService)

	// Register services
	todov1.RegisterTodoServiceServer(grpcServer, todoServer)

	// Health check
	healthCheck := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthCheck)
	healthCheck.SetServingStatus("todo.v1.TodoService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection in development
	if cfg.Reflection {
		reflection.Register(grpcServer)
	}

	return &Server{
		grpcServer:  grpcServer,
		todoServer:  todoServer,
		healthCheck: healthCheck,
		config:      cfg,
		logger:      logger,
	}
}

// Start starts the gRPC server.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.logger.Info("gRPC server starting", slog.String("addr", addr))

	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("gRPC server failed: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server.
func (s *Server) Stop() {
	s.logger.Info("gRPC server stopping")
	s.healthCheck.SetServingStatus("todo.v1.TodoService", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	s.grpcServer.GracefulStop()
}

// Shutdown immediately stops the gRPC server.
func (s *Server) Shutdown(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		s.Stop()
		close(done)
	}()

	select {
	case <-ctx.Done():
		s.grpcServer.Stop()
		return ctx.Err()
	case <-done:
		return nil
	}
}
