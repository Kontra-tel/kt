# Filesystem layout migration

Newly scaffolded projects install artifacts and writable state into separate
Linux filesystem locations:

```text
old: /opt/<app>/
new artifacts: /usr/lib/<app>/
new mutable data: /var/lib/<app>/
new logs: /var/log/<app>/
```

Packaged systemd units now install to `/usr/lib/systemd/system/` instead of
`/etc/systemd/system/`. Files under `/etc/systemd/system/` are reserved for
administrator-managed overrides.

## Existing projects

`kt update-tools` refreshes shared `.kt/mk/` files. It does
not rewrite project-specific templates. Existing projects should migrate
deliberately:

1. Update `nfpm.yaml` artifact destinations and packaged unit destinations.
2. Update launch scripts under `deploy/bin/`.
3. Update systemd `WorkingDirectory` and `ReadWritePaths`.
4. Update package hooks to create `/var/lib/<app>/` and `/var/log/<app>/`.
5. Copy existing mutable data from `/opt/<app>/` into `/var/lib/<app>/`.
6. Build the desired package formats and inspect their contents.
7. Back up service data, install the package, restart explicitly, and verify
   service health.

Keep the previous package available. If validation fails, reinstall the
known-good version and restore data from the backup when required.
