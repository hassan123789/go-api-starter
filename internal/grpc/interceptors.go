package grpc

import (
	"context"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthInterceptor creates a unary interceptor for JWT authentication.
func AuthInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract and validate token
		newCtx, err := authenticate(ctx, jwtSecret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		return handler(newCtx, req)
	}
}

// StreamAuthInterceptor creates a stream interceptor for JWT authentication.
func StreamAuthInterceptor(jwtSecret string) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Extract and validate token
		newCtx, err := authenticate(ss.Context(), jwtSecret)
		if err != nil {
			return status.Error(codes.Unauthenticated, err.Error())
		}

		// Wrap the stream with the new context
		wrapped := &wrappedStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		return handler(srv, wrapped)
	}
}

// authenticate validates the JWT token and returns a context with user ID.
func authenticate(ctx context.Context, jwtSecret string) (context.Context, error) {
	tokenString, err := getTokenFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, status.Error(codes.Unauthenticated, "invalid signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid user_id in token")
	}

	return setUserIDInContext(ctx, int64(userIDFloat)), nil
}

// wrappedStream wraps a grpc.ServerStream with a custom context.
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// LoggingInterceptor creates a unary interceptor for logging.
func LoggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		code := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
			}
		}

		logger.Info("gRPC request",
			slog.String("method", info.FullMethod),
			slog.String("code", code.String()),
			slog.Duration("duration", duration),
		)

		return resp, err
	}
}

// RecoveryInterceptor creates a unary interceptor for panic recovery.
func RecoveryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("gRPC panic recovered",
					slog.Any("panic", r),
					slog.String("method", info.FullMethod),
				)
				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}
