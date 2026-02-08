// Package archmage provides a system for loading and managing configurations
// from JSON files with support for internationalization, references, durations,
// and layered overrides.
//
// The core component is the Atlas system, which loads configuration items from
// an index file and populates them from JSON files. It supports three mapping
// strategies: unique (one-to-one), single (variant-based), and multiple
// (multi-file merging). Configuration overrides can be applied from additional
// directories or file systems.
//
// Key features include:
//   - Atlas-based configuration loading with flexible mapping strategies
//   - I18n for multi-language text management with automatic fallbacks
//   - Ref type for deferred reference resolution between config items
//   - Duration type with compact JSON array format and unit optimization
//   - Vector types (Vec2, Vec3, Vec4) with JSON array serialization
//   - Tuple types (Tuple1-7) for grouping mixed-type elements
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
