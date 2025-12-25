package mocks

import (
	"context"

	"github.com/basilex/promenade/pkg/bus"
)

type MockEventBus struct {
	publishedEvents []bus.Event
	publishErr      error
}

func NewMockEventBus() *MockEventBus {
	return &MockEventBus{
		publishedEvents: make([]bus.Event, 0),
	}
}

func (m *MockEventBus) Publish(ctx context.Context, topic string, event bus.Event) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.publishedEvents = append(m.publishedEvents, event)
	return nil
}

func (m *MockEventBus) Subscribe(topic string, handler bus.Handler) error {
	return nil
}

func (m *MockEventBus) Unsubscribe(topic string, handler bus.Handler) error {
	return nil
}

func (m *MockEventBus) Close(ctx context.Context) error {
	return nil
}

func (m *MockEventBus) Health(ctx context.Context) error {
	return nil
}

func (m *MockEventBus) GetPublishedEventCount() int {
	return len(m.publishedEvents)
}

func (m *MockEventBus) Reset() {
	m.publishedEvents = make([]bus.Event, 0)
	m.publishErr = nil
}
