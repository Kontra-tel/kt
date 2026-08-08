# Release & maintenance

## Releasing kt

Releases are driven by immutable annotated Git tags. Push a strict SemVer tag in
the form `v<version>` and the `Release` workflow builds artifacts and creates
the matching Gitea release.

```bash
kt release next patch
kt release push 1.3.0
```

The `push` command requires a clean working tree and refuses a version that is
already tagged locally or on `origin`. It never changes source files.

### Prereleases

Prerelease tags have a SemVer suffix and are marked as prereleases in Gitea:

```bash
kt release push 1.4.0-rc.1
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

`kt update` downloads the matching binary for the current OS and architecture from Gitea and atomically replaces the running executable. If the install location requires elevated permissions (e.g. `/usr/local/bin`) it re-runs automatically with `sudo`.

By default, updates only install stable releases. Plain `kt update --check`
into prerelease channels.

Dev builds (version = `dev`) skip the check.

## Updating project tooling

After upgrading `kt`, run this in each project to pull in the latest `.kt/mk/`:

```bash
kt update-tools
```

## Embedded asset layout

Templates and shared tooling are embedded into the `kt` binary at build time:

```text
internal/assets/
  common/mk/
  templates/projects/
```

The root `deploy/` folder only packages the `kt` binary itself; it is unrelated
to generated project templates.
