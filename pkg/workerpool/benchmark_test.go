package workerpool

import (
	"context"
	"testing"
	"time"
)

// Benchmark tests for worker pool

func BenchmarkPool_Submit(b *testing.B) {
	pool := New(4, func(ctx context.Context, n int) (int, error) {
		return n * 2, nil
	})
	ctx := context.Background()
	pool.Start(ctx)

	// Drain results
	go func() {
		for range pool.Results() {
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Submit(ctx, i)
	}
	b.StopTimer()
	pool.Stop()
}

func BenchmarkProcess(b *testing.B) {
	sizes := []int{10, 100, 1000}
	workers := []int{1, 4, 8}

	for _, size := range sizes {
		for _, w := range workers {
			inputs := make([]int, size)
			for i := range inputs {
				inputs[i] = i
			}

			b.Run(
				"size="+string(rune(size))+"_workers="+string(rune(w)),
				func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						_, _ = Process(
							context.Background(),
							w,
							inputs,
							func(ctx context.Context, n int) (int, error) {
								return n * 2, nil
							},
						)
					}
				},
			)
		}
	}
}

func BenchmarkProcess_WithWork(b *testing.B) {
	inputs := make([]int, 100)
	for i := range inputs {
		inputs[i] = i
	}

	b.Run("NoWork", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Process(
				context.Background(),
				4,
				inputs,
				func(ctx context.Context, n int) (int, error) {
					return n * 2, nil
				},
			)
		}
	})

	b.Run("LightWork", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Process(
				context.Background(),
				4,
				inputs,
				func(ctx context.Context, n int) (int, error) {
					// Simulate light work
					sum := 0
					for j := 0; j < 100; j++ {
						sum += j
					}
					return sum, nil
				},
			)
		}
	})

	b.Run("HeavyWork", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Process(
				context.Background(),
				4,
				inputs,
				func(ctx context.Context, n int) (int, error) {
					// Simulate heavy work
					time.Sleep(time.Microsecond)
					return n * 2, nil
				},
			)
		}
	})
}

func BenchmarkWorkerCount(b *testing.B) {
	inputs := make([]int, 1000)
	for i := range inputs {
		inputs[i] = i
	}

	for _, workers := range []int{1, 2, 4, 8, 16, 32} {
		b.Run("workers="+string(rune(workers)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = Process(
					context.Background(),
					workers,
					inputs,
					func(ctx context.Context, n int) (int, error) {
						// Light CPU work
						sum := 0
						for j := 0; j < 10; j++ {
							sum += j * n
						}
						return sum, nil
					},
				)
			}
		})
	}
}
