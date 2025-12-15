package healthcheck

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestChecker_CheckAll(t *testing.T) {
	t.Run("no checks registered", func(t *testing.T) {
		checker := NewChecker()
		results := checker.CheckAll(context.Background())
		if len(results) != 0 {
			t.Errorf("expected no results, got %d", len(results))
		}
	})

	t.Run("all checks pass", func(t *testing.T) {
		checker := NewChecker()
		checker.Register("test1", func(_ context.Context) CheckResult {
			return CheckResult{Status: StatusUp}
		})
		checker.Register("test2", func(_ context.Context) CheckResult {
			return CheckResult{Status: StatusUp}
		})

		results := checker.CheckAll(context.Background())
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}

		for name, result := range results {
			if result.Status != StatusUp {
				t.Errorf("%s: expected StatusUp, got %s", name, result.Status)
			}
		}
	})

	t.Run("check fails", func(t *testing.T) {
		checker := NewChecker()
		checker.Register("failing", func(_ context.Context) CheckResult {
			return CheckResult{Status: StatusDown, Message: "connection failed"}
		})

		results := checker.CheckAll(context.Background())
		if results["failing"].Status != StatusDown {
			t.Errorf("expected StatusDown, got %s", results["failing"].Status)
		}
	})

	t.Run("parallel execution", func(t *testing.T) {
		checker := NewChecker(WithParallel(true))

		for i := 0; i < 5; i++ {
			checker.Register(string(rune('A'+i)), func(_ context.Context) CheckResult {
				time.Sleep(50 * time.Millisecond)
				return CheckResult{Status: StatusUp}
			})
		}

		start := time.Now()
		results := checker.CheckAll(context.Background())
		duration := time.Since(start)

		if len(results) != 5 {
			t.Errorf("expected 5 results, got %d", len(results))
		}

		// Parallel execution should take ~50ms, not ~250ms
		if duration > 200*time.Millisecond {
			t.Errorf("parallel execution took too long: %v", duration)
		}
	})

	t.Run("sequential execution", func(t *testing.T) {
		checker := NewChecker(WithParallel(false))

		for i := 0; i < 3; i++ {
			checker.Register(string(rune('A'+i)), func(_ context.Context) CheckResult {
				time.Sleep(20 * time.Millisecond)
				return CheckResult{Status: StatusUp}
			})
		}

		start := time.Now()
		results := checker.CheckAll(context.Background())
		duration := time.Since(start)

		if len(results) != 3 {
			t.Errorf("expected 3 results, got %d", len(results))
		}

		// Sequential execution should take ~60ms
		if duration < 50*time.Millisecond {
			t.Errorf("sequential execution was too fast: %v", duration)
		}
	})
}

func TestChecker_Handler(t *testing.T) {
	t.Run("returns 200 when all up", func(t *testing.T) {
		checker := NewChecker()
		checker.Register("test", func(_ context.Context) CheckResult {
			return CheckResult{Status: StatusUp}
		})

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		checker.Handler("1.0.0")(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("returns 503 when down", func(t *testing.T) {
		checker := NewChecker()
		checker.Register("test", func(_ context.Context) CheckResult {
			return CheckResult{Status: StatusDown}
		})

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		checker.Handler("1.0.0")(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", rec.Code)
		}
	})

	t.Run("returns 200 when degraded", func(t *testing.T) {
		checker := NewChecker()
		checker.Register("test", func(_ context.Context) CheckResult {
			return CheckResult{Status: StatusDegraded}
		})

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		checker.Handler("1.0.0")(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200 for degraded, got %d", rec.Code)
		}
	})
}

func TestLivenessHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	rec := httptest.NewRecorder()

	LivenessHandler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestReadinessHandler(t *testing.T) {
	t.Run("ready when all up", func(t *testing.T) {
		checker := NewChecker()
		checker.Register("test", func(_ context.Context) CheckResult {
			return CheckResult{Status: StatusUp}
		})

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		checker.ReadinessHandler()(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("not ready when down", func(t *testing.T) {
		checker := NewChecker()
		checker.Register("test", func(_ context.Context) CheckResult {
			return CheckResult{Status: StatusDown}
		})

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		checker.ReadinessHandler()(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", rec.Code)
		}
	})
}

func TestMemoryCheck(t *testing.T) {
	check := MemoryCheck(0) // No limit

	result := check(context.Background())

	if result.Status != StatusUp {
		t.Errorf("expected StatusUp, got %s", result.Status)
	}

	if result.Details["alloc_mb"] == nil {
		t.Error("expected alloc_mb in details")
	}
}

func TestCustomCheck(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		check := CustomCheck("custom", func(_ context.Context) error {
			return nil
		})

		result := check(context.Background())
		if result.Status != StatusUp {
			t.Errorf("expected StatusUp, got %s", result.Status)
		}
	})

	t.Run("failure", func(t *testing.T) {
		check := CustomCheck("custom", func(_ context.Context) error {
			return errors.New("custom error")
		})

		result := check(context.Background())
		if result.Status != StatusDown {
			t.Errorf("expected StatusDown, got %s", result.Status)
		}
		if result.Message != "custom error" {
			t.Errorf("expected 'custom error', got %s", result.Message)
		}
	})
}

func TestHTTPCheck(t *testing.T) {
	t.Run("healthy endpoint", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		check := HTTPCheck(server.URL, 5*time.Second)
		result := check(context.Background())

		if result.Status != StatusUp {
			t.Errorf("expected StatusUp, got %s", result.Status)
		}
	})

	t.Run("unhealthy endpoint", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		check := HTTPCheck(server.URL, 5*time.Second)
		result := check(context.Background())

		if result.Status != StatusDown {
			t.Errorf("expected StatusDown, got %s", result.Status)
		}
	})

	t.Run("connection refused", func(t *testing.T) {
		check := HTTPCheck("http://localhost:99999", 1*time.Second)
		result := check(context.Background())

		if result.Status != StatusDown {
			t.Errorf("expected StatusDown, got %s", result.Status)
		}
	})
}
