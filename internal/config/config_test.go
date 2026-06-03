package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	content := `
[backup]
compress = true
dump_method = "binary"

[database.mysql]
host = "db.example.com"
port = 3307
user = "admin"
password = "pass123"
databases = ["app", "analytics"]

[storage.r2]
account_id = "acc123"
access_key_id = "key123"
access_key_secret = "secret123"
bucket = "backups"
path_prefix = "prod/"

[retention]
strategy = "days"
days = 14

[notification.telegram]
enabled = true
bot_token = "123:ABC"
chat_id = "-100123"
`
	path := writeTempFile(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Backup.Compress != true {
		t.Error("expected compress=true")
	}
	if cfg.Backup.DumpMethod != "binary" {
		t.Errorf("expected dump_method=binary, got %s", cfg.Backup.DumpMethod)
	}
	if cfg.Database.MySQL.Host != "db.example.com" {
		t.Errorf("expected host=db.example.com, got %s", cfg.Database.MySQL.Host)
	}
	if cfg.Database.MySQL.Port != 3307 {
		t.Errorf("expected port=3307, got %d", cfg.Database.MySQL.Port)
	}
	if len(cfg.Database.MySQL.Databases) != 2 {
		t.Errorf("expected 2 databases, got %d", len(cfg.Database.MySQL.Databases))
	}
	if cfg.Retention.Strategy != "days" {
		t.Errorf("expected strategy=days, got %s", cfg.Retention.Strategy)
	}
	if cfg.Retention.Days != 14 {
		t.Errorf("expected days=14, got %d", cfg.Retention.Days)
	}
	if cfg.Notification.Telegram.BotToken != "123:ABC" {
		t.Errorf("unexpected bot_token: %s", cfg.Notification.Telegram.BotToken)
	}
}

func TestLoad_Defaults(t *testing.T) {
	content := `
[database.mysql]
user = "root"
password = "x"
databases = ["test"]

[storage.r2]
account_id = "a"
access_key_id = "b"
access_key_secret = "c"
bucket = "d"
`
	path := writeTempFile(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Backup.DumpMethod != "go" {
		t.Errorf("expected default dump_method=go, got %s", cfg.Backup.DumpMethod)
	}
	if cfg.Database.MySQL.Host != "localhost" {
		t.Errorf("expected default host=localhost, got %s", cfg.Database.MySQL.Host)
	}
	if cfg.Database.MySQL.Port != 3306 {
		t.Errorf("expected default port=3306, got %d", cfg.Database.MySQL.Port)
	}
	if cfg.Retention.Strategy != "count" {
		t.Errorf("expected default strategy=count, got %s", cfg.Retention.Strategy)
	}
	if cfg.Retention.Count != 7 {
		t.Errorf("expected default count=7, got %d", cfg.Retention.Count)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.toml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoad_InvalidTOML(t *testing.T) {
	path := writeTempFile(t, "this is [[[not valid toml")
	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid TOML")
	}
}

func TestDSN(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			MySQL: MySQLConfig{
				Host:     "myhost",
				Port:     3307,
				User:     "admin",
				Password: "secret",
			},
		},
	}
	want := "admin:secret@tcp(myhost:3307)/"
	if got := cfg.DSN(); got != want {
		t.Errorf("DSN() = %s, want %s", got, want)
	}
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
