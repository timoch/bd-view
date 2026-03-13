BINARY := bd-view
ENTRY  := ./cmd/bd-view
PREFIX ?= $(HOME)/.local

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: build test vet lint clean install

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(ENTRY)

test:
	go test ./...

vet:
	go vet ./...

lint:
	@which golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

clean:
	rm -f $(BINARY)
	rm -rf dist/

install: build
	mkdir -p $(PREFIX)/bin
	cp -f $(BINARY) $(PREFIX)/bin/$(BINARY)
