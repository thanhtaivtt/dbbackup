package engine

import (
	"context"
	"fmt"
	"io"
	"log/slog"
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
	store, err := storage.NewR2(cfg.Storage.R2)
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
	for _, db := range e.cfg.Database.MySQL.Databases {
		if err := e.backupDatabase(ctx, db); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) backupDatabase(ctx context.Context, database string) error {
	start := time.Now()
	e.logger.Info("starting backup", "database", database)

	// 1. Dump
	reader, err := e.dumper.Dump(ctx, database)
	if err != nil {
		e.notify(ctx, database, "", 0, start, err)
		return fmt.Errorf("dump %s: %w", database, err)
	}
	defer reader.Close()

	// 2. Compress (optional)
	var body io.Reader = reader
	ext := ".sql"
	if e.cfg.Backup.Compress {
		body = compress.NewGzipReader(reader)
		ext = ".sql.gz"
	}

	// 3. Upload
	key := fmt.Sprintf("%s%s_%s%s", e.cfg.Storage.R2.PathPrefix, database, time.Now().Format("20060102_150405"), ext)
	e.logger.Info("uploading", "key", key)

	cr := &countingReader{Reader: body}
	if err := e.storage.Upload(ctx, key, cr); err != nil {
		e.notify(ctx, database, key, 0, start, err)
		return fmt.Errorf("upload %s: %w", key, err)
	}

	// 4. Retention
	deleted, err := e.retention.Apply(ctx, e.cfg.Storage.R2.PathPrefix+database+"_")
	if err != nil {
		e.logger.Warn("retention cleanup failed", "error", err)
	} else if deleted > 0 {
		e.logger.Info("retention cleanup", "deleted", deleted)
	}

	// 5. Notify success
	duration := time.Since(start)
	e.logger.Info("backup complete", "database", database, "size", cr.n, "duration", duration)
	e.notify(ctx, database, key, cr.n, start, nil)

	return nil
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

type countingReader struct {
	io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.Reader.Read(p)
	c.n += int64(n)
	return n, err
}
