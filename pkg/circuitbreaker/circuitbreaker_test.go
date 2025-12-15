package circuitbreaker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestCircuitBreaker_StateClosed(t *testing.T) {
	t.Run("stays closed on success", func(t *testing.T) {
		cb := New(Options{MaxFailures: 3, Timeout: time.Second})

		for i := 0; i < 10; i++ {
			err := cb.Execute(context.Background(), func(_ context.Context) error {
				return nil
			})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}

		if cb.State() != StateClosed {
			t.Errorf("expected StateClosed, got %s", cb.State())
		}
	})

	t.Run("opens after max failures", func(t *testing.T) {
		cb := New(Options{MaxFailures: 3, Timeout: time.Second})
		testErr := errors.New("test error")

		for i := 0; i < 3; i++ {
			_ = cb.Execute(context.Background(), func(_ context.Context) error {
				return testErr
			})
		}

		if cb.State() != StateOpen {
			t.Errorf("expected StateOpen, got %s", cb.State())
		}
	})

	t.Run("resets failure count on success", func(t *testing.T) {
		cb := New(Options{MaxFailures: 3, Timeout: time.Second})
		testErr := errors.New("test error")

		// Two failures
		for i := 0; i < 2; i++ {
			_ = cb.Execute(context.Background(), func(_ context.Context) error {
				return testErr
			})
		}

		// One success resets the counter
		_ = cb.Execute(context.Background(), func(_ context.Context) error {
			return nil
		})

		// Two more failures should not open
		for i := 0; i < 2; i++ {
			_ = cb.Execute(context.Background(), func(_ context.Context) error {
				return testErr
			})
		}

		if cb.State() != StateClosed {
			t.Errorf("expected StateClosed, got %s", cb.State())
		}
	})
}

