# Migrating to 1.5

`kt` 1.5 makes the project manifest enforceable at the nFPM boundary. Existing projects remain readable and `kt deploy check` can validate their current package mappings. It does not rewrite `nfpm.yaml` automatically.

## Before changing files

```bash
kt config validate
kt deploy plan --json
kt deploy check
```

Fix existing manifest or deploy errors first. `kt deploy check` now requires every planned mapping in `nfpm.yaml`: application artifacts, commands, service runners, units, and config examples.

## Adopt managed nFPM entries

New 1.5 projects surround only their generated `contents:` entries with markers. Preserve package entries you own after the closing marker:

```yaml
contents:
  # kt:contents:start
  - src: dist/app
    dst: /usr/lib/my-app
  # generated command, runner, unit, and config mappings
  # kt:contents:end

  # User-owned package entries remain outside the managed block.
  - src: resources
    dst: /usr/lib/my-app/resources
```

The markers must be indented under the one existing `contents:` key. Do not create a second `contents:` key.

Then review and apply the generated entries:

```bash
kt deploy sync --dry-run
kt deploy sync
kt deploy check --strict
```

`sync` replaces only the text from `kt:contents:start` through `kt:contents:end`. Without both markers it refuses to write, so legacy manifests cannot be overwritten accidentally.

## Add commands to older service manifests

1.5 validates every manifest command. Service templates before 1.5 can declare their default metadata command explicitly:

```yaml
commands:
  - name: my-app
    path: deploy/bin/my-app
```

This is optional for an unchanged legacy service: kt derives the same default. Add entries when a project packages additional commands.

## Add the verification target

Update shared tooling, then add this user-owned target to the project `Makefile` if it is missing:

```bash
kt update-tools --apply
```

```make
.PHONY: verify
verify: config-check ## Validate the manifest, deploy files, and package mappings
	@kt deploy check --strict
```

Run `make config-init && make verify` after adapting the project. `verify` intentionally does not invoke the scaffold's placeholder test command.

## Release prefixes

`release.tag_prefix` is now active. The default remains `v`; a project may use a different safe prefix:

```yaml
release:
  tag_prefix: release-
```

With that configuration, release tags are `release-1.5.0` and `kt release validate release-1.5.0` succeeds. The prefix must match `[A-Za-z][A-Za-z0-9._-]*`.
