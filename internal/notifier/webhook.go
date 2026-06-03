package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/thanhtaivtt/dbbackup/internal/config"
)

type WebhookNotifier struct {
	url    string
	client *http.Client
}

func NewWebhook(cfg config.WebhookConfig) *WebhookNotifier {
	return &WebhookNotifier{
		url:    cfg.URL,
		client: &http.Client{},
	}
}

func (w *WebhookNotifier) Name() string { return "webhook" }

func (w *WebhookNotifier) Notify(ctx context.Context, msg Message) error {
	payload := map[string]interface{}{
		"status":   msg.Status,
		"database": msg.Database,
		"file":     msg.FileName,
		"size":     msg.FileSize,
		"duration": msg.Duration,
		"error":    msg.Error,
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned %d", resp.StatusCode)
	}
	return nil
}
