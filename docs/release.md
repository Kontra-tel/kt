# Release & maintenance

## Releasing kt

Releases are driven by immutable annotated Git tags. Push a strict SemVer tag in the form `v<version>` and the `Release` workflow builds artifacts and creates the matching Gitea release.

```bash
kt release plan minor
kt release next minor --pre rc
kt release push 1.4.0-rc.1
kt release next stable
kt release push 1.4.0
```

The `plan` command previews the current reachable release, next version, tag, dirty state, and local/remote tag conflicts. The `push` command requires a clean working tree and refuses a version that is already tagged locally or on `origin`. It never changes source files.

### Release notes

`kt release notes` prints Markdown bullets from `git log --oneline`.

```bash
kt release notes --since latest
kt release notes v1.3.0..HEAD
```

With `--since latest`, kt finds the latest reachable tag before `HEAD` and logs that range. If no prior tag is reachable, it logs the available history. The release workflow uses this command when creating or updating the hosted release body.

### Prereleases

Prerelease tags have a SemVer suffix and are marked as prereleases in Gitea:

```bash
kt release next minor --pre rc # 1.4.0-rc.1
kt release next pre            # 1.4.0-rc.2 when rc.1 is latest
kt release next stable         # 1.4.0 when an rc is latest
```

### Release artifacts

```text
kt-linux-amd64
kt-linux-arm64
kt-darwin-amd64
kt-darwin-arm64
kt_<version>_amd64.deb
kt_<version>_arm64.deb
kt-<version>.amd64.rpm
kt-<version>.arm64.rpm
SHA256SUMS
```

### Build binaries locally

```bash
make release                 # uses an exact v<semver> tag, or a dev version
make release VERSION=1.3.0   # explicit local package/build version
```

### Install prereleases with the script

```bash
KT_VERSION=1.3.0-rc.1 bash <(curl -sL https://git.kontra.tel/kontra.tel/Kt/raw/branch/main/scripts/install.sh)
KT_PRERELEASE=1 bash <(curl -sL https://git.kontra.tel/kontra.tel/Kt/raw/branch/main/scripts/install.sh)
```

## Updating kt

```bash
kt update          # check for a newer release and apply it
kt update --check  # check only; also informs you when a newer prerelease exists
kt update --prerelease
```

`kt update` downloads the matching binary for the current OS and architecture from Gitea, verifies it against `SHA256SUMS`, and atomically replaces the running executable. If the install location requires elevated permissions (e.g. `/usr/local/bin`) it re-runs automatically with `sudo`.

By default, updates only install stable releases. Plain `kt update --check` still reports newer prereleases, but it does not opt into installing them. Use `--prerelease` to install from the prerelease channel.

Dev builds (version = `dev`) skip the check.

## Updating project tooling

After upgrading `kt`, run this in each project to pull in the latest `.kt/mk/`:

```bash
kt update-tools --check
kt update-tools --diff
kt update-tools --apply
```

Use `--check` in CI to detect drift, `--diff` to inspect local changes, and `--apply` with either mode to refresh the tooling.

## Embedded asset layout

Templates and shared tooling are embedded into the `kt` binary at build time:

```text
internal/assets/
  common/mk/
  templates/projects/
```

The root `deploy/` folder only packages the `kt` binary itself; it is unrelated to generated project templates.
