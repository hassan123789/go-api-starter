// Package ctxutil provides context utilities for safe and idiomatic context handling.
package ctxutil

import (
	"context"
	"time"
)

// DefaultTimeout is the default timeout for database operations.
const DefaultTimeout = 5 * time.Second

// WithTimeout returns a context with the specified timeout.
// If timeout is 0, DefaultTimeout is used.
// Always remember to call the cancel function when done.
func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	return context.WithTimeout(parent, timeout)
}

// WithDeadline returns a context that will be cancelled at the specified deadline.
func WithDeadline(parent context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(parent, deadline)
}

// ContextKey is a type-safe key for context values.
type ContextKey string

// Common context keys
const (
	KeyUserID    ContextKey = "user_id"
	KeyRequestID ContextKey = "request_id"
	KeyLogger    ContextKey = "logger"
	KeyTraceID   ContextKey = "trace_id"
	KeySpanID    ContextKey = "span_id"
)

// WithValue sets a typed value in the context.
func WithValue[T any](ctx context.Context, key ContextKey, value T) context.Context {
	return context.WithValue(ctx, key, value)
}

// Value retrieves a typed value from the context.
func Value[T any](ctx context.Context, key ContextKey) (T, bool) {
	val, ok := ctx.Value(key).(T)
	return val, ok
}

// MustValue retrieves a typed value from the context, panicking if not found.
func MustValue[T any](ctx context.Context, key ContextKey) T {
	val, ok := Value[T](ctx, key)
	if !ok {
		var zero T
		return zero
	}
	return val
}

// IsDone returns true if the context is done (cancelled or timed out).
func IsDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// Sleep sleeps for the specified duration or until the context is cancelled.
// Returns true if the sleep completed, false if cancelled.
func Sleep(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// After returns a channel that sends the current time after the specified duration,
// respecting context cancellation.
func After(ctx context.Context, duration time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	go func() {
		timer := time.NewTimer(duration)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			close(ch)
		case t := <-timer.C:
			ch <- t
		}
	}()
	return ch
}

// Merge creates a new context that is cancelled when either parent context is cancelled.
func Merge(ctx1, ctx2 context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx1)
	go func() {
		select {
		case <-ctx2.Done():
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}
