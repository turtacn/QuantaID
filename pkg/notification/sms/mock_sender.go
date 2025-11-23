package sms

import (
	"context"
	"fmt"

	"github.com/turtacn/QuantaID/pkg/notification"
)

// MockSender implements notification.Notifier for SMS (Mock)
type MockSender struct {
}

// NewMockSender creates a new MockSender
func NewMockSender() *MockSender {
	return &MockSender{}
}

// Send logs the SMS message instead of sending it
func (s *MockSender) Send(ctx context.Context, msg notification.Message) error {
	fmt.Printf("[Mock SMS Sender] Sending SMS to %s: %s\n", msg.Recipient, msg.Body)
	return nil
}

// Type returns the type of the notifier
func (s *MockSender) Type() string {
	return "sms"
}
