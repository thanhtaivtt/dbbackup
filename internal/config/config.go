package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Backup       BackupConfig       `toml:"backup"`
	Database     DatabaseConfig     `toml:"database"`
	Storage      StorageConfig      `toml:"storage"`
	Retention    RetentionConfig    `toml:"retention"`
	Notification NotificationConfig `toml:"notification"`
}

type BackupConfig struct {
	Compress   bool   `toml:"compress"`
	DumpMethod string `toml:"dump_method"` // "go" or "binary"
}

type DatabaseConfig struct {
	MySQL MySQLConfig `toml:"mysql"`
}

type MySQLConfig struct {
	Host      string   `toml:"host"`
	Port      int      `toml:"port"`
	User      string   `toml:"user"`
	Password  string   `toml:"password"`
	Databases []string `toml:"databases"`
}

type StorageConfig struct {
	R2 R2Config `toml:"r2"`
}

type R2Config struct {
	AccountID       string `toml:"account_id"`
	AccessKeyID     string `toml:"access_key_id"`
	AccessKeySecret string `toml:"access_key_secret"`
	Bucket          string `toml:"bucket"`
	PathPrefix      string `toml:"path_prefix"`
}

type RetentionConfig struct {
	Strategy string `toml:"strategy"` // "count" or "days"
	Count    int    `toml:"count"`
	Days     int    `toml:"days"`
}

type NotificationConfig struct {
	Telegram TelegramConfig `toml:"telegram"`
	Slack    SlackConfig    `toml:"slack"`
	Webhook  WebhookConfig  `toml:"webhook"`
}

type TelegramConfig struct {
	Enabled  bool   `toml:"enabled"`
	BotToken string `toml:"bot_token"`
	ChatID   string `toml:"chat_id"`
}

type SlackConfig struct {
	Enabled    bool   `toml:"enabled"`
	WebhookURL string `toml:"webhook_url"`
}

type WebhookConfig struct {
	Enabled bool   `toml:"enabled"`
	URL     string `toml:"url"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg := &Config{
		Backup: BackupConfig{
			DumpMethod: "go",
		},
		Database: DatabaseConfig{
			MySQL: MySQLConfig{
				Host: "localhost",
				Port: 3306,
			},
		},
		Retention: RetentionConfig{
			Strategy: "count",
			Count:    7,
		},
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Database.MySQL.User == "" {
		return fmt.Errorf("config: database.mysql.user is required")
	}
	if len(c.Database.MySQL.Databases) == 0 {
		return fmt.Errorf("config: database.mysql.databases is required")
	}
	if c.Storage.R2.AccountID == "" {
		return fmt.Errorf("config: storage.r2.account_id is required")
	}
	if c.Storage.R2.AccessKeyID == "" {
		return fmt.Errorf("config: storage.r2.access_key_id is required")
	}
	if c.Storage.R2.AccessKeySecret == "" {
		return fmt.Errorf("config: storage.r2.access_key_secret is required")
	}
	if c.Storage.R2.Bucket == "" {
		return fmt.Errorf("config: storage.r2.bucket is required")
	}
	if c.Notification.Telegram.Enabled {
		if c.Notification.Telegram.BotToken == "" {
			return fmt.Errorf("config: notification.telegram.bot_token is required when enabled")
		}
		if c.Notification.Telegram.ChatID == "" {
			return fmt.Errorf("config: notification.telegram.chat_id is required when enabled")
		}
	}
	if c.Notification.Slack.Enabled {
		if c.Notification.Slack.WebhookURL == "" {
			return fmt.Errorf("config: notification.slack.webhook_url is required when enabled")
		}
	}
	if c.Notification.Webhook.Enabled {
		if c.Notification.Webhook.URL == "" {
			return fmt.Errorf("config: notification.webhook.url is required when enabled")
		}
	}
	return nil
}

func (c *Config) DSN() string {
	m := c.Database.MySQL
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/", m.User, m.Password, m.Host, m.Port)
}

// DumpDSN returns DSN with database name for mysqldump library.
// The jarvanstack/mysqldump library requires ?parseTime=true or similar
// query parameter in DSN to correctly parse the database name.
func (c *Config) DumpDSN(database string) string {
	m := c.Database.MySQL
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", m.User, m.Password, m.Host, m.Port, database)
}
