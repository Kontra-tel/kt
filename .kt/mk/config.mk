CONFIG_DIR ?= deploy/config
CONFIG_EXAMPLE_SUFFIX ?= .example

.PHONY: config-init config-check config-diff

config-init: ## Create config files from *.example without overwriting
	@find $(CONFIG_DIR) -name '*$(CONFIG_EXAMPLE_SUFFIX)' -print | while read example; do \
		target="$${example%$(CONFIG_EXAMPLE_SUFFIX)}"; \
		if [ -f "$$target" ]; then echo "exists: $$target"; else cp "$$example" "$$target"; echo "created: $$target"; fi; \
	done

config-check: ## Fail if configs based on examples are missing
	@missing=0; \
	find $(CONFIG_DIR) -name '*$(CONFIG_EXAMPLE_SUFFIX)' -print | while read example; do \
		target="$${example%$(CONFIG_EXAMPLE_SUFFIX)}"; \
		if [ ! -f "$$target" ]; then echo "missing: $$target"; missing=1; fi; \
	done; exit $$missing

config-diff: ## Diff *.example files against actual config files
	@find $(CONFIG_DIR) -name '*$(CONFIG_EXAMPLE_SUFFIX)' -print | while read example; do \
		target="$${example%$(CONFIG_EXAMPLE_SUFFIX)}"; \
		if [ -f "$$target" ]; then echo "=== $$example -> $$target ==="; diff -u "$$example" "$$target" || true; else echo "missing: $$target"; fi; \
	done
