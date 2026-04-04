package notify

import (
	"context"
	"fmt"
	"net/smtp"

	"flowpilot/internal/models"
)

type EmailNotifier struct {
	host string
	port string
	user string
	pass string
	to   string
}

func NewEmailNotifier(host, port, user, pass, to string) *EmailNotifier {
	return &EmailNotifier{host: host, port: port, user: user, pass: pass, to: to}
}

func (e *EmailNotifier) Name() string { return "email" }

func (e *EmailNotifier) Send(ctx context.Context, event models.AlertEvent) error {
	subject := fmt.Sprintf("FlowPilot Alert [%s]: %s", event.Severity, event.Message)
	body := fmt.Sprintf("Subject: %s\r\nFrom: %s\r\nTo: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\n\nEvidence: %s",
		subject, e.user, e.to, event.Message, event.Evidence)

	auth := smtp.PlainAuth("", e.user, e.pass, e.host)
	addr := e.host + ":" + e.port
	return smtp.SendMail(addr, auth, e.user, []string{e.to}, []byte(body))
}
