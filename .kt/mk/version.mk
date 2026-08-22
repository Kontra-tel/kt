.PHONY: version release-next release-tag release-push

version: ## Print the version derived from the exact release tag
	@printf '%s\n' "$(VERSION)"

release-next: ## Print the next version (KIND=patch|minor|major)
	@kt release next $(or $(KIND),patch)

release-tag: ## Create an annotated release tag (VERSION=x.y.z)
	@kt release tag "$(VERSION)"

release-push: ## Create and push an annotated release tag (VERSION=x.y.z)
	@kt release push "$(VERSION)"
