// Package audit provides audit logging for security and compliance.
package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"time"
)

// Action represents an auditable action.
type Action string

// Predefined audit actions
const (
	// Authentication actions
	ActionLogin         Action = "auth.login"
	ActionLoginFailed   Action = "auth.login_failed"
	ActionLogout        Action = "auth.logout"
	ActionTokenRefresh  Action = "auth.token_refresh"
	ActionPasswordReset Action = "auth.password_reset"

	// User actions
	ActionUserCreate Action = "user.create"
	ActionUserUpdate Action = "user.update"
	ActionUserDelete Action = "user.delete"

	// Todo actions
	ActionTodoCreate Action = "todo.create"
	ActionTodoUpdate Action = "todo.update"
	ActionTodoDelete Action = "todo.delete"

	// Admin actions
	ActionAdminAccess    Action = "admin.access"
	ActionRoleChange     Action = "admin.role_change"
	ActionConfigChange   Action = "admin.config_change"
	ActionAuditLogExport Action = "admin.audit_export"
)

// Severity represents the severity level of an audit event.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// Event represents an audit log event.
type Event struct {
	ID           int64                  `json:"id,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	Action       Action                 `json:"action"`
	Severity     Severity               `json:"severity"`
	UserID       *int64                 `json:"user_id,omitempty"`
	UserEmail    string                 `json:"user_email,omitempty"`
	ResourceType string                 `json:"resource_type,omitempty"`
	ResourceID   *int64                 `json:"resource_id,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	IPAddress    net.IP                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	RequestID    string                 `json:"request_id,omitempty"`
	Success      bool                   `json:"success"`
	Error        string                 `json:"error,omitempty"`
}

// Logger is the interface for audit logging.
type Logger interface {
	Log(ctx context.Context, event *Event) error
	Query(ctx context.Context, filter *Filter) ([]*Event, error)
}

// Filter represents query filters for audit logs.
type Filter struct {
	UserID       *int64
	Action       *Action
	ResourceType *string
	StartTime    *time.Time
	EndTime      *time.Time
	Severity     *Severity
	Limit        int
	Offset       int
}

// SlogLogger implements Logger using slog.
type SlogLogger struct {
	logger *slog.Logger
	store  Store
}

// Store is the interface for persisting audit logs.
type Store interface {
	Save(ctx context.Context, event *Event) error
	Query(ctx context.Context, filter *Filter) ([]*Event, error)
}

// NewSlogLogger creates a new audit logger using slog.
func NewSlogLogger(logger *slog.Logger, store Store) *SlogLogger {
	return &SlogLogger{
		logger: logger.With(slog.String("component", "audit")),
		store:  store,
	}
}

// Log logs an audit event.
func (l *SlogLogger) Log(ctx context.Context, event *Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Log to slog
	attrs := []any{
		slog.String("action", string(event.Action)),
		slog.String("severity", string(event.Severity)),
		slog.Bool("success", event.Success),
	}

	if event.UserID != nil {
		attrs = append(attrs, slog.Int64("user_id", *event.UserID))
	}
	if event.UserEmail != "" {
		attrs = append(attrs, slog.String("user_email", event.UserEmail))
	}
	if event.ResourceType != "" {
		attrs = append(attrs, slog.String("resource_type", event.ResourceType))
	}
	if event.ResourceID != nil {
		attrs = append(attrs, slog.Int64("resource_id", *event.ResourceID))
	}
	if event.IPAddress != nil {
		attrs = append(attrs, slog.String("ip_address", event.IPAddress.String()))
	}
	if event.RequestID != "" {
		attrs = append(attrs, slog.String("request_id", event.RequestID))
	}
	if event.Error != "" {
		attrs = append(attrs, slog.String("error", event.Error))
	}
	if len(event.Details) > 0 {
		detailsJSON, _ := json.Marshal(event.Details)
		attrs = append(attrs, slog.String("details", string(detailsJSON)))
	}

	switch event.Severity {
	case SeverityCritical, SeverityError:
		l.logger.Error("audit event", attrs...)
	case SeverityWarning:
		l.logger.Warn("audit event", attrs...)
	default:
		l.logger.Info("audit event", attrs...)
	}

	// Persist to store if available
	if l.store != nil {
		if err := l.store.Save(ctx, event); err != nil {
			l.logger.Error("failed to persist audit event", slog.String("error", err.Error()))
			return err
		}
	}

	return nil
}

