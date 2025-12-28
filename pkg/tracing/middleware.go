// Package tracing provides OpenTelemetry distributed tracing integration.
package tracing

import (
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Middleware returns an Echo middleware for OpenTelemetry tracing.
func Middleware(serviceName string) echo.MiddlewareFunc {
	tracer := otel.Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			// Extract trace context from incoming request headers
			ctx := propagator.Extract(req.Context(), propagation.HeaderCarrier(req.Header))

			// Start a new span for this request
			spanName := req.Method + " " + c.Path()
			ctx, span := tracer.Start(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.method", req.Method),
					attribute.String("http.url", req.URL.String()),
					attribute.String("http.route", c.Path()),
					attribute.String("http.scheme", c.Scheme()),
					attribute.String("net.host.name", req.Host),
					attribute.String("user_agent.original", req.UserAgent()),
					attribute.String("http.client_ip", c.RealIP()),
				),
			)
			defer span.End()

			// Store trace context in request
			c.SetRequest(req.WithContext(ctx))

			// Add request ID if available
			if requestID := c.Request().Header.Get(echo.HeaderXRequestID); requestID != "" {
				span.SetAttributes(AttrRequestID.String(requestID))
			}

			// Process request
			err := next(c)

			// Record response status
			status := c.Response().Status
			span.SetAttributes(attribute.Int("http.status_code", status))

			// Handle errors
			if err != nil {
				span.RecordError(err)
				span.SetAttributes(attribute.String("error.message", err.Error()))
			}

			// Set span status based on HTTP status code
			if status >= 500 {
				span.SetStatus(codes.Error, "Server Error")
			} else if status >= 400 {
				span.SetStatus(codes.Error, "Client Error")
			} else {
				span.SetStatus(codes.Ok, "")
			}

			// Inject trace context into response headers for debugging
			propagator.Inject(ctx, propagation.HeaderCarrier(c.Response().Header()))

			return err
		}
	}
}

// TraceIDFromContext returns the trace ID from the context.
func TraceIDFromContext(c echo.Context) string {
	span := trace.SpanFromContext(c.Request().Context())
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// SpanIDFromContext returns the span ID from the context.
func SpanIDFromContext(c echo.Context) string {
	span := trace.SpanFromContext(c.Request().Context())
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// AddUserToSpan adds user information to the current span.
func AddUserToSpan(c echo.Context, userID int64, email, role string) {
	span := trace.SpanFromContext(c.Request().Context())
	span.SetAttributes(UserAttributes(userID, email, role)...)
}
