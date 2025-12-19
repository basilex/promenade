package bus_test

import (
	"context"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/domain/event"
	"github.com/basilex/promenade/internal/infrastructure/notification"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/bus/memory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventBusIntegration demonstrates the full event-driven flow:
// 1. User registers
// 2. Event is published to bus
// 3. Email service picks up the event
// 4. Welcome email is sent asynchronously
func TestEventBusIntegration(t *testing.T) {
	// Initialize in-memory event bus
	eventBus := memory.NewDefaultMemoryBus()
	defer eventBus.Close(context.Background())

	// Initialize mock email sender
	emailSender := notification.NewMockEmailSender()
	emailService, err := notification.NewEmailService(
		eventBus,
		emailSender,
		"", // templates path
		"noreply@test.com",
		"Test Team",
		"http://localhost:8081",
		"TestApp",
	)
	require.NoError(t, err)

	// Start email service (subscribes to events)
	err = emailService.Start(context.Background())
	require.NoError(t, err)

	// Simulate user registration event
	userID := uuid.New()
	email := "john.doe@example.com"
	name := "John Doe"

	userRegisteredEvent := event.NewUserRegisteredEvent(userID, email, name)

	// Publish event (this is what happens in AuthUseCase.Register)
	err = eventBus.Publish(context.Background(), bus.TopicUserRegistered, userRegisteredEvent)
	require.NoError(t, err)

	// Wait for async processing (in real app, this happens in background)
	time.Sleep(200 * time.Millisecond)

	// Verify email was sent
	sentEmails := emailSender.GetSentEmails()
	assert.Len(t, sentEmails, 1, "should have sent 1 email")

	if len(sentEmails) > 0 {
		assert.Equal(t, email, sentEmails[0].To)
		assert.Equal(t, "Welcome to TestApp!", sentEmails[0].Subject)
		assert.Contains(t, sentEmails[0].HTML, name)
	}
}

// TestMultipleEvents demonstrates handling multiple user events
func TestMultipleEvents(t *testing.T) {
	eventBus := memory.NewDefaultMemoryBus()
	defer eventBus.Close(context.Background())

	emailSender := notification.NewMockEmailSender()
	emailService, err := notification.NewEmailService(
		eventBus,
		emailSender,
		"", // templates path
		"noreply@test.com",
		"Test Team",
		"http://localhost:8081",
		"TestApp",
	)
	require.NoError(t, err)
	require.NoError(t, emailService.Start(context.Background()))

	userID := uuid.New()
	email := "test@example.com"

	// Simulate multiple events
	events := []bus.Event{
		event.NewUserRegisteredEvent(userID, email, "Test User"),
		event.NewUserEmailVerifiedEvent(userID, email),
		event.NewUserPasswordChangedEvent(userID, email),
	}

	ctx := context.Background()
	for _, e := range events {
		err := eventBus.Publish(ctx, e.Type(), e)
		require.NoError(t, err)
	}

	// Wait for async processing
	time.Sleep(300 * time.Millisecond)

	// Verify all emails were sent
	sentEmails := emailSender.GetSentEmails()
	assert.Len(t, sentEmails, 3, "should have sent 3 emails")

	expectedSubjects := []string{
		"Welcome to TestApp!",
		"TestApp - Email Verified Successfully",
		"TestApp - Password Changed - Security Alert",
	}

	// Check that all expected subjects are present (order may vary due to async processing)
	actualSubjects := make([]string, len(sentEmails))
	for i, email := range sentEmails {
		actualSubjects[i] = email.Subject
	}

	for _, expectedSubject := range expectedSubjects {
		assert.Contains(t, actualSubjects, expectedSubject, "Expected subject not found: %s", expectedSubject)
	}
}

// TestBusPerformance demonstrates that event publishing doesn't block
func TestBusPerformance(t *testing.T) {
	eventBus := memory.NewDefaultMemoryBus()
	defer eventBus.Close(context.Background())

	emailSender := notification.NewMockEmailSender()
	emailSender.Delay = 500 * time.Millisecond // Simulate slow email sending

	emailService, err := notification.NewEmailService(
		eventBus,
		emailSender,
		"", // templates path
		"noreply@test.com",
		"Test Team",
		"http://localhost:8081",
		"TestApp",
	)
	require.NoError(t, err)
	require.NoError(t, emailService.Start(context.Background()))

	// Measure time to publish event
	start := time.Now()

	userEvent := event.NewUserRegisteredEvent(uuid.New(), "fast@example.com", "Fast User")
	err = eventBus.Publish(context.Background(), bus.TopicUserRegistered, userEvent)
	require.NoError(t, err)

	elapsed := time.Since(start)

	// Publishing should be fast (< 50ms), even though email sending takes 500ms
	assert.Less(t, elapsed, 50*time.Millisecond, "event publishing should not block")

	// Wait for email to be sent
	time.Sleep(600 * time.Millisecond)

	// Verify email was eventually sent
	assert.Len(t, emailSender.GetSentEmails(), 1)
}
