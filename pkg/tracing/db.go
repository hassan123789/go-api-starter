// Package tracing provides database tracing utilities for OpenTelemetry.
package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// DBTracer provides tracing for database operations.
type DBTracer struct {
	tracer trace.Tracer
	dbName string
}

// NewDBTracer creates a new database tracer.
func NewDBTracer(serviceName, dbName string) *DBTracer {
	return &DBTracer{
		tracer: otel.Tracer(serviceName + "/db"),
		dbName: dbName,
	}
}

// TraceQuery traces a database query operation.
func (t *DBTracer) TraceQuery(ctx context.Context, operation, table, query string) (newCtx context.Context, finish func(error)) {
	spanName := operation + " " + table
	ctx, span := t.tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			AttrDBSystem.String("postgresql"),
			AttrDBOperation.String(operation),
			AttrDBTable.String(table),
			attribute.String("db.name", t.dbName),
		),
	)

	// Only add statement in development (avoid PII in production)
	if query != "" {
		span.SetAttributes(AttrDBStatement.String(query))
	}

	return ctx, func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "success")
		}
		span.End()
	}
}

// TraceCreate traces a CREATE operation.
func (t *DBTracer) TraceCreate(ctx context.Context, table string, resourceID int64) (newCtx context.Context, finish func(error)) {
	ctx, span := t.tracer.Start(ctx, "INSERT "+table,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			AttrDBSystem.String("postgresql"),
			AttrDBOperation.String("INSERT"),
			AttrDBTable.String(table),
			AttrResourceID.Int64(resourceID),
		),
	)

	return ctx, func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "created")
		}
		span.End()
	}
}

// TraceRead traces a READ operation.
func (t *DBTracer) TraceRead(ctx context.Context, table string, resourceID int64) (newCtx context.Context, finish func(error)) {
	ctx, span := t.tracer.Start(ctx, "SELECT "+table,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			AttrDBSystem.String("postgresql"),
			AttrDBOperation.String("SELECT"),
			AttrDBTable.String(table),
			AttrResourceID.Int64(resourceID),
		),
	)

	return ctx, func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "found")
		}
		span.End()
	}
}

// TraceUpdate traces an UPDATE operation.
func (t *DBTracer) TraceUpdate(ctx context.Context, table string, resourceID int64) (newCtx context.Context, finish func(error)) {
	ctx, span := t.tracer.Start(ctx, "UPDATE "+table,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			AttrDBSystem.String("postgresql"),
			AttrDBOperation.String("UPDATE"),
			AttrDBTable.String(table),
			AttrResourceID.Int64(resourceID),
		),
	)

	return ctx, func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "updated")
		}
		span.End()
	}
}

// TraceDelete traces a DELETE operation.
func (t *DBTracer) TraceDelete(ctx context.Context, table string, resourceID int64) (newCtx context.Context, finish func(error)) {
	ctx, span := t.tracer.Start(ctx, "DELETE "+table,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			AttrDBSystem.String("postgresql"),
			AttrDBOperation.String("DELETE"),
			AttrDBTable.String(table),
			AttrResourceID.Int64(resourceID),
		),
	)

	return ctx, func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "deleted")
		}
		span.End()
	}
}

// TraceList traces a LIST operation.
func (t *DBTracer) TraceList(ctx context.Context, table string, limit, offset int) (newCtx context.Context, finish func(error, int)) {
	ctx, span := t.tracer.Start(ctx, "SELECT ALL "+table,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			AttrDBSystem.String("postgresql"),
			AttrDBOperation.String("SELECT"),
			AttrDBTable.String(table),
			attribute.Int("db.limit", limit),
			attribute.Int("db.offset", offset),
		),
	)

	return ctx, func(err error, count int) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "success")
			span.SetAttributes(attribute.Int("db.rows_affected", count))
		}
		span.End()
	}
}

// TraceTransaction traces a database transaction.
func (t *DBTracer) TraceTransaction(ctx context.Context, name string) (newCtx context.Context, finish func(error)) {
	ctx, span := t.tracer.Start(ctx, "TX "+name,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			AttrDBSystem.String("postgresql"),
			AttrDBOperation.String("TRANSACTION"),
			attribute.String("db.transaction.name", name),
		),
	)

	return ctx, func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "rollback: "+err.Error())
			span.SetAttributes(attribute.Bool("db.transaction.committed", false))
		} else {
			span.SetStatus(codes.Ok, "committed")
			span.SetAttributes(attribute.Bool("db.transaction.committed", true))
		}
		span.End()
	}
}
