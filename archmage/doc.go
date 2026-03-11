// Package archmage provides a system for loading and managing game
// configurations from JSON files with support for internationalization,
// cross-table references, and layered overrides.
//
// The core component is Atlas, a registry of configuration tables.
// Each config key maps to its JSON files in an index file, atlas.json.
// It supports three mapping strategies: unique (one-to-one), single (one from
// a set), and multiple (multi-file merging).
//
// Key features include:
//   - I18n for multi-language text management with automatic fallback
//   - XRef type for cross-table reference resolution via Atlas.BindRefs
//   - Duration type with compact JSON array format and unit optimization
//   - Whitelist / blacklist to load only a subset of configurations
//   - Layered overrides: additional directories or filesystems supply JSON that
//     is merged into the base data at load time, field by field
//
// Example usage:
//
//	atlas := conf.NewConfigAtlas()
//	err := archmage.LoadAtlas("atlas.json", "config", atlas,
//	    archmage.WithOverrideRoot("overrides"),
//	    archmage.WithWhitelist([]string{"item", "hero"}),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
package archmage
