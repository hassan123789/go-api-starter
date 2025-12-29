package grpc

import (
	"sync"

	todov1 "github.com/zareh/go-api-starter/gen/go/todo/v1"
)

// EventHub manages todo event subscriptions for streaming.
type EventHub struct {
	mu          sync.RWMutex
	subscribers map[chan *todov1.TodoEvent]struct{}
}

// NewEventHub creates a new EventHub.
func NewEventHub() *EventHub {
	return &EventHub{
		subscribers: make(map[chan *todov1.TodoEvent]struct{}),
	}
}

// Subscribe creates a new subscription channel.
func (h *EventHub) Subscribe() chan *todov1.TodoEvent {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan *todov1.TodoEvent, 100)
	h.subscribers[ch] = struct{}{}
	return ch
}

// Unsubscribe removes a subscription channel.
func (h *EventHub) Unsubscribe(ch chan *todov1.TodoEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.subscribers, ch)
	close(ch)
}

// Publish sends an event to all subscribers.
func (h *EventHub) Publish(event *todov1.TodoEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.subscribers {
		select {
		case ch <- event:
		default:
			// Drop message if subscriber is slow
		}
	}
}