// Query queries audit logs.
func (l *SlogLogger) Query(ctx context.Context, filter *Filter) ([]*Event, error) {
	if l.store == nil {
		return nil, nil
	}
	return l.store.Query(ctx, filter)
}

// Builder helps construct audit events fluently.
type Builder struct {
	event *Event
}

// NewEvent creates a new audit event builder.
func NewEvent(action Action) *Builder {
	return &Builder{
		event: &Event{
			Timestamp: time.Now(),
			Action:    action,
			Severity:  SeverityInfo,
			Success:   true,
		},
	}
}

// WithUser sets the user information.
func (b *Builder) WithUser(userID int64, email string) *Builder {
	b.event.UserID = &userID
	b.event.UserEmail = email
	return b
}

// WithResource sets the resource information.
func (b *Builder) WithResource(resourceType string, resourceID int64) *Builder {
	b.event.ResourceType = resourceType
	b.event.ResourceID = &resourceID
	return b
}

// WithDetails sets additional details.
func (b *Builder) WithDetails(details map[string]interface{}) *Builder {
	b.event.Details = details
	return b
}

// WithRequest sets request information.
func (b *Builder) WithRequest(requestID, ipAddress, userAgent string) *Builder {
	b.event.RequestID = requestID
	b.event.UserAgent = userAgent
	if ip := net.ParseIP(ipAddress); ip != nil {
		b.event.IPAddress = ip
	}
	return b
}

// WithSeverity sets the severity level.
func (b *Builder) WithSeverity(severity Severity) *Builder {
	b.event.Severity = severity
	return b
}

// WithError marks the event as failed with an error.
func (b *Builder) WithError(err error) *Builder {
	b.event.Success = false
	if err != nil {
		b.event.Error = err.Error()
	}
	return b
}

// Failed marks the event as failed.
func (b *Builder) Failed() *Builder {
	b.event.Success = false
	return b
}

// Build returns the constructed event.
func (b *Builder) Build() *Event {
	return b.event
}

// Helper functions for common audit events

// LoginSuccess creates a successful login audit event.
func LoginSuccess(userID int64, email, ipAddress, userAgent, requestID string) *Event {
	return NewEvent(ActionLogin).
		WithUser(userID, email).
		WithRequest(requestID, ipAddress, userAgent).
		Build()
}

// LoginFailed creates a failed login audit event.
func LoginFailed(email, ipAddress, userAgent, requestID string, reason error) *Event {
	event := NewEvent(ActionLoginFailed).
		WithRequest(requestID, ipAddress, userAgent).
		WithSeverity(SeverityWarning).
		WithError(reason).
		Build()
	event.UserEmail = email
	return event
}

// ResourceCreated creates a resource creation audit event.
func ResourceCreated(action Action, userID int64, email, resourceType string, resourceID int64, requestID string) *Event {
	return NewEvent(action).
		WithUser(userID, email).
		WithResource(resourceType, resourceID).
		WithRequest(requestID, "", "").
		Build()
}

// ResourceDeleted creates a resource deletion audit event.
func ResourceDeleted(action Action, userID int64, email, resourceType string, resourceID int64, requestID string) *Event {
	return NewEvent(action).
		WithUser(userID, email).
		WithResource(resourceType, resourceID).
		WithRequest(requestID, "", "").
		WithSeverity(SeverityWarning).
		Build()
}
