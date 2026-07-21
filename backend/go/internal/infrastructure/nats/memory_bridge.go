package nats

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// InMemoryBridge is a non-NATS implementation for development/testing
// Replace with real NATS client in production
type InMemoryBridge struct {
	subscriptions map[string][]EventHandler
	mu            sync.RWMutex
	closed        bool
}

func NewInMemoryBridge() *InMemoryBridge {
	return &InMemoryBridge{
		subscriptions: make(map[string][]EventHandler),
	}
}

func (b *InMemoryBridge) Publish(subject string, event *Event) error {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return fmt.Errorf("bridge is closed")
	}
	handlers := b.subscriptions[subject]
	b.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(event); err != nil {
			log.Printf("Event handler error for subject %s: %v", subject, err)
		}
	}

	// Also publish to wildcard subscribers
	b.mu.RLock()
	wildcardHandlers := b.subscriptions[subject+".*"]
	b.mu.RUnlock()

	for _, handler := range wildcardHandlers {
		if err := handler(event); err != nil {
			log.Printf("Wildcard handler error for subject %s: %v", subject, err)
		}
	}

	return nil
}

func (b *InMemoryBridge) Subscribe(subject string, handler EventHandler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return fmt.Errorf("bridge is closed")
	}

	b.subscriptions[subject] = append(b.subscriptions[subject], handler)
	return nil
}

func (b *InMemoryBridge) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	b.subscriptions = make(map[string][]EventHandler)
}

// PublishWithContext publishes with context support
func (b *InMemoryBridge) PublishWithContext(ctx context.Context, subject string, event *Event) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return b.Publish(subject, event)
	}
}

// SubscribeWithTimeout subscribes with automatic unsubscribe after timeout
func (b *InMemoryBridge) SubscribeWithTimeout(subject string, handler EventHandler, timeout time.Duration) error {
	if err := b.Subscribe(subject, handler); err != nil {
		return err
	}

	go func() {
		time.Sleep(timeout)
		b.mu.Lock()
		defer b.mu.Unlock()
		if handlers, ok := b.subscriptions[subject]; ok {
			if len(handlers) > 1 {
				b.subscriptions[subject] = handlers[1:]
			} else {
				delete(b.subscriptions, subject)
			}
		}
	}()

	return nil
}
