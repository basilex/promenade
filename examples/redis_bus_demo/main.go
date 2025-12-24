package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/pkg/bus"
	_ "github.com/basilex/promenade/pkg/bus/memory" // Register memory adapter
	_ "github.com/basilex/promenade/pkg/bus/redis"  // Register redis adapter
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// TestEvent is a simple event for testing
type TestEvent struct {
	*bus.BaseEvent
	Message string `json:"message"`
}

func main() {
	// Initialize logger
	logger.Init(logger.Config{
		Level:      "info",
		Format:     "text",
		AddSource:  false,
		TimeFormat: time.RFC3339,
	})
	log := logger.Default().Logger

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Error("Failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	log.Info("Testing Event IBus", slog.String("adapter", cfg.IBus.Adapter))

	// Create event bus using factory
	eventBus, err := bus.NewBus(cfg.IBus)
	if err != nil {
		log.Error("Failed to create event bus", slog.Any("error", err))
		os.Exit(1)
	}
	defer eventBus.Close(context.Background())

	// Check health
	ctx := context.Background()
	if err := eventBus.Health(ctx); err != nil {
		log.Error("IBus health check failed", slog.Any("error", err))
		os.Exit(1)
	}
	log.Info("[OK] IBus health check passed")

	// Subscribe to test topic
	messagesReceived := 0
	handler := func(ctx context.Context, event bus.Event) error {
		messagesReceived++
		log.Info("[<<] Received event",
			slog.String("type", event.Type()),
			slog.Int("count", messagesReceived))
		return nil
	}

	topic := "test.messages"
	if err := eventBus.Subscribe(topic, handler); err != nil {
		log.Error("Failed to subscribe", slog.Any("error", err))
		os.Exit(1)
	}
	log.Info("[OK] Subscribed to topic", slog.String("topic", topic))

	// Give subscriber time to register (important for Redis Pub/Sub)
	time.Sleep(500 * time.Millisecond)

	// Publish test events
	numMessages := 5
	log.Info("[>>] Publishing test events", slog.Int("count", numMessages))

	for i := 1; i <= numMessages; i++ {
		aggregateID := uuidv7.New()
		event := &TestEvent{
			BaseEvent: bus.NewBaseEvent("test.message", aggregateID),
			Message:   fmt.Sprintf("Test message #%d", i),
		}

		if err := eventBus.Publish(ctx, topic, event); err != nil {
			log.Error("Failed to publish event",
				slog.Int("message_num", i),
				slog.Any("error", err))
			continue
		}

		log.Info("[OK] Published event", slog.Int("num", i))
		time.Sleep(100 * time.Millisecond)
	}

	// Wait for messages to be processed
	log.Info("⏳ Waiting for messages to be processed...")
	time.Sleep(2 * time.Second)

	// Results
	log.Info("[RESULTS] Test Results",
		slog.Int("published", numMessages),
		slog.Int("received", messagesReceived))

	if messagesReceived == numMessages {
		log.Info("[SUCCESS] All messages received!")
	} else {
		log.Warn("[WARNING] Message count mismatch",
			slog.Int("expected", numMessages),
			slog.Int("actual", messagesReceived))
	}

	// Graceful shutdown
	log.Info("🛑 Shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := eventBus.Close(shutdownCtx); err != nil {
		log.Error("Failed to close bus", slog.Any("error", err))
	} else {
		log.Info("[OK] IBus closed gracefully")
	}
}
