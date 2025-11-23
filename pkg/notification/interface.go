package notification

import (
	"context"
)

// MessageType defines the type of message (e.g., OTP, Alert)
type MessageType string

const (
	MessageTypeOTP   MessageType = "otp"
	MessageTypeAlert MessageType = "alert"
)

// Message represents a notification message
type Message struct {
	Recipient string
	Subject   string
	Body      string // Supports HTML
	Type      MessageType
	Metadata  map[string]string
}

// Notifier defines the interface for sending notifications
type Notifier interface {
	// Send sends a notification
	Send(ctx context.Context, msg Message) error
	// Type returns the notifier type (e.g., "email", "sms")
	Type() string
}

// Manager manages multiple Notifiers
type Manager interface {
	GetNotifier(method string) (Notifier, error)
}
