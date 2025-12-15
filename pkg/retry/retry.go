// Package retry provides a flexible retry mechanism with configurable backoff strategies.
//
// # Overview
//
// This package implements retry logic for unreliable operations with:
//   - Configurable retry attempts
//   - Multiple backoff strategies (constant, linear, exponential)
//   - Jitter for distributed systems
//   - Context cancellation support
//   - Retry condition customization
//
// # Basic Usage
//
//	result, err := retry.Do(func() (string, error) {
//	    return callUnreliableService()
//	})
//
// # With Options
//
//	result, err := retry.Do(
//	    func() (string, error) {
//	        return callService()
//	    },
//	    retry.WithAttempts(5),
//	    retry.WithDelay(time.Second),
//	    retry.WithBackoff(retry.ExponentialBackoff),
//	)
//
// # Context Support
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	result, err := retry.DoWithContext(ctx, func(ctx context.Context) (string, error) {
//	    return callServiceWithContext(ctx)
//	})
//
// # Custom Retry Conditions
//
//	result, err := retry.Do(
//	    func() (*Response, error) {
//	        return callAPI()
//	    },
//	    retry.WithRetryIf(func(err error) bool {
//	        // Only retry on transient errors
//	        return isTransient(err)
//	    }),
//	)
package retry

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// BackoffStrategy defines the type of backoff to use between retries.
type BackoffStrategy int

const (
	// ConstantBackoff uses a fixed delay between retries.
	ConstantBackoff BackoffStrategy = iota
	// LinearBackoff increases delay linearly (delay * attempt).
	LinearBackoff
	// ExponentialBackoff doubles the delay after each retry.
	ExponentialBackoff
)

// Config holds retry configuration.
type Config struct {
	Attempts     int
	Delay        time.Duration
	MaxDelay     time.Duration
	Backoff      BackoffStrategy
	Jitter       bool
	JitterFactor float64
	RetryIf      func(error) bool
	OnRetry      func(attempt int, err error)
}

// Option is a functional option for configuring retry behavior.
type Option func(*Config)

// DefaultConfig returns the default retry configuration.
func DefaultConfig() *Config {
	return &Config{
		Attempts:     3,
		Delay:        100 * time.Millisecond,
		MaxDelay:     30 * time.Second,
		Backoff:      ExponentialBackoff,
		Jitter:       true,
		JitterFactor: 0.2,
		RetryIf:      func(err error) bool { return err != nil },
	}
}

// WithAttempts sets the maximum number of retry attempts.
func WithAttempts(n int) Option {
	return func(c *Config) {
		if n > 0 {
			c.Attempts = n
		}
	}
}

// WithDelay sets the initial delay between retries.
func WithDelay(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.Delay = d
		}
	}
}

// WithMaxDelay sets the maximum delay between retries.
func WithMaxDelay(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.MaxDelay = d
		}
	}
}

// WithBackoff sets the backoff strategy.
func WithBackoff(strategy BackoffStrategy) Option {
	return func(c *Config) {
		c.Backoff = strategy
	}
}

// WithJitter enables or disables jitter.
func WithJitter(enabled bool) Option {
	return func(c *Config) {
		c.Jitter = enabled
	}
}

// WithJitterFactor sets the jitter factor (0.0 to 1.0).
func WithJitterFactor(factor float64) Option {
	return func(c *Config) {
		if factor >= 0 && factor <= 1 {
			c.JitterFactor = factor
		}
	}
}

// WithRetryIf sets a custom function to determine if an error should be retried.
func WithRetryIf(fn func(error) bool) Option {
	return func(c *Config) {
		if fn != nil {
			c.RetryIf = fn
		}
	}
}

// WithOnRetry sets a callback function called on each retry.
func WithOnRetry(fn func(attempt int, err error)) Option {
	return func(c *Config) {
		c.OnRetry = fn
	}
}

// Do executes the function with retry logic.
func Do[T any](fn func() (T, error), opts ...Option) (T, error) {
	return DoWithContext(context.Background(), func(ctx context.Context) (T, error) {
		return fn()
	}, opts...)
}

// DoWithContext executes the function with retry logic and context support.
func DoWithContext[T any](ctx context.Context, fn func(context.Context) (T, error), opts ...Option) (T, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	var lastErr error
	var zero T

	for attempt := 1; attempt <= cfg.Attempts; attempt++ {
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		default:
		}

		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if we should retry
		if !cfg.RetryIf(err) {
			return zero, err
		}

		// Don't wait after the last attempt
		if attempt == cfg.Attempts {
			break
		}

		// Call onRetry callback if set
		if cfg.OnRetry != nil {
			cfg.OnRetry(attempt, err)
		}

		// Calculate delay
		delay := calculateDelay(cfg, attempt)

		// Wait with context
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return zero, ctx.Err()
		case <-timer.C:
		}
	}

	return zero, lastErr
}

// calculateDelay calculates the delay for the given attempt.
func calculateDelay(cfg *Config, attempt int) time.Duration {
	var delay time.Duration

	switch cfg.Backoff {
	case ConstantBackoff:
		delay = cfg.Delay
	case LinearBackoff:
		delay = cfg.Delay * time.Duration(attempt)
	case ExponentialBackoff:
		delay = cfg.Delay * time.Duration(math.Pow(2, float64(attempt-1)))
	}

	// Apply max delay cap
	if delay > cfg.MaxDelay {
		delay = cfg.MaxDelay
	}

	// Apply jitter
	if cfg.Jitter && cfg.JitterFactor > 0 {
		jitter := float64(delay) * cfg.JitterFactor * (rand.Float64()*2 - 1)
		delay = time.Duration(float64(delay) + jitter)
		if delay < 0 {
			delay = 0
		}
	}

	return delay
}

// Retry is a convenience wrapper that retries without returning a value.
func Retry(fn func() error, opts ...Option) error {
	_, err := Do(func() (struct{}, error) {
		return struct{}{}, fn()
	}, opts...)
	return err
}

// RetryWithContext is a convenience wrapper that retries without returning a value.
func RetryWithContext(ctx context.Context, fn func(context.Context) error, opts ...Option) error {
	_, err := DoWithContext(ctx, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, fn(ctx)
	}, opts...)
	return err
}
