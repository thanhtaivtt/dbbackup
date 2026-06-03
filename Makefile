APP_NAME := dbbackup
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X github.com/thanhtaivtt/dbbackup/cmd.Version=$(VERSION) -X github.com/thanhtaivtt/dbbackup/cmd.Commit=$(COMMIT) -X github.com/thanhtaivtt/dbbackup/cmd.Date=$(DATE)

.PHONY: build test lint vet clean release

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

release:
	@latest=$$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//'); \
	if [ -z "$$latest" ]; then next="0.1.0"; \
	else next=$$(echo $$latest | awk -F. '{print $$1"."$$2"."$$3+1}'); fi; \
	if [ -n "$(v)" ]; then next=$(v); fi; \
	echo "Releasing v$$next..."; \
	git tag -a "v$$next" -m "Release v$$next" && \
	git push origin "v$$next"
