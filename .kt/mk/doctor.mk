DOCTOR_TOOLS ?= git make nfpm

.PHONY: doctor

doctor: ## Check required tools
	@ok=1; \
	for t in $(DOCTOR_TOOLS); do \
		if command -v $$t >/dev/null 2>&1; then echo "✓ $$t"; else echo "✗ missing $$t"; ok=0; fi; \
	done; \
	test $$ok -eq 1
