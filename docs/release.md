# Release & maintenance

## Releasing kt

Releases now use a two-step flow:

1. Run `Prepare Release Tag` from Gitea's Actions UI on the branch you want to release from.
2. That workflow commits `version.txt` if needed and pushes `v...`.
3. The pushed tag triggers `Publish Release`, which builds artifacts and creates the Gitea release.

```text
Actions → Prepare Release Tag → Run workflow
```

The manual workflow (`.gitea/workflows/release.yaml`) supports two version modes:

- `bump`: increment `version.txt` by `patch`, `minor`, or `major`
- `set-version`: set an exact version such as `1.3.0-rc.1`

The publish workflow (`.gitea/workflows/publish.yaml`) determines prerelease vs stable automatically from the tag.

### Branch policy

- `main` may only produce RC tags such as `v1.3.0-rc.1`
- `release/*` branches may only produce stable tags such as `v1.3.0`

This is validated both when preparing the tag and when publishing it.

### Example: 1.3.0-rc.1

Run `Prepare Release Tag` on `main` with:

- `mode`: `set-version`
- `version`: `1.3.0-rc.1`

The workflows will:

1. Set or bump `version.txt` and commit it when required
2. Run `go test ./...`
3. Push an annotated `v1.3.0-rc.1` tag
4. Build cross-platform binaries via `scripts/release.sh`
5. Build Linux packages (`.deb` + `.rpm`) for amd64 and arm64 via nFPM
6. Create a prerelease in Gitea and upload all artifacts

### Example: 1.3.0

Run `Prepare Release Tag` on `release/1.3` with either:

- `mode`: `bump`
- `bump`: `patch`

or:

- `mode`: `set-version`
- `version`: `1.3.0`

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
