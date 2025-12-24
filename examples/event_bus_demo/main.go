package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/basilex/promenade/internal/domain/event"
	"github.com/basilex/promenade/internal/infrastructure/notification"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/bus/memory"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func main() {
	fmt.Println("* Event IBus Demo - Async Email Notifications")
	fmt.Println("================================================")

	// Create memory bus with config
	eventBus := memory.NewMemoryBus(bus.BusConfig{
		WorkerPoolSize: 4,
		BufferSize:     100,
	})
	defer eventBus.Close(context.Background())

	// Initialize email service
	emailSender := notification.NewMockEmailSender()
	emailService, err := notification.NewEmailService(
		eventBus,
		emailSender,
		"templates/email",
		"noreply@promenade.com",
		"Promenade Demo",
		"http://localhost:8081",
		"Promenade Demo App",
	)
	if err != nil {
		log.Fatalf("Failed to create email service: %v", err)
	}

	// Start email service (subscribes to events)
	ctx := context.Background()
	if err := emailService.Start(ctx); err != nil {
		log.Fatalf("Failed to start email service: %v", err)
	}

	fmt.Println("\nEmail service started and listening for events...")

	// Simulate user registration
	userID := uuidv7.New()
	email := "demo@example.com"
	name := "Demo User"

	fmt.Printf("\nSimulating user registration: %s (%s)\n", name, email)

	// Publish user registered event
	userEvent := event.NewUserRegisteredEvent(userID, email, name)
	if err := eventBus.Publish(ctx, userEvent.Type(), userEvent); err != nil {
		log.Fatalf("Failed to publish event: %v", err)
	}

	fmt.Println("Event published to bus (returns immediately)")
	fmt.Println("Email being sent in background goroutine...")

	// Wait for email to be processed
	time.Sleep(200 * time.Millisecond)

	// Check sent emails
	sentEmails := emailSender.GetSentEmails()
	fmt.Printf("\nEmails sent: %d\n", len(sentEmails))
	if len(sentEmails) > 0 {
		fmt.Println("\nEmail details:")
		for i, e := range sentEmails {
			fmt.Printf("  %d. To: %s\n", i+1, e.To)
			fmt.Printf("     Subject: %s\n", e.Subject)
		}
	}

	// IBus stats
	stats := eventBus.Stats()
	fmt.Printf("\nEvent IBus Stats:\n")
	fmt.Printf("  Topics: %d\n", stats["total_topics"])
	fmt.Printf("  Subscribers: %d\n", stats["total_subscribers"])
	fmt.Printf("  Worker Pool Size: %d\n", stats["worker_pool_size"])

	fmt.Println("\nKey Takeaways:")
	fmt.Println("   • Publish() returns immediately - non-blocking")
	fmt.Println("   • Email sent asynchronously in worker pool")
	fmt.Println("   • User doesn't wait for email delivery")
	fmt.Println("   • Same pattern works for Redis, NATS, Kafka")
	fmt.Println("   • Easy evolution from monolith to microservices")
}
