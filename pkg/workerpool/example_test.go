package workerpool_test

import (
	"context"
	"fmt"

	"github.com/zareh/go-api-starter/pkg/workerpool"
)

// ExampleNew demonstrates creating and using a generic worker pool.
func ExampleNew() {
	// Create a worker pool that doubles numbers
	pool := workerpool.New(4, func(ctx context.Context, n int) (int, error) {
		return n * 2, nil
	})

	ctx := context.Background()
	pool.Start(ctx)

	// Submit tasks
	for i := 1; i <= 5; i++ {
		pool.Submit(ctx, i)
	}

	// Collect results
	go func() {
		for range pool.Results() {
			// Process results
		}
	}()

	pool.Stop()
	fmt.Println("Worker pool stopped")

	// Output:
	// Worker pool stopped
}

// ExampleProcess demonstrates batch processing with worker pool.
func ExampleProcess() {
	numbers := []int{1, 2, 3, 4, 5}

	// Process all numbers in parallel
	results, errors := workerpool.Process(
		context.Background(),
		4, // worker count
		numbers,
		func(ctx context.Context, n int) (int, error) {
			return n * 2, nil
		},
	)

	// Note: Results order may vary due to concurrent execution
	fmt.Printf("Results count: %d\n", len(results))
	fmt.Printf("Errors: %d\n", len(errors))

	// Output:
	// Results count: 5
	// Errors: 0
}
