// Package tracing provides OpenTelemetry distributed tracing integration.
package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Config holds the configuration for tracing.
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Endpoint       string // OTLP endpoint (e.g., "localhost:4318")
	Insecure       bool
	SampleRate     float64 // 0.0 to 1.0
}

// DefaultConfig returns the default tracing configuration.
func DefaultConfig() *Config {
	return &Config{
		ServiceName:    "go-api-starter",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		Endpoint:       "localhost:4318",
		Insecure:       true,
		SampleRate:     1.0, // Sample all traces in development
	}
}

// Provider manages the OpenTelemetry tracer provider.
type Provider struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
}

// NewProvider creates a new tracing provider.
func NewProvider(cfg *Config) (*Provider, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Create OTLP exporter
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	client := otlptracehttp.NewClient(opts...)
	exporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP exporter: %w", err)
	}

	// Create resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating resource: %w", err)
	}

	// Create sampler
	var sampler sdktrace.Sampler
	switch cfg.SampleRate {
	case 0:
		sampler = sdktrace.NeverSample()
	case 1:
		sampler = sdktrace.AlwaysSample()
	default:
		sampler = sdktrace.TraceIDRatioBased(cfg.SampleRate)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Provider{
		provider: tp,
		tracer:   tp.Tracer(cfg.ServiceName),
	}, nil
}

// Tracer returns the tracer instance.
func (p *Provider) Tracer() trace.Tracer {
	return p.tracer
}

// Shutdown shuts down the tracer provider.
func (p *Provider) Shutdown(ctx context.Context) error {
	return p.provider.Shutdown(ctx)
}

// SpanFromContext returns the current span from context.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// StartSpan starts a new span with the given name.
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer("").Start(ctx, name, opts...)
}

// SetSpanAttributes sets attributes on the current span.
func SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// SetSpanError records an error on the current span.
func SetSpanError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// SetSpanOK marks the span as successful.
func SetSpanOK(ctx context.Context, message string) {
	span := trace.SpanFromContext(ctx)
	span.SetStatus(codes.Ok, message)
}

// Common attribute keys
var (
	// HTTP attributes
	AttrHTTPMethod     = attribute.Key("http.method")
	AttrHTTPURL        = attribute.Key("http.url")
	AttrHTTPStatusCode = attribute.Key("http.status_code")
	AttrHTTPRoute      = attribute.Key("http.route")
	AttrHTTPClientIP   = attribute.Key("http.client_ip")
	AttrHTTPUserAgent  = attribute.Key("http.user_agent")

	// User attributes
	AttrUserID    = attribute.Key("user.id")
	AttrUserEmail = attribute.Key("user.email")
	AttrUserRole  = attribute.Key("user.role")

	// Database attributes
	AttrDBSystem    = attribute.Key("db.system")
	AttrDBStatement = attribute.Key("db.statement")
	AttrDBOperation = attribute.Key("db.operation")
	AttrDBTable     = attribute.Key("db.table")

	// Custom attributes
	AttrRequestID    = attribute.Key("request.id")
	AttrResourceType = attribute.Key("resource.type")
	AttrResourceID   = attribute.Key("resource.id")
)

// HTTPAttributes creates common HTTP span attributes.
func HTTPAttributes(method, url, route, clientIP, userAgent string, statusCode int) []attribute.KeyValue {
	return []attribute.KeyValue{
		AttrHTTPMethod.String(method),
		AttrHTTPURL.String(url),
		AttrHTTPRoute.String(route),
		AttrHTTPClientIP.String(clientIP),
		AttrHTTPUserAgent.String(userAgent),
		AttrHTTPStatusCode.Int(statusCode),
	}
}

// DBAttributes creates common database span attributes.
func DBAttributes(system, operation, table, statement string) []attribute.KeyValue {
	return []attribute.KeyValue{
		AttrDBSystem.String(system),
		AttrDBOperation.String(operation),
		AttrDBTable.String(table),
		AttrDBStatement.String(statement),
	}
}

// UserAttributes creates user-related span attributes.
func UserAttributes(userID int64, email, role string) []attribute.KeyValue {
	return []attribute.KeyValue{
		AttrUserID.Int64(userID),
		AttrUserEmail.String(email),
		AttrUserRole.String(role),
	}
}
