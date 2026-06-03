APP_NAME := dbbackup
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X github.com/thanhtaivtt/dbbackup/cmd.Version=$(VERSION) -X github.com/thanhtaivtt/dbbackup/cmd.Commit=$(COMMIT) -X github.com/thanhtaivtt/dbbackup/cmd.Date=$(DATE)

.PHONY: build test lint vet clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(APP_NAME) .

test:
	go test -race ./...

vet:
	go vet ./...

lint:
	golangci-lint run

clean:
	rm -f $(APP_NAME)
