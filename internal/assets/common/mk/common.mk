SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

APP ?= app
VERSION_FILE ?= version.txt
VERSION ?= $(shell test -f $(VERSION_FILE) && cat $(VERSION_FILE) || echo 0.1.0)
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
