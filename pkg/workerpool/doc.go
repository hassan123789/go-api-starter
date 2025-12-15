// Package workerpool provides a generic concurrent worker pool implementation.
//
// # Overview
//
// This package implements a bounded worker pool pattern with:
//   - Generic type support for input and output
//   - Configurable number of workers
//   - Graceful shutdown
//   - Context cancellation support
//
// # Basic Usage
//
//	// Create pool with 4 workers
//	pool := workerpool.New[int, int](4, func(ctx context.Context, n int) (int, error) {
//	    return n * 2, nil
//	})
//
//	// Process items
//	inputs := []int{1, 2, 3, 4, 5}
//	results := workerpool.Process(context.Background(), pool, inputs)
//
//	for _, r := range results {
//	    if r.Err != nil {
//	        log.Printf("error: %v", r.Err)
//	        continue
//	    }
//	    fmt.Println(r.Value)
//	}
//
// # Error Handling
//
// Each result includes both a value and an error, allowing individual
// task failures without affecting the entire batch:
//
//	pool := workerpool.New[string, *Response](10, func(ctx context.Context, url string) (*Response, error) {
//	    return fetchURL(ctx, url)
//	})
//
//	results := workerpool.Process(ctx, pool, urls)
//	for _, r := range results {
//	    if r.Err != nil {
//	        log.Printf("failed to fetch: %v", r.Err)
//	        continue
//	    }
//	    handleResponse(r.Value)
//	}
//
// # Cancellation
//
// The pool respects context cancellation:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	results := workerpool.Process(ctx, pool, largeDataset)
//
// # Resource Management
//
// Always close the pool when done:
//
//	pool := workerpool.New[int, int](4, processFunc)
//	defer pool.Close()
//
// # Use Cases
//
//   - Batch API requests with rate limiting
//   - Parallel file processing
//   - Concurrent database operations
//   - Image/video processing pipelines
package workerpool
