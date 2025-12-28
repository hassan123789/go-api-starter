package ctxutil

import (
	"context"
	"testing"
	"time"
)

func TestWithTimeout(t *testing.T) {
	t.Run("default timeout", func(t *testing.T) {
		ctx, cancel := WithTimeout(context.Background(), 0)
		defer cancel()

		deadline, ok := ctx.Deadline()
		if !ok {
			t.Error("expected deadline to be set")
		}

		// Should be approximately DefaultTimeout from now
		expected := time.Now().Add(DefaultTimeout)
		if deadline.Before(expected.Add(-100*time.Millisecond)) || deadline.After(expected.Add(100*time.Millisecond)) {
			t.Errorf("deadline not within expected range")
		}
	})

	t.Run("custom timeout", func(t *testing.T) {
		ctx, cancel := WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		deadline, ok := ctx.Deadline()
		if !ok {
			t.Error("expected deadline to be set")
		}

		expected := time.Now().Add(2 * time.Second)
		if deadline.Before(expected.Add(-100*time.Millisecond)) || deadline.After(expected.Add(100*time.Millisecond)) {
			t.Errorf("deadline not within expected range")
		}
	})
}

func TestContextValue(t *testing.T) {
	t.Run("WithValue and Value", func(t *testing.T) {
		ctx := WithValue(context.Background(), KeyUserID, int64(123))

		val, ok := Value[int64](ctx, KeyUserID)
		if !ok {
			t.Error("expected value to be found")
		}
		if val != 123 {
			t.Errorf("expected 123, got %d", val)
		}
	})

	t.Run("Value not found", func(t *testing.T) {
		ctx := context.Background()

		val, ok := Value[int64](ctx, KeyUserID)
		if ok {
			t.Error("expected value to not be found")
		}
		if val != 0 {
			t.Errorf("expected zero value, got %d", val)
		}
	})

	t.Run("MustValue", func(t *testing.T) {
		ctx := WithValue(context.Background(), KeyRequestID, "req-123")

		val := MustValue[string](ctx, KeyRequestID)
		if val != "req-123" {
			t.Errorf("expected req-123, got %s", val)
		}
	})
}

func TestIsDone(t *testing.T) {
	t.Run("not done", func(t *testing.T) {
		ctx := context.Background()
		if IsDone(ctx) {
			t.Error("expected context to not be done")
		}
	})

	t.Run("canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if !IsDone(ctx) {
			t.Error("expected context to be done")
		}
	})
}

func TestSleep(t *testing.T) {
	t.Run("sleep completes", func(t *testing.T) {
		ctx := context.Background()
		start := time.Now()
		completed := Sleep(ctx, 50*time.Millisecond)
		duration := time.Since(start)

		if !completed {
			t.Error("expected sleep to complete")
		}
		if duration < 50*time.Millisecond {
			t.Errorf("expected duration >= 50ms, got %v", duration)
		}
	})

	t.Run("sleep canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(25 * time.Millisecond)
			cancel()
		}()

		start := time.Now()
		completed := Sleep(ctx, 1*time.Second)
		duration := time.Since(start)

		if completed {
			t.Error("expected sleep to be canceled")
		}
		if duration > 100*time.Millisecond {
			t.Errorf("expected quick cancellation, got %v", duration)
		}
	})
}

func TestMerge(t *testing.T) {
	t.Run("cancel first context", func(t *testing.T) {
		ctx1, cancel1 := context.WithCancel(context.Background())
		ctx2 := context.Background()

		merged, cancelMerged := Merge(ctx1, ctx2)
		defer cancelMerged()

		cancel1()

		select {
		case <-merged.Done():
			// Expected
		case <-time.After(100 * time.Millisecond):
			t.Error("merged context should be canceled")
		}
	})

	t.Run("cancel second context", func(t *testing.T) {
		ctx1 := context.Background()
		ctx2, cancel2 := context.WithCancel(context.Background())

		merged, cancelMerged := Merge(ctx1, ctx2)
		defer cancelMerged()

		cancel2()

		select {
		case <-merged.Done():
			// Expected
		case <-time.After(100 * time.Millisecond):
			t.Error("merged context should be canceled")
		}
	})
}
