// Package healthcheck provides a comprehensive health check system for monitoring application health.
package healthcheck

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// Status represents the health status of a component.
type Status string

const (
	StatusUp       Status = "up"
	StatusDown     Status = "down"
	StatusDegraded Status = "degraded"
)

// CheckResult represents the result of a health check.
type CheckResult struct {
	Status    Status         `json:"status"`
	Latency   time.Duration  `json:"latency_ms"`
	Message   string         `json:"message,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// Check is a function that performs a health check.
type Check func(ctx context.Context) CheckResult

// Checker manages and executes health checks.
type Checker struct {
	mu       sync.RWMutex
	checks   map[string]Check
	timeout  time.Duration
	parallel bool
}

// Option configures the health checker.
type Option func(*Checker)

// WithTimeout sets the timeout for health checks.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Checker) {
		c.timeout = timeout
	}
}

// WithParallel enables parallel execution of health checks.
func WithParallel(parallel bool) Option {
	return func(c *Checker) {
		c.parallel = parallel
	}
}

// NewChecker creates a new health checker.
func NewChecker(opts ...Option) *Checker {
	c := &Checker{
		checks:   make(map[string]Check),
		timeout:  5 * time.Second,
		parallel: true,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Register registers a health check with the given name.
func (c *Checker) Register(name string, check Check) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = check
}

// CheckAll runs all registered health checks and returns the results.
func (c *Checker) CheckAll(ctx context.Context) map[string]CheckResult {
	c.mu.RLock()
	checks := make(map[string]Check, len(c.checks))
	for k, v := range c.checks {
		checks[k] = v
	}
	c.mu.RUnlock()

	if len(checks) == 0 {
		return map[string]CheckResult{}
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	results := make(map[string]CheckResult, len(checks))

	if c.parallel {
		var mu sync.Mutex
		var wg sync.WaitGroup

		for name, check := range checks {
			wg.Add(1)
			go func(name string, check Check) {
				defer wg.Done()
				result := check(ctx)
				mu.Lock()
				results[name] = result
				mu.Unlock()
			}(name, check)
		}

		wg.Wait()
	} else {
		for name, check := range checks {
			results[name] = check(ctx)
		}
	}

	return results
}

// Response is the HTTP response format for health checks.
type Response struct {
	Status  Status                 `json:"status"`
	Checks  map[string]CheckResult `json:"checks"`
	Version string                 `json:"version,omitempty"`
}

// Handler returns an HTTP handler for health checks.
func (c *Checker) Handler(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := c.CheckAll(r.Context())

		overallStatus := StatusUp
		for _, result := range results {
			if result.Status == StatusDown {
				overallStatus = StatusDown
				break
			}
			if result.Status == StatusDegraded {
				overallStatus = StatusDegraded
			}
		}

		response := Response{
			Status:  overallStatus,
			Checks:  results,
			Version: version,
		}

		w.Header().Set("Content-Type", "application/json")
		if overallStatus == StatusDown {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else if overallStatus == StatusDegraded {
			w.WriteHeader(http.StatusOK) // Still return 200 for degraded
		}

		json.NewEncoder(w).Encode(response)
	}
}

// LivenessHandler returns a simple liveness probe handler.
func LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
	}
}

// ReadinessHandler returns a readiness probe handler that checks all components.
func (c *Checker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := c.CheckAll(r.Context())

		for _, result := range results {
			if result.Status == StatusDown {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]string{"status": "not ready"})
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	}
}

// Database health check builders

// DatabaseCheck creates a health check for a SQL database.
func DatabaseCheck(db *sql.DB) Check {
	return func(ctx context.Context) CheckResult {
		start := time.Now()
		result := CheckResult{
			Status:    StatusUp,
			Timestamp: start,
		}

		err := db.PingContext(ctx)
		result.Latency = time.Since(start)

		if err != nil {
			result.Status = StatusDown
			result.Message = err.Error()
			return result
		}

		// Get connection pool stats
		stats := db.Stats()
		result.Details = map[string]any{
			"open_connections":     stats.OpenConnections,
			"in_use":               stats.InUse,
			"idle":                 stats.Idle,
			"max_open_connections": stats.MaxOpenConnections,
		}

		// Check if connections are exhausted
		if stats.MaxOpenConnections > 0 && stats.InUse >= stats.MaxOpenConnections {
			result.Status = StatusDegraded
			result.Message = "connection pool exhausted"
		}

		return result
	}
}

// HTTPCheck creates a health check for an HTTP endpoint.
func HTTPCheck(url string, timeout time.Duration) Check {
	client := &http.Client{Timeout: timeout}

	return func(ctx context.Context) CheckResult {
		start := time.Now()
		result := CheckResult{
			Status:    StatusUp,
			Timestamp: start,
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			result.Status = StatusDown
			result.Message = err.Error()
			result.Latency = time.Since(start)
			return result
		}

		resp, err := client.Do(req)
		result.Latency = time.Since(start)

		if err != nil {
			result.Status = StatusDown
			result.Message = err.Error()
			return result
		}
		defer resp.Body.Close()

		result.Details = map[string]any{
			"status_code": resp.StatusCode,
		}

		if resp.StatusCode >= 500 {
			result.Status = StatusDown
			result.Message = "server error"
		} else if resp.StatusCode >= 400 {
			result.Status = StatusDegraded
			result.Message = "client error"
		}

		return result
	}
}

// MemoryCheck creates a health check that monitors memory usage.
func MemoryCheck(maxMemoryMB uint64) Check {
	return func(ctx context.Context) CheckResult {
		start := time.Now()
		result := CheckResult{
			Status:    StatusUp,
			Timestamp: start,
		}

		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		allocMB := m.Alloc / 1024 / 1024
		result.Latency = time.Since(start)
		result.Details = map[string]any{
			"alloc_mb":       allocMB,
			"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
			"sys_mb":         m.Sys / 1024 / 1024,
			"num_gc":         m.NumGC,
		}

		if maxMemoryMB > 0 && allocMB > maxMemoryMB {
			result.Status = StatusDegraded
			result.Message = "high memory usage"
		}

		return result
	}
}

// CustomCheck creates a health check with a custom function.
func CustomCheck(name string, fn func(ctx context.Context) error) Check {
	return func(ctx context.Context) CheckResult {
		start := time.Now()
		result := CheckResult{
			Status:    StatusUp,
			Timestamp: start,
		}

		err := fn(ctx)
		result.Latency = time.Since(start)

		if err != nil {
			result.Status = StatusDown
			result.Message = err.Error()
		}

		return result
	}
}
