NFPM_CONFIG ?= nfpm.yaml
# Auto-detect package format from the system package manager if not overridden
NFPM_PACKAGER ?= $(shell command -v dpkg >/dev/null 2>&1 && echo deb || (command -v rpm >/dev/null 2>&1 && echo rpm || (command -v pacman >/dev/null 2>&1 && echo archlinux || echo deb)))
# Map uname -m to the arch names nFPM expects
NFPM_ARCH ?= $(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/;s/armv7l/armv7/')

.PHONY: package package-deb package-rpm local-install

package: build build-metadata ## Build package with nFPM (override: NFPM_PACKAGER=deb|rpm|archlinux)
	@mkdir -p dist
	APP=$(APP) ARCH=$(NFPM_ARCH) VERSION=$(VERSION) nfpm package -f $(NFPM_CONFIG) -p $(NFPM_PACKAGER) -t dist/

package-deb: build build-metadata ## Build .deb package explicitly
	@mkdir -p dist
	APP=$(APP) ARCH=$(NFPM_ARCH) VERSION=$(VERSION) nfpm package -f $(NFPM_CONFIG) -p deb -t dist/

package-rpm: build build-metadata ## Build .rpm package explicitly
	@mkdir -p dist
	APP=$(APP) ARCH=$(NFPM_ARCH) VERSION=$(VERSION) nfpm package -f $(NFPM_CONFIG) -p rpm -t dist/

local-install: package ## Install built package on this machine
	@if [ "$(NFPM_PACKAGER)" = "rpm" ]; then sudo rpm -U $$(ls dist/*.rpm | tail -1); \
	elif [ "$(NFPM_PACKAGER)" = "archlinux" ]; then sudo pacman -U $$(ls dist/*.pkg.tar.zst | tail -1); \
	else sudo dpkg -i $$(ls dist/*.deb | tail -1); fi