func TestCircuitBreaker_StateOpen(t *testing.T) {
	t.Run("rejects requests when open", func(t *testing.T) {
		cb := New(Options{MaxFailures: 1, Timeout: time.Hour})
		testErr := errors.New("test error")

		// Open the circuit
		_ = cb.Execute(context.Background(), func(_ context.Context) error {
			return testErr
		})

		// Subsequent requests should fail
		err := cb.Execute(context.Background(), func(_ context.Context) error {
			return nil
		})

		if !errors.Is(err, ErrCircuitOpen) {
			t.Errorf("expected ErrCircuitOpen, got %v", err)
		}
	})

	t.Run("transitions to half-open after timeout", func(t *testing.T) {
		cb := New(Options{MaxFailures: 1, Timeout: 50 * time.Millisecond, MaxHalfOpenRequests: 1})
		testErr := errors.New("test error")

		// Open the circuit
		_ = cb.Execute(context.Background(), func(_ context.Context) error {
			return testErr
		})

		// Wait for timeout
		time.Sleep(100 * time.Millisecond)

		// Should allow request and close after success (MaxHalfOpenRequests: 1)
		var executed bool
		err := cb.Execute(context.Background(), func(_ context.Context) error {
			executed = true
			return nil
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !executed {
			t.Error("expected function to be executed")
		}
		if cb.State() != StateClosed {
			t.Errorf("expected StateClosed after successful half-open request, got %s", cb.State())
		}
	})
}

func TestCircuitBreaker_StateHalfOpen(t *testing.T) {
	t.Run("reopens on failure", func(t *testing.T) {
		cb := New(Options{MaxFailures: 1, Timeout: 50 * time.Millisecond, MaxHalfOpenRequests: 3})
		testErr := errors.New("test error")

		// Open the circuit
		_ = cb.Execute(context.Background(), func(_ context.Context) error {
			return testErr
		})

		// Wait for timeout
		time.Sleep(100 * time.Millisecond)

		// Fail in half-open state
		_ = cb.Execute(context.Background(), func(_ context.Context) error {
			return testErr
		})

		if cb.State() != StateOpen {
			t.Errorf("expected StateOpen after half-open failure, got %s", cb.State())
		}
	})

	t.Run("closes after successful requests", func(t *testing.T) {
		cb := New(Options{MaxFailures: 1, Timeout: 50 * time.Millisecond, MaxHalfOpenRequests: 2})
		testErr := errors.New("test error")

		// Open the circuit
		_ = cb.Execute(context.Background(), func(_ context.Context) error {
			return testErr
		})

		// Wait for timeout
		time.Sleep(100 * time.Millisecond)

		// Two successful requests should close the circuit
		for i := 0; i < 2; i++ {
			_ = cb.Execute(context.Background(), func(_ context.Context) error {
				return nil
			})
		}

		if cb.State() != StateClosed {
			t.Errorf("expected StateClosed, got %s", cb.State())
		}
	})
}

func TestCircuitBreaker_ExecuteWithFallback(t *testing.T) {
	cb := New(Options{MaxFailures: 1, Timeout: time.Hour})
	testErr := errors.New("test error")

	// Open the circuit
	_ = cb.Execute(context.Background(), func(_ context.Context) error {
		return testErr
	})

	// Execute with fallback
	var fallbackCalled bool
	err := cb.ExecuteWithFallback(context.Background(),
		func(_ context.Context) error {
			return nil
		},
		func(_ context.Context, err error) error {
			fallbackCalled = true
			return nil
		},
	)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !fallbackCalled {
		t.Error("expected fallback to be called")
	}
}

func TestCircuitBreaker_OnStateChange(t *testing.T) {
	var changes []struct {
		from, to State
	}

	cb := New(Options{
		MaxFailures:         1,
		Timeout:             50 * time.Millisecond,
		MaxHalfOpenRequests: 1,
		OnStateChange: func(from, to State) {
			changes = append(changes, struct{ from, to State }{from, to})
		},
	})

	testErr := errors.New("test error")

	// Open the circuit
	_ = cb.Execute(context.Background(), func(_ context.Context) error {
		return testErr
	})

	// Wait for callback
	time.Sleep(10 * time.Millisecond)

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	// Transition to half-open and then closed
	_ = cb.Execute(context.Background(), func(_ context.Context) error {
		return nil
	})

	// Wait for callbacks
	time.Sleep(10 * time.Millisecond)

	if len(changes) < 2 {
		t.Errorf("expected at least 2 state changes, got %d", len(changes))
	}
}

func TestRegistry(t *testing.T) {
	t.Run("creates circuit breakers on demand", func(t *testing.T) {
		reg := NewRegistry(DefaultOptions())

		cb1 := reg.Get("service-a")
		cb2 := reg.Get("service-a")
		cb3 := reg.Get("service-b")

		if cb1 != cb2 {
			t.Error("expected same circuit breaker for same name")
		}
		if cb1 == cb3 {
			t.Error("expected different circuit breakers for different names")
		}
	})

	t.Run("execute with registry", func(t *testing.T) {
		reg := NewRegistry(Options{MaxFailures: 1, Timeout: time.Hour})
		testErr := errors.New("test error")

		// Open the circuit for service-a
		_ = reg.Execute(context.Background(), "service-a", func(_ context.Context) error {
			return testErr
		})

		// service-a should be open
		err := reg.Execute(context.Background(), "service-a", func(_ context.Context) error {
			return nil
		})
		if !errors.Is(err, ErrCircuitOpen) {
			t.Errorf("expected ErrCircuitOpen for service-a, got %v", err)
		}

		// service-b should be closed
		err = reg.Execute(context.Background(), "service-b", func(_ context.Context) error {
			return nil
		})
		if err != nil {
			t.Errorf("unexpected error for service-b: %v", err)
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		reg := NewRegistry(DefaultOptions())
		var ops atomic.Int64

		done := make(chan struct{})
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					cb := reg.Get("concurrent-test")
					_ = cb.Execute(context.Background(), func(_ context.Context) error {
						ops.Add(1)
						return nil
					})
				}
				done <- struct{}{}
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}

		if ops.Load() != 1000 {
			t.Errorf("expected 1000 operations, got %d", ops.Load())
		}
	})
}

func TestMetrics(t *testing.T) {
	cb := New(Options{MaxFailures: 5, Timeout: time.Hour})

	// Initial metrics
	metrics := cb.Metrics()
	if metrics.State != StateClosed {
		t.Errorf("expected StateClosed, got %s", metrics.State)
	}
	if !metrics.IsAllowing {
		t.Error("expected IsAllowing to be true")
	}

	// After some failures
	for i := 0; i < 3; i++ {
		_ = cb.Execute(context.Background(), func(_ context.Context) error {
			return errors.New("error")
		})
	}

	metrics = cb.Metrics()
	if metrics.Failures != 3 {
		t.Errorf("expected 3 failures, got %d", metrics.Failures)
	}
}
