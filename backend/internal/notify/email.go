package notify

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"github.com/pingan/hydra/internal/model"
)

// EmailChannel sends alert notifications via SMTP email.
type EmailChannel struct {
	smtpHost string
	smtpPort string
	username string
	password string
	from     string
	to       []string
}

// NewEmailChannel creates an email notification channel.
func NewEmailChannel(smtpHost, smtpPort, username, password, from string, to []string) *EmailChannel {
	return &EmailChannel{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		username: username,
		password: password,
		from:     from,
		to:       to,
	}
}

func (e *EmailChannel) Name() string { return "email" }

// Send delivers an alert as an email.
func (e *EmailChannel) Send(alert *model.Alert) error {
	subject := fmt.Sprintf("[%s] %s - %s: %s",
		alert.Severity, alert.RuleName, alert.Host, alert.Metric)

	body := fmt.Sprintf(`Alert: %s
Severity: %s
Host: %s (%s)
Metric: %s = %.2f (threshold: %.2f)
Message: %s
Time: %s
---
Sent by Hydra Monitoring System`,
		alert.RuleName,
		alert.Severity,
		alert.Host, alert.Region,
		alert.Metric, alert.Value, alert.Threshold,
		alert.Message,
		alert.FiredAt.Format("2006-01-02 15:04:05"),
	)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		e.from, strings.Join(e.to, ", "), subject, body)

	addr := fmt.Sprintf("%s:%s", e.smtpHost, e.smtpPort)
	auth := smtp.PlainAuth("", e.username, e.password, e.smtpHost)

	if err := smtp.SendMail(addr, auth, e.from, e.to, []byte(msg)); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	log.Printf("email: sent alert %s to %v", alert.ID, e.to)
	return nil
}
