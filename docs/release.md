# Release & maintenance

## Releasing kt

Releases are triggered manually from Gitea's Actions UI:

```text
Actions → Release → Run workflow → choose bump type (patch / minor / major)
```

The workflow (`gitea/workflows/release.yaml`) will:

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

## Updating kt

```bash
kt update          # check for a newer release and apply it
kt update --check  # check only, exits 1 if a newer version is available
```

`kt update` downloads the matching binary for the current OS and architecture from Gitea and atomically replaces the running executable. If the install location requires elevated permissions (e.g. `/usr/local/bin`) it re-runs automatically with `sudo`.

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
      multi/
```

The `deploy/` folder at the repository root is only for packaging the `kt` binary itself and is unrelated to the project templates.
