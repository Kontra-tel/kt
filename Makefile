VERSION ?= $(shell tag=$$(git describe --tags --exact-match HEAD 2>/dev/null || true); if [ -n "$$tag" ]; then printf '%s' "$${tag#v}"; else printf '0.0.0-dev.%s' "$$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"; fi)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

.PHONY: build install test release clean

build:
	go build -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o kt ./cmd/kt

install: build
	sudo install -m 755 kt /usr/local/bin/kt
	sudo install -D -m 644 docs/man/kt.1 /usr/local/share/man/man1/kt.1

test:
	go test ./...

release:
	VERSION="$(VERSION)" ./scripts/release.sh

clean:
	rm -rf dist kt
