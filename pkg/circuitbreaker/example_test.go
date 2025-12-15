package circuitbreaker_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zareh/go-api-starter/pkg/circuitbreaker"
)

// ExampleNew demonstrates basic circuit breaker usage.
func ExampleNew() {
	// Create a circuit breaker with default options
	cb := circuitbreaker.New(circuitbreaker.DefaultOptions())

	// Execute operations through the circuit breaker
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		// Simulated service call
		return nil
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Success\n")
	fmt.Printf("State: %s\n", cb.State())

	// Output:
	// Success
	// State: closed
}

// ExampleCircuitBreaker_Execute demonstrates error handling.
func ExampleCircuitBreaker_Execute() {
	opts := circuitbreaker.DefaultOptions()
	opts.MaxFailures = 3
	opts.Timeout = 1 * time.Second
	cb := circuitbreaker.New(opts)

	// Simulate a failing operation
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("service unavailable")
	})

	if err != nil {
		fmt.Printf("First error: service unavailable\n")
	}

	// Check the state
	fmt.Printf("State after failure: %s\n", cb.State())

	// Output:
	// First error: service unavailable
	// State after failure: closed
}

// ExampleCircuitBreaker_ExecuteWithFallback demonstrates using fallback functions.
func ExampleCircuitBreaker_ExecuteWithFallback() {
	opts := circuitbreaker.DefaultOptions()
	opts.MaxFailures = 1
	cb := circuitbreaker.New(opts)

	// First, trigger the circuit breaker to open
	_ = cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("service unavailable")
	})

	// Now use fallback when circuit is open
	err := cb.ExecuteWithFallback(
		context.Background(),
		func(ctx context.Context) error {
			return errors.New("service unavailable")
		},
		func(ctx context.Context, e error) error {
			fmt.Println("Using fallback")
			return nil
		},
	)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Fallback succeeded")

	// Output:
	// Using fallback
	// Fallback succeeded
}

// ExampleDefaultOptions demonstrates custom configuration.
func ExampleDefaultOptions() {
	opts := circuitbreaker.DefaultOptions()
	opts.MaxFailures = 5            // Open after 5 failures
	opts.Timeout = 30 * time.Second // Try half-open after 30s
	opts.MaxHalfOpenRequests = 3    // Allow 3 requests in half-open

	cb := circuitbreaker.New(opts)

	fmt.Printf("Initial state: %s\n", cb.State())

	// Output:
	// Initial state: closed
}
