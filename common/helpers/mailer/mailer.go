package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// Mailer interface.
// This interface is used to define the mailer methods.
//
// Methods:
//   - SendMail(to []string, subject string, body string) error: This method is used to send a plain text email.
//   - SendMailWithTemplate(to []string, subject string, templatePath string, data interface{}) error: This method is used to send an HTML email with a template.
type Mailer interface {
	SendMail(to []string, subject string, body string) error
	SendMailWithTemplate(to []string, subject string, templatePath string, data interface{}) error
}

// mailer struct.
// This struct is used to implement the Mailer interface.
//
// Attributes:
//   - auth (smtp.Auth): The SMTP authentication.
//   - from (string): The sender email address.
//   - smtpServer (string): The SMTP server address.
type mailer struct {
	auth       smtp.Auth
	from       string
	smtpServer string
	log        *logrus.Logger
}

// NewMailer function.
// This function is used to create a new mailer.
//
// Parameters:
//   - smtpUser (string): The SMTP user.
//   - smtpPassword (string): The SMTP password.
//   - smtpHost (string): The SMTP host.
//   - smtpPort (int): The SMTP port.
//   - from (string): The sender email address.
//
// Returns:
//   - Mailer: The mailer.
func NewMailer(smtpUser, smtpPassword, smtpHost string, smtpPort int, from string, log *logrus.Logger) Mailer {
	smtpServer := fmt.Sprintf("%s:%d", smtpHost, smtpPort)
	var auth smtp.Auth
	if smtpUser != "" && smtpPassword != "" {
		auth = smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)
	} else {
		auth = nil
	}

	return &mailer{
		auth:       auth,
		from:       from,
		smtpServer: smtpServer,
		log:        log,
	}
}

// SendMail method.
// This method is used to send a plain text email.
//
// Parameters:
//   - to ([]string): The recipient email addresses.
//   - subject (string): The email subject.
//   - body (string): The email body.
//
// Returns:
//   - error: The error.
func (m *mailer) SendMail(to []string, subject string, body string) error {
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	msg := []byte("To: " + to[0] + "\r\n" +
		"From: " + m.from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		mime + "\r\n" +
		body)

	err := smtp.SendMail(m.smtpServer, m.auth, m.from, to, msg)
	if err != nil {
		m.log.WithFields(logrus.Fields{"to": to, "subject": subject, "error": err}).Error("Failed to Send Email!")
		return fmt.Errorf("failed to send email: %w", err)
	}

	m.log.WithFields(logrus.Fields{"to": to, "subject": subject}).Info("Email Sent Successfully!")
	return nil
}

// SendMailWithTemplate method.
// This method is used to send an HTML email with a template.
//
// Parameters:
//   - to ([]string): The recipient email addresses.
//   - subject (string): The email subject.
//   - templatePath (string): The path to the HTML template.
//   - data (interface{}): The data to be passed to the template.
//
// Returns:
//   - error: The error.
func (m *mailer) SendMailWithTemplate(to []string, subject string, templatePath string, data interface{}) error {
	tmplPath, err := filepath.Abs(templatePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for template: %w", err)
	}

	if _, err := os.Stat(tmplPath); os.IsNotExist(err) {
		return fmt.Errorf("template file not found: %s", tmplPath)
	}

	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte("To: " + to[0] + "\r\n" +
		"From: " + m.from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		mime + body.String())

	err = smtp.SendMail(m.smtpServer, m.auth, m.from, to, msg)
	if err != nil {
		m.log.WithFields(logrus.Fields{"to": to, "subject": subject, "templatePath": templatePath, "error": err}).Error("Failed to Send Email with Template!")
		return fmt.Errorf("failed to send email with template: %w", err)
	}

	m.log.WithFields(logrus.Fields{"to": to, "subject": subject, "templatePath": templatePath}).Info("Email Sent Successfully with Template!")
	return nil
}
