package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"flowpilot/internal/models"
)

type DiscordNotifier struct {
	webhookURL string
	client     *http.Client
}

func NewDiscordNotifier(webhookURL string) *DiscordNotifier {
	return &DiscordNotifier{webhookURL: webhookURL, client: &http.Client{}}
}

func (d *DiscordNotifier) Name() string { return "discord" }

func (d *DiscordNotifier) Send(ctx context.Context, event models.AlertEvent) error {
	payload := map[string]string{
		"content": fmt.Sprintf("**FlowPilot Alert** [%s]\n\n%s\n\n*%s*", event.Severity, event.Message, event.Evidence),
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", d.webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord webhook returned %d", resp.StatusCode)
	}
	return nil
}
