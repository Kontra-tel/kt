# kt

Tiny scaffolding tool for projects that use Make, nFPM, and systemd.

```bash
kt init                  # interactive — choose a template and name
kt init service my-api   # preferred single-service template
kt init cli my-tool      # command-line app
kt init mixed my-suite   # CLI + companion service
cd my-api
make doctor && make build && make install
```

## Install

**One-liner** (Linux and macOS):

```bash
curl -sL https://git.kontra.tel/kontra.tel/Kt/raw/branch/main/scripts/install.sh | bash
```

Installs the right binary for your OS and architecture to `/usr/local/bin/kt`.
Override the destination with `BIN_DIR=~/bin bash <(curl ...)`.

**With Go:**

```bash
go install git.kontra.tel/kontra.tel/Kt/cmd/kt@latest
```

**From source:**

```bash
git clone https://git.kontra.tel/kontra.tel/Kt
cd kt
make install
```

**Binary releases** are on the [releases page](https://git.kontra.tel/kontra.tel/Kt/releases).

## What it does

`kt init` scaffolds a project with:

- A `Makefile` wired to shared `.kt/mk/` includes for building, packaging, versioning, and config
- A `.kt/project.yaml` that stores the app name and other metadata — read by Make and `kt config`
- An `nfpm.yaml` for building `.deb`, `.rpm`, and Arch Linux packages with [nFPM](https://nfpm.goreleaser.com)
- systemd service units with security hardening (for service templates)
- `deploy/` scripts for install and removal lifecycle
- A `.gitignore` and `version.txt`

It is **not** a deployment framework — it sets up the structure, then normal Linux tools do the rest.

## Available templates

| Template | Description |
| --- | --- |
| `service` | Single-service application package |
| `app` | Single-service application package (legacy name; use `service`) |
| `cli` | Command-line application with no service unit |
| `mixed` | CLI application package with a companion service |
| `multi` | Multi-service application package (backend + frontend) |

```bash
kt templates   # list all available templates
```

## 1.3 changes

`1.3.0` changes the public scaffold model in a few important ways:

- `service` is now the preferred single-service template name
- `app` still works, but it is now the legacy name for that same scaffold
- pure service packages no longer use `/usr/bin/<app>` as the systemd runtime entrypoint
- `.kt/project.yaml` now carries explicit `kind` and `services` fields
- prerelease updates require `kt update --prerelease`

If you are upgrading an existing project, read [1.3 migration](docs/migration-1.3.md).

## Self-update

```bash
kt update         # update kt to the latest release
kt update --check # check only, exits 1 if a stable or prerelease update is available
kt update --prerelease
```

## Documentation

- [Commands](docs/commands.md)
- [Templates](docs/templates.md)
- [Packaging](docs/packaging.md)
- [1.3 migration](docs/migration-1.3.md)
- [Filesystem layout migration](docs/migration-fhs.md)
- [Release & maintenance](docs/release.md)
