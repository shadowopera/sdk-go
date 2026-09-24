# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.18.0] - 2026-09-24

### Changed

- Rebuilt the integration tests on the self-contained SDK test data, and now verify every `XRef` binding after each golden scenario.

## [0.17.0] - 2026-09-09

### Added

- Added `WithVariant(key, variant)` option for non-default variant selection.
- Added `Duration` support to `MinMax[T]`, including the `Sample` method.

## [0.16.0] - 2026-07-31

### Changed

- Renamed the `single`/`multiple` mapping strategies to `variant`/`many` in `atlas.json`.

## [0.15.2] - 2026-07-12

### Added

- Added `SampleWith` and `SampleIndexWith` methods to `WeightedPool[T]`.

## [0.15.0] - 2026-06-11

### Added

- Added `WeightedPool[T]` for weighted random sampling.

## [0.14.0] - 2026-05-31

### Added

- Added `MinMax[T]` struct for representing a numeric range with `Min` and `Max` values, including a `Sample` method supporting all numeric types.

## [0.12.0] - 2026-05-07

### Changed

- Changed vector (`Vec2[T]`/`Vec3[T]`/`Vec4[T]`) types to use object-based JSON serialization instead of arrays.

## [0.9.0] - 2026-04-07

### Added

- `RGBA` struct with hex string parsing (`#RRGGBB` / `#RRGGBBAA`) and JSON marshaling; alpha is omitted from output when `A == 0xFF`.

## [0.7.0] - 2026-04-01

### Changed

- Renamed `XRef.RawValue` to `XRef.CfgID`.

## [0.6.0] - 2026-03-29

### Changed

- Config tables now include `id` in JSON serialization output.

## [0.1.0] - 2026-03-11

### Added

- Initial release
- Atlas loading system with layered overrides and whitelist/blacklist filtering
- JSON serialization via `encoding/json/v2`
- `Duration` type with nanosecond precision
- `Vec2[T]`, `Vec3[T]`, `Vec4[T]` typed vectors
- `Tuple1`–`Tuple7` heterogeneous tuples
- `XRef[V, T]` type for cross-table references
- I18n internationalization with language fallback
