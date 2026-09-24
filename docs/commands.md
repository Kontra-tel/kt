# Commands

## Global flags

Global flags may appear before the command or before a subcommand.

| Flag | Description |
| --- | --- |
| `--json` | Machine-readable output where supported (`config show`, `deploy`, `release plan`, `doctor`) |
| `--quiet`, `-q` | Suppress non-error status output |
| `--no-color` | Disable ANSI styling |
| `--color auto\|always\|never` | Control styling |

Unknown commands include a nearest-command suggestion when the typo is close enough. Use `kt help <topic>` for contextual help without expanding the top-level help page.

```bash
kt help init
kt help config
kt help deploy
kt help release
kt help completion
```

## kt init

Scaffold a new project from a template. Running without arguments starts an interactive prompt to choose a template, enter a valid app name, confirm the target directory, and fill package/service defaults.

```bash
kt init                                              # interactive
kt init <template> <app> [options]                   # explicit
```

| Option | Default | Description |
| --- | --- | --- |
| `--dir` | `.` | Target directory |
| `--maintainer` | Git author | Package maintainer; useful for non-interactive scaffolding |
| `--service-user` | App name | Service user for service-bearing templates |
| `--service-group` | Service user | Service group for service-bearing templates |
| `--force` | `false` | Overwrite existing files |
| `--dry-run` | `false` | Print create/keep/overwrite actions without writing files |

```bash
kt init service my-api
kt init cli my-tool
kt init mixed my-suite
kt init multi my-platform
kt init cli my-tool --dir /srv/projects
kt init service my-api --maintainer 'Ops <ops@example.invalid>' --service-user my-api --service-group my-api
kt init service my-api --dry-run
```

Package maintainer is derived from Git config automatically when `--maintainer` is omitted. Service templates derive service user and group from the app name unless overridden.

## kt templates

List all available project templates.

```bash
kt templates
```

## kt completion

Print shell completion snippets with contextual root commands, subcommands, templates, and supported flags.

```bash
kt completion bash
kt completion zsh
kt completion fish
```

Load the printed snippet through your shell's normal completion mechanism.

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

All config commands accept `--dir DIR`, which selects a project without changing the calling shell.

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
- `release`: release tag settings, including `tag_prefix`
- `kt`: scaffold metadata

Legacy comma-separated `services` plus top-level `user`/`group` still load. `kt config migrate --to kt.project/v1` rewrites an existing manifest into the structured format; see [1.4 migration](migration-1.4.md) for the recommended upgrade sequence.

```bash
kt config show                  # print all top-level scalar keys and values
kt config show --json           # print normalized project contract as JSON
kt config shape                 # print kind-aware summary from .kt/project.yaml
kt config get <key>             # print a scalar, including nested fields such as release.tag_prefix
kt config set <key> <value>     # update a top-level scalar value in .kt/project.yaml
kt config validate              # check manifest consistency
kt config edit                  # open $VISUAL/$EDITOR on .kt/project.yaml, then validate
kt config schema                # print a JSON schema for kt.project/v1
kt config migrate --to kt.project/v1
kt config --dir /srv/projects/my-api validate
```

### Deploy config

```bash
kt config init    # copy deploy/config/*.example files to actual config (no-clobber)
kt config check   # exit 1 if any config file derived from an example is missing
kt config diff    # diff each *.example against its actual counterpart
```

These delegate to the `config-init`, `config-check`, and `config-diff` Make targets.

## kt deploy

Inspect, plan, synchronize, and validate deploy files against `.kt/project.yaml`. All deploy commands accept `--dir DIR`.

```bash
kt deploy inspect
kt deploy inspect --json
kt deploy metadata --json
kt deploy metadata --json --output dist/app/meta/deploy.json
kt deploy plan
kt deploy plan --json
kt deploy check
kt deploy check --strict --json
kt deploy sync --dry-run
kt deploy sync
```

`inspect --json` emits runtime deploy metadata including structured services, config/data/log directories, packaged unit names, and installed runner paths. `metadata` emits the same JSON-only contract and can write it directly to a file. The generated `make build-metadata` target writes the data to `dist/app/meta/deploy.json`.

`plan` derives the package source-to-destination contract from the normalized manifest. It includes application artifacts, declared commands, service runners and units, and config examples.

`check` verifies the manifest, every declared command executable, `nfpm.yaml` mappings, deploy config examples, service runners, systemd units, expected `ExecStart` paths, executable bits, and stale lifecycle scripts. `--strict` also fails on warnings, including incomplete unit hardening or missing config examples.

New 1.5 scaffolds own only the entries between `kt:contents:start` and `kt:contents:end` in `nfpm.yaml`. `sync` refreshes that interval; `--dry-run` prints a unified diff. Entries below the closing marker remain user-owned. Sync refuses older unmarked manifests; see [1.5 migration](migration-1.5.md).

## kt release

Create immutable annotated release tags. Release commands accept `--dir DIR` and use `release.tag_prefix` from the selected project; new projects default to `v`.

```bash
kt release next patch              # prints the next stable patch version
kt release next minor --pre rc     # prints the first RC for the next minor
kt release next pre                # increments the current prerelease number
kt release next stable             # promotes the current prerelease to stable
kt release plan minor              # preview tag, dirty state, and tag conflicts
kt release plan 1.5.0-rc.1 --json  # machine-readable release plan
kt release notes --since latest    # bullet git log since the previous matching tag
kt release validate v1.5.0         # strict tag validation for CI with the default prefix
kt release tag 1.5.0               # create local annotated tag using the configured prefix
kt release push 1.5.0              # create and push the configured tag
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

Check required tools listed in `DOCTOR_TOOLS`.

```bash
kt doctor
kt doctor --dir /srv/projects/my-api --json
```

The default command delegates to `make doctor`. `--json` reports each tool's name, resolved path, first `--version` output line, and whether all required tools are present.

## kt version

Print the current `kt` version, commit, and build date.

```bash
kt version
```
