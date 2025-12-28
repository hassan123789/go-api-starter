package audit_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zareh/go-api-starter/pkg/audit"
)

// MockStore is a mock implementation of audit.Store for testing.
type MockStore struct {
	events []*audit.Event
}

func (m *MockStore) Save(_ context.Context, event *audit.Event) error {
	m.events = append(m.events, event)
	return nil
}

func (m *MockStore) Query(_ context.Context, _ *audit.Filter) ([]*audit.Event, error) {
	return m.events, nil
}

func TestNewEvent_Builder(t *testing.T) {
	event := audit.NewEvent(audit.ActionLogin).
		WithUser(1, "test@example.com").
		WithRequest("req-123", "192.168.1.1", "Mozilla/5.0").
		WithDetails(map[string]interface{}{"method": "password"}).
		Build()

	assert.Equal(t, audit.ActionLogin, event.Action)
	assert.Equal(t, int64(1), *event.UserID)
	assert.Equal(t, "test@example.com", event.UserEmail)
	assert.Equal(t, "req-123", event.RequestID)
	assert.Equal(t, "192.168.1.1", event.IPAddress.String())
	assert.Equal(t, "Mozilla/5.0", event.UserAgent)
	assert.True(t, event.Success)
	assert.Equal(t, audit.SeverityInfo, event.Severity)
}

func TestNewEvent_WithResource(t *testing.T) {
	event := audit.NewEvent(audit.ActionTodoCreate).
		WithUser(1, "test@example.com").
		WithResource("todo", 42).
		Build()

	assert.Equal(t, "todo", event.ResourceType)
	assert.Equal(t, int64(42), *event.ResourceID)
}

func TestNewEvent_WithError(t *testing.T) {
	event := audit.NewEvent(audit.ActionLoginFailed).
		WithUser(1, "test@example.com").
		WithError(assert.AnError).
		Build()

	assert.False(t, event.Success)
	assert.Contains(t, event.Error, "assert.AnError")
}

func TestNewEvent_Failed(t *testing.T) {
	event := audit.NewEvent(audit.ActionLogin).
		Failed().
		Build()

	assert.False(t, event.Success)
}

func TestLoginSuccess(t *testing.T) {
	event := audit.LoginSuccess(1, "test@example.com", "10.0.0.1", "Chrome", "req-456")

	assert.Equal(t, audit.ActionLogin, event.Action)
	assert.Equal(t, int64(1), *event.UserID)
	assert.Equal(t, "test@example.com", event.UserEmail)
	assert.Equal(t, "10.0.0.1", event.IPAddress.String())
	assert.True(t, event.Success)
}

func TestLoginFailed(t *testing.T) {
	event := audit.LoginFailed("test@example.com", "10.0.0.1", "Chrome", "req-789", assert.AnError)

	assert.Equal(t, audit.ActionLoginFailed, event.Action)
	assert.Equal(t, "test@example.com", event.UserEmail)
	assert.False(t, event.Success)
	assert.Equal(t, audit.SeverityWarning, event.Severity)
}

func TestResourceCreated(t *testing.T) {
	event := audit.ResourceCreated(audit.ActionTodoCreate, 1, "test@example.com", "todo", 42, "req-123")

	assert.Equal(t, audit.ActionTodoCreate, event.Action)
	assert.Equal(t, "todo", event.ResourceType)
	assert.Equal(t, int64(42), *event.ResourceID)
	assert.True(t, event.Success)
}

func TestResourceDeleted(t *testing.T) {
	event := audit.ResourceDeleted(audit.ActionTodoDelete, 1, "test@example.com", "todo", 42, "req-123")

	assert.Equal(t, audit.ActionTodoDelete, event.Action)
	assert.Equal(t, audit.SeverityWarning, event.Severity)
}

func TestSlogLogger_Log(t *testing.T) {
	store := &MockStore{}
	logger := audit.NewSlogLogger(slog.New(slog.NewTextHandler(os.Stdout, nil)), store)

	event := audit.NewEvent(audit.ActionLogin).
		WithUser(1, "test@example.com").
		Build()

	err := logger.Log(context.Background(), event)
	require.NoError(t, err)

	assert.Len(t, store.events, 1)
	assert.Equal(t, event, store.events[0])
}

func TestSlogLogger_Query(t *testing.T) {
	store := &MockStore{}
	logger := audit.NewSlogLogger(slog.New(slog.NewTextHandler(os.Stdout, nil)), store)

	// Add some events
	for i := 0; i < 3; i++ {
		event := audit.NewEvent(audit.ActionLogin).
			WithUser(int64(i+1), "test@example.com").
			Build()
		err := logger.Log(context.Background(), event)
		require.NoError(t, err)
	}

	// Query
	events, err := logger.Query(context.Background(), &audit.Filter{})
	require.NoError(t, err)

	assert.Len(t, events, 3)
}

func TestEventTimestamp(t *testing.T) {
	event := audit.NewEvent(audit.ActionLogin).Build()

	assert.False(t, event.Timestamp.IsZero())
}
