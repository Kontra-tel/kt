# 1.4 migration

Use this guide when upgrading existing kt-generated projects from 1.3.x to 1.4.x. kt 1.4 keeps legacy project manifests working, but new scaffolds write a more structured `.kt/project.yaml` and the shared tooling can now check drift before updating.

## What changed

- Top-level help is compact; use `kt help <topic>` for command details.
- `kt init --dry-run` shows scaffold write actions before creating files.
- New scaffolds write structured `package`, `services`, `commands`, `config`, `release`, and `kt` manifest blocks.
- Legacy comma-separated `services` plus top-level `user` and `group` still load.
- `kt config migrate --to kt.project/v1` rewrites legacy manifests to the structured format.
- `kt config schema` prints a JSON schema for editor or CI integration.
- `kt deploy metadata --json` emits the deploy metadata contract used by `dist/app/meta/deploy.json`.
- `kt update-tools --check` and `kt update-tools --diff` detect local `.kt/mk/` drift before applying updates.
- `kt completion bash|zsh|fish` prints shell completion snippets.

## Compatibility

Existing projects do not need an immediate manifest rewrite. kt 1.4 normalizes older manifests in memory and keeps `kt config get services` returning comma-separated service names for Make compatibility.

Do migrate deliberately when you want the manifest checked in using the 1.4 structured format.

## Recommended upgrade path

1. Upgrade `kt`.
2. Check shared tooling drift:

   ```bash
   kt update-tools --check
   kt update-tools --diff
   ```

3. If the diff is expected, apply it:

   ```bash
   kt update-tools --check --apply
   ```

4. Validate the current project before changing the manifest:

   ```bash
   kt config validate
   kt deploy check
   ```

5. Preview the normalized manifest:

   ```bash
   kt config show --json
   ```

6. Rewrite `.kt/project.yaml` when ready:

   ```bash
   kt config migrate --to kt.project/v1
   ```

7. Review and commit the manifest diff.
8. Re-run validation:

   ```bash
   kt config validate
   kt deploy check
   make build-metadata
   ```

## Manifest before and after

Legacy 1.3 style:

```yaml
schema: kt.project/v1
template: service
app: my-api
kind: service
services: my-api
user: my-api
group: my-api
```

Structured 1.4 style:

```yaml
schema: kt.project/v1
template: service
app: my-api
kind: service
package:
  name: my-api
services:
  - name: my-api
    runner: deploy/run/my-api
    unit: deploy/systemd/my-api.service
    user: my-api
    group: my-api
commands:
  - name: my-api
    path: deploy/bin/my-api
config:
  dir: deploy/config
  install_dir: /etc/my-api
  example_suffix: .example
release:
  tag_prefix: v
kt:
  scaffold_version: "1.4"
```

## Deploy metadata

`make build-metadata` writes both:

```text
dist/app/meta/build.json
dist/app/meta/deploy.json
```

The deploy metadata JSON can also be generated directly:

```bash
kt deploy metadata --json
kt deploy metadata --json --output dist/app/meta/deploy.json
```

## Shell completions

Install completions through your shell's normal mechanism:

```bash
kt completion bash
kt completion zsh
kt completion fish
```

## Rollback

The manifest migration is a file rewrite only. If a review or validation fails, restore the previous `.kt/project.yaml` from git and continue using the legacy format. kt 1.4 still supports it.
