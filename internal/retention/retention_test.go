package retention

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/thanhtaivtt/dbbackup/internal/config"
	"github.com/thanhtaivtt/dbbackup/internal/storage"
)

type mockStorage struct {
	objects []storage.Object
	deleted []string
}

func (m *mockStorage) Name() string { return "mock" }

func (m *mockStorage) Upload(_ context.Context, _ string, _ io.Reader) error { return nil }

func (m *mockStorage) List(_ context.Context, _ string) ([]storage.Object, error) {
	return m.objects, nil
}

func (m *mockStorage) Delete(_ context.Context, keys []string) error {
	m.deleted = append(m.deleted, keys...)
	return nil
}

func TestRetentionByCount(t *testing.T) {
	now := time.Now()
	mock := &mockStorage{
		objects: []storage.Object{
			{Key: "backup_1.sql", LastModified: now.Add(-4 * time.Hour)},
			{Key: "backup_2.sql", LastModified: now.Add(-3 * time.Hour)},
			{Key: "backup_3.sql", LastModified: now.Add(-2 * time.Hour)},
			{Key: "backup_4.sql", LastModified: now.Add(-1 * time.Hour)},
			{Key: "backup_5.sql", LastModified: now},
		},
	}

	mgr := New(mock, config.RetentionConfig{Strategy: "count", Count: 3})
	deleted, err := mgr.Apply(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if deleted != 2 {
		t.Errorf("expected 2 deletions, got %d", deleted)
	}
	// Should delete the 2 oldest
	deletedSet := map[string]bool{}
	for _, k := range mock.deleted {
		deletedSet[k] = true
	}
	if !deletedSet["backup_1.sql"] || !deletedSet["backup_2.sql"] {
		t.Errorf("expected backup_1.sql and backup_2.sql deleted, got %v", mock.deleted)
	}
}

func TestRetentionByCount_NoDelete(t *testing.T) {
	mock := &mockStorage{
		objects: []storage.Object{
			{Key: "backup_1.sql", LastModified: time.Now()},
		},
	}

	mgr := New(mock, config.RetentionConfig{Strategy: "count", Count: 5})
	deleted, err := mgr.Apply(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deletions, got %d", deleted)
	}
}

func TestRetentionByDays(t *testing.T) {
	now := time.Now()
	mock := &mockStorage{
		objects: []storage.Object{
			{Key: "old.sql", LastModified: now.AddDate(0, 0, -10)},
			{Key: "recent.sql", LastModified: now.AddDate(0, 0, -2)},
			{Key: "today.sql", LastModified: now},
		},
	}

	mgr := New(mock, config.RetentionConfig{Strategy: "days", Days: 7})
	deleted, err := mgr.Apply(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if deleted != 1 {
		t.Errorf("expected 1 deletion, got %d", deleted)
	}
	if mock.deleted[0] != "old.sql" {
		t.Errorf("expected old.sql deleted, got %v", mock.deleted)
	}
}
