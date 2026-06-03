# dbbackup

[![CI](https://github.com/thanhtaivtt/dbbackup/actions/workflows/ci.yml/badge.svg)](https://github.com/thanhtaivtt/dbbackup/actions/workflows/ci.yml)
[![Release](https://github.com/thanhtaivtt/dbbackup/actions/workflows/release.yml/badge.svg)](https://github.com/thanhtaivtt/dbbackup/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/thanhtaivtt/dbbackup)](https://goreportcard.com/report/github.com/thanhtaivtt/dbbackup)
[![GitHub release](https://img.shields.io/github/v/release/thanhtaivtt/dbbackup)](https://github.com/thanhtaivtt/dbbackup/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A CLI tool for backing up databases to cloud storage. Currently supports MySQL → Cloudflare R2, designed to be extensible for additional databases and storage backends.

## Features

- **MySQL backup** via pure Go library (no external dependencies) or `mysqldump` binary
- **Cloudflare R2** storage (S3-compatible)
- **Gzip compression** (optional)
- **Retention policies** — keep N latest backups or delete older than X days
- **Telegram notifications** on success/failure
- **TOML configuration** with CLI flag overrides
- **Extensible architecture** — easy to add new DB engines, storage backends, or notification channels

## Installation

### Quick install (Linux/macOS)

```bash
curl -sSL https://raw.githubusercontent.com/thanhtaivtt/dbbackup/main/install.sh | sh
```

### From releases

Download the latest binary from the [Releases](https://github.com/thanhtaivtt/dbbackup/releases) page.

### From source

```bash
go install github.com/thanhtaivtt/dbbackup@latest
```

### Build locally

```bash
make build
```

## Usage

```bash
# Generate config file
dbbackup init

# Edit config
vim config.toml

# Run backup
dbbackup backup --config config.toml

# With CLI overrides
dbbackup backup --config config.toml --compress --retention-count 5

# Verbose output
dbbackup backup --config config.toml -v
```

### Cron setup (daily at 2AM)

```cron
0 2 * * * /usr/local/bin/dbbackup backup --config /etc/dbbackup/config.toml
```

## Configuration

See [config.example.toml](config.example.toml) for a full example.

### `[backup]`

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `compress` | bool | `false` | Enable gzip compression for backup files |
| `dump_method` | string | `"go"` | MySQL dump method: `"go"` (pure Go library, supports remote DB) or `"binary"` (requires `mysqldump` CLI installed on host) |

### `[database.mysql]`

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `host` | string | `"localhost"` | MySQL server hostname or IP |
| `port` | int | `3306` | MySQL server port |
| `user` | string | *required* | MySQL user |
| `password` | string | *required* | MySQL password |
| `databases` | []string | *required* | List of database names to back up |

### `[storage.r2]`

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `account_id` | string | *required* | Cloudflare account ID |
| `access_key_id` | string | *required* | R2 API token Access Key ID |
| `access_key_secret` | string | *required* | R2 API token Secret Access Key |
| `bucket` | string | *required* | R2 bucket name |
| `path_prefix` | string | `""` | Prefix for object keys (e.g. `"mysql/"` → `mysql/mydb_20240101_020000.sql.gz`) |

### `[retention]`

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `strategy` | string | `"count"` | Retention strategy: `"count"` (keep N latest) or `"days"` (keep for N days) |
| `count` | int | `7` | Number of backups to keep (when `strategy = "count"`) |
| `days` | int | `30` | Days to keep backups (when `strategy = "days"`) |

### `[notification.telegram]`

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | `false` | Enable Telegram notifications |
| `bot_token` | string | *required* | Telegram Bot API token (from [@BotFather](https://t.me/BotFather)) |
| `chat_id` | string | *required* | Target chat/group ID (use [@userinfobot](https://t.me/userinfobot) to find) |

### Minimal config example

```toml
[database.mysql]
user = "root"
password = "secret"
databases = ["mydb"]

[storage.r2]
account_id = "your-account-id"
access_key_id = "your-access-key-id"
access_key_secret = "your-access-key-secret"
bucket = "db-backups"
```

### Full config example

```toml
[backup]
compress = true
dump_method = "go"

[database.mysql]
host = "localhost"
port = 3306
user = "root"
password = "secret"
databases = ["mydb", "analytics"]

[storage.r2]
account_id = "your-account-id"
access_key_id = "your-access-key-id"
access_key_secret = "your-access-key-secret"
bucket = "db-backups"
path_prefix = "mysql/"

[retention]
strategy = "count"
count = 7

[notification.telegram]
enabled = true
bot_token = "123456:ABC-DEF"
chat_id = "-1001234567890"
```

## CLI Flags

| Flag | Description |
|------|-------------|
| `--config` | Config file path (default: `config.toml`) |
| `--compress` | Enable gzip compression |
| `--retention-count N` | Keep N latest backups |
| `--retention-days N` | Keep backups for N days |
| `-v, --verbose` | Enable debug logging |
| `--version` | Print version info |

## Commands

| Command | Description |
|---------|-------------|
| `init` | Generate a config file |
| `backup` | Run database backup |
| `update` | Check for new version on GitHub |

## Extending

### Add a new database

Implement the `Dumper` interface in `internal/dumper/`:

```go
type Dumper interface {
    Dump(ctx context.Context, database string) (io.ReadCloser, error)
    Name() string
}
```

### Add a new storage backend

Implement the `Storage` interface in `internal/storage/`:

```go
type Storage interface {
    Upload(ctx context.Context, key string, reader io.Reader) error
    List(ctx context.Context, prefix string) ([]Object, error)
    Delete(ctx context.Context, keys []string) error
    Name() string
}
```

### Add a new notification channel

Implement the `Notifier` interface in `internal/notifier/`:

```go
type Notifier interface {
    Notify(ctx context.Context, msg Message) error
    Name() string
}
```

## License

MIT — see [LICENSE](LICENSE).
