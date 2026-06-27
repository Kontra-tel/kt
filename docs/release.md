# Release & maintenance

## Releasing kt

Releases are driven by `version.txt`.

When `version.txt` changes on `main` or `release/*`, the `Release` workflow in
[.gitea/workflows/release.yaml](/home/markus/Projektit/projects/kt/.gitea/workflows/release.yaml:1)
automatically:

1. validates the version against the branch
2. restores `version.txt` if the new value is invalid
3. runs `go test ./...`
4. creates and pushes tag `v<version>`
5. builds artifacts and creates the Gitea release

### Branch policy

- `main` may only release RC versions such as `1.3.0-rc.1`
- `release/*` branches may only release stable versions such as `1.3.0`

If the branch/version combination is invalid, the workflow does not push a tag.
It restores `version.txt` to the previous value in a follow-up bot commit.

### Example: 1.3.0-rc.1

1. Change `version.txt` to `1.3.0-rc.1` on `main`
2. Push the commit

The workflow will:

1. validate that `main` is only releasing an RC
2. run `go test ./...`
3. push annotated tag `v1.3.0-rc.1`
4. build cross-platform binaries via `scripts/release.sh`
5. build Linux packages (`.deb` + `.rpm`) for amd64 and arm64 via nFPM
6. create a prerelease in Gitea and upload all artifacts

### Example: 1.3.0

1. Change `version.txt` to `1.3.0` on `release/1.3`
2. Push the commit

The same workflow will publish a stable release.

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
make release   # produces dist/kt-* binaries and SHA256SUMS (no packages)
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
still informs you when a newer prerelease exists. Pass `--prerelease` to opt
into channels such as `1.3.0-rc.1`. `--check` and `--prerelease` are kept as
separate modes on purpose.

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
  common/
    mk/
      common.mk
      config.mk
      doctor.mk
      nfpm.mk
      version.mk
  templates/
    projects/
      app/
      cli/
      mixed/
      multi/
```

The `deploy/` folder at the repository root is only for packaging the `kt` binary itself and is unrelated to the project templates.
