// Package ctxutil provides context utilities for request-scoped data.
//
// # Overview
//
// This package provides type-safe context value management for:
//   - Request IDs for distributed tracing
//   - User IDs for authentication
//   - Correlation data for logging
//
// # Basic Usage
//
//	// Setting values
//	ctx = ctxutil.WithRequestID(ctx, "req-123")
//	ctx = ctxutil.WithUserID(ctx, 456)
//
//	// Getting values
//	requestID := ctxutil.GetRequestID(ctx)
//	userID := ctxutil.GetUserID(ctx)
//
// # Middleware Integration
//
//	func RequestIDMiddleware() echo.MiddlewareFunc {
//	    return func(next echo.HandlerFunc) echo.HandlerFunc {
//	        return func(c echo.Context) error {
//	            requestID := c.Request().Header.Get("X-Request-ID")
//	            if requestID == "" {
//	                requestID = uuid.New().String()
//	            }
//	            ctx := ctxutil.WithRequestID(c.Request().Context(), requestID)
//	            c.SetRequest(c.Request().WithContext(ctx))
//	            return next(c)
//	        }
//	    }
//	}
//
// # Logging Integration
//
//	func logWithContext(ctx context.Context, msg string) {
//	    logger.Info(msg,
//	        "request_id", ctxutil.GetRequestID(ctx),
//	        "user_id", ctxutil.GetUserID(ctx),
//	    )
//	}
package ctxutil
