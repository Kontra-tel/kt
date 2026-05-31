# Gitea Debian publishing

Gitea can host Debian packages built by `kt` projects. This is an opt-in
example: `kt` remains package-registry agnostic and continues to support
`.deb`, `.rpm`, and Arch Linux packages.

Copy `examples/gitea-actions/release-deb.yml` into a project's
`.gitea/workflows/` directory and adjust it for the runner environment.

## Required Actions configuration

Configure these non-secret variables:

```text
PACKAGE_BASE_URL=https://gitea.example.com
PACKAGE_OWNER=<package-owner>
PACKAGE_DISTRIBUTION=<distribution>
PACKAGE_COMPONENT=<component>
PACKAGE_USER=<package-publisher>
```

Configure this secret:

```text
PACKAGE_TOKEN=<personal-access-token>
```

Gitea reserves the `GITEA_` and `GITHUB_` prefixes for built-in values, so do
not use either prefix for custom secrets.

## Runner requirements

The example expects these commands on the runner:

```text
git
make
nfpm
curl
```

Provision and pin the nFPM version as part of runner management. This avoids
downloading an unreviewed tool version during every release.

## Deployment and rollback

Publishing does not deploy or restart a service. Install packages through the
target system's normal `apt` configuration, run migrations when needed, and
restart the service explicitly.

Keep known-good versions available so rollback can select an exact version:

```bash
sudo apt install <app>=<previous-version>
sudo systemctl restart <app>
sudo systemctl status <app>
```
