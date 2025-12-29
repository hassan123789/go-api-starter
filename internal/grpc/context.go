package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/metadata"
)

// ctxKey is a type for context keys.
type ctxKey string

const (
	// userIDKey is the context key for user ID.
	userIDKey ctxKey = "user_id"
)

// ErrNoUserID is returned when user ID is not found in context.
var ErrNoUserID = errors.New("user ID not found in context")

// getUserIDFromContext extracts user ID from the gRPC context.
// The user ID is expected to be set by the auth interceptor.
func getUserIDFromContext(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value(userIDKey).(int64)
	if !ok {
		return 0, ErrNoUserID
	}
	return userID, nil
}

// setUserIDInContext sets the user ID in the context.
func setUserIDInContext(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// getTokenFromMetadata extracts the bearer token from gRPC metadata.
func getTokenFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", errors.New("missing authorization header")
	}

	auth := values[0]
	const prefix = "Bearer "
	if len(auth) < len(prefix) || auth[:len(prefix)] != prefix {
		return "", errors.New("invalid authorization format")
	}

	return auth[len(prefix):], nil
}
