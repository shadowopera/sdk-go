// Package archmage is the runtime library through which Go applications load and
// access config data exported by Archmage.
//
// [Archmage] is a configuration solution for
// game development: specifications for how to structure config data, define
// fields, and fill in each value; pipelines that export runtime data and
// generate strongly-typed code; multi-language SDKs for loading and accessing
// that data at runtime; and a collaborative editing workflow for teams.
//
// The SDK is built around the concept of an [Atlas] — a registry that maps named
// keys to configurations. Each key is associated with one or more JSON files.
// At runtime, the SDK reads these files, deserializes them into instances of
// generated Go types, resolves cross-table references, and calls post-load hooks.
//
// Key features:
//   - [I18n] — multi-language text management with automatic fallback
//   - [XRef] — cross-table reference resolution via [Atlas.BindRefs]
//   - [MinMax] — random value selection within a range
//   - [Duration] — nanosecond precision; formats as human-readable strings
//   - [WeightedPool] — weighted random selection with probability proportional
//     to item weight
//   - Variants — switch an item to an alternative data set at load time via
//     [WithVariant]
//   - Whitelist/Blacklist — load only a subset of atlas items
//   - Layered overrides — merge files with matching relative paths from
//     additional override sources (a directory path or an [fs.FS]) into the
//     base configs, field by field, at load time
//   - Pluggable load strategies — parallel loading via [WithLoadStrategy]
//   - Versioning — VCS metadata (branch, commit, timestamp, etc.), when present
//     in atlas.json, is available on the loaded atlas
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
//
// [Archmage]: https://shadop.dev/archmage/
package archmage
