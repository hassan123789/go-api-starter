// Package metrics provides Prometheus metrics for HTTP API observability.
//
// # Overview
//
// This package provides comprehensive metrics collection for HTTP APIs:
//   - HTTP request metrics (count, duration, in-flight, response size)
//   - Business metrics (active users, todos, auth attempts)
//   - Custom metrics support
//   - Echo middleware integration
//
// # Basic Usage
//
//	m := metrics.New()
//	e := echo.New()
//	e.Use(m.Middleware())
//	e.GET("/metrics", echo.WrapHandler(m.Handler()))
//
// # With Custom Registry
//
//	registry := prometheus.NewRegistry()
//	m := metrics.NewWithRegistry(registry)
//
// # Available HTTP Metrics
//
//   - http_requests_total: Counter of HTTP requests by method, path, status
//   - http_request_duration_seconds: Histogram of request latencies
//   - http_requests_in_flight: Gauge of currently processing requests
//   - http_response_size_bytes: Histogram of response sizes
//
// # Business Metrics
//
// Record business events:
//
//	m.RecordActiveUsers(150)
//	m.RecordTodoCreated()
//	m.RecordAuthAttempt(true)  // successful login
//	m.RecordAuthAttempt(false) // failed login
//
// # Custom Metrics
//
// Register and record custom metrics:
//
//	m.CustomCounter("api_calls_total", "Total API calls", "service")
//	m.IncrementCustomCounter("api_calls_total", "user-service")
//
// # Integration with Prometheus
//
// Configure Prometheus to scrape the /metrics endpoint:
//
//	scrape_configs:
//	  - job_name: 'go-api-starter'
//	    static_configs:
//	      - targets: ['localhost:8080']
//	    metrics_path: /metrics
//
// # Grafana Dashboard
//
// The metrics are designed to work with standard Prometheus/Grafana
// dashboards for API monitoring.
package metrics
