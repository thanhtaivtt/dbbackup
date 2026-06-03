package notifier

import "context"

type Message struct {
	Status   string // "success" or "error"
	Database string
	FileName string
	FileSize int64
	Duration string
	Error    string
}

type Notifier interface {
	Notify(ctx context.Context, msg Message) error
	Name() string
}
