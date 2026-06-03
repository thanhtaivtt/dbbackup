package retention

import (
	"context"
	"sort"
	"time"

	"github.com/thanhtaivtt/dbbackup/internal/config"
	"github.com/thanhtaivtt/dbbackup/internal/storage"
)

type Manager struct {
	storage  storage.Storage
	strategy string
	count    int
	days     int
}

func New(s storage.Storage, cfg config.RetentionConfig) *Manager {
	return &Manager{
		storage:  s,
		strategy: cfg.Strategy,
		count:    cfg.Count,
		days:     cfg.Days,
	}
}

func (m *Manager) Apply(ctx context.Context, prefix string) (int, error) {
	objects, err := m.storage.List(ctx, prefix)
	if err != nil {
		return 0, err
	}

	var toDelete []string

	switch m.strategy {
	case "count":
		toDelete = m.byCount(objects)
	case "days":
		toDelete = m.byDays(objects)
	}

	if len(toDelete) == 0 {
		return 0, nil
	}

	return len(toDelete), m.storage.Delete(ctx, toDelete)
}

func (m *Manager) byCount(objects []storage.Object) []string {
	if len(objects) <= m.count {
		return nil
	}

	sort.Slice(objects, func(i, j int) bool {
		return objects[i].LastModified.After(objects[j].LastModified)
	})

	var keys []string
	for _, obj := range objects[m.count:] {
		keys = append(keys, obj.Key)
	}
	return keys
}

func (m *Manager) byDays(objects []storage.Object) []string {
	cutoff := time.Now().AddDate(0, 0, -m.days)
	var keys []string
	for _, obj := range objects {
		if obj.LastModified.Before(cutoff) {
			keys = append(keys, obj.Key)
		}
	}
	return keys
}
