package dumper

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/thanhtaivtt/dbbackup/internal/config"
)

type MySQLBinaryDumper struct {
	cfg config.MySQLConfig
}

func NewMySQLBinary(cfg *config.Config) *MySQLBinaryDumper {
	return &MySQLBinaryDumper{cfg: cfg.Database.MySQL}
}

func (d *MySQLBinaryDumper) Name() string { return "mysql-binary" }

func (d *MySQLBinaryDumper) Dump(ctx context.Context, database string) (io.ReadCloser, error) {
	args := []string{
		"-h", d.cfg.Host,
		"-P", fmt.Sprintf("%d", d.cfg.Port),
		"-u", d.cfg.User,
		fmt.Sprintf("-p%s", d.cfg.Password),
		database,
	}

	cmd := exec.CommandContext(ctx, "mysqldump", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting mysqldump: %w", err)
	}

	return &cmdReadCloser{ReadCloser: stdout, cmd: cmd}, nil
}

type cmdReadCloser struct {
	io.ReadCloser
	cmd *exec.Cmd
}

func (c *cmdReadCloser) Close() error {
	c.ReadCloser.Close()
	return c.cmd.Wait()
}
