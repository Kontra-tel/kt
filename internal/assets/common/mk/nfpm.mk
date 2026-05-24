NFPM_CONFIG ?= nfpm.yaml
NFPM_PACKAGER ?= deb

.PHONY: package

package: build ## Build package with nFPM
	@mkdir -p dist
	nfpm package -f $(NFPM_CONFIG) -p $(NFPM_PACKAGER) -t dist/
