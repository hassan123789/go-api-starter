package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/golang-jwt/jwt/v5"

	todov1 "github.com/zareh/go-api-starter/gen/go/todo/v1"
)

const testJWTSecret = "test-secret-key"

// generateTestToken creates a JWT token for testing.
func generateTestToken(userID int64) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(userID),
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(testJWTSecret))
	return tokenString
}

// contextWithAuth adds authorization metadata to context.
func contextWithAuth(ctx context.Context, token string) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"authorization", "Bearer "+token,
	))
}

func TestTodoServer_CreateTodo_Unauthenticated(t *testing.T) {
	// This test verifies that unauthenticated requests are rejected
	ctx := context.Background()

	// Create a mock server that doesn't have a real service
	// We're testing the auth interceptor behavior
	server := &TodoServer{
		eventHub: NewEventHub(),
	}

	// Without proper context setup, this should fail
	_, err := server.CreateTodo(ctx, &todov1.CreateTodoRequest{
		Title: "Test Todo",
	})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestTodoServer_CreateTodo_ValidationError(t *testing.T) {
	server := &TodoServer{
		eventHub: NewEventHub(),
	}

	// Set user ID in context
	ctx := setUserIDInContext(context.Background(), 1)

	tests := []struct {
		name    string
		title   string
		wantErr codes.Code
	}{
		{
			name:    "empty title",
			title:   "",
			wantErr: codes.InvalidArgument,
		},
		{
			name:    "title too long",
			title:   string(make([]byte, 256)),
			wantErr: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := server.CreateTodo(ctx, &todov1.CreateTodoRequest{
				Title: tt.title,
			})

			require.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, tt.wantErr, st.Code())
		})
	}
}

func TestEventHub(t *testing.T) {
	hub := NewEventHub()

	// Subscribe
	sub1 := hub.Subscribe()
	sub2 := hub.Subscribe()

	// Publish event
	event := &todov1.TodoEvent{
		Type: todov1.EventType_EVENT_TYPE_CREATED,
		Todo: &todov1.Todo{
			Id:    1,
			Title: "Test",
		},
	}

	hub.Publish(event)

	// Both subscribers should receive the event
	select {
	case received := <-sub1:
		assert.Equal(t, event.Type, received.Type)
		assert.Equal(t, event.Todo.Id, received.Todo.Id)
	case <-time.After(time.Second):
		t.Fatal("subscriber 1 did not receive event")
	}

	select {
	case received := <-sub2:
		assert.Equal(t, event.Type, received.Type)
	case <-time.After(time.Second):
		t.Fatal("subscriber 2 did not receive event")
	}

	// Unsubscribe
	hub.Unsubscribe(sub1)

	// Publish another event
	hub.Publish(event)

	// Only sub2 should receive
	select {
	case <-sub2:
		// OK
	case <-time.After(time.Second):
		t.Fatal("subscriber 2 did not receive event after unsubscribe")
	}
}

func TestContextHelpers(t *testing.T) {
	t.Run("getUserIDFromContext success", func(t *testing.T) {
		ctx := setUserIDInContext(context.Background(), 123)
		userID, err := getUserIDFromContext(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(123), userID)
	})

	t.Run("getUserIDFromContext failure", func(t *testing.T) {
		ctx := context.Background()
		_, err := getUserIDFromContext(ctx)
		require.Error(t, err)
	})

	t.Run("getTokenFromMetadata success", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			"authorization", "Bearer test-token",
		))
		token, err := getTokenFromMetadata(ctx)
		require.NoError(t, err)
		assert.Equal(t, "test-token", token)
	})

	t.Run("getTokenFromMetadata missing metadata", func(t *testing.T) {
		ctx := context.Background()
		_, err := getTokenFromMetadata(ctx)
		require.Error(t, err)
	})

	t.Run("getTokenFromMetadata invalid format", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
			"authorization", "InvalidFormat",
		))
		_, err := getTokenFromMetadata(ctx)
		require.Error(t, err)
	})
}

// Integration test helper - requires running server
func TestIntegration_TodoService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// This would be used for full integration tests with a running server
	t.Skip("integration tests require running server")

	conn, err := grpc.NewClient("localhost:9090",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Logf("failed to close connection: %v", closeErr)
		}
	}()

	client := todov1.NewTodoServiceClient(conn)

	ctx := contextWithAuth(context.Background(), generateTestToken(1))

	// Create
	created, err := client.CreateTodo(ctx, &todov1.CreateTodoRequest{
		Title: "Test Todo",
	})
	require.NoError(t, err)
	assert.Equal(t, "Test Todo", created.Title)

	// Get
	got, err := client.GetTodo(ctx, &todov1.GetTodoRequest{
		Id: created.Id,
	})
	require.NoError(t, err)
	assert.Equal(t, created.Id, got.Id)

	// List
	list, err := client.ListTodos(ctx, &todov1.ListTodosRequest{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, list.Total, int32(1))

	// Delete
	_, err = client.DeleteTodo(ctx, &todov1.DeleteTodoRequest{
		Id: created.Id,
	})
	require.NoError(t, err)
}
