# Templates

## Available templates

| Template | Description |
| --- | --- |
| `generic-service` | Language-agnostic systemd service skeleton |
| `generic-cli` | Language-agnostic CLI binary skeleton |
| `go-cli` | Go CLI binary packaged with nFPM |
| `java-service` | Java systemd service packaged with nFPM |
| `node-service` | Node.js systemd service packaged with nFPM (Nuxt 3) |
| `multi-service` | Java backend + Node.js frontend as a single package |

## Template variables

| Variable | Default | Description |
| --- | --- | --- |
| `{{.App}}` | *(required)* | Application name |
| `{{.Template}}` | *(required)* | Template name |
| `{{.Author}}` | git config user.name + email | Package maintainer name |
| `{{.Port}}` | `8080` | Primary listen port |
| `{{.ServiceUser}}` | `<app>` | systemd service user |
| `{{.ServiceGroup}}` | `<user>` | systemd service group |

## File renames

Files ending in `.tmpl` are rendered through Go `text/template` and written without the suffix.
The following path segments are also renamed:

| Source | Output |
| --- | --- |
| `app.service.tmpl` | `<app>.service` |
| `backend.service.tmpl` | `<app>-backend.service` |
| `frontend.service.tmpl` | `<app>-frontend.service` |
| `cmd/app/` | `cmd/<app>/` |
| `.gitignore.tmpl` | `.gitignore` |

## Scaffolded structure

Every template produces:

```text
Makefile
nfpm.yaml
version.txt
.gitignore
deploy/
  config/
    app.env.example
.kt/
  mk/
  scripts/
```

Service templates (`java-service`, `node-service`, `multi-service`) additionally produce:

```text
deploy/
  scripts/
    postinstall.sh
    preremove.sh
  systemd/
    <app>.service
```

The `go-cli` template additionally produces `go.mod` and `cmd/<app>/main.go`.

The `generic-cli` template produces only the packaging skeleton (no systemd units or install scripts) — fill in the `build` target in the `Makefile` for your language/toolchain.

## Examples

### Java service

```bash
kt init java-service my-api --port 4002 --user myapp --group myapp
cd my-api
make doctor
make config-init   # creates deploy/config/app.env from the example
make build
make install
```

### Multi-service app

Scaffolds a Java backend and a Node.js frontend packaged and deployed as a single unit.

```bash
kt init multi-service my-app --port 4002
cd my-app
make build
make install
```

### Generic CLI

Language-agnostic skeleton — fill in the `build` target in `Makefile` for your toolchain.

```bash
kt init generic-cli my-tool
cd my-tool
# edit Makefile — add your build command
make build
make install
```

### Go CLI

```bash
kt init go-cli my-tool
cd my-tool
make build
make install
```

## Config files

Config example files live in `deploy/config/` and are named `<name>.<format>.example`, e.g.:

```text
deploy/config/app.env.example       # KEY=VALUE env vars (default)
deploy/config/app.yaml.example      # YAML
deploy/config/app.ini.example       # INI
deploy/config/app.properties.example  # Java-style properties
```

Rules:

- All `*.example` files are **tracked in git** and packaged into the `.deb`/`.rpm` under `/etc/<app>/`.
- All non-`.example` files in `deploy/config/` are **gitignored** (actual runtime config).
- `make config-init` copies every `*.example` → the same name without `.example`, without overwriting existing files.
- `postinstall.sh` does the same on the target machine after package install.

The systemd `EnvironmentFile` directive only supports `KEY=VALUE` format. For YAML/INI/properties config, have your application read the file directly (e.g. pass the path via `ExecStart` argument or an env var pointing to the file).

## systemd hardening

All generated service units include:

```ini
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/opt/<app>
```

Adjust `ReadWritePaths` if the service writes data outside `/opt/<app>`.

The Node.js service unit includes comments showing the correct `ExecStart` path for common frameworks:

```ini
# Nuxt 3:  /usr/bin/node /opt/<app>/server/index.mjs
# Express: /usr/bin/node /opt/<app>/index.js
```

## Adding a new template

1. Create a directory under `internal/assets/templates/projects/<name>/`
2. Add a `template.yaml` with `name` and `description` fields
3. Add `.tmpl` files — they have access to all template variables above
4. Run `go build ./...` to embed the new template into the binary

`kt init <name> <app>` will pick it up automatically.
