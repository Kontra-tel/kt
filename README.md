# kt / build-tools

`kt` is a tiny scaffolding tool for projects that use:

- Make for local build/test/package commands
- nFPM for `.deb` and `.rpm` packages
- systemd for services
- Gitea Actions for manual releases
- Cockpit or systemctl/journalctl for operations

It is intentionally **not** a deployment framework. It creates and updates project structure and local tooling, then normal Linux tools do the real work.

## Install

```bash
go build -o kt ./cmd/kt
sudo install -m 755 kt /usr/local/bin/kt
```

Or from this repo:

```bash
make install
```

## Commands

```bash
kt templates
kt init <template> <app> [--dir .] [--port 8080] [--user app] [--group app] [--author Name] [--force]
kt install-tools [--dir .] [--force]
kt update-tools  [--dir .] [--force]
kt config init
kt config diff
kt config check
kt release patch
kt release minor
kt release major
kt update
kt update --check
kt doctor
kt version
```

## Templates

Templates live under:

```text
internal/assets/templates/projects/<template-name>/
```

Currently included:

| Template | Description |
| --- | --- |
| `go-cli` | Go CLI packaged with nFPM |
| `java-service` | Java systemd service packaged with nFPM |
| `node-service` | Node.js systemd service packaged with nFPM (Nuxt 3) |
| `multi-service` | Java backend + Node.js frontend as a single package |

Files ending in `.tmpl` are rendered with Go `text/template` and written without the `.tmpl` suffix.

Available template variables:

```text
{{.App}}           application name
{{.Template}}      template name
{{.Author}}        maintainer name (default: Kontra)
{{.Port}}          primary port (default: 8080)
{{.ServiceUser}}   systemd service user (default: <app>)
{{.ServiceGroup}}  systemd service group (default: <user>)
```

Special file renames applied during scaffolding:

```text
app.service.tmpl       -> <app>.service
backend.service.tmpl   -> <app>-backend.service
frontend.service.tmpl  -> <app>-frontend.service
cmd/app/               -> cmd/<app>/
.gitignore.tmpl        -> .gitignore
```

Every scaffolded project includes:

```text
Makefile
nfpm.yaml
version.txt
.gitignore
deploy/
  config/
    app.env.example   (commented, copy to app.env)
.kt/
  mk/                 (shared Makefile includes)
  scripts/            (shared helper scripts)
```

Service templates (`java-service`, `node-service`, `multi-service`) additionally include:

```text
deploy/
  scripts/
    postinstall.sh
    preremove.sh
  systemd/
    <app>.service     (multi-service generates -backend and -frontend variants)
```

The `go-cli` template additionally generates `go.mod` and `cmd/<app>/main.go`.

## Create a Java service

```bash
kt init java-service kontra-api --port 4002 --user kontra --group kontra
cd kontra-api
make doctor
make config-init   # creates deploy/config/app.env from example
make build
make install       # detects dpkg/rpm/pacman and installs
```

## Create a multi-service app

Useful for apps like a Java backend + Node.js frontend pair.

```bash
kt init multi-service knetlog --port 4002 --user kontra --group kontra
cd knetlog
make build
make install
```

## Packaging

nFPM is used for packaging. Package format and architecture are auto-detected but can be overridden:

```bash
make package                    # auto-detects deb or rpm from system
make package-deb                # always build .deb
make package-rpm                # always build .rpm
NFPM_PACKAGER=rpm make package  # override at call site
NFPM_ARCH=arm64 make package    # cross-package for arm64
make local-install              # build then install with the right package manager
```

The `ARCH` environment variable is passed to nFPM and must match the binary you built.
The `arch` field in `nfpm.yaml` is `${ARCH}` — it is always set by the Makefile from `NFPM_ARCH`.

### Supported package managers

| Platform | Packager | Format |
| --- | --- | --- |
| Debian / Ubuntu | `dpkg` | `.deb` |
| RHEL / Fedora / openSUSE | `rpm` | `.rpm` |
| Arch Linux | `pacman` | `.pkg.tar.zst` (build only) |

## Config management

Generated projects use `deploy/config/app.env.example` as a documented reference.
Actual config lives at `deploy/config/app.env` (gitignored).

```bash
make config-init   # create app.env from example (no-clobber)
make config-check  # fail if app.env is missing
make config-diff   # diff example vs actual
```

## systemd service hardening

All generated systemd service units include:

```ini
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/opt/<app>
```

Adjust `ReadWritePaths` if the service writes data outside its default install directory.

## Project philosophy

Per-project `Makefile` handles only:

```text
test / build / run
package / install
restart / logs / status
```

nFPM handles packaging and package lifecycle scripts (`postinstall`, `preremove`).
systemd handles runtime.
Gitea Actions handles manual release automation.

## Release kt

Trigger a release from Gitea's Actions UI:

```text
Actions → Release → Run workflow → choose bump type (patch / minor / major)
```

Workflow file: `.gitea/workflows/release.yaml`

The workflow:

1. Bumps `version.txt` by the chosen increment
2. Runs `go test ./...`
3. Builds cross-platform binaries via `scripts/release.sh`
4. Builds Linux packages (`.deb` + `.rpm`) for amd64 and arm64
5. Commits and tags the version bump
6. Creates a Gitea release with all artifacts

Release artifacts:

```text
dist/kt-linux-amd64
dist/kt-linux-arm64
dist/kt-darwin-amd64
dist/kt-darwin-arm64
dist/kt_<version>_amd64.deb
dist/kt_<version>_arm64.deb
dist/kt-<version>.amd64.rpm
dist/kt-<version>.arm64.rpm
dist/SHA256SUMS
```

Build binaries locally without the workflow:

```bash
make release   # produces dist/kt-* binaries and SHA256SUMS only, no packages
```

## Updating kt

```bash
kt update          # check for a newer release and apply it
kt update --check  # check only, exits 1 if a newer version exists
```

`kt update` downloads the release binary for the current OS and architecture from Gitea and atomically replaces the running executable. Dev builds (`version = "dev"`) skip the check.

## Updating local tooling

After upgrading `kt`, update the `.kt/` tooling in each project:

```bash
kt update-tools
```

This overwrites `.kt/mk/` and `.kt/scripts/` with the latest versions from the installed `kt` binary.

## Template layout

Built-in templates and shared tooling are embedded in the binary at build time:

```text
internal/assets/
  common/
    mk/
      common.mk
      config.mk
      doctor.mk
      nfpm.mk
      version.mk
    scripts/
      postinstall-systemd.sh
      preremove-systemd.sh
  templates/
    projects/
      go-cli/
      java-service/
      node-service/
      multi-service/
```

To add a new project template, create a directory under `internal/assets/templates/projects/<name>/`.
`kt init <name> <app>` will pick it up automatically.

The `deploy/` folder at the repository root is only for packaging the `kt` CLI itself and is not related to the templates.
