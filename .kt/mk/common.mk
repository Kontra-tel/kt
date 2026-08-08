SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

APP ?= $(shell kt config get app 2>/dev/null)
# Releases derive their version from an exact v<semver> tag. Override VERSION for local builds.
VERSION ?= $(shell tag=$$(git describe --tags --exact-match HEAD 2>/dev/null || true); if [[ $$tag =~ ^v ]]; then printf '%s' "$${tag#v}"; else printf '0.0.0-dev.%s' "$$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"; fi)
DIST_DIR ?= dist
DEPLOY_DIR ?= deploy

.PHONY: help env-print clean

help:
	@echo "Available targets:"
	@grep -hE '^[a-zA-Z0-9_.-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-24s %s\n", $$1, $$2}'

env-print: ## Print resolved variables
	@echo "APP=$(APP)"
	@echo "VERSION=$(VERSION)"
	@echo "DIST_DIR=$(DIST_DIR)"
	@echo "DEPLOY_DIR=$(DEPLOY_DIR)"

clean: ## Remove build output
	rm -rf $(DIST_DIR)
