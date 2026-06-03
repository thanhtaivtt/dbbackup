package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/thanhtaivtt/dbbackup/internal/config"
)

type TelegramNotifier struct {
	botToken string
	chatID   string
	client   *http.Client
}

func NewTelegram(cfg config.TelegramConfig) *TelegramNotifier {
	return &TelegramNotifier{
		botToken: cfg.BotToken,
		chatID:   cfg.ChatID,
		client:   &http.Client{},
	}
}

func (t *TelegramNotifier) Name() string { return "telegram" }

func (t *TelegramNotifier) Notify(ctx context.Context, msg Message) error {
	text := formatMessage(msg)

	payload := map[string]string{
		"chat_id":    t.chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned %d", resp.StatusCode)
	}
	return nil
}

func formatMessage(msg Message) string {
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

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
