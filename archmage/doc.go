// Package archmage is the runtime library for the Archmage game configuration system, which loads
// and manages configurations from JSON files with support for i18n, cross-table references,
// durations, and data types ranging from primitives to recursive structures.
//
// It is built around the concept of an Atlas — a registry that maps named keys to
// configurations. Each key points to one or more JSON files; Archmage reads and
// deserializes them into generated C# objects, resolves cross-table references, and
// calls post-load hooks.
//
// Key features:
//   - I18n for multi-language text management with automatic fallback
//   - XRef for cross-table reference resolution via IAtlas.BindRefs
//   - Duration with nanosecond precision and compact JSON array encoding
//   - Whitelist / blacklist to load only a subset of items
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
