package circuitbreaker

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Benchmark tests for circuit breaker

func BenchmarkCircuitBreaker_Execute_Success(b *testing.B) {
	opts := DefaultOptions()
	cb := New(opts)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Execute(ctx, func(ctx context.Context) error {
			return nil
		})
	}
}

func BenchmarkCircuitBreaker_Execute_Failure(b *testing.B) {
	opts := DefaultOptions()
	opts.MaxFailures = 1000000 // Prevent opening
	cb := New(opts)
	ctx := context.Background()
	testErr := errors.New("test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}
}

func BenchmarkCircuitBreaker_Execute_Open(b *testing.B) {
	opts := DefaultOptions()
	opts.MaxFailures = 1
	opts.Timeout = time.Hour // Keep open
	cb := New(opts)
	ctx := context.Background()

	// Open the circuit
	_ = cb.Execute(ctx, func(ctx context.Context) error {
		return errors.New("error")
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Execute(ctx, func(ctx context.Context) error {
			return nil
		})
	}
}

func BenchmarkCircuitBreaker_State(b *testing.B) {
	cb := New(DefaultOptions())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.State()
	}
}

func BenchmarkCircuitBreaker_ExecuteWithFallback(b *testing.B) {
	opts := DefaultOptions()
	opts.MaxFailures = 1
	opts.Timeout = time.Hour
	cb := New(opts)
	ctx := context.Background()

	// Open the circuit
	_ = cb.Execute(ctx, func(ctx context.Context) error {
		return errors.New("error")
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.ExecuteWithFallback(
			ctx,
			func(ctx context.Context) error {
				return nil
			},
			func(ctx context.Context, err error) error {
				return nil
			},
		)
	}
}

func BenchmarkRegistry_Get(b *testing.B) {
	registry := NewRegistry(DefaultOptions())

	// Pre-populate
	for i := 0; i < 100; i++ {
		registry.Get("service-" + string(rune(i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.Get("service-50")
	}
}

func BenchmarkRegistry_Get_New(b *testing.B) {
	registry := NewRegistry(DefaultOptions())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Always creates new (not realistic but tests creation)
		b.StopTimer()
		registry = NewRegistry(DefaultOptions())
		b.StartTimer()
		_ = registry.Get("new-service")
	}
}

func BenchmarkParallel_Execute(b *testing.B) {
	cb := New(DefaultOptions())
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = cb.Execute(ctx, func(ctx context.Context) error {
				return nil
			})
		}
	})
}

func BenchmarkParallel_Registry(b *testing.B) {
	registry := NewRegistry(DefaultOptions())
	ctx := context.Background()

	// Pre-populate some services
	services := []string{"service-a", "service-b", "service-c", "service-d"}
	for _, s := range services {
		registry.Get(s)
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cb := registry.Get(services[i%len(services)])
			_ = cb.Execute(ctx, func(ctx context.Context) error {
				return nil
			})
			i++
		}
	})
}
