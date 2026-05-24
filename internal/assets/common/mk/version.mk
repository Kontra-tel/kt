VERSION_FILE ?= version.txt
VERSION ?= $(shell test -f $(VERSION_FILE) && cat $(VERSION_FILE) || echo 0.1.0)

.PHONY: version version-init version-patch version-minor version-major version-set version-tag

version: ## Print current version
	@cat $(VERSION_FILE)

version-init: ## Create version.txt if missing
	@test -f $(VERSION_FILE) || echo "0.1.0" > $(VERSION_FILE)
	@cat $(VERSION_FILE)

version-patch: ## Bump patch version
	@awk -F. '{printf "%d.%d.%d\n", $$1, $$2, $$3+1}' $(VERSION_FILE) > $(VERSION_FILE).tmp
	@mv $(VERSION_FILE).tmp $(VERSION_FILE)
	@echo "New version: $$(cat $(VERSION_FILE))"

version-minor: ## Bump minor version
	@awk -F. '{printf "%d.%d.0\n", $$1, $$2+1}' $(VERSION_FILE) > $(VERSION_FILE).tmp
	@mv $(VERSION_FILE).tmp $(VERSION_FILE)
	@echo "New version: $$(cat $(VERSION_FILE))"

version-major: ## Bump major version
	@awk -F. '{printf "%d.0.0\n", $$1+1}' $(VERSION_FILE) > $(VERSION_FILE).tmp
	@mv $(VERSION_FILE).tmp $(VERSION_FILE)
	@echo "New version: $$(cat $(VERSION_FILE))"

version-set: ## Set VERSION=x.y.z
	@test -n "$(VERSION)" || (echo "VERSION required"; exit 1)
	@echo "$(VERSION)" > $(VERSION_FILE)
	@echo "Set version to $(VERSION)"

version-tag: ## Create git tag v<version>
	@git tag -a "v$$(cat $(VERSION_FILE))" -m "Release v$$(cat $(VERSION_FILE))"
