package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

type EventPublisher struct {
	bridge   EventBridge
	source   string
	mu       sync.RWMutex
}

func NewEventPublisher(bridge EventBridge, source string) *EventPublisher {
	return &EventPublisher{
		bridge: bridge,
		source: source,
	}
}

func (p *EventPublisher) Publish(ctx context.Context, subject string, eventType string, data map[string]interface{}) error {
	event := NewEvent(eventType, p.source, subject, data)

	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.bridge.Publish(subject, event)
}

func (p *EventPublisher) PublishAsync(ctx context.Context, subject string, eventType string, data map[string]interface{}) {
	go func() {
		if err := p.Publish(ctx, subject, eventType, data); err != nil {
			log.Printf("Failed to publish event %s: %v", eventType, err)
		}
	}()
}

type EventSubscriber struct {
	bridge    EventBridge
	source    string
	handlers  map[string]EventHandler
	mu        sync.RWMutex
	closed    bool
}

func NewEventSubscriber(bridge EventBridge, source string) *EventSubscriber {
	return &EventSubscriber{
		bridge:   bridge,
		source:   source,
		handlers: make(map[string]EventHandler),
	}
}

func (s *EventSubscriber) Subscribe(ctx context.Context, subject string, handler EventHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("subscriber is closed")
	}

	wrappedHandler := func(event *Event) error {
		if event.Source == s.source {
			return nil
		}
		return handler(event)
	}

	s.handlers[subject] = wrappedHandler
	return s.bridge.Subscribe(subject, wrappedHandler)
}

func (s *EventSubscriber) SubscribeMultiple(ctx context.Context, handlers map[string]EventHandler) error {
	for subject, handler := range handlers {
		if err := s.Subscribe(ctx, subject, handler); err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
		}
	}
	return nil
}

func (s *EventSubscriber) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.handlers = make(map[string]EventHandler)
}

// EventProcessor handles event processing with deduplication and retry
type EventProcessor struct {
	publisher  *EventPublisher
	subscriber *EventSubscriber
	processed  map[string]time.Time
	mu         sync.RWMutex
	ttl        time.Duration
}

func NewEventProcessor(publisher *EventPublisher, subscriber *EventSubscriber) *EventProcessor {
	return &EventProcessor{
		publisher:  publisher,
		subscriber: subscriber,
		processed:  make(map[string]time.Time),
		ttl:        24 * time.Hour,
	}
}

func (ep *EventProcessor) Process(ctx context.Context, subject string, handler EventHandler) error {
	return ep.subscriber.Subscribe(ctx, subject, func(event *Event) error {
		if ep.isDuplicate(event.ID) {
			return nil
		}

		if err := handler(event); err != nil {
			ep.publishDeadLetter(ctx, event, err)
			return err
		}

		ep.markProcessed(event.ID)
		return nil
	})
}

func (ep *EventProcessor) isDuplicate(eventID string) bool {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	if _, exists := ep.processed[eventID]; exists {
		return true
	}

	for id, t := range ep.processed {
		if time.Since(t) > ep.ttl {
			delete(ep.processed, id)
		}
	}

	return false
}

func (ep *EventProcessor) markProcessed(eventID string) {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	ep.processed[eventID] = time.Now()
}

func (ep *EventProcessor) publishDeadLetter(ctx context.Context, event *Event, err error) {
	deadLetterData := map[string]interface{}{
		"original_event": event.Data,
		"error":          err.Error(),
		"event_type":     event.Type,
		"source":         event.Source,
	}

	deadLetterJSON, _ := json.Marshal(deadLetterData)
	_ = deadLetterJSON

	ep.publisher.PublishAsync(ctx, SubjectEvent+".dead_letter", "event.dead_letter", deadLetterData)
}
