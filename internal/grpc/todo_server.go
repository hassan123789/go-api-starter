package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	todov1 "github.com/zareh/go-api-starter/gen/go/todo/v1"
	"github.com/zareh/go-api-starter/internal/model"
	"github.com/zareh/go-api-starter/internal/service"
)

// TodoServer implements the gRPC TodoService.
type TodoServer struct {
	todov1.UnimplementedTodoServiceServer
	todoService *service.TodoService
	eventHub    *EventHub
}

// NewTodoServer creates a new TodoServer.
func NewTodoServer(todoService *service.TodoService) *TodoServer {
	return &TodoServer{
		todoService: todoService,
		eventHub:    NewEventHub(),
	}
}

// CreateTodo creates a new todo.
func (s *TodoServer) CreateTodo(ctx context.Context, req *todov1.CreateTodoRequest) (*todov1.Todo, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	if len(req.Title) > 255 {
		return nil, status.Error(codes.InvalidArgument, "title must be at most 255 characters")
	}

	todo, err := s.todoService.Create(ctx, userID, req.Title)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create todo")
	}

	protoTodo := modelToProto(todo)

	// Publish event
	s.eventHub.Publish(&todov1.TodoEvent{
		Type:      todov1.EventType_EVENT_TYPE_CREATED,
		Todo:      protoTodo,
		Timestamp: timestamppb.Now(),
	})

	return protoTodo, nil
}

// GetTodo retrieves a todo by ID.
func (s *TodoServer) GetTodo(ctx context.Context, req *todov1.GetTodoRequest) (*todov1.Todo, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	todo, err := s.todoService.GetByID(ctx, req.Id, userID)
	if err != nil {
		if errors.Is(err, service.ErrTodoNotFound) {
			return nil, status.Error(codes.NotFound, "todo not found")
		}
		if errors.Is(err, service.ErrUnauthorized) {
			return nil, status.Error(codes.NotFound, "todo not found")
		}
		return nil, status.Error(codes.Internal, "failed to get todo")
	}

	return modelToProto(todo), nil
}

// ListTodos retrieves all todos for the authenticated user.
func (s *TodoServer) ListTodos(ctx context.Context, req *todov1.ListTodosRequest) (*todov1.ListTodosResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	todos, err := s.todoService.ListByUserID(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list todos")
	}

	// Apply filter if specified
	var filtered []model.Todo
	for _, t := range todos {
		if req.Completed != nil && t.Completed != req.Completed.Value {
			continue
		}
		filtered = append(filtered, t)
	}

	protoTodos := make([]*todov1.Todo, len(filtered))
	for i, t := range filtered {
		protoTodos[i] = modelToProto(&t)
	}

	return &todov1.ListTodosResponse{
		Todos: protoTodos,
		Total: int32(len(protoTodos)),
	}, nil
}

// UpdateTodo updates a todo.
func (s *TodoServer) UpdateTodo(ctx context.Context, req *todov1.UpdateTodoRequest) (*todov1.Todo, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	if req.Title != nil && len(req.Title.Value) > 255 {
		return nil, status.Error(codes.InvalidArgument, "title must be at most 255 characters")
	}

	updateReq := model.UpdateTodoRequest{}
	if req.Title != nil {
		updateReq.Title = &req.Title.Value
	}
	if req.Completed != nil {
		updateReq.Completed = &req.Completed.Value
	}

	todo, err := s.todoService.Update(ctx, req.Id, userID, updateReq)
	if err != nil {
		if errors.Is(err, service.ErrTodoNotFound) {
			return nil, status.Error(codes.NotFound, "todo not found")
		}
		if errors.Is(err, service.ErrUnauthorized) {
			return nil, status.Error(codes.NotFound, "todo not found")
		}
		return nil, status.Error(codes.Internal, "failed to update todo")
	}

	protoTodo := modelToProto(todo)

	// Publish event
	s.eventHub.Publish(&todov1.TodoEvent{
		Type:      todov1.EventType_EVENT_TYPE_UPDATED,
		Todo:      protoTodo,
		Timestamp: timestamppb.Now(),
	})

	return protoTodo, nil
}

// DeleteTodo removes a todo.
func (s *TodoServer) DeleteTodo(ctx context.Context, req *todov1.DeleteTodoRequest) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	// Get todo before deletion for event
	todo, err := s.todoService.GetByID(ctx, req.Id, userID)
	if err != nil {
		if errors.Is(err, service.ErrTodoNotFound) {
			return nil, status.Error(codes.NotFound, "todo not found")
		}
		if errors.Is(err, service.ErrUnauthorized) {
			return nil, status.Error(codes.NotFound, "todo not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete todo")
	}

	if err := s.todoService.Delete(ctx, req.Id, userID); err != nil {
		if errors.Is(err, service.ErrTodoNotFound) {
			return nil, status.Error(codes.NotFound, "todo not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete todo")
	}

	// Publish event
	s.eventHub.Publish(&todov1.TodoEvent{
		Type:      todov1.EventType_EVENT_TYPE_DELETED,
		Todo:      modelToProto(todo),
		Timestamp: timestamppb.Now(),
	})

	return &emptypb.Empty{}, nil
}

// StreamTodos streams todo events to the client.
func (s *TodoServer) StreamTodos(req *todov1.StreamTodosRequest, stream todov1.TodoService_StreamTodosServer) error {
	_, err := getUserIDFromContext(stream.Context())
	if err != nil {
		return status.Error(codes.Unauthenticated, "authentication required")
	}

	// Subscribe to events
	sub := s.eventHub.Subscribe()
	defer s.eventHub.Unsubscribe(sub)

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case event := <-sub:
			// Filter by event type if specified
			if len(req.EventTypes) > 0 {
				found := false
				for _, et := range req.EventTypes {
					if et == event.Type {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			if err := stream.Send(event); err != nil {
				return err
			}
		}
	}
}

// modelToProto converts a model.Todo to a proto Todo.
func modelToProto(todo *model.Todo) *todov1.Todo {
	return &todov1.Todo{
		Id:        todo.ID,
		UserId:    todo.UserID,
		Title:     todo.Title,
		Completed: todo.Completed,
		CreatedAt: timestamppb.New(todo.CreatedAt),
		UpdatedAt: timestamppb.New(todo.UpdatedAt),
	}
}

// protoToUpdateRequest converts proto update request to model.
func protoToUpdateRequest(req *todov1.UpdateTodoRequest) model.UpdateTodoRequest {
	result := model.UpdateTodoRequest{}
	if req.Title != nil {
		title := req.Title.Value
		result.Title = &title
	}
	if req.Completed != nil {
		completed := req.Completed.Value
		result.Completed = &completed
	}
	return result
}

// Ensure interface compatibility at compile time.
var _ = (*wrapperspb.StringValue)(nil) // Used for optional fields
