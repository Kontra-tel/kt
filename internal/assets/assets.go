package assets

import "embed"

// FS contains all built-in project templates and shared local tooling.
//
// Keep this as the single source of truth for embedded scaffold files.
// To add a new template, add files under:
//
//	internal/assets/templates/projects/<template-name>/
//
// To add shared Makefile/script tooling, add files under:
//
//	internal/assets/common/
//
//go:embed all:common all:templates
var FS embed.FS
