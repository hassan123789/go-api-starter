// Package circuitbreaker provides a circuit breaker implementation for resilient service calls.
package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
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

// Common errors
var (
	ErrCircuitOpen    = errors.New("circuit breaker is open")
	ErrTooManyRequest = errors.New("too many requests in half-open state")
)

// Options configures the circuit breaker.
type Options struct {
	// MaxFailures is the maximum number of failures before opening the circuit.
	MaxFailures int

	// Timeout is the duration the circuit stays open before transitioning to half-open.
	Timeout time.Duration

	// MaxHalfOpenRequests is the maximum number of requests allowed in half-open state.
	MaxHalfOpenRequests int

	// OnStateChange is called when the circuit breaker state changes.
	OnStateChange func(from, to State)

	// IsSuccessful determines if an error should be counted as a failure.
	// If nil, all non-nil errors are counted as failures.
	IsSuccessful func(err error) bool
}

// DefaultOptions returns sensible default options.
func DefaultOptions() Options {
	return Options{
		MaxFailures:         5,
		Timeout:             30 * time.Second,
		MaxHalfOpenRequests: 3,
	}
}

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	opts Options

	mu              sync.RWMutex
	state           State
	failures        int
	successes       int
	halfOpenCounter int
	lastFailure     time.Time
}

// New creates a new circuit breaker with the given options.
func New(opts Options) *CircuitBreaker {
	if opts.MaxFailures <= 0 {
		opts.MaxFailures = DefaultOptions().MaxFailures
	}
	if opts.Timeout <= 0 {
		opts.Timeout = DefaultOptions().Timeout
	}
	if opts.MaxHalfOpenRequests <= 0 {
		opts.MaxHalfOpenRequests = DefaultOptions().MaxHalfOpenRequests
	}

	return &CircuitBreaker{
		opts:  opts,
		state: StateClosed,
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Execute runs the function if the circuit allows it.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	if !cb.allowRequest() {
		return ErrCircuitOpen
	}

	err := fn(ctx)
	cb.recordResult(err)

	return err
}

// ExecuteWithFallback runs the function with a fallback if the circuit is open.
func (cb *CircuitBreaker) ExecuteWithFallback(ctx context.Context, fn func(context.Context) error, fallback func(context.Context, error) error) error {
	err := cb.Execute(ctx, fn)
	if errors.Is(err, ErrCircuitOpen) || errors.Is(err, ErrTooManyRequest) {
		return fallback(ctx, err)
	}
	return err
}

// allowRequest checks if a request is allowed based on the current state.
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		// Check if timeout has passed
		if time.Since(cb.lastFailure) >= cb.opts.Timeout {
			cb.transitionTo(StateHalfOpen)
			cb.halfOpenCounter = 1
			return true
		}
		return false

	case StateHalfOpen:
		if cb.halfOpenCounter < cb.opts.MaxHalfOpenRequests {
			cb.halfOpenCounter++
			return true
		}
		return false

	default:
		return false
	}
}

// recordResult records the result of an execution.
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	isSuccess := err == nil
	if cb.opts.IsSuccessful != nil {
		isSuccess = cb.opts.IsSuccessful(err)
	}

	switch cb.state {
	case StateClosed:
		if !isSuccess {
			cb.failures++
			cb.lastFailure = time.Now()
			if cb.failures >= cb.opts.MaxFailures {
				cb.transitionTo(StateOpen)
			}
		} else {
			cb.failures = 0
		}

	case StateHalfOpen:
		if isSuccess {
			cb.successes++
			if cb.successes >= cb.opts.MaxHalfOpenRequests {
				cb.transitionTo(StateClosed)
			}
		} else {
			cb.lastFailure = time.Now()
			cb.transitionTo(StateOpen)
		}
	}
}

// transitionTo changes the circuit breaker state.
func (cb *CircuitBreaker) transitionTo(newState State) {
	if cb.state == newState {
		return
	}

	oldState := cb.state
	cb.state = newState
	cb.failures = 0
	cb.successes = 0
	cb.halfOpenCounter = 0

	if cb.opts.OnStateChange != nil {
		go cb.opts.OnStateChange(oldState, newState)
	}
}

// Reset resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.halfOpenCounter = 0
}

// Metrics returns current circuit breaker metrics.
type Metrics struct {
	State      State
	Failures   int
	Successes  int
	IsAllowing bool
}

// Metrics returns current metrics for monitoring.
func (cb *CircuitBreaker) Metrics() Metrics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return Metrics{
		State:      cb.state,
		Failures:   cb.failures,
		Successes:  cb.successes,
		IsAllowing: cb.state == StateClosed || (cb.state == StateHalfOpen && cb.halfOpenCounter < cb.opts.MaxHalfOpenRequests),
	}
}

// Registry manages multiple circuit breakers.
type Registry struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker
	opts     Options
}

// NewRegistry creates a new circuit breaker registry.
func NewRegistry(defaultOpts Options) *Registry {
	return &Registry{
		breakers: make(map[string]*CircuitBreaker),
		opts:     defaultOpts,
	}
}

// Get returns the circuit breaker for the given name, creating it if necessary.
func (r *Registry) Get(name string) *CircuitBreaker {
	r.mu.RLock()
	cb, ok := r.breakers[name]
	r.mu.RUnlock()

	if ok {
		return cb
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if cb, ok := r.breakers[name]; ok {
		return cb
	}

	cb = New(r.opts)
	r.breakers[name] = cb
	return cb
}

// Execute runs a function with the circuit breaker for the given name.
func (r *Registry) Execute(ctx context.Context, name string, fn func(context.Context) error) error {
	return r.Get(name).Execute(ctx, fn)
}

// AllMetrics returns metrics for all circuit breakers.
func (r *Registry) AllMetrics() map[string]Metrics {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := make(map[string]Metrics, len(r.breakers))
	for name, cb := range r.breakers {
		metrics[name] = cb.Metrics()
	}
	return metrics
}
