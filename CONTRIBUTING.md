# Contributing

Thanks for your interest in contributing to dbbackup!

## Getting Started

```bash
git clone https://github.com/thanhtaivtt/dbbackup.git
cd dbbackup
go mod download
make build
```

## Development Workflow

1. Fork the repo and create a branch from `main`
2. Make your changes
3. Run checks before committing:

```bash
make vet
make lint
make test
```

4. Open a pull request against `main`

## Running Tests

```bash
# All tests
make test

# Single package
go test ./internal/retention/...

# With coverage
go test -race -coverprofile=coverage.out ./...
```

Some tests (e.g. dumper integration) require a running MySQL instance. CI provides one automatically. For local integration testing, set:

```bash
export TEST_MYSQL_DSN="root:password@tcp(localhost:3306)/testdb"
```

## Adding a New Storage Backend

1. Create `internal/storage/yourbackend.go` implementing the `Storage` interface
2. Add config struct to `internal/config/config.go`
3. Add validation for the new backend in `config.validate()`
4. Wire it up in `internal/engine/engine.go` switch on `cfg.Storage.Backend`
5. Update `config.example.toml` and `cmd/init.go` template

## Adding a New Notification Channel

1. Create `internal/notifier/yourchannel.go` implementing the `Notifier` interface
2. Add config struct to `internal/config/config.go`
3. Add validation when `enabled = true`
4. Wire it up in `internal/engine/engine.go` notifier initialization
5. Update `config.example.toml` and `cmd/init.go` template

## Adding a New Database Engine

1. Create `internal/dumper/yourengine.go` implementing the `Dumper` interface
2. Add config struct to `internal/config/config.go`
3. Wire it up in `internal/engine/engine.go` dumper selection

## Code Style

- Run `golangci-lint run` before submitting
- Keep functions focused and small
- Use `context.Context` for cancellation in I/O operations
- HTTP clients must have a timeout set
- Streaming with `io.Pipe()` is preferred over buffering entire payloads in memory

## Releases

Releases are automated via GoReleaser when a tag is pushed:

```bash
make release           # auto-increment patch version
make release v=1.2.0   # specific version
```

## Reporting Issues

Open an issue on GitHub with:

- What you expected to happen
- What actually happened
- Steps to reproduce
- dbbackup version (`dbbackup --version`)
