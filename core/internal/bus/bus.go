package bus

import (
	"sync"
)

// InternalBus represents the internal event bus of the Core.
// It uses a Pub/Sub model where multiple channels can subscribe to events.
type InternalBus struct {
	mu          sync.RWMutex
	subscribers []chan Event
}

// NewInternalBus creates a new internal event bus.
func NewInternalBus() *InternalBus {
	return &InternalBus{
		subscribers: make([]chan Event, 0),
	}
}

// Subscribe adds a new subscriber and returns a channel that will receive events.
// The caller is responsible for consuming the channel.
func (b *InternalBus) Subscribe(bufferSize int) <-chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan Event, bufferSize)
	b.subscribers = append(b.subscribers, ch)
	return ch
}

// Publish broadcasts an event to all subscribers.
func (b *InternalBus) Publish(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.subscribers {
		// Non-blocking send
		select {
		case ch <- e:
		default:
			// Subscriber is too slow, event is dropped.
		}
	}
}
