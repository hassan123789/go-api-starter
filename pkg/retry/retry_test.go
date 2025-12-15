package retry_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/zareh/go-api-starter/pkg/retry"
)

func TestDo(t *testing.T) {
	t.Parallel()

	t.Run("succeeds on first attempt", func(t *testing.T) {
		t.Parallel()

		attempts := 0
		result, err := retry.Do(func() (int, error) {
			attempts++
			return 42, nil
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != 42 {
			t.Errorf("expected 42, got %d", result)
		}
		if attempts != 1 {
			t.Errorf("expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("succeeds on third attempt", func(t *testing.T) {
		t.Parallel()

		attempts := 0
		result, err := retry.Do(func() (int, error) {
			attempts++
			if attempts < 3 {
				return 0, errors.New("temporary error")
			}
			return 42, nil
		}, retry.WithAttempts(3), retry.WithDelay(time.Millisecond))

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != 42 {
			t.Errorf("expected 42, got %d", result)
		}
		if attempts != 3 {
			t.Errorf("expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("fails after max attempts", func(t *testing.T) {
		t.Parallel()

		attempts := 0
		_, err := retry.Do(func() (int, error) {
			attempts++
			return 0, errors.New("persistent error")
		}, retry.WithAttempts(3), retry.WithDelay(time.Millisecond))

		if err == nil {
			t.Error("expected error, got nil")
		}
		if attempts != 3 {
			t.Errorf("expected 3 attempts, got %d", attempts)
		}
	})
}

func TestDoWithContext(t *testing.T) {
	t.Parallel()

	t.Run("respects context cancellation", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		attempts := 0

		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		_, err := retry.DoWithContext(ctx, func(ctx context.Context) (int, error) {
			attempts++
			return 0, errors.New("error")
		}, retry.WithAttempts(100), retry.WithDelay(30*time.Millisecond))

		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})

	t.Run("respects context timeout", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_, err := retry.DoWithContext(ctx, func(ctx context.Context) (int, error) {
			return 0, errors.New("error")
		}, retry.WithAttempts(100), retry.WithDelay(30*time.Millisecond))

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("expected context.DeadlineExceeded, got %v", err)
		}
	})
}

func TestBackoffStrategies(t *testing.T) {
	t.Parallel()

	t.Run("constant backoff", func(t *testing.T) {
		t.Parallel()

		delays := []time.Duration{}
		_, _ = retry.Do(func() (int, error) {
			return 0, errors.New("error")
		},
			retry.WithAttempts(4),
			retry.WithDelay(10*time.Millisecond),
			retry.WithBackoff(retry.ConstantBackoff),
			retry.WithJitter(false),
			retry.WithOnRetry(func(attempt int, err error) {
				delays = append(delays, 10*time.Millisecond)
			}),
		)

		// Should have 3 delays (between 4 attempts)
		if len(delays) != 3 {
			t.Errorf("expected 3 delays, got %d", len(delays))
		}
	})
}

func TestRetryIf(t *testing.T) {
	t.Parallel()

	t.Run("only retries matching errors", func(t *testing.T) {
		t.Parallel()

		transientErr := errors.New("transient")
		permanentErr := errors.New("permanent")

		attempts := 0
		_, err := retry.Do(func() (int, error) {
			attempts++
			if attempts == 1 {
				return 0, transientErr
			}
			return 0, permanentErr
		},
			retry.WithAttempts(5),
			retry.WithDelay(time.Millisecond),
			retry.WithRetryIf(func(err error) bool {
				return errors.Is(err, transientErr)
			}),
		)

		if !errors.Is(err, permanentErr) {
			t.Errorf("expected permanentErr, got %v", err)
		}
		if attempts != 2 {
			t.Errorf("expected 2 attempts, got %d", attempts)
		}
	})
}

func TestOnRetry(t *testing.T) {
	t.Parallel()

	t.Run("calls onRetry callback", func(t *testing.T) {
		t.Parallel()

		retries := []int{}
		_, _ = retry.Do(func() (int, error) {
			return 0, errors.New("error")
		},
			retry.WithAttempts(4),
			retry.WithDelay(time.Millisecond),
			retry.WithOnRetry(func(attempt int, err error) {
				retries = append(retries, attempt)
			}),
		)

		expected := []int{1, 2, 3}
		if len(retries) != len(expected) {
			t.Errorf("expected %d retries, got %d", len(expected), len(retries))
		}
		for i, v := range expected {
			if retries[i] != v {
				t.Errorf("retry %d: expected attempt %d, got %d", i, v, retries[i])
			}
		}
	})
}

func TestRetry(t *testing.T) {
	t.Parallel()

	t.Run("succeeds without return value", func(t *testing.T) {
		t.Parallel()

		attempts := 0
		err := retry.Retry(func() error {
			attempts++
			if attempts < 2 {
				return errors.New("error")
			}
			return nil
		}, retry.WithAttempts(3), retry.WithDelay(time.Millisecond))

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if attempts != 2 {
			t.Errorf("expected 2 attempts, got %d", attempts)
		}
	})
}

// Examples
func ExampleDo() {
	result, err := retry.Do(func() (string, error) {
		return "success", nil
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result)
	// Output: success
}

func ExampleDo_withOptions() {
	attempts := 0
	result, _ := retry.Do(func() (int, error) {
		attempts++
		if attempts < 3 {
			return 0, errors.New("not yet")
		}
		return 42, nil
	},
		retry.WithAttempts(5),
		retry.WithDelay(time.Millisecond),
	)
	fmt.Println(result)
	// Output: 42
}

// Benchmarks
func BenchmarkDoSuccess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = retry.Do(func() (int, error) {
			return 42, nil
		})
	}
}

func BenchmarkDoWithRetries(b *testing.B) {
	for i := 0; i < b.N; i++ {
		attempts := 0
		_, _ = retry.Do(func() (int, error) {
			attempts++
			if attempts < 3 {
				return 0, errors.New("error")
			}
			return 42, nil
		}, retry.WithDelay(time.Microsecond))
	}
}
