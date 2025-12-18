package notification

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path/filepath"
	"time"

	"github.com/basilex/promenade/internal/domain/event"
	"github.com/basilex/promenade/pkg/bus"
)

// EmailService handles email notifications based on domain events.
// It subscribes to user-related events and sends appropriate emails.
type EmailService struct {
	bus           bus.Bus
	sender        EmailSender
	templates     *template.Template
	templatesPath string
}

// EmailSender is an interface for actual email sending implementation.
// This allows swapping between real SMTP, SendGrid, AWS SES, or mock for testing.
type EmailSender interface {
	Send(ctx context.Context, email Email) error
}

// Email represents an email message.
type Email struct {
	To      string
	Subject string
	Body    string
	HTML    string // Optional HTML body
}

// NewEmailService creates a new email notification service.
// templatesPath should point to the directory containing email templates.
// If templatesPath is empty, uses fallback templates (useful for testing).
func NewEmailService(eventBus bus.Bus, sender EmailSender, templatesPath string) (*EmailService, error) {
	var templates *template.Template

	// Load templates if path provided
	if templatesPath != "" {
		tmplPattern := filepath.Join(templatesPath, "*.html")
		tmpl, err := template.ParseGlob(tmplPattern)
		if err != nil {
			return nil, fmt.Errorf("failed to load email templates from %s: %w", tmplPattern, err)
		}
		templates = tmpl
	}

	return &EmailService{
		bus:           eventBus,
		sender:        sender,
		templates:     templates,
		templatesPath: templatesPath,
	}, nil
}

// renderTemplate renders an email template with the given data.
func (s *EmailService) renderTemplate(templateName string, data interface{}) (string, error) {
	// If templates not loaded (e.g., in tests), return simple fallback
	if s.templates == nil {
		return s.renderFallbackTemplate(templateName, data), nil
	}

	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, templateName, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", templateName, err)
	}
	return buf.String(), nil
}

// renderFallbackTemplate provides simple HTML for tests when templates aren't loaded.
func (s *EmailService) renderFallbackTemplate(templateName string, data interface{}) string {
	d, _ := data.(map[string]interface{})

	switch templateName {
	case "welcome.html":
		if name, ok := d["Name"].(string); ok {
			return fmt.Sprintf("<html><body><h1>Welcome %s!</h1></body></html>", name)
		}
	case "email_verified.html":
		return "<html><body><h1>Email Verified!</h1></body></html>"
	case "password_changed.html":
		return "<html><body><h1>Password Changed</h1></body></html>"
	case "account_suspended.html":
		if reason, ok := d["Reason"].(string); ok {
			return fmt.Sprintf("<html><body><h1>Suspended</h1><p>%s</p></body></html>", reason)
		}
	case "account_banned.html":
		if reason, ok := d["Reason"].(string); ok {
			return fmt.Sprintf("<html><body><h1>Banned</h1><p>%s</p></body></html>", reason)
		}
	}
	return "<html><body>Email notification</body></html>"
}

// Start subscribes to relevant events and starts processing.
func (s *EmailService) Start(ctx context.Context) error {
	// Subscribe to user events
	if err := s.bus.Subscribe(bus.TopicUserRegistered, s.handleUserRegistered); err != nil {
		return fmt.Errorf("failed to subscribe to user.registered: %w", err)
	}

	if err := s.bus.Subscribe(bus.TopicUserEmailVerified, s.handleUserEmailVerified); err != nil {
		return fmt.Errorf("failed to subscribe to user.email.verified: %w", err)
	}

	if err := s.bus.Subscribe(bus.TopicUserPasswordChanged, s.handleUserPasswordChanged); err != nil {
		return fmt.Errorf("failed to subscribe to user.password.changed: %w", err)
	}

	if err := s.bus.Subscribe(bus.TopicUserSuspended, s.handleUserSuspended); err != nil {
		return fmt.Errorf("failed to subscribe to user.suspended: %w", err)
	}

	if err := s.bus.Subscribe(bus.TopicUserBanned, s.handleUserBanned); err != nil {
		return fmt.Errorf("failed to subscribe to user.banned: %w", err)
	}

	return nil
}

// handleUserRegistered sends a welcome email to newly registered users.
func (s *EmailService) handleUserRegistered(ctx context.Context, e bus.Event) error {
	evt, ok := e.(*event.UserRegisteredEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", e)
	}

	data := map[string]interface{}{
		"Name":     evt.Name,
		"Email":    evt.Email,
		"UserID":   evt.UserID.String(),
		"LoginURL": "https://promenade.app/login", // TODO: make configurable
		"Year":     time.Now().Year(),
	}

	html, err := s.renderTemplate("welcome.html", data)
	if err != nil {
		return fmt.Errorf("failed to render welcome template: %w", err)
	}

	email := Email{
		To:      evt.Email,
		Subject: "Welcome to Promenade!",
		HTML:    html,
	}

	return s.sender.Send(ctx, email)
}

