package dumper

import (
	"context"
	"io"
)

type Dumper interface {
	Dump(ctx context.Context, database string) (io.ReadCloser, error)
	Name() string
}
