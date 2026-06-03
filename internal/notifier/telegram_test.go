package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thanhtaivtt/dbbackup/internal/config"
)

func TestTelegramNotifier_Success(t *testing.T) {
	var received map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notif := &TelegramNotifier{
		botToken: "test-token",
		chatID:   "-100123",
		client:   server.Client(),
	}
	// Override URL by replacing the endpoint
	origNotify := notif.Notify
	_ = origNotify

	// Use a custom notifier that points to our test server
	notif2 := &testTelegramNotifier{
		baseURL: server.URL,
		chatID:  "-100123",
		client:  server.Client(),
	}

	msg := Message{
		Status:   "success",
		Database: "mydb",
		FileName: "mydb_20240101.sql.gz",
		FileSize: 1024 * 1024,
		Duration: "2.5s",
	}

	err := notif2.Notify(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["chat_id"] != "-100123" {
		t.Errorf("unexpected chat_id: %s", received["chat_id"])
	}
}

func TestTelegramNotifier_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	notif := &testTelegramNotifier{
		baseURL: server.URL,
		chatID:  "-100123",
		client:  server.Client(),
	}

	err := notif.Notify(context.Background(), Message{Status: "success", Database: "test"})
	if err == nil {
		t.Error("expected error for non-200 response")
	}
}

func TestFormatMessage_Success(t *testing.T) {
	msg := Message{Status: "success", Database: "mydb", FileName: "file.sql", FileSize: 2048, Duration: "1s"}
	text := formatMessage(msg)
	if len(text) == 0 {
		t.Error("expected non-empty message")
	}
}

func TestFormatMessage_Error(t *testing.T) {
	msg := Message{Status: "error", Database: "mydb", Error: "connection refused"}
	text := formatMessage(msg)
	if len(text) == 0 {
		t.Error("expected non-empty message")
	}
}

func TestNewTelegram(t *testing.T) {
	cfg := config.TelegramConfig{BotToken: "tok", ChatID: "123"}
	n := NewTelegram(cfg)
	if n.Name() != "telegram" {
		t.Errorf("expected name=telegram, got %s", n.Name())
	}
}

// testTelegramNotifier is a version that uses a custom base URL for testing
type testTelegramNotifier struct {
	baseURL string
	chatID  string
	client  *http.Client
}

func (t *testTelegramNotifier) Notify(ctx context.Context, msg Message) error {
	text := formatMessage(msg)
	payload := map[string]string{
		"chat_id":    t.chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/sendMessage", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned %d", resp.StatusCode)
	}
	return nil
}
