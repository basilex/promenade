package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/basilex/promenade/internal/domain/event"
	"github.com/basilex/promenade/internal/infrastructure/notification"
	"github.com/basilex/promenade/pkg/bus/memory"
	"github.com/google/uuid"
)

func main() {
	fmt.Println("🚀 Event Bus Demo - Async Email Notifications")
	fmt.Println("================================================")

	// Initialize event bus
	eventBus := memory.NewDefaultMemoryBus()
	defer eventBus.Close(context.Background())

	// Initialize email service
	emailSender := notification.NewMockEmailSender()
	emailService, err := notification.NewEmailService(eventBus, emailSender, "templates/email")
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
	userID := uuid.New()
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

	// Bus stats
	stats := eventBus.Stats()
	fmt.Printf("\nEvent Bus Stats:\n")
	fmt.Printf("  Topics: %d\n", stats["total_topics"])
	fmt.Printf("  Subscribers: %d\n", stats["total_subscribers"])
	fmt.Printf("  Messages Published: %d\n", stats["messages_published"])
	fmt.Printf("  Messages Processed: %d\n", stats["messages_processed"])

	fmt.Println("\nKey Takeaways:")
	fmt.Println("   • Publish() returns immediately - non-blocking")
	fmt.Println("   • Email sent asynchronously in worker pool")
	fmt.Println("   • User doesn't wait for email delivery")
	fmt.Println("   • Same pattern works for Redis, NATS, Kafka")
	fmt.Println("   • Easy evolution from monolith to microservices")
}
