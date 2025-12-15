package workerpool

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	t.Run("basic processing", func(t *testing.T) {
		processor := func(_ context.Context, n int) (int, error) {
			return n * 2, nil
		}

		pool := New(3, processor)
		ctx := context.Background()
		pool.Start(ctx)

		// Submit tasks
		go func() {
			for i := 1; i <= 5; i++ {
				pool.Submit(ctx, i)
			}
			pool.Stop()
		}()

		// Collect results
		sum := 0
		for task := range pool.Results() {
			if task.Err != nil {
				t.Errorf("unexpected error: %v", task.Err)
			}
			sum += task.Result
		}

		// 2 + 4 + 6 + 8 + 10 = 30
		if sum != 30 {
			t.Errorf("expected sum 30, got %d", sum)
		}
	})

	t.Run("error handling", func(t *testing.T) {
		expectedErr := errors.New("processing error")
		processor := func(_ context.Context, n int) (int, error) {
			if n == 3 {
				return 0, expectedErr
			}
			return n * 2, nil
		}

		pool := New(2, processor)
		ctx := context.Background()
		pool.Start(ctx)

		go func() {
			for i := 1; i <= 5; i++ {
				pool.Submit(ctx, i)
			}
			pool.Stop()
		}()

		errCount := 0
		for task := range pool.Results() {
			if task.Err != nil {
				errCount++
			}
		}

		if errCount != 1 {
			t.Errorf("expected 1 error, got %d", errCount)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		var processedCount atomic.Int32
		processor := func(ctx context.Context, n int) (int, error) {
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case <-time.After(100 * time.Millisecond):
				processedCount.Add(1)
				return n * 2, nil
			}
		}

		pool := New(2, processor)
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		pool.Start(ctx)

		go func() {
			for i := 1; i <= 100; i++ {
				if !pool.Submit(ctx, i) {
					break
				}
			}
			pool.Stop()
		}()

		for range pool.Results() {
		}

		// Should process very few items due to timeout
		if processedCount.Load() > 10 {
			t.Errorf("expected few processed items due to timeout, got %d", processedCount.Load())
		}
	})
}

func TestProcess(t *testing.T) {
	t.Run("batch processing", func(t *testing.T) {
		inputs := []int{1, 2, 3, 4, 5}
		processor := func(_ context.Context, n int) (int, error) {
			return n * 2, nil
		}

		results, errs := Process(context.Background(), 3, inputs, processor)

		if len(errs) != 0 {
			t.Errorf("expected no errors, got %v", errs)
		}

		sum := 0
		for _, r := range results {
			sum += r
		}

		if sum != 30 {
			t.Errorf("expected sum 30, got %d", sum)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		var inputs []int
		processor := func(_ context.Context, n int) (int, error) {
			return n * 2, nil
		}

		results, errs := Process(context.Background(), 3, inputs, processor)

		if results != nil {
			t.Errorf("expected nil results, got %v", results)
		}
		if errs != nil {
			t.Errorf("expected nil errors, got %v", errs)
		}
	})
}

func TestMapConcurrent(t *testing.T) {
	t.Run("preserves order", func(t *testing.T) {
		inputs := []int{1, 2, 3, 4, 5}
		fn := func(_ context.Context, n int) (int, error) {
			return n * 2, nil
		}

		results, err := MapConcurrent(context.Background(), 3, inputs, fn)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expected := []int{2, 4, 6, 8, 10}
		for i, r := range results {
			if r != expected[i] {
				t.Errorf("index %d: expected %d, got %d", i, expected[i], r)
			}
		}
	})
}

func TestPipeline(t *testing.T) {
	t.Run("multiple stages", func(t *testing.T) {
		pipeline := NewPipeline[int]().
			AddStage(func(_ context.Context, n int) (int, error) {
				return n + 1, nil
			}).
			AddStage(func(_ context.Context, n int) (int, error) {
				return n * 2, nil
			}).
			AddStage(func(_ context.Context, n int) (int, error) {
				return n - 1, nil
			})

		result, err := pipeline.Execute(context.Background(), 5)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// (5 + 1) * 2 - 1 = 11
		if result != 11 {
			t.Errorf("expected 11, got %d", result)
		}
	})

	t.Run("error in stage", func(t *testing.T) {
		expectedErr := errors.New("stage error")
		pipeline := NewPipeline[int]().
			AddStage(func(_ context.Context, n int) (int, error) {
				return n + 1, nil
			}).
			AddStage(func(_ context.Context, _ int) (int, error) {
				return 0, expectedErr
			}).
			AddStage(func(_ context.Context, n int) (int, error) {
				return n * 2, nil
			})

		_, err := pipeline.Execute(context.Background(), 5)

		if !errors.Is(err, expectedErr) {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
	})
}

func TestFanOut(t *testing.T) {
	t.Run("distributes work", func(t *testing.T) {
		fn := func(_ context.Context, input int, workerID int) (int, error) {
			return input + workerID, nil
		}

		results, err := FanOut(context.Background(), 10, 5, fn)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(results) != 5 {
			t.Errorf("expected 5 results, got %d", len(results))
		}

		// Results should be [10, 11, 12, 13, 14]
		for i, r := range results {
			if r != 10+i {
				t.Errorf("worker %d: expected %d, got %d", i, 10+i, r)
			}
		}
	})
}
