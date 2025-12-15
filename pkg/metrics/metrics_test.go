package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNew(t *testing.T) {
	m := New()
	if m == nil {
		t.Fatal("expected Metrics to be non-nil")
	}
	if m.registry == nil {
		t.Fatal("expected registry to be non-nil")
	}
}

func TestNewWithRegistry(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := New(WithRegistry(reg))
	if m.registry != reg {
		t.Fatal("expected custom registry to be used")
	}
}

func TestMiddleware(t *testing.T) {
	m := New()
	e := echo.New()
	e.Use(m.Middleware())
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// Verify metrics were recorded
	metrics, err := m.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	found := false
	for _, mf := range metrics {
		if strings.HasPrefix(mf.GetName(), "http_requests_total") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected http_requests_total metric to be recorded")
	}
}

func TestBusinessMetrics(t *testing.T) {
	m := New()

	// Test IncTodoCreated
	m.IncTodoCreated()
	m.IncTodoCreated()

	// Test IncTodoCompleted
	m.IncTodoCompleted()

	// Test SetActiveUsers
	m.SetActiveUsers(42)

	// Test RecordAuthAttempt
	m.RecordAuthAttempt("login", true)
	m.RecordAuthAttempt("login", false)

	// Test SetCircuitBreakerState
	m.SetCircuitBreakerState("db", 0)
	m.SetCircuitBreakerState("db", 1)

	// Test SetWorkerPoolActive
	m.SetWorkerPoolActive("default", 5)

	// Gather and verify
	metrics, err := m.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	expectedMetrics := []string{
		"app_todos_created_total",
		"app_todos_completed_total",
		"app_active_users",
		"app_auth_attempts_total",
		"app_circuit_breaker_state",
		"app_worker_pool_active_workers",
	}

	foundMetrics := make(map[string]bool)
	for _, mf := range metrics {
		foundMetrics[mf.GetName()] = true
	}

	for _, expected := range expectedMetrics {
		if !foundMetrics[expected] {
			t.Errorf("expected metric %s not found", expected)
		}
	}
}

func TestCustomMetrics(t *testing.T) {
	m := New()

	// Test CustomCounter
	counter := m.CustomCounter("test_counter", "A test counter", "label1")
	counter.WithLabelValues("value1").Inc()

	// Test CustomGauge
	gauge := m.CustomGauge("test_gauge", "A test gauge", "label1")
	gauge.WithLabelValues("value1").Set(123)

	// Test CustomHistogram
	histogram := m.CustomHistogram("test_histogram", "A test histogram", []float64{1, 5, 10}, "label1")
	histogram.WithLabelValues("value1").Observe(3)

	// Verify all were registered
	metrics, err := m.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	expectedMetrics := []string{
		"custom_test_counter",
		"custom_test_gauge",
		"custom_test_histogram",
	}

	foundMetrics := make(map[string]bool)
	for _, mf := range metrics {
		foundMetrics[mf.GetName()] = true
	}

	for _, expected := range expectedMetrics {
		if !foundMetrics[expected] {
			t.Errorf("expected metric %s not found", expected)
		}
	}
}

func TestHandler(t *testing.T) {
	m := New()
	handler := m.Handler()
	if handler == nil {
		t.Fatal("expected handler to be non-nil")
	}
}

func BenchmarkMiddleware(b *testing.B) {
	m := New()
	e := echo.New()
	e.Use(m.Middleware())
	e.GET("/bench", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/bench", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
	}
}
