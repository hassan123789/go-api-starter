// Package resilience provides patterns for graceful degradation and fault tolerance.
package resilience

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is open.
	ErrCircuitOpen = errors.New("circuit breaker is open")
	// ErrTimeout is returned when an operation times out.
	ErrTimeout = errors.New("operation timed out")
	// ErrRejected is returned when rate limiting rejects a request.
	ErrRejected = errors.New("request rejected by rate limiter")
)

// State represents the state of a circuit breaker.
type State int

const (
	StateClosed   State = iota // Normal operation
	StateOpen                  // Failing, reject requests
	StateHalfOpen              // Testing recovery
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	name          string
	maxFailures   int32
	failureCount  int32
	successCount  int32
	timeout       time.Duration
	halfOpenMax   int32
	state         State
	lastFailure   time.Time
	mu            sync.RWMutex
	onStateChange func(from, to State)
	fallback      func(ctx context.Context, err error) (interface{}, error)
}

// CircuitBreakerOption configures a circuit breaker.
type CircuitBreakerOption func(*CircuitBreaker)

// WithMaxFailures sets the maximum failures before opening.
func WithMaxFailures(n int32) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.maxFailures = n
	}
}

// WithTimeout sets the timeout before trying to recover.
func WithTimeout(d time.Duration) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.timeout = d
	}
}

// WithHalfOpenMax sets the max requests in half-open state.
func WithHalfOpenMax(n int32) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.halfOpenMax = n
	}
}

// WithStateChangeCallback sets a callback for state changes.
func WithStateChangeCallback(fn func(from, to State)) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.onStateChange = fn
	}
}

// WithFallback sets a fallback function when the circuit is open.
func WithFallback(fn func(ctx context.Context, err error) (interface{}, error)) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.fallback = fn
	}
}

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(name string, opts ...CircuitBreakerOption) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:        name,
		maxFailures: 5,
		timeout:     30 * time.Second,
		halfOpenMax: 3,
		state:       StateClosed,
	}

	for _, opt := range opts {
		opt(cb)
	}

	return cb
}

// Execute runs the given function with circuit breaker protection.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	if !cb.allowRequest() {
		if cb.fallback != nil {
			return cb.fallback(ctx, ErrCircuitOpen)
		}
		return nil, ErrCircuitOpen
	}

	result, err := fn(ctx)

	if err != nil {
		cb.recordFailure()
		if cb.fallback != nil && cb.State() == StateOpen {
			return cb.fallback(ctx, err)
		}
		return result, err
	}

	cb.recordSuccess()
	return result, nil
}

// allowRequest checks if a request should be allowed.
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.RLock()
	state := cb.state
	lastFailure := cb.lastFailure
	cb.mu.RUnlock()

	switch state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if timeout has passed
		if time.Since(lastFailure) > cb.timeout {
			cb.transition(StateHalfOpen)
			return true
		}
		return false
	case StateHalfOpen:
		// Allow limited requests
		return atomic.LoadInt32(&cb.successCount) < cb.halfOpenMax
	default:
		return true
	}
}

// recordSuccess records a successful request.
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		atomic.StoreInt32(&cb.failureCount, 0)
	case StateHalfOpen:
		count := atomic.AddInt32(&cb.successCount, 1)
		if count >= cb.halfOpenMax {
			cb.transitionLocked(StateClosed)
		}
	case StateOpen:
		// No action needed in open state
	}
}

// recordFailure records a failed request.
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		count := atomic.AddInt32(&cb.failureCount, 1)
		if count >= cb.maxFailures {
			cb.transitionLocked(StateOpen)
		}
	case StateHalfOpen:
		cb.transitionLocked(StateOpen)
	case StateOpen:
		// Already open, no action needed
	}
}

// transition changes the circuit breaker state.
func (cb *CircuitBreaker) transition(to State) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.transitionLocked(to)
}

func (cb *CircuitBreaker) transitionLocked(to State) {
	from := cb.state
	if from == to {
		return
	}

	cb.state = to
	atomic.StoreInt32(&cb.failureCount, 0)
	atomic.StoreInt32(&cb.successCount, 0)

	if cb.onStateChange != nil {
		go cb.onStateChange(from, to)
	}
}

// State returns the current state.
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Name returns the circuit breaker name.
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// Reset resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.transition(StateClosed)
}

// Stats returns circuit breaker statistics.
func (cb *CircuitBreaker) Stats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"name":         cb.name,
		"state":        cb.state.String(),
		"failures":     atomic.LoadInt32(&cb.failureCount),
		"successes":    atomic.LoadInt32(&cb.successCount),
		"max_failures": cb.maxFailures,
		"timeout":      cb.timeout.String(),
		"last_failure": cb.lastFailure,
	}
}
