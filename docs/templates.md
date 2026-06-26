# Templates

## Available templates

| Template | Description |
| --- | --- |
| `service` | Single-service application package |
| `app` | Single-service application package (legacy name; use `service`) |
| `cli` | Command-line application with no service unit |
| `mixed` | CLI application package with a companion service |
| `multi` | Multi-service application package (backend + frontend) |

## Project contract

Every scaffold writes `.kt/project.yaml`. `kt`, Make, and nFPM treat it as the
project contract.

| Key | Meaning |
| --- | --- |
| `template` | Template name as chosen by the user |
| `app` | Package / application name |
| `kind` | `cli`, `service`, `mixed`, or `multi-service` |
| `services` | Comma-separated packaged service names, blank for `cli` |
| `user` | Service user for service-bearing templates |
| `group` | Service group for service-bearing templates |

Use `kt config show --json` or `kt config shape` to inspect the normalized
contract from an existing project.

## App shapes

- `cli`: `/usr/bin/<app>` is the runnable command.
- `service`: `/usr/bin/<app>` prints package metadata, while systemd runs `/usr/lib/<app>/bin/<app>`.
- `mixed`: `/usr/bin/<app>` is the runnable CLI, while `/usr/bin/<app>-service` prints service metadata and systemd runs `/usr/lib/<app>/bin/<app>-service`.
- `multi`: `/usr/bin/<app>` prints package metadata, while systemd runs dedicated backend/frontend runners under `/usr/lib/<app>/bin/`.

This split keeps service packages safe to inspect manually while giving the
service manager a dedicated runtime entrypoint.

## Deploy layout

### `cli`

| Path | Installed to | Purpose |
| --- | --- | --- |
| `deploy/bin/<app>` | `/usr/bin/<app>` | Runnable user command |
| `deploy/config/*.example` | `/etc/<app>/` | Runtime config examples |
| `dist/app/` | `/usr/lib/<app>/` | Application artifacts |

### `service`

| Path | Installed to | Purpose |
| --- | --- | --- |
| `deploy/bin/<app>` | `/usr/bin/<app>` | Metadata command with `--json` support |
| `deploy/run/<app>` | `/usr/lib/<app>/bin/<app>` | Service runner |
| `deploy/systemd/<app>.service` | `/usr/lib/systemd/system/` | Packaged service unit |
| `deploy/config/*.example` | `/etc/<app>/` | Runtime config examples |
| `deploy/hooks-examples/*.sh` | not packaged | Example local lifecycle hooks |
| `deploy/scripts/postinstall.sh` | nFPM hook | Host prep + optional local extension |
| `deploy/scripts/preremove.sh` | nFPM hook | Generic removal hook + optional local extension |
| `dist/app/` | `/usr/lib/<app>/` | Application artifacts |

### `mixed`

| Path | Installed to | Purpose |
| --- | --- | --- |
| `deploy/bin/<app>` | `/usr/bin/<app>` | Runnable CLI command |
| `deploy/bin/<app>-service` | `/usr/bin/<app>-service` | Service metadata command with `--json` support |
| `deploy/run/<app>-service` | `/usr/lib/<app>/bin/<app>-service` | Service runner |
| `deploy/systemd/<app>-service.service` | `/usr/lib/systemd/system/` | Packaged service unit |
| `deploy/config/*.example` | `/etc/<app>/` | Runtime config examples |
| `deploy/hooks-examples/*.sh` | not packaged | Example local lifecycle hooks |
| `deploy/scripts/postinstall.sh` | nFPM hook | Host prep + optional local extension |
| `deploy/scripts/preremove.sh` | nFPM hook | Generic removal hook + optional local extension |
| `dist/app/` | `/usr/lib/<app>/` | Application artifacts |

### `multi`

| Path | Installed to | Purpose |
| --- | --- | --- |
| `deploy/bin/<app>` | `/usr/bin/<app>` | Metadata command with `--json` support |
| `deploy/run/<app>-backend` | `/usr/lib/<app>/bin/<app>-backend` | Backend runner |
| `deploy/run/<app>-frontend` | `/usr/lib/<app>/bin/<app>-frontend` | Frontend runner |
| `deploy/systemd/<app>-backend.service` | `/usr/lib/systemd/system/` | Backend unit |
| `deploy/systemd/<app>-frontend.service` | `/usr/lib/systemd/system/` | Frontend unit |
| `deploy/config/*.example` | `/etc/<app>/` | Runtime config examples |
| `deploy/hooks-examples/*.sh` | not packaged | Example local lifecycle hooks |
| `deploy/scripts/postinstall.sh` | nFPM hook | Host prep + optional local extension |
| `deploy/scripts/preremove.sh` | nFPM hook | Generic removal hook + optional local extension |
| `dist/app/` | `/usr/lib/<app>/` | Application artifacts |

Service-bearing templates also create:

```text
/var/lib/<app>/    mutable service data and working directory
/var/log/<app>/    service logs when not using the journal
/etc/<app>/hooks/  deployment-local lifecycle extension directory
```

## Hook extensions

Generated package hooks deliberately do not enable, restart, stop, or disable
services. If an environment needs that behavior, copy the scaffolded examples
from `deploy/hooks-examples/` to the target host:

```text
/etc/<app>/hooks/postinstall.local.sh
/etc/<app>/hooks/preremove.local.sh
```

The package hooks invoke them when present and pass these environment variables:

- `KT_APP`
- `KT_KIND`
- `KT_SERVICES`
- `KT_SERVICE_USER`
- `KT_SERVICE_GROUP`

Maintainer-script arguments are passed through unchanged.

## File renames

Files ending in `.tmpl` are rendered through Go `text/template` and written without the suffix.
The following path segments are also renamed at scaffold time:

| Source pattern | Output |
| --- | --- |
| `deploy/bin/app` | `deploy/bin/<app>` |
| `deploy/run/app` | `deploy/run/<app>` |
| `deploy/run/service` | `deploy/run/<app>-service` |
| `deploy/run/backend` | `deploy/run/<app>-backend` |
| `deploy/run/frontend` | `deploy/run/<app>-frontend` |
| `deploy/systemd/app.service.tmpl` | `deploy/systemd/<app>.service` |
| `deploy/systemd/service.service.tmpl` | `deploy/systemd/<app>-service.service` |
| `deploy/systemd/backend.service.tmpl` | `deploy/systemd/<app>-backend.service` |
| `deploy/systemd/frontend.service.tmpl` | `deploy/systemd/<app>-frontend.service` |
| `.gitignore.tmpl` | `.gitignore` |

## Examples

### Single service

```bash
kt init service my-api
cd my-api
make doctor
make build
make print-info
make package
```

### CLI app

```bash
kt init cli my-tool
cd my-tool
make build
make run
make package
```

### Mixed CLI + service

```bash
kt init mixed my-suite
cd my-suite
make build
make run
make print-info
make package
```

### Multi-service app

```bash
kt init multi my-platform
cd my-platform
make build
make print-info
make package
```

## Adding a new template

1. Create a directory under `internal/assets/templates/projects/<name>/`
2. Add a `template.yaml` with `name` and `description`
3. Add `.tmpl` files using the template variables available in the scaffold context
4. Run `go build ./...` to embed the new template into the binary

`kt init <name> <app>` will pick it up automatically.
