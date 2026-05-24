# kt / build-tools

`kt` is a tiny scaffolding tool for projects that use:

- Make for local build/test/package commands
- nFPM for `.deb` packages
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
kt init <template> <app>
kt install-tools
kt update-tools
kt config init
kt config diff
kt config check
kt release patch
kt release minor
kt release major
kt doctor
kt version
```

## Templates

Templates live under:

```text
assets/templates/projects/<template-name>/
```

Currently included:

```text
java-service
node-service
go-cli
multi-service
```

Add a new project template by creating another directory under `assets/templates/projects/`.
Files ending in `.tmpl` are rendered with Go `text/template` and written without the `.tmpl` suffix.

Available template variables:

```text
{{.App}}
{{.Template}}
{{.Author}}
{{.Port}}
{{.ServiceUser}}
{{.ServiceGroup}}
```

Special file renames:

```text
app.service.tmpl      -> <app>.service
backend.service.tmpl  -> <app>-backend.service
frontend.service.tmpl -> <app>-frontend.service
cmd/app/              -> cmd/<app>/
```

## Create a Java service

```bash
kt init java-service kontra-api --port 4002 --user kontra --group kontra
cd kontra-api
make doctor
make build
make package
sudo dpkg -i dist/*.deb
```

## Create a multi-service app

Useful for apps like a backend + frontend pair.

```bash
kt init multi-service knetlog --port 4002 --user kontra --group kontra
cd knetlog
make build
make package
sudo dpkg -i dist/*.deb
```

This creates:

```text
Makefile
nfpm.yaml
version.txt
.kt/
  mk/
  scripts/
deploy/
  config/
  scripts/
  systemd/
```

## Project philosophy

Per-project `Makefile` should only handle:

```text
test
build
package
install/restart/log helpers
```

nFPM handles packaging and package lifecycle scripts.
systemd handles runtime.
Gitea Actions handles manual release automation.

## Release kt

Manual release workflow:

```text
.gitea/workflows/release.yml
```

It builds:

```text
dist/kt-linux-amd64
dist/kt-linux-arm64
dist/kt-darwin-amd64
dist/kt-darwin-arm64
dist/SHA256SUMS
```

Run locally:

```bash
make release
```


## Template layout

Built-in templates and shared tooling live in one place only:

```text
internal/assets/
  common/
    mk/
    scripts/
  templates/
    projects/
      java-service/
      node-service/
      go-cli/
      multi-service/
```

To add a new project template, add a directory under `internal/assets/templates/projects/<name>/`.
`kt init <name> <app>` will then be able to use it.

The `deploy/` folder in this repository is only for packaging the `kt` CLI itself.
Application-specific deploy folders are generated into projects from templates.
