package smtp

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"testing"

	"github.com/turtacn/QuantaID/pkg/notification"
)

func TestSMTPSender_Type(t *testing.T) {
	cfg := SMTPConfig{}
	sender := NewSMTPSender(cfg)
	if sender.Type() != "email" {
		t.Errorf("expected type email, got %s", sender.Type())
	}
}

func TestNewSMTPSender(t *testing.T) {
	cfg := SMTPConfig{
		Host: "localhost",
		Port: 2525,
	}
	sender := NewSMTPSender(cfg)
	if sender == nil {
		t.Error("NewSMTPSender returned nil")
	}
	if sender.config.Host != "localhost" {
		t.Errorf("expected host localhost, got %s", sender.config.Host)
	}
}

func TestSMTPSender_Send_Mock(t *testing.T) {
	cfg := SMTPConfig{
		Host: "localhost",
		Port: 2525,
		From: "sender@example.com",
	}
	sender := NewSMTPSender(cfg)

	// Mock sendMailFunc
	var sentAddr string
	var sentFrom string
	var sentTo []string
	var sentMsg []byte

	sender.sendMailFunc = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		sentAddr = addr
		sentFrom = from
		sentTo = to
		sentMsg = msg
		return nil
	}

	msg := notification.Message{
		Recipient: "recipient@example.com",
		Subject:   "Test Subject",
		Body:      "Test Body",
	}

	err := sender.Send(context.Background(), msg)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	if sentAddr != "localhost:2525" {
		t.Errorf("expected addr localhost:2525, got %s", sentAddr)
	}
	if sentFrom != "sender@example.com" {
		t.Errorf("expected from sender@example.com, got %s", sentFrom)
	}
	if len(sentTo) != 1 || sentTo[0] != "recipient@example.com" {
		t.Errorf("expected to [recipient@example.com], got %v", sentTo)
	}

	msgStr := string(sentMsg)
	if !strings.Contains(msgStr, "From: sender@example.com") {
		t.Error("expected From header in message body")
	}
	if !strings.Contains(msgStr, "To: recipient@example.com") {
		t.Error("expected To header in message body")
	}
	if !strings.Contains(msgStr, "Subject: Test Subject") {
		t.Error("expected Subject header in message body")
	}
}

func TestSMTPSender_Send_ConnectionError(t *testing.T) {
	// This test attempts to send an email with a failing send function
	cfg := SMTPConfig{
		Host: "localhost",
		Port: 12345,
	}
	sender := NewSMTPSender(cfg)

	sender.sendMailFunc = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return fmt.Errorf("connection failed")
	}

	msg := notification.Message{
		Recipient: "test@example.com",
		Subject:   "Test",
		Body:      "Test Body",
	}

	err := sender.Send(context.Background(), msg)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if err.Error() != "failed to send email: connection failed" {
		t.Errorf("expected error message 'failed to send email: connection failed', got '%s'", err.Error())
	}
}
