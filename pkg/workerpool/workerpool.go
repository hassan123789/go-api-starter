// Package workerpool provides a generic worker pool implementation for concurrent task processing.
package workerpool

import (
	"context"
	"sync"
)

// Task represents a unit of work to be processed.
type Task[T any, R any] struct {
	Input  T
	Result R
	Err    error
}

// Pool represents a worker pool that processes tasks concurrently.
type Pool[T any, R any] struct {
	workers    int
	taskQueue  chan *Task[T, R]
	resultChan chan *Task[T, R]
	processor  func(context.Context, T) (R, error)
	wg         sync.WaitGroup
}

// New creates a new worker pool with the specified number of workers.
func New[T any, R any](workers int, processor func(context.Context, T) (R, error)) *Pool[T, R] {
	if workers <= 0 {
		workers = 1
	}

	return &Pool[T, R]{
		workers:    workers,
		taskQueue:  make(chan *Task[T, R], workers*2),
		resultChan: make(chan *Task[T, R], workers*2),
		processor:  processor,
	}
}

// Start starts the worker pool.
func (p *Pool[T, R]) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}
}

// worker is the main worker goroutine.
func (p *Pool[T, R]) worker(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-p.taskQueue:
			if !ok {
				return
			}

			result, err := p.processor(ctx, task.Input)
			task.Result = result
			task.Err = err

			select {
			case p.resultChan <- task:
			case <-ctx.Done():
				return
			}
		}
	}
}

// Submit submits a task to the worker pool.
func (p *Pool[T, R]) Submit(ctx context.Context, input T) bool {
	task := &Task[T, R]{Input: input}

	select {
	case p.taskQueue <- task:
		return true
	case <-ctx.Done():
		return false
	}
}

// Results returns the channel for receiving results.
func (p *Pool[T, R]) Results() <-chan *Task[T, R] {
	return p.resultChan
}

// Stop stops the worker pool and waits for all workers to finish.
func (p *Pool[T, R]) Stop() {
	close(p.taskQueue)
	p.wg.Wait()
	close(p.resultChan)
}

// Process processes all inputs and returns results.
// This is a convenience method for batch processing.
func Process[T any, R any](ctx context.Context, workers int, inputs []T, processor func(context.Context, T) (R, error)) ([]R, []error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	pool := New(workers, processor)
	pool.Start(ctx)

	// Submit all tasks
	go func() {
		for _, input := range inputs {
			if !pool.Submit(ctx, input) {
				break
			}
		}
		pool.Stop()
	}()

	// Collect results
	results := make([]R, 0, len(inputs))
	errors := make([]error, 0)

	for task := range pool.Results() {
		if task.Err != nil {
			errors = append(errors, task.Err)
		} else {
			results = append(results, task.Result)
		}
	}

	return results, errors
}

// MapConcurrent applies a function to each element concurrently.
func MapConcurrent[T any, R any](ctx context.Context, workers int, items []T, fn func(context.Context, T) (R, error)) ([]R, error) {
	if len(items) == 0 {
		return nil, nil
	}

	type indexedResult struct {
		index  int
		result R
		err    error
	}

	resultChan := make(chan indexedResult, len(items))
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	for i, item := range items {
		wg.Add(1)
		go func(idx int, input T) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				resultChan <- indexedResult{index: idx, err: ctx.Err()}
				return
			}

			result, err := fn(ctx, input)
			resultChan <- indexedResult{index: idx, result: result, err: err}
		}(i, item)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	results := make([]R, len(items))
	var firstErr error

	for ir := range resultChan {
		if ir.err != nil && firstErr == nil {
			firstErr = ir.err
		}
		results[ir.index] = ir.result
	}

	return results, firstErr
}

// Pipeline represents a stage in a processing pipeline.
type Pipeline[T any] struct {
	stages []func(context.Context, T) (T, error)
}

// NewPipeline creates a new processing pipeline.
func NewPipeline[T any]() *Pipeline[T] {
	return &Pipeline[T]{
		stages: make([]func(context.Context, T) (T, error), 0),
	}
}

// AddStage adds a processing stage to the pipeline.
func (p *Pipeline[T]) AddStage(stage func(context.Context, T) (T, error)) *Pipeline[T] {
	p.stages = append(p.stages, stage)
	return p
}

// Execute runs the pipeline on the input.
func (p *Pipeline[T]) Execute(ctx context.Context, input T) (T, error) {
	current := input
	var err error

	for _, stage := range p.stages {
		select {
		case <-ctx.Done():
			return current, ctx.Err()
		default:
		}

		current, err = stage(ctx, current)
		if err != nil {
			return current, err
		}
	}

	return current, nil
}

// FanOut distributes work across multiple goroutines and collects results.
func FanOut[T any, R any](ctx context.Context, input T, workers int, fn func(context.Context, T, int) (R, error)) ([]R, error) {
	resultChan := make(chan struct {
		index  int
		result R
		err    error
	}, workers)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			result, err := fn(ctx, input, workerID)
			resultChan <- struct {
				index  int
				result R
				err    error
			}{index: workerID, result: result, err: err}
		}(i)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	results := make([]R, workers)
	var firstErr error

	for r := range resultChan {
		if r.err != nil && firstErr == nil {
			firstErr = r.err
		}
		results[r.index] = r.result
	}

	return results, firstErr
}
