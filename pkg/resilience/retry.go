// Package resilience provides retry mechanisms with various backoff strategies.
package resilience

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

// RetryableError wraps an error that should trigger a retry.
type RetryableError struct {
	Err       error
	Retryable bool
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// IsRetryable checks if an error should be retried.
func IsRetryable(err error) bool {
	var retryable *RetryableError
	if errors.As(err, &retryable) {
		return retryable.Retryable
	}
	// Default: retry on all errors except context cancellation
	return !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
}

// BackoffStrategy defines how to calculate delay between retries.
type BackoffStrategy interface {
	// Delay returns the delay before the nth retry (0-indexed).
	Delay(attempt int) time.Duration
}

// ConstantBackoff returns a constant delay.
type ConstantBackoff struct {
	Delay_ time.Duration
}

func (b *ConstantBackoff) Delay(_ int) time.Duration {
	return b.Delay_
}

// ExponentialBackoff returns exponentially increasing delays.
type ExponentialBackoff struct {
	Initial    time.Duration
	Max        time.Duration
	Multiplier float64
	Jitter     bool
}

func (b *ExponentialBackoff) Delay(attempt int) time.Duration {
	if attempt == 0 {
		return b.Initial
	}

	delay := float64(b.Initial) * math.Pow(b.Multiplier, float64(attempt))
	if delay > float64(b.Max) {
		delay = float64(b.Max)
	}

	if b.Jitter {
		// Add up to 20% jitter
		jitter := delay * 0.2 * rand.Float64()
		delay += jitter
	}

	return time.Duration(delay)
}

// LinearBackoff returns linearly increasing delays.
type LinearBackoff struct {
	Initial   time.Duration
	Increment time.Duration
	Max       time.Duration
}

func (b *LinearBackoff) Delay(attempt int) time.Duration {
	delay := b.Initial + time.Duration(attempt)*b.Increment
	if delay > b.Max {
		return b.Max
	}
	return delay
}

// Retryer provides retry functionality.
type Retryer struct {
	maxAttempts int
	backoff     BackoffStrategy
	onRetry     func(attempt int, err error, delay time.Duration)
}

// RetryerOption configures a retryer.
type RetryerOption func(*Retryer)

// WithMaxAttempts sets the maximum retry attempts.
func WithMaxAttempts(n int) RetryerOption {
	return func(r *Retryer) {
		r.maxAttempts = n
	}
}

// WithBackoff sets the backoff strategy.
func WithBackoff(b BackoffStrategy) RetryerOption {
	return func(r *Retryer) {
		r.backoff = b
	}
}

// WithOnRetry sets a callback for each retry.
func WithOnRetry(fn func(attempt int, err error, delay time.Duration)) RetryerOption {
	return func(r *Retryer) {
		r.onRetry = fn
	}
}

// NewRetryer creates a new retryer.
func NewRetryer(opts ...RetryerOption) *Retryer {
	r := &Retryer{
		maxAttempts: 3,
		backoff: &ExponentialBackoff{
			Initial:    100 * time.Millisecond,
			Max:        10 * time.Second,
			Multiplier: 2.0,
			Jitter:     true,
		},
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Do executes the function with retry logic.
func (r *Retryer) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	var lastErr error

	for attempt := 0; attempt < r.maxAttempts; attempt++ {
		err := fn(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry
		if !IsRetryable(err) {
			return err
		}

		// Check if this was the last attempt
		if attempt == r.maxAttempts-1 {
			break
		}

		// Calculate delay
		delay := r.backoff.Delay(attempt)

		// Call retry callback
		if r.onRetry != nil {
			r.onRetry(attempt+1, err, delay)
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return lastErr
}

// DefaultRetryer returns a retryer with sensible defaults.
func DefaultRetryer() *Retryer {
	return NewRetryer(
		WithMaxAttempts(3),
		WithBackoff(&ExponentialBackoff{
			Initial:    100 * time.Millisecond,
			Max:        5 * time.Second,
			Multiplier: 2.0,
			Jitter:     true,
		}),
	)
}

// DoWithValue executes the function with retry logic and returns a value.
// This is a standalone generic function since Go doesn't allow type parameters on methods.
func DoWithValue[T any](ctx context.Context, r *Retryer, fn func(ctx context.Context) (T, error)) (T, error) {
	var result T
	var lastErr error

	for attempt := 0; attempt < r.maxAttempts; attempt++ {
		value, err := fn(ctx)
		if err == nil {
			return value, nil
		}

		lastErr = err

		// Check if we should retry
		if !IsRetryable(err) {
			return result, err
		}

		// Check if this was the last attempt
		if attempt == r.maxAttempts-1 {
			break
		}

		// Calculate delay
		delay := r.backoff.Delay(attempt)

		// Call retry callback
		if r.onRetry != nil {
			r.onRetry(attempt+1, err, delay)
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(delay):
		}
	}

	return result, lastErr
}
