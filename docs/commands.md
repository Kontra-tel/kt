# Commands

## kt init

Scaffold a new project from a template. Running without arguments starts an interactive prompt to choose a template and enter an app name.

```bash
kt init                                              # interactive
kt init <template> <app> [options]                   # explicit
```

| Option | Default | Description |
| --- | --- | --- |
| `--dir` | `.` | Target directory |
| `--force` | `false` | Overwrite existing files |

```bash
kt init service my-api
kt init cli my-tool
kt init mixed my-suite
kt init multi my-platform
kt init cli my-tool --dir /srv/projects
```

Package maintainer is derived from git config automatically. Service templates also derive service user and group from the app name. Edit the generated files directly if you need to override them.

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

`kt config` has two responsibilities: reading and writing the project's `.kt/project.yaml`, and managing runtime config files under `deploy/config/`.

### Project config

`.kt/project.yaml` is created by `kt init` and holds the project contract used by Make, nFPM, and `kt` itself. The key fields are:

- `template`: scaffold template name as chosen by the user
- `app`: package / application name
- `kind`: `cli`, `service`, `mixed`, or `multi-service`
- `services`: comma-separated packaged service names, blank for `cli`
- `user` / `group`: service account values for service-bearing templates

```bash
kt config show              # print all keys and values
kt config show --json       # print normalized project contract as JSON
kt config shape             # print kind-aware summary from .kt/project.yaml
kt config get <key>         # print a single value (used by Makefile: APP := $(shell kt config get app))
kt config set <key> <value> # update a value in .kt/project.yaml
```

### Deploy config

```bash
kt config init    # copy deploy/config/*.example files to actual config (no-clobber)
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
kt release set 2.0.0-rc.1
```

## kt update

Update `kt` itself to the latest release.

```bash
kt update          # check and apply
kt update --check  # check only; also informs about newer prereleases
kt update --prerelease
```

Automatically re-runs with `sudo` if the install location requires elevated permissions.
Has no effect on dev builds.
By default, `kt update` only installs stable releases. `kt update --check`
still informs you when a newer prerelease exists. Use `--prerelease` to opt
into downloading prerelease versions such as `2.0.0-rc.1`. `--check` and
`--prerelease` are intentionally separate modes.

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
