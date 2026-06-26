# Release & maintenance

## Releasing kt

Releases are triggered manually from Gitea's Actions UI:

```text
Actions → Release → Run workflow → choose mode, version strategy, and whether the release is marked prerelease
```

The workflow (`gitea/workflows/release.yaml`) supports two release modes:

- `bump`: increment `version.txt` by `patch`, `minor`, or `major`
- `set-version`: set an exact version such as `1.3.0-rc.1`

Mark prerelease builds with the `prerelease` input when creating release
candidates, betas, or alphas.

### Example: 1.3.0-rc.1

Use these workflow inputs:

- `mode`: `set-version`
- `version`: `1.3.0-rc.1`
- `prerelease`: `true`

The workflow will:

1. Bump `version.txt` by the chosen increment and commit it
2. Run `go test ./...`
3. Build cross-platform binaries via `scripts/release.sh`
4. Build Linux packages (`.deb` + `.rpm`) for amd64 and arm64 via nFPM
5. Push the commit and create an annotated git tag
6. Create a Gitea release and upload all artifacts

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

After upgrading `kt`, run this in each project to pull in the latest `.kt/mk/` and `.kt/scripts/`:

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
    scripts/
      postinstall-systemd.sh
      preremove-systemd.sh
  templates/
    projects/
      app/
      cli/
      mixed/
      multi/
```

The `deploy/` folder at the repository root is only for packaging the `kt` binary itself and is unrelated to the project templates.
