package mail

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"
)

// EmailSender interface for sending emails
type EmailSender interface {
	SendPasswordResetEmail(toEmail, resetLink string) error
}

// SMTPEmailSender implements EmailSender for SMTP
type SMTPEmailSender struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// NewSMTPEmailSender creates a new SMTPEmailSender
func NewSMTPEmailSender(host, port, username, password, from string) *SMTPEmailSender {
	return &SMTPEmailSender{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}

// SendPasswordResetEmail sends a password reset email
func (s *SMTPEmailSender) SendPasswordResetEmail(toEmail, resetLink string) error {
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)

	t, err := template.New("password_reset").Parse(passwordResetEmailTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	if err := t.Execute(&body, struct{ ResetLink string }{ResetLink: resetLink}); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	msg := []byte(
		"From: " + s.From + "\r\n" +
			"To: " + toEmail + "\r\n" +
			"Subject: Password Reset Request\r\n" +
			"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n" +
			"\r\n" +
			body.String())

	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)
	err = smtp.SendMail(addr, auth, s.From, []string{toEmail}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

const passwordResetEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Password Reset</title>
</head>
<body>
    <p>Hello,</p>
    <p>You have requested a password reset for your account. Please click on the following link to reset your password:</p>
    <p><a href="{{.ResetLink}}">Reset Your Password</a></p>
    <p>This link will expire in 1 hour.</p>
    <p>If you did not request a password reset, please ignore this email.</p>
    <p>Thanks,</p>
    <p>Your Application Team</p>
</body>
</html>
`
