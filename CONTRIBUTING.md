# Contributing

`kt` is maintained mainly for the maintainer's personal workflow.

That means:

- changes are accepted only when they fit that workflow and keep the tool boring
- issues and pull requests are handled best-effort, with no support SLA
- compatibility is considered, but the maintainer may choose the simpler project direction over broad use cases
- feature requests that need different defaults, extra abstractions, or broader platform support may be better served by a fork

## Forks and releases

You are welcome to fork the project, adjust it for your own workflow, and publish your own builds or releases from that fork once the repository has an explicit open-source license file.

If you publish forked releases, use distinct names, release notes, package metadata, and update endpoints so users can tell your fork apart from the maintainer's builds.

## Contributions

Good contributions are small, boring, and easy to review:

1. Open an issue or discussion first for behavior changes.
2. Keep pull requests focused on one problem.
3. Preserve existing scaffold behavior unless the change is intentional and documented.
4. Add or update tests for observable behavior changes.
5. Update user-facing docs for command, template, release, or migration changes.
6. Run `go test ./...` before submitting.

## Maintainer decisions

The maintainer may decline a contribution even when it is technically correct. Common reasons:

- it adds maintenance burden for a use case the maintainer does not run
- it makes generated projects less explicit or harder to inspect
- it changes release, packaging, or deployment assumptions too broadly
- it needs long-term support the maintainer cannot commit to

Forking is the preferred path when your needs diverge.

## License recommendation

Recommended license: **MIT**.

MIT fits this project because it is short, permissive, common for Go CLI tools, and allows people to:

- use the code privately or commercially
- fork and modify it
- publish their own binaries and releases
- redistribute source or packaged builds

while preserving attribution and warranty disclaimers.

Until a `LICENSE` file is added, this recommendation is not itself a legal grant. Add an explicit `LICENSE` file before telling users they may redistribute forked releases.
