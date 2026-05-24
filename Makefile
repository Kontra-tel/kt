VERSION := $(shell cat version.txt)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

.PHONY: build install test release clean

build:
	go build -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o kt ./cmd/kt

install: build
	sudo install -m 755 kt /usr/local/bin/kt

test:
	go test ./...

release:
	./scripts/release.sh

clean:
	rm -rf dist kt
