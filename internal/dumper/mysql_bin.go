package dumper

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

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
		database,
	}

	cmd := exec.CommandContext(ctx, "mysqldump", args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", d.cfg.Password))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting mysqldump: %w", err)
	}

	return &cmdReadCloser{ReadCloser: stdout, cmd: cmd, stderr: &stderr}, nil
}

type cmdReadCloser struct {
	io.ReadCloser
	cmd    *exec.Cmd
	stderr *bytes.Buffer
}

func (c *cmdReadCloser) Close() error {
	c.ReadCloser.Close()
	if err := c.cmd.Wait(); err != nil {
		if c.stderr.Len() > 0 {
			return fmt.Errorf("mysqldump: %s", strings.TrimSpace(c.stderr.String()))
		}
		return fmt.Errorf("mysqldump: %w", err)
	}
	return nil
}
