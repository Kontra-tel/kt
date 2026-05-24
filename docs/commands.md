# Commands

## kt init

Scaffold a new project from a template.

```bash
kt init <template> <app> [options]
```

| Option | Default | Description |
| --- | --- | --- |
| `--dir` | `.` | Target directory |
| `--port` | `8080` | Primary listen port |
| `--user` | `<app>` | systemd service user |
| `--group` | `<user>` | systemd service group |
| `--author` | git config user.name + email | Package maintainer name |
| `--force` | `false` | Overwrite existing files |

```bash
kt init java-service my-api --port 4002 --user myapp --group myapp
kt init go-cli my-tool --author "Alice"
kt init multi-service my-app --port 4002 --force
```

## kt templates

List all available project templates.

```bash
kt templates
```

## kt install-tools / kt update-tools

Install or update the shared `.kt/mk/` and `.kt/scripts/` tooling into the current project directory.

```bash
kt install-tools [--dir .] [--force]
kt update-tools  [--dir .] [--force]
```

Run `kt update-tools` in an existing project after upgrading `kt` to pull in the latest shared Makefile includes.

## kt config

Manage config files under `deploy/config/`.

```bash
kt config init    # copy *.example files to actual config (no-clobber)
kt config check   # exit 1 if any config file derived from an example is missing
kt config diff    # diff each *.example against its actual counterpart
```

These delegate to the `config-init`, `config-check`, and `config-diff` Make targets.

## kt release

Bump `version.txt` and print the new version.

```bash
kt release patch   # 1.2.3 -> 1.2.4
kt release minor   # 1.2.3 -> 1.3.0
kt release major   # 1.2.3 -> 2.0.0
```

## kt update

Update `kt` itself to the latest release.

```bash
kt update          # check and apply
kt update --check  # check only, exits 1 if a newer version is available
```

Automatically re-runs with `sudo` if the install location requires elevated permissions.
Has no effect on dev builds.

## kt doctor

Check that all required tools listed in `DOCTOR_TOOLS` are installed and on PATH.

```bash
kt doctor
```

Equivalent to running `make doctor`.

## kt version

Print the current `kt` version, commit, and build date.

```bash
kt version
```
