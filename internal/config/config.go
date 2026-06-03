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
}

type TelegramConfig struct {
	Enabled  bool   `toml:"enabled"`
	BotToken string `toml:"bot_token"`
	ChatID   string `toml:"chat_id"`
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

	return cfg, nil
}

func (c *Config) DSN() string {
	m := c.Database.MySQL
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/", m.User, m.Password, m.Host, m.Port)
}
