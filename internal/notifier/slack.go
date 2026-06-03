package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/thanhtaivtt/dbbackup/internal/config"
)

type SlackNotifier struct {
	webhookURL string
	client     *http.Client
}

func NewSlack(cfg config.SlackConfig) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: cfg.WebhookURL,
		client:     &http.Client{},
	}
}

func (s *SlackNotifier) Name() string { return "slack" }

func (s *SlackNotifier) Notify(ctx context.Context, msg Message) error {
	text := formatSlackMessage(msg)

	payload := map[string]string{"text": text}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("slack send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned %d", resp.StatusCode)
	}
	return nil
}

func formatSlackMessage(msg Message) string {
	icon := "✅"
	status := "Success"
	if msg.Status == "error" {
		icon = "❌"
		status = "Error"
	}

	text := fmt.Sprintf("%s *DB Backup %s*\n\n", icon, status)
	text += fmt.Sprintf("📦 Database: `%s`\n", msg.Database)

	if msg.FileName != "" {
		text += fmt.Sprintf("📄 File: `%s`\n", msg.FileName)
	}
	if msg.FileSize > 0 {
		text += fmt.Sprintf("📏 Size: %s\n", formatSize(msg.FileSize))
	}
	if msg.Duration != "" {
		text += fmt.Sprintf("⏱ Duration: %s\n", msg.Duration)
	}
	if msg.Error != "" {
		text += fmt.Sprintf("⚠️ Error: %s\n", msg.Error)
	}
	return text
}
