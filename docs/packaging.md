# Packaging

## nFPM

All templates use [nFPM](https://nfpm.goreleaser.com) for packaging. The `nfpm.yaml` in each project controls what goes into the package.

The `arch` field uses `${ARCH}` and the `version` field uses `${VERSION}` — both are set automatically by the Makefile when you run `make package`.

## Make targets

```bash
make package                     # build package, format and arch auto-detected
make package-deb                 # always build .deb
make package-rpm                 # always build .rpm
make local-install               # build then install with the right package manager
NFPM_PACKAGER=rpm make package   # override format at call site
NFPM_PACKAGER=archlinux make package
NFPM_ARCH=arm64 make package     # cross-package for a different architecture
```

## Supported package managers

Auto-detection checks for `dpkg`, then `rpm`, then `pacman` on PATH and picks accordingly.

| Platform | Format | NFPM_PACKAGER |
| --- | --- | --- |
| Debian / Ubuntu | `.deb` | `deb` |
| RHEL / Fedora / openSUSE | `.rpm` | `rpm` |
| Arch Linux | `.pkg.tar.zst` | `archlinux` |

## Config management

Each project ships a documented `deploy/config/app.env.example`. The actual config file (`app.env`) is gitignored and must be created on each machine.

```bash
make config-init    # copy *.example → actual config (will not overwrite)
make config-check   # fail if any config file is missing (useful in CI)
make config-diff    # diff example vs actual to spot drift
```

Config files are placed under `deploy/config/`. When packaged with nFPM, the example is installed to `/etc/<app>/app.env.example` with `type: config|noreplace` so upgrades never overwrite a live config.

The postinstall script copies the example to `app.env` on first install if the file does not already exist.

## Filesystem layout

Generated service packages use a distribution-neutral Linux layout:

```text
/etc/<app>/                  runtime configuration
/usr/lib/<app>/              packaged application artifacts
/usr/lib/systemd/system/     packaged systemd service units
/var/lib/<app>/              mutable service data
/var/log/<app>/              service logs when not using the journal
```

Use `/etc/systemd/system/` for administrator-managed overrides, not packaged
units.

## Service lifecycle

Generated hooks create writable directories, initialize config without
overwriting existing files, and reload systemd. They deliberately do not
enable, restart, stop, or disable services.

Use deployment automation to perform upgrade-specific migrations, restart the
service, verify health, and roll back to a known-good package when necessary.

## Publishing

Package publishing is separate from scaffolding. Use the registry and package
format appropriate for your environment. For an opt-in Gitea Debian package
registry example, see [Gitea Debian publishing](gitea-debian.md).
