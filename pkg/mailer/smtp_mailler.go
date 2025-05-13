package mailer

import (
	"errors"

	"gopkg.in/gomail.v2"
)

type EmailSender interface {
	Send(to []string, subject, body string) error
}

type GomailSenderConfig struct {
	From     string
	Host     string
	Port     int
	Username string
	Password string
}

type GomailSender struct {
	config GomailSenderConfig
}

func NewGomailSender(config GomailSenderConfig) *GomailSender {
	return &GomailSender{config: config}
}

// Send sends an email using the SMTP configuration
// Parameters:
//   - to: slice of recipient email addresses
//   - subject: email subject line
//   - body: html email body content
//
// Returns:
//   - error if sending fails, nil on success
func (s *GomailSender) Send(to []string, subject, plainText, html string) error {

	// Validate inputs
	if len(to) == 0 {
		return errors.New("recipient list cannot be empty")
	}
	if subject == "" {
		return errors.New("email subject cannot be empty")
	}
	if plainText == "" && html == "" {
		return errors.New("either plain text or HTML content must be provided")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.config.From)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	// Add plain text and HTML alternatives
	if plainText != "" {
		m.SetBody("text/plain", plainText)
	}
	if html != "" {
		m.AddAlternative("text/html", html)
	}

	// Create a new dialer and send the email
	dialer := gomail.NewDialer(s.config.Host, s.config.Port, s.config.Username, s.config.Password)
	if err := dialer.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
