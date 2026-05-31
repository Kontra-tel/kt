# Templates

## Available templates

| Template | Description |
| --- | --- |
| `app` | Language-agnostic application — service, daemon, or CLI |
| `multi` | Two-service application packaged as a single unit (backend + frontend) |

## Template variables

| Variable | Default | Description |
| --- | --- | --- |
| `{{.App}}` | *(required)* | Application name |
| `{{.Template}}` | *(required)* | Template name |
| `{{.Author}}` | git config user.name + email | Package maintainer name |
| `{{.ServiceUser}}` | `<app>` | systemd service user |
| `{{.ServiceGroup}}` | `<user>` | systemd service group |

## Deploy folder layout

The `deploy/` directory controls what nFPM installs and where. Remove any sub-directory you don't need:

| Directory | Installed to | Purpose |
| --- | --- | --- |
| `deploy/bin/<app>` | `/usr/bin/<app>` | Launch script — invoked by systemd and directly from the shell |
| `dist/app/` | `/usr/lib/<app>/` | Application artifacts |
| `deploy/systemd/<app>.service` | `/usr/lib/systemd/system/` | Packaged systemd service unit |
| `deploy/config/*.example` | `/etc/<app>/` | Runtime config examples (all formats) |
| `deploy/scripts/postinstall.sh` | nFPM hook | Creates the service user and writable directories, copies config, and reloads systemd |
| `deploy/scripts/preremove.sh` | nFPM hook | Project-specific removal hook placeholder |

Service templates also create writable runtime directories:

```text
/var/lib/<app>/    mutable service data and working directory
/var/log/<app>/    service logs when not using the journal
```

Package files belong under `/usr/lib/<app>/`. Administrator-managed unit
overrides belong under `/etc/systemd/system/`, not in the package.

### CLI-only apps

Remove `deploy/systemd/` and `deploy/scripts/`, then delete the corresponding
blocks from `nfpm.yaml`. The launch script in `deploy/bin/` is still packaged
to `/usr/bin/<app>` so the binary is on PATH.

### Service-only apps (no CLI entry point)

Remove `deploy/bin/` and its nfpm.yaml content block. Update `ExecStart` in
the service unit to invoke the binary or interpreter directly.

## File renames

Files ending in `.tmpl` are rendered through Go `text/template` and written without the suffix.
The following path segments are also renamed at scaffold time:

| Source pattern | Output |
| --- | --- |
| `deploy/bin/app` | `deploy/bin/<app>` |
| `deploy/bin/app-backend` | `deploy/bin/<app>-backend` |
| `deploy/bin/app-frontend` | `deploy/bin/<app>-frontend` |
| `deploy/systemd/app.service.tmpl` | `deploy/systemd/<app>.service` |
| `deploy/systemd/backend.service.tmpl` | `deploy/systemd/<app>-backend.service` |
| `deploy/systemd/frontend.service.tmpl` | `deploy/systemd/<app>-frontend.service` |
| `.gitignore.tmpl` | `.gitignore` |

## Scaffolded structure

Every template produces:

```text
Makefile
nfpm.yaml
version.txt
.gitignore
.kt/
  project.yaml
  mk/
  scripts/
deploy/
  bin/
    <app>              # launch script — edit to set your runtime command
  systemd/
    <app>.service
  config/
    app.env.example
  scripts/
    postinstall.sh
    preremove.sh
```

The `multi` template additionally produces:

```text
deploy/
  bin/
    <app>-backend
    <app>-frontend
  systemd/
    <app>-backend.service
    <app>-frontend.service
```

## Launch script

`deploy/bin/<app>` is a thin shell wrapper installed to `/usr/bin/<app>`.
It is the single point of invocation for both systemd and manual use:

```sh
# Edit this line to match your runtime:
exec /usr/lib/myapp/myapp "$@"                                        # native binary
exec java ${JAVA_OPTS:--Xmx512m} -jar /usr/lib/myapp/myapp.jar "$@"  # Java
exec /usr/bin/node /usr/lib/myapp/server/index.mjs "$@"               # Node (Nuxt 3)
exec /usr/bin/python3 /usr/lib/myapp/main.py "$@"                     # Python
```

The systemd unit simply does `ExecStart=/usr/bin/<app>` — interpreter-specific
flags belong in the launch script, not the unit file.

## Examples

### Simple service

```bash
kt init app my-api
cd my-api
# Edit deploy/bin/my-api — set exec to your binary or interpreter
# Edit Makefile — add your build command
make doctor
make config-init
make build
make package
```

### CLI tool (no daemon)

```bash
kt init app my-tool
cd my-tool
# Delete deploy/systemd/ and deploy/scripts/
# Remove the systemd and scripts sections from nfpm.yaml
# Edit deploy/bin/my-tool — set exec to your binary
make build
make package
```

### Multi-service app

```bash
kt init multi my-platform
cd my-platform
# Edit deploy/bin/my-platform-backend and deploy/bin/my-platform-frontend
make build
make package
```

## Config files

Config example files live in `deploy/config/` and are named `<name>.<format>.example`:

```text
deploy/config/app.env.example         # KEY=VALUE env vars (default)
deploy/config/app.yaml.example        # YAML
deploy/config/app.ini.example         # INI
deploy/config/app.properties.example  # Java-style properties
```

Rules:

- All `*.example` files are **tracked in git** and packaged under `/etc/<app>/`.
- All non-`.example` files in `deploy/config/` are **gitignored** (actual runtime config).
- `make config-init` copies every `*.example` → the same name without `.example`, without overwriting existing files.
- `postinstall.sh` does the same on the target machine after package install.

The systemd `EnvironmentFile` directive only supports `KEY=VALUE` format. For structured config (YAML/INI/properties), have your application read the file directly and pass its path via an env var in `app.env` or an argument in the launch script.

## systemd hardening

All generated service units include:

```ini
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/var/lib/<app> /var/log/<app>
```

Adjust `ReadWritePaths` if the service writes data elsewhere.

## Service activation

Generated package hooks reload systemd but do not automatically enable,
restart, stop, or disable services. Package managers may run hooks during both
installation and upgrade, while applications may need migration and health
check steps before a restart.

Handle activation in your deployment process:

```bash
sudo systemctl enable --now <app>
sudo systemctl restart <app>
sudo systemctl status <app>
```

## Adding a new template

1. Create a directory under `internal/assets/templates/projects/<name>/`
2. Add a `template.yaml` with `name` and `description` fields
3. Add `.tmpl` files — they have access to all template variables above
4. Run `go build ./...` to embed the new template into the binary

`kt init <name> <app>` will pick it up automatically.
