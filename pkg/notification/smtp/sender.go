package smtp

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/turtacn/QuantaID/pkg/notification"
)

// SMTPConfig holds configuration for the SMTP sender
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// SMTPSender implements notification.Notifier for email via SMTP
type SMTPSender struct {
	config       SMTPConfig
	sendMailFunc func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

// NewSMTPSender creates a new SMTPSender
func NewSMTPSender(cfg SMTPConfig) *SMTPSender {
	return &SMTPSender{
		config:       cfg,
		sendMailFunc: smtp.SendMail,
	}
}

// Send sends an email using the configured SMTP server
func (s *SMTPSender) Send(ctx context.Context, msg notification.Message) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// Set up authentication information.
	var auth smtp.Auth
	if s.config.Username != "" && s.config.Password != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	to := []string{msg.Recipient}

	// Create the email message
	// Added From header
	body := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", s.config.From, strings.Join(to, ","), msg.Subject, msg.Body))

	err := s.sendMailFunc(addr, auth, s.config.From, to, body)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// Type returns the type of the notifier
func (s *SMTPSender) Type() string {
	return "email"
}
