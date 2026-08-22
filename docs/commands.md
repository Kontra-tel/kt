# Commands

## Global flags

Global flags may appear before the command or before a subcommand.

| Flag | Description |
| --- | --- |
| `--json` | Machine-readable output where supported (`config show`, `deploy`, `release plan`) |
| `--quiet`, `-q` | Suppress non-error status output |
| `--no-color` | Disable ANSI styling |
| `--color auto\|always\|never` | Control styling |

Unknown commands include a nearest-command suggestion when the typo is close enough.

## kt init

Scaffold a new project from a template. Running without arguments starts an interactive prompt to choose a template, enter an app name, confirm the target directory, and fill package/service defaults.

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

Package maintainer is derived from git config automatically. Service templates derive service user and group from the app name unless overridden in the interactive form.

## kt templates

List all available project templates.

```bash
kt templates
```

## kt install-tools / kt update-tools

Install or update the shared `.kt/mk/` tooling into a project directory.

```bash
kt install-tools [--dir .] [--force] [--check|--diff] [--apply]
kt update-tools  [--dir .] [--force] [--check|--diff] [--apply]
```

`kt install-tools` copies `.kt/mk/` into an existing directory without overwriting local files unless you pass `--force`.

Run `kt update-tools` in an existing project after upgrading `kt` to refresh the shared `.kt/mk/` files. Unlike `install-tools`, `update-tools` overwrites those shared files by default so the checked-in tooling actually updates.

Use `--check` in CI to fail when local `.kt/mk/` differs from the embedded version. Use `--diff` to print the recursive diff. Add `--apply` to update after a failed check or diff.

## kt config

`kt config` has two responsibilities: reading and writing the project's `.kt/project.yaml`, and managing runtime config files under `deploy/config/`.

### Project config

`.kt/project.yaml` is created by `kt init` and holds the project contract used by Make, nFPM, and `kt` itself. The key fields are:

- `schema`: project manifest schema (`kt.project/v1` for new scaffolds)
- `template`: scaffold template name as chosen by the user
- `app`: package / application name
- `kind`: `cli`, `service`, `mixed`, or `multi-service`
- `package`: package name, maintainer, description, section, and optional license
- `services`: structured service list for service-bearing projects; `[]` for `cli`
- `commands`: packaged user-facing commands
- `config`: source config directory, install directory, and example suffix
- `release`: release tag settings
- `kt`: scaffold metadata

Legacy comma-separated `services` plus top-level `user`/`group` still load, but new scaffolds write structured service entries with `name`, `role`, `runner`, `unit`, `user`, and `group`.

```bash
kt config show              # print all top-level scalar keys and values
kt config show --json       # print normalized project contract as JSON
kt config shape             # print kind-aware summary from .kt/project.yaml
kt config get <key>         # print a single value (used by Makefile: APP := $(shell kt config get app))
kt config set <key> <value> # update a top-level scalar value in .kt/project.yaml
kt config validate          # check app/kind/services/user/group consistency
```

### Deploy config

```bash
kt config init    # copy deploy/config/*.example files to actual config (no-clobber)
kt config check   # exit 1 if any config file derived from an example is missing
kt config diff    # diff each *.example against its actual counterpart
```

These delegate to the `config-init`, `config-check`, and `config-diff` Make targets.

## kt deploy

Inspect and validate generated deploy files against `.kt/project.yaml`.

```bash
kt deploy inspect
kt deploy inspect --json
kt deploy check
kt deploy check --json
```

`inspect --json` emits deploy metadata including structured services, config/data/log directories, packaged unit names, and installed runner paths. The generated `make build-metadata` target writes the same data to `dist/app/meta/deploy.json`.

`check` verifies the project manifest, `nfpm.yaml`, deploy config examples, service runners, systemd units, expected `ExecStart` paths, executable bits, and stale lifecycle scripts.

## kt release

Create immutable annotated release tags. Releases and deployment workflows should trigger from pushed `v<semver>` tags.

```bash
kt release next patch              # prints the next stable patch version
kt release next minor --pre rc     # prints the first RC for the next minor
kt release next pre                # increments the current prerelease number
kt release next stable             # promotes the current prerelease to stable
kt release plan minor              # preview tag, dirty state, and tag conflicts
kt release plan 1.4.0-rc.1 --json  # machine-readable release plan
kt release notes --since latest    # bullet git log since the previous tag
kt release notes v1.3.0..HEAD      # bullet git log for an explicit range
kt release validate v1.4.0         # strict tag validation for CI
kt release tag 1.4.0               # create local annotated tag v1.4.0 at HEAD
kt release push 1.4.0              # create and push v1.4.0
```

`tag` and `push` require a clean working tree, reject invalid SemVer versions, and refuse tags that already exist locally or on `origin`.

## kt update

Update `kt` itself to the latest release.

```bash
kt update          # check and apply
kt update --check  # check only; also informs about newer prereleases
kt update --prerelease
```

Automatically re-runs with `sudo` if the install location requires elevated permissions. Has no effect on dev builds.

By default, `kt update` only installs stable releases. `kt update --check` still informs you when a newer prerelease exists. Use `--prerelease` to opt into downloading prerelease versions such as `1.3.0-rc.1`. `--check` and `--prerelease` are intentionally separate modes.

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
