package engine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/thanhtaivtt/dbbackup/internal/compress"
	"github.com/thanhtaivtt/dbbackup/internal/config"
	"github.com/thanhtaivtt/dbbackup/internal/dumper"
	"github.com/thanhtaivtt/dbbackup/internal/notifier"
	"github.com/thanhtaivtt/dbbackup/internal/retention"
	"github.com/thanhtaivtt/dbbackup/internal/storage"
)

type Engine struct {
	cfg       *config.Config
	dumper    dumper.Dumper
	storage   storage.Storage
	retention *retention.Manager
	notifiers []notifier.Notifier
	logger    *slog.Logger
}

func New(cfg *config.Config, logger *slog.Logger) (*Engine, error) {
	// Dumper
	var d dumper.Dumper
	switch cfg.Backup.DumpMethod {
	case "binary":
		d = dumper.NewMySQLBinary(cfg)
	default:
		d = dumper.NewMySQL(cfg)
	}

	// Storage
	var (
		store storage.Storage
		err   error
	)
	switch cfg.Storage.Backend {
	case "s3":
		store, err = storage.NewS3(cfg.Storage.S3)
	default:
		store, err = storage.NewR2(cfg.Storage.R2)
	}
	if err != nil {
		return nil, fmt.Errorf("init storage: %w", err)
	}

	// Retention
	ret := retention.New(store, cfg.Retention)

	// Notifiers
	var notifiers []notifier.Notifier
	if cfg.Notification.Telegram.Enabled {
		notifiers = append(notifiers, notifier.NewTelegram(cfg.Notification.Telegram))
	}
	if cfg.Notification.Slack.Enabled {
		notifiers = append(notifiers, notifier.NewSlack(cfg.Notification.Slack))
	}
	if cfg.Notification.Webhook.Enabled {
		notifiers = append(notifiers, notifier.NewWebhook(cfg.Notification.Webhook))
	}

	return &Engine{
		cfg:       cfg,
		dumper:    d,
		storage:   store,
		retention: ret,
		notifiers: notifiers,
		logger:    logger,
	}, nil
}

func (e *Engine) Run(ctx context.Context) error {
	var errs []error
	for _, db := range e.cfg.Database.MySQL.Databases {
		if err := e.backupDatabase(ctx, db); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("backup failed for %d database(s): %w", len(errs), errors.Join(errs...))
	}
	return nil
}

func (e *Engine) backupDatabase(ctx context.Context, database string) error {
	start := time.Now()
	e.logger.Info("starting backup", "database", database)

	// 1. Dump
	e.logger.Debug("dumping database", "database", database, "method", e.dumper.Name())
	reader, err := e.dumper.Dump(ctx, database)
	if err != nil {
		e.notify(ctx, database, "", 0, start, err)
		return fmt.Errorf("dump %s: %w", database, err)
	}

	// 2. Compress (optional) + buffer to temp file
	ext := ".sql"
	var body io.Reader = reader
	if e.cfg.Backup.Compress {
		body = compress.NewGzipReader(reader)
		ext = ".sql.gz"
	}

	// Buffer to temp file (needed for: detecting dump errors early, seekable upload)
	tmp, err := os.CreateTemp("", "dbbackup-dump-*")
	if err != nil {
		reader.Close()
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := io.Copy(tmp, body); err != nil {
		reader.Close()
		e.notify(ctx, database, "", 0, start, err)
		return fmt.Errorf("dump %s: %w", database, err)
	}
	reader.Close()

	fileSize, _ := tmp.Seek(0, io.SeekEnd)
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seeking temp file: %w", err)
	}

	// 3. Upload
	prefix := e.pathPrefix()
	key := fmt.Sprintf("%s%s_%s%s", prefix, database, time.Now().Format("20060102_150405"), ext)
	e.logger.Info("uploading", "key", key, "size", fileSize)

	if err := e.storage.Upload(ctx, key, tmp); err != nil {
		e.notify(ctx, database, key, 0, start, err)
		return fmt.Errorf("upload %s: %w", key, err)
	}

	// 4. Retention
	deleted, err := e.retention.Apply(ctx, prefix+database+"_")
	if err != nil {
		e.logger.Warn("retention cleanup failed", "error", err)
	} else if deleted > 0 {
		e.logger.Info("retention cleanup", "deleted", deleted)
	}

	// 5. Notify success
	duration := time.Since(start)
	e.logger.Info("backup complete", "database", database, "size", fileSize, "duration", duration)
	e.notify(ctx, database, key, fileSize, start, nil)

	return nil
}

func (e *Engine) pathPrefix() string {
	if e.cfg.Storage.Backend == "s3" {
		return e.cfg.Storage.S3.PathPrefix
	}
	return e.cfg.Storage.R2.PathPrefix
}

func (e *Engine) notify(ctx context.Context, database, fileName string, size int64, start time.Time, backupErr error) {
	msg := notifier.Message{
		Status:   "success",
		Database: database,
		FileName: fileName,
		FileSize: size,
		Duration: time.Since(start).Round(time.Millisecond).String(),
	}
	if backupErr != nil {
		msg.Status = "error"
		msg.Error = backupErr.Error()
	}

	for _, n := range e.notifiers {
		if err := n.Notify(ctx, msg); err != nil {
			e.logger.Warn("notification failed", "notifier", n.Name(), "error", err)
		}
	}
}

