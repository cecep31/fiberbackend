package email

import (
	"context"
	"strings"
	"testing"
	"time"

	"fiberbackend/config"
	"fiberbackend/internal/platform/queue"
)

func TestEmailService_Unconfigured(t *testing.T) {
	svc := NewService(config.EmailConfig{}, nil)

	if svc.IsConfigured() {
		t.Fatal("expected unconfigured service to report IsConfigured() == false")
	}

	err := svc.EnqueuePasswordResetEmail("user@example.com", "http://localhost/reset")
	if err == nil {
		t.Fatal("expected error when enqueuing on unconfigured service")
	}

	err = svc.SendPasswordResetEmail(context.Background(), "user@example.com", "http://localhost/reset")
	if err == nil {
		t.Fatal("expected error when sending on unconfigured service")
	}

	if err := svc.Close(); err != nil {
		t.Fatalf("unexpected error on Close: %v", err)
	}
}

func TestEmailService_SMTPWithoutQueueIsUnconfigured(t *testing.T) {
	// Fiber's email delivery goes through the Asynq queue: SMTP settings alone
	// are not enough for the service to report itself as configured.
	cfg := config.EmailConfig{
		SMTPHost:     "127.0.0.1",
		SMTPPort:     2525,
		SMTPUsername: "user",
		SMTPPassword: "password",
		From:         "noreply@example.com",
		Timeout:      2 * time.Second,
		TaskTimeout:  5 * time.Second,
	}

	svc := NewService(cfg, queue.NewService(config.QueueConfig{}))
	if svc.IsConfigured() {
		t.Fatal("expected service without a configured queue to report IsConfigured() == false")
	}

	if err := svc.EnqueuePasswordResetEmail("test@example.com", "http://localhost/reset?token=abc"); err == nil {
		t.Fatal("expected error when enqueuing without a configured queue")
	}

	if err := svc.Close(); err != nil {
		t.Fatalf("unexpected error closing service: %v", err)
	}
}

func TestEmailService_Template(t *testing.T) {
	textBody, htmlBody := passwordResetTemplate("http://example.com/reset?token=123", "1 hour")

	if !strings.Contains(textBody, "http://example.com/reset?token=123") {
		t.Error("text body does not contain reset link")
	}
	if !strings.Contains(htmlBody, "http://example.com/reset?token=123") {
		t.Error("html body does not contain reset link")
	}
}

func TestEmailService_BuildMessage(t *testing.T) {
	msg, err := buildMessage("sender@example.com", "recipient@example.com", "Test Subject", "Hello text", "<p>Hello html</p>")
	if err != nil {
		t.Fatalf("unexpected error building message: %v", err)
	}

	msgStr := string(msg)
	if !strings.Contains(msgStr, "Subject: Test Subject") {
		t.Error("message missing subject header")
	}
	if !strings.Contains(msgStr, "sender@example.com") {
		t.Error("message missing from header")
	}
	if !strings.Contains(msgStr, "recipient@example.com") {
		t.Error("message missing to header")
	}
}
