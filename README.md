# kt

Tiny scaffolding tool for projects that use Make, nFPM, and systemd.

```bash
kt init                      # interactive — choose a template and name
kt init java-service my-api  # or explicit
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
GOPRIVATE=git.kontra.tel go install git.kontra.tel/kontra.tel/build-tools/cmd/kt@latest
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
- An `nfpm.yaml` for building `.deb` and `.rpm` packages with [nFPM](https://nfpm.goreleaser.com)
- systemd service units with security hardening (for service templates)
- `deploy/` scripts for install and removal lifecycle
- A `.gitignore` and `version.txt`

It is **not** a deployment framework — it sets up the structure, then normal Linux tools do the rest.

## Available templates

| Template | Description |
| --- | --- |
| `generic-service` | Language-agnostic systemd service skeleton |
| `generic-cli` | Language-agnostic CLI binary skeleton |
| `go-cli` | Go CLI binary packaged with nFPM |
| `java-service` | Java systemd service |
| `node-service` | Node.js systemd service (Nuxt 3) |
| `multi-service` | Java backend + Node.js frontend as one package |

```bash
kt templates   # list all available templates
```

## Self-update

```bash
kt update         # update kt to the latest release
kt update --check # check only, exits 1 if an update is available
```

## Documentation

- [Commands](docs/commands.md)
- [Templates](docs/templates.md)
- [Packaging](docs/packaging.md)
- [Release & maintenance](docs/release.md)
