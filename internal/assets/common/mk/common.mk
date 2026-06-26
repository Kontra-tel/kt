SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

APP ?= $(shell kt config get app 2>/dev/null)
VERSION_FILE ?= version.txt
VERSION ?= $(shell test -f $(VERSION_FILE) && cat $(VERSION_FILE) || echo 0.1.0)
DIST_DIR ?= dist
DEPLOY_DIR ?= deploy
BUILD_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_FULL_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILD_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)
BUILD_DIRTY ?= $(shell test -n "$$(git status --porcelain 2>/dev/null)" && echo true || echo false)
PROJECT_TEMPLATE ?= $(shell kt config get template 2>/dev/null || echo unknown)
PROJECT_KIND ?= $(shell kt config get kind 2>/dev/null || echo unknown)
PROJECT_SERVICES ?= $(shell kt config get services 2>/dev/null || echo)
PROJECT_USER ?= $(shell kt config get user 2>/dev/null || echo)
PROJECT_GROUP ?= $(shell kt config get group 2>/dev/null || echo)
BUILD_HOST ?= $(shell hostname 2>/dev/null || echo unknown)
BUILD_OS ?= $(shell uname -s 2>/dev/null | tr '[:upper:]' '[:lower:]' || echo unknown)
BUILD_ARCH ?= $(shell uname -m 2>/dev/null || echo unknown)
BUILD_METADATA_DIR ?= $(DIST_DIR)/app/meta
BUILD_METADATA_FILE ?= $(BUILD_METADATA_DIR)/build.json

.PHONY: help env-print clean build-metadata

help:
	@echo "Available targets:"
	@grep -hE '^[a-zA-Z0-9_.-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-24s %s\n", $$1, $$2}'

env-print: ## Print resolved variables
	@echo "APP=$(APP)"
	@echo "VERSION=$(VERSION)"
	@echo "DIST_DIR=$(DIST_DIR)"
	@echo "DEPLOY_DIR=$(DEPLOY_DIR)"
	@echo "BUILD_COMMIT=$(BUILD_COMMIT)"
	@echo "BUILD_DATE=$(BUILD_DATE)"
	@echo "BUILD_BRANCH=$(BUILD_BRANCH)"
	@echo "BUILD_DIRTY=$(BUILD_DIRTY)"
	@echo "PROJECT_TEMPLATE=$(PROJECT_TEMPLATE)"
	@echo "PROJECT_KIND=$(PROJECT_KIND)"
	@echo "PROJECT_SERVICES=$(PROJECT_SERVICES)"
	@echo "PROJECT_USER=$(PROJECT_USER)"
	@echo "PROJECT_GROUP=$(PROJECT_GROUP)"

build-metadata: ## Write packaged build metadata to dist/app/meta/build.json
	@mkdir -p "$(BUILD_METADATA_DIR)"
	@APP="$(APP)" VERSION="$(VERSION)" BUILD_COMMIT="$(BUILD_COMMIT)" BUILD_FULL_COMMIT="$(BUILD_FULL_COMMIT)" BUILD_BRANCH="$(BUILD_BRANCH)" BUILD_DATE="$(BUILD_DATE)" BUILD_DIRTY="$(BUILD_DIRTY)" PROJECT_TEMPLATE="$(PROJECT_TEMPLATE)" PROJECT_KIND="$(PROJECT_KIND)" PROJECT_SERVICES="$(PROJECT_SERVICES)" PROJECT_USER="$(PROJECT_USER)" PROJECT_GROUP="$(PROJECT_GROUP)" BUILD_HOST="$(BUILD_HOST)" BUILD_OS="$(BUILD_OS)" BUILD_ARCH="$(BUILD_ARCH)" BUILD_METADATA_FILE="$(BUILD_METADATA_FILE)" python3 -c 'import json, os, pathlib; path = pathlib.Path(os.environ["BUILD_METADATA_FILE"]); services = [s.strip() for s in os.environ["PROJECT_SERVICES"].split(",") if s.strip()]; payload = {"app": os.environ["APP"], "version": os.environ["VERSION"], "template": os.environ["PROJECT_TEMPLATE"], "kind": os.environ["PROJECT_KIND"], "services": services, "service_user": os.environ["PROJECT_USER"], "service_group": os.environ["PROJECT_GROUP"], "commit": os.environ["BUILD_COMMIT"], "full_commit": os.environ["BUILD_FULL_COMMIT"], "branch": os.environ["BUILD_BRANCH"], "build_date": os.environ["BUILD_DATE"], "dirty": os.environ["BUILD_DIRTY"] == "true", "build_host": os.environ["BUILD_HOST"], "build_os": os.environ["BUILD_OS"], "build_arch": os.environ["BUILD_ARCH"]}; path.write_text(json.dumps(payload, indent=2) + "\n")'

clean: ## Remove build output
	rm -rf $(DIST_DIR)
