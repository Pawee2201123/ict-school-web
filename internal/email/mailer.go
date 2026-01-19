package email

import (
	"fmt"
	"log"
	"strconv"

	"gopkg.in/gomail.v2"
)

// Config holds SMTP configuration
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// Mailer handles email sending
type Mailer struct {
	config Config
}

// NewMailer creates a new Mailer instance
func NewMailer(config Config) *Mailer {
	return &Mailer{config: config}
}

// Send sends an email
func (m *Mailer) Send(to, subject, htmlBody string) error {
	// Skip sending if SMTP is not configured
	if m.config.Host == "" || m.config.From == "" {
		log.Printf("SMTP not configured, skipping email to %s", to)
		return nil
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.config.From)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	// gomail uses UTF-8 by default for HTML body
	msg.SetBody("text/html", htmlBody)

	dialer := gomail.NewDialer(m.config.Host, m.config.Port, m.config.Username, m.config.Password)

	if err := dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent successfully to %s", to)
	return nil
}

// ParsePort converts port string to int, returns default 587 if invalid
func ParsePort(portStr string) int {
	if portStr == "" {
		return 587
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 587
	}
	return port
}
