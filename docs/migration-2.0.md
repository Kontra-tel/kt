# 2.0 migration

Use this guide when validating or adopting `2.0.0-rc.1` and later `2.0.x`
releases. `kt` 2.0 changes both the scaffold contract and the default service
packaging shape.

## Breaking changes at a glance

- `service` is now the preferred name for single-service scaffolds
- `app` is now a legacy alias for that same scaffold
- pure service packages no longer use `/usr/bin/<app>` as `ExecStart`
- `.kt/project.yaml` now explicitly carries `kind` and `services`
- prerelease updater opt-in now requires `kt update --prerelease`

## Template naming

- New projects should prefer `kt init service <app>`.
- `kt init app <app>` still works, but `app` is now the legacy name for the same single-service scaffold.
- New first-class shapes are:
  - `cli`
  - `mixed`
  - `multi`

## Project contract

New scaffolds write explicit shape metadata to `.kt/project.yaml`:

```yaml
template: service
app: my-api
kind: service
services: my-api
user: my-api
group: my-api
```

Existing projects can add `kind` and `services` manually if they want `kt config show --json` and `kt config shape` to reflect the exact scaffold shape. `kt` also derives sensible defaults for older projects that do not have those keys yet.

## Service runner split

Before 2.0, service templates used `/usr/bin/<app>` as both:

- the user-facing command
- the systemd `ExecStart` target

2.0 separates those concerns:

- `/usr/bin/<app>` now prints package metadata for pure service packages
- `/usr/lib/<app>/bin/<app>` is the service manager runtime entrypoint

For multi-service packages:

- `/usr/bin/<app>` prints metadata
- `/usr/lib/<app>/bin/<app>-backend` and `/usr/lib/<app>/bin/<app>-frontend` are the service runners

For mixed packages:

- `/usr/bin/<app>` remains the runnable CLI command
- `/usr/bin/<app>-service` prints service metadata
- `/usr/lib/<app>/bin/<app>-service` is the service runner

## Lifecycle hook extensions

Generated package hooks are still generic by default. 2.0 formalizes the local
extension point:

```text
/etc/<app>/hooks/postinstall.local.sh
/etc/<app>/hooks/preremove.local.sh
```

Scaffolded examples now live under `deploy/hooks-examples/`.

## Migrating an existing service project

1. Move the current runtime command from `deploy/bin/<app>` into `deploy/run/<app>`.
2. Change `deploy/bin/<app>` to a metadata command, or re-scaffold from `service` and port your runtime line over.
3. Update packaged paths in `nfpm.yaml` so the runner lands under `/usr/lib/<app>/bin/`.
4. Update `ExecStart` in systemd units to point at the new runner path.
5. Add `kind` and `services` to `.kt/project.yaml`.
6. If you want package-triggered lifecycle behavior, copy and adapt the sample hook scripts into `/etc/<app>/hooks/` on target hosts.
7. Build the package, inspect contents, and validate an upgrade on a test host before rolling out broadly.

Keep the previous package available so rollback can reinstall a known-good build if needed.

## RC validation checklist

Before shipping `2.0.0-rc.1` or promoting it to `2.0.0`, verify these on a
real host or CI runner:

1. `make test` passes
2. `kt init service`, `kt init cli`, `kt init mixed`, and `kt init multi` all scaffold successfully
3. each scaffold builds and packages cleanly
4. service packages install the expected runner paths under `/usr/lib/<app>/bin/`
5. metadata commands print the expected text and `--json` output
6. local hook examples can be copied into `/etc/<app>/hooks/` and invoked successfully when desired