// handleUserEmailVerified sends a confirmation email after email verification.
func (s *EmailService) handleUserEmailVerified(ctx context.Context, e bus.Event) error {
	evt, ok := e.(*event.UserEmailVerifiedEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", e)
	}

	data := map[string]interface{}{
		"Email":  evt.Email,
		"UserID": evt.UserID.String(),
		"Year":   time.Now().Year(),
	}

	html, err := s.renderTemplate("email_verified.html", data)
	if err != nil {
		return fmt.Errorf("failed to render email_verified template: %w", err)
	}

	email := Email{
		To:      evt.Email,
		Subject: "Email Verified Successfully",
		HTML:    html,
	}

	return s.sender.Send(ctx, email)
}

// handleUserPasswordChanged sends a security notification.
func (s *EmailService) handleUserPasswordChanged(ctx context.Context, e bus.Event) error {
	evt, ok := e.(*event.UserPasswordChangedEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", e)
	}

	data := map[string]interface{}{
		"Email":     evt.Email,
		"UserID":    evt.UserID.String(),
		"Timestamp": time.Now().Format("January 2, 2006 at 3:04 PM MST"),
		"Year":      time.Now().Year(),
	}

	html, err := s.renderTemplate("password_changed.html", data)
	if err != nil {
		return fmt.Errorf("failed to render password_changed template: %w", err)
	}

	email := Email{
		To:      evt.Email,
		Subject: "Password Changed - Security Alert",
		HTML:    html,
	}

	return s.sender.Send(ctx, email)
}

// handleUserSuspended notifies user about suspension.
func (s *EmailService) handleUserSuspended(ctx context.Context, e bus.Event) error {
	evt, ok := e.(*event.UserSuspendedEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", e)
	}

	var expiresAt string
	if evt.ExpiresAt != nil {
		expiresAt = evt.ExpiresAt.Format("January 2, 2006 at 3:04 PM MST")
	}

	data := map[string]interface{}{
		"Email":     evt.Email,
		"UserID":    evt.UserID.String(),
		"Reason":    evt.Reason,
		"ExpiresAt": expiresAt,
		"Year":      time.Now().Year(),
	}

	html, err := s.renderTemplate("account_suspended.html", data)
	if err != nil {
		return fmt.Errorf("failed to render account_suspended template: %w", err)
	}

	email := Email{
		To:      evt.Email,
		Subject: "Account Suspended - Action Required",
		HTML:    html,
	}

	return s.sender.Send(ctx, email)
}

// handleUserBanned notifies user about permanent ban.
func (s *EmailService) handleUserBanned(ctx context.Context, e bus.Event) error {
	evt, ok := e.(*event.UserBannedEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", e)
	}

	data := map[string]interface{}{
		"Email":  evt.Email,
		"UserID": evt.UserID.String(),
		"Reason": evt.Reason,
		"Year":   time.Now().Year(),
	}

	html, err := s.renderTemplate("account_banned.html", data)
	if err != nil {
		return fmt.Errorf("failed to render account_banned template: %w", err)
	}

	email := Email{
		To:      evt.Email,
		Subject: "Account Permanently Banned",
		HTML:    html,
	}

	return s.sender.Send(ctx, email)
}

// MockEmailSender is a simple mock for testing and development.
type MockEmailSender struct {
	SentEmails []Email
	Delay      time.Duration // Simulate network delay
}

// NewMockEmailSender creates a mock email sender.
func NewMockEmailSender() *MockEmailSender {
	return &MockEmailSender{
		SentEmails: make([]Email, 0),
		Delay:      100 * time.Millisecond,
	}
}

// Send implements EmailSender interface.
func (m *MockEmailSender) Send(ctx context.Context, email Email) error {
	// Simulate network delay
	time.Sleep(m.Delay)

	m.SentEmails = append(m.SentEmails, email)
	fmt.Printf("[MOCK EMAIL] To: %s, Subject: %s\n", email.To, email.Subject)
	return nil
}

// GetSentEmails returns all sent emails (for testing).
func (m *MockEmailSender) GetSentEmails() []Email {
	return m.SentEmails
}

// Clear clears the sent emails list.
func (m *MockEmailSender) Clear() {
	m.SentEmails = make([]Email, 0)
}
