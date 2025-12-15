// Package metrics provides Prometheus metrics for HTTP servers.
//
// This package implements production-ready observability with:
//   - HTTP request counters with labels
//   - Response time histograms
//   - Active connection gauges
//   - Custom business metrics support
//
// Example usage:
//
//	m := metrics.New()
//	e.Use(m.Middleware())
//	e.GET("/metrics", echo.WrapHandler(m.Handler()))
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics for the application.
type Metrics struct {
	registry *prometheus.Registry

	// HTTP metrics
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight *prometheus.GaugeVec
	httpResponseSize     *prometheus.HistogramVec

	// Business metrics
	activeUsers      prometheus.Gauge
	todosCreated     prometheus.Counter
	todosCompleted   prometheus.Counter
	authAttempts     *prometheus.CounterVec
	circuitBreakers  *prometheus.GaugeVec
	workerPoolActive *prometheus.GaugeVec
}

// Option is a functional option for configuring Metrics.
type Option func(*Metrics)

// WithRegistry sets a custom Prometheus registry.
func WithRegistry(r *prometheus.Registry) Option {
	return func(m *Metrics) {
		m.registry = r
	}
}

// New creates a new Metrics instance with default configuration.
func New(opts ...Option) *Metrics {
	m := &Metrics{
		registry: prometheus.NewRegistry(),
	}

	for _, opt := range opts {
		opt(m)
	}

	// Register Go runtime metrics
	m.registry.MustRegister(collectors.NewGoCollector())
	m.registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// Initialize HTTP metrics
	m.initHTTPMetrics()

	// Initialize business metrics
	m.initBusinessMetrics()

	return m
}

func (m *Metrics) initHTTPMetrics() {
	m.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests processed",
		},
		[]string{"method", "path", "status"},
	)

	m.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request latency in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path", "status"},
	)

	m.httpRequestsInFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "http",
			Name:      "requests_in_flight",
			Help:      "Current number of HTTP requests being processed",
		},
		[]string{"method"},
	)

	m.httpResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "http",
			Name:      "response_size_bytes",
			Help:      "HTTP response size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 7), // 100B to 100MB
		},
		[]string{"method", "path"},
	)

	m.registry.MustRegister(
		m.httpRequestsTotal,
		m.httpRequestDuration,
		m.httpRequestsInFlight,
		m.httpResponseSize,
	)
}

func (m *Metrics) initBusinessMetrics() {
	m.activeUsers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "app",
			Name:      "active_users",
			Help:      "Current number of active users",
		},
	)

	m.todosCreated = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "app",
			Name:      "todos_created_total",
			Help:      "Total number of TODOs created",
		},
	)

	m.todosCompleted = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "app",
			Name:      "todos_completed_total",
			Help:      "Total number of TODOs completed",
		},
	)

	m.authAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Name:      "auth_attempts_total",
			Help:      "Total number of authentication attempts",
		},
		[]string{"type", "success"},
	)

	m.circuitBreakers = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "app",
			Name:      "circuit_breaker_state",
			Help:      "Circuit breaker state (0=closed, 1=open, 2=half-open)",
		},
		[]string{"name"},
	)

	m.workerPoolActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "app",
			Name:      "worker_pool_active_workers",
			Help:      "Current number of active workers in pool",
		},
		[]string{"pool"},
	)

	m.registry.MustRegister(
		m.activeUsers,
		m.todosCreated,
		m.todosCompleted,
		m.authAttempts,
		m.circuitBreakers,
		m.workerPoolActive,
	)
}

// Middleware returns an Echo middleware that records HTTP metrics.
func (m *Metrics) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			method := req.Method
			path := c.Path() // Use route pattern, not actual path (prevents cardinality explosion)

			// Track in-flight requests
			m.httpRequestsInFlight.WithLabelValues(method).Inc()
			defer m.httpRequestsInFlight.WithLabelValues(method).Dec()

			start := time.Now()
			err := next(c)
			duration := time.Since(start).Seconds()

			status := strconv.Itoa(c.Response().Status)
			size := float64(c.Response().Size)

			// Record metrics
			m.httpRequestsTotal.WithLabelValues(method, path, status).Inc()
			m.httpRequestDuration.WithLabelValues(method, path, status).Observe(duration)
			m.httpResponseSize.WithLabelValues(method, path).Observe(size)

			return err
		}
	}
}

// Handler returns an HTTP handler for the /metrics endpoint.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// Registry returns the underlying Prometheus registry.
func (m *Metrics) Registry() *prometheus.Registry {
	return m.registry
}

// IncTodoCreated increments the todos created counter.
func (m *Metrics) IncTodoCreated() {
	m.todosCreated.Inc()
}

// IncTodoCompleted increments the todos completed counter.
func (m *Metrics) IncTodoCompleted() {
	m.todosCompleted.Inc()
}

// SetActiveUsers sets the current number of active users.
func (m *Metrics) SetActiveUsers(count float64) {
	m.activeUsers.Set(count)
}

// RecordAuthAttempt records an authentication attempt.
func (m *Metrics) RecordAuthAttempt(authType string, success bool) {
	m.authAttempts.WithLabelValues(authType, strconv.FormatBool(success)).Inc()
}

// SetCircuitBreakerState sets the circuit breaker state metric.
// state: 0=closed, 1=open, 2=half-open
func (m *Metrics) SetCircuitBreakerState(name string, state int) {
	m.circuitBreakers.WithLabelValues(name).Set(float64(state))
}

// SetWorkerPoolActive sets the number of active workers in a pool.
func (m *Metrics) SetWorkerPoolActive(pool string, count int) {
	m.workerPoolActive.WithLabelValues(pool).Set(float64(count))
}

// CustomCounter creates a new custom counter metric.
func (m *Metrics) CustomCounter(name, help string, labels ...string) *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "custom",
			Name:      name,
			Help:      help,
		},
		labels,
	)
	m.registry.MustRegister(counter)
	return counter
}

// CustomGauge creates a new custom gauge metric.
func (m *Metrics) CustomGauge(name, help string, labels ...string) *prometheus.GaugeVec {
	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "custom",
			Name:      name,
			Help:      help,
		},
		labels,
	)
	m.registry.MustRegister(gauge)
	return gauge
}

// CustomHistogram creates a new custom histogram metric.
func (m *Metrics) CustomHistogram(name, help string, buckets []float64, labels ...string) *prometheus.HistogramVec {
	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "custom",
			Name:      name,
			Help:      help,
			Buckets:   buckets,
		},
		labels,
	)
	m.registry.MustRegister(histogram)
	return histogram
}
