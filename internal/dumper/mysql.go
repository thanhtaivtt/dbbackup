package dumper

import (
	"context"
	"fmt"
	"io"

	"github.com/jarvanstack/mysqldump"
	"github.com/thanhtaivtt/dbbackup/internal/config"
)

type MySQLDumper struct {
	dsn string
}

func NewMySQL(cfg *config.Config) *MySQLDumper {
	return &MySQLDumper{dsn: cfg.DSN()}
}

func (d *MySQLDumper) Name() string { return "mysql-go" }

func (d *MySQLDumper) Dump(_ context.Context, database string) (io.ReadCloser, error) {
	pr, pw := io.Pipe()

	go func() {
		dsn := d.dsn + database
		err := mysqldump.Dump(dsn, mysqldump.WithWriter(pw), mysqldump.WithAllTable(), mysqldump.WithData())
		if err != nil {
			pw.CloseWithError(fmt.Errorf("mysqldump: %w", err))
			return
		}
		pw.Close()
	}()

	return pr, nil
}
