# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go SDK (`shadop.dev/pkg/sdk-go`) for Shadow Opera game configuration management. The `archmage` package is the public API; `internal/` contains test infrastructure and generated configuration types.

Requires **Go 1.25+** (uses `encoding/json/v2`, iterators, generics).

## Common Commands

```bash
# Run all tests
go test ./...

# Run a single test
go test ./internal -run TestAtlasGolden

# Update golden files (after intentional output changes)
go test ./internal -run TestAtlasGolden -update

# Cross-SDK porting validation (compare Go vs C# golden files)
go test ./internal -run TestPorting -porting

# Coverage report (opens HTML)
cd archmage && bash coverage.sh
```

## Architecture

**`archmage/`** — Public SDK package:
- `atlas.go` — Core `Atlas` interface and `LoadAtlas` function. Three mapping strategies: Unique (1:1), Single (pick one from array), Multiple (merge arrays)
- `ref.go` — Generic `Ref[V, T]` for cross-table lazy reference resolution
- `duration.go` — Custom `Duration` type with compact JSON array format `[unitType, value]`
- `vec.go` — Generic `Vec2/Vec3/Vec4` types with JSON array marshaling
- `tuple.go` — Generic `Tuple1`–`Tuple7` fixed-size typed collections
- `i18n.go` — Multi-language text loading with fallback support
- `logger.go` — Minimal `Logger` interface (single `Info` method)

**`internal/conf/`** — Generated code (DO NOT EDIT). Game config types and `ConfigAtlas` produced by the external `archmage` code-gen tool.

**`internal/`** — Test infrastructure with golden file testing, test data in `testdata/`, expected outputs in `golden/`, and override configs in `override/`.

## Key Patterns

- **Functional options**: `LoadAtlas(..., WithLogger(l), WithWhitelist(...))`
- **Interface compliance checks**: `var _ Logger = (*defaultLogger)(nil)`
- **Custom JSON marshaling** via `json/v2`: `MarshalJSONTo` / `UnmarshalJSONFrom`
- **Error format**: `fmt.Errorf("<archmage> context message | %w", err)`
- **Unexported sentinel errors**: `_errInvalidDurationShardsType`

## Commit Message Style

- Short imperative subject, no trailing period
