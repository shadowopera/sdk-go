package archmage

import (
	"cmp"
	"context"
	"encoding/json/v2"
	"fmt"
	"io/fs"
	"iter"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const (
	// MappingUnique indicates a one-to-one mapping between a key and a file.
	MappingUnique = "unique"
	// MappingVariant indicates that a key maps to multiple file variants, only one of which is loaded.
	MappingVariant = "variant"
	// MappingMany indicates that a key maps to multiple files loaded separately and merged into one.
	MappingMany = "many"
)

// Atlas is the interface a configuration collection must implement to be
// loaded by LoadAtlas.
type Atlas interface {
	// SetDataVersion stores the version metadata from atlas.json.
	SetDataVersion(v *VersionInfo)
	// AtlasItems returns all registered items.
	AtlasItems() map[string]*AtlasItem
	// BindRefs resolves cross-table references after all items are loaded.
	BindRefs()
	// OnLoaded is called once after all items are loaded and refs are bound.
	// Returning an error aborts the load.
	OnLoaded() error
}

// AtlasItem represents a configuration item within atlas.json.
type AtlasItem struct {
	// Cfg is a pointer to the configuration struct that receives unmarshaled data.
	Cfg any
	// Mapping specifies how this item maps to files (MappingUnique, MappingVariant, or MappingMany).
	Mapping string
	// Key is the item's key in atlas.json.
	Key string
	// Variant is the variant used by a variant-mapped item, defaulting to "/".
	// It is empty for other mappings.
	Variant string
	// Ready reports whether the item was successfully loaded.
	Ready bool
}

// AtlasJSON defines the structure of atlas.json.
type AtlasJSON struct {
	// Version holds the VCS version metadata.
	Version *VersionInfo `json:"version"`
	// Unique maps each key to a unique file path (one-to-one).
	Unique map[string]string `json:"unique"`
	// Variant maps each key to its file variants, where "/" denotes the default.
	Variant map[string]map[string]string `json:"variant"`
	// Many maps each key to an ordered list of files to merge.
	Many map[string][]string `json:"many"`
}

func (atlas *AtlasJSON) pickFromVariant(key, variant string) (string, bool) {
	m, ok := atlas.Variant[key]
	if ok {
		f, ok := m[variant]
		if ok {
			return f, true
		}
	}
	return "", false
}

// LoadAtlas reads atlasFile, loads each configuration item from cfgRoot,
// applies any overrides, calls BindRefs to resolve cross-table references,
// and finally calls OnLoaded on the atlas.
func LoadAtlas(atlasFile string, cfgRoot string, atlas Atlas, opts ...Option) error {
	if atlas.AtlasItems() == nil {
		return fmt.Errorf("<archmage> atlas must be created using NewConfigAtlas")
	}
	atlasOpts := newAtlasOptions()
	atlasOpts.loadStrategy = func(seq iter.Seq2[string, *AtlasItem], itemLoadFunc AtlasItemLoadFunc) error {
		for key, item := range seq {
			err := itemLoadFunc(context.Background(), key, item)
			if err != nil {
				return err
			}
		}
		return nil
	}
	for _, opt := range opts {
		opt(atlasOpts)
	}

	err := loadAtlasImpl(atlasFile, cfgRoot, atlas, atlasOpts)
	if err != nil {
		return err
	}

	return atlas.OnLoaded()
}

func loadAtlasImpl(atlasFile string, cfgRoot string, atlas Atlas, opts *atlasOptions) error {
	start := time.Now()
	for _, cfg := range opts.overrideConfigs {
		if cfg.fsys != nil {
			continue
		}
		stat, err := os.Stat(cfg.root)
		if err != nil {
			return fmt.Errorf("<archmage> invalid override root directory %q | %w", cfg.root, err)
		}
		if !stat.IsDir() {
			return fmt.Errorf("<archmage> override root %q is not a directory", cfg.root)
		}
	}

	atlasData, err := os.ReadFile(atlasFile)
	if err != nil {
		return err
	}

	var atlasJSON AtlasJSON
	err = json.Unmarshal(atlasData, &atlasJSON)
	if err != nil {
		return fmt.Errorf("<archmage> invalid %q | %w", atlasFile, err)
	}

	opts.cbAtlasModifier(&atlasJSON)
	atlas.SetDataVersion(atlasJSON.Version)

	items := atlas.AtlasItems()
	for _, v := range opts.whitelist {
		if _, ok := items[v]; !ok {
			return fmt.Errorf("<archmage> atlas whitelist: unknown item %q", v)
		}
	}
	for _, v := range opts.blacklist {
		if _, ok := items[v]; !ok {
			return fmt.Errorf("<archmage> atlas blacklist: unknown item %q", v)
		}
	}
	for _, v := range slices.SortedFunc(maps.Keys(opts.variants), compareLower) {
		if _, ok := items[v]; !ok {
			return fmt.Errorf("<archmage> atlas variant: unknown item %q", v)
		}
		if opts.variants[v] == "" {
			return fmt.Errorf("<archmage> atlas variant: empty variant for item %q", v)
		}
	}

	keys := slices.SortedFunc(maps.Keys(items), compareLower)
	filtered := slices.DeleteFunc(keys, func(k string) bool {
		cause, yes := opts.shouldSkip(k)
		if yes {
			opts.Info(fmt.Sprintf("<archmage> Skipping atlas item: %s. cause: %s", k, cause))
		}
		return yes
	})
	filteredItemSeq := func(yield func(string, *AtlasItem) bool) {
		for _, k := range filtered {
			if !yield(k, items[k]) {
				break
			}
		}
	}
	err = opts.loadStrategy(filteredItemSeq, func(ctx context.Context, key string, item *AtlasItem) error {
		if err := loadItem(ctx, key, item, &atlasJSON, atlasFile, cfgRoot, opts); err != nil {
			return fmt.Errorf("<archmage> failed to load atlas item %q. atlasFile: %s, cfgRoot: %s | %w",
				key, atlasFile, cfgRoot, err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	atlas.BindRefs()

	elapsed := time.Since(start).Milliseconds()
	opts.Info(fmt.Sprintf("<archmage> Loaded %d config items in %dms", len(filtered), elapsed))
	return nil
}

// loadItem loads a single atlas item from mapped files and applies any overrides.
// Merge rules:
//   - null → resets the target field to its default value or raises an error.
//   - JSON object → recursively merges: only fields present in the input are updated, others remain unchanged.
//   - Any other value → overwrites the field.
func loadItem(ctx context.Context, key string, item *AtlasItem,
	atlasJSON *AtlasJSON, atlasFile string, cfgRoot string, opts *atlasOptions,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	var fd struct {
		paths         string
		overrideFiles []string
		overrides     [][]byte
	}

	readOverrides := func(file string) error {
		for _, cfg := range opts.overrideConfigs {
			if cfg.fsys != nil {
				if _, err := fs.Stat(cfg.fsys, file); err != nil {
					continue
				}
			} else {
				ovr := filepath.Join(cfg.root, file)
				if _, err := os.Stat(ovr); err != nil {
					continue
				}
			}
			ovrFile, ovrData, err := readOverrideFile(cfg, file)
			if err != nil {
				return fmt.Errorf("failed to read override file: %s | %w", file, err)
			}
			fd.overrideFiles = append(fd.overrideFiles, ovrFile)
			fd.overrides = append(fd.overrides, ovrData)
		}
		return nil
	}

	item.Key = key
	start := time.Now()

	var files []string
	var keyPath string
	switch item.Mapping {
	case MappingUnique:
		if f, ok := atlasJSON.Unique[key]; ok {
			files = []string{f}
		} else {
			keyPath = fmt.Sprintf("$.unique['%s']", key)
		}
	case MappingVariant:
		variant := cmp.Or(opts.variants[key], "/")
		item.Variant = variant
		if f, ok := atlasJSON.pickFromVariant(key, variant); ok {
			files = []string{f}
		} else {
			keyPath = fmt.Sprintf("$.variant['%s']['%s']", key, variant)
		}
	case MappingMany:
		files = atlasJSON.Many[key]
		if len(files) == 0 {
			keyPath = fmt.Sprintf("$.many['%s']", key)
		}
	default:
		return fmt.Errorf("unsupported mapping: %s", item.Mapping)
	}

	if len(files) == 0 {
		return fmt.Errorf("could not find %s in %s", keyPath, atlasFile)
	}

	for i, f := range files {
		fp := filepath.Join(cfgRoot, f)
		data, err := os.ReadFile(fp)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, item.Cfg)
		if err != nil {
			return fmt.Errorf("failed to unmarshal %q | %w", fp, err)
		}
		if err = readOverrides(f); err != nil {
			return err
		}
		if i > 0 {
			fd.paths += ", "
		}
		fd.paths += fp
	}

	for i, data := range fd.overrides {
		err := json.Unmarshal(data, item.Cfg)
		if err != nil {
			return fmt.Errorf("failed to apply override %q | %w", fd.overrideFiles[i], err)
		}
	}

	applier, ok := item.Cfg.(interface{ ApplyKeys() })
	if ok {
		applier.ApplyKeys()
	}

	var supplement string
	switch len(fd.overrides) {
	case 0:
	case 1:
		supplement = " with 1 override"
	default:
		supplement = fmt.Sprintf(" with %d overrides", len(fd.overrides))
	}
	elapsed := time.Since(start).Milliseconds()
	opts.Info(fmt.Sprintf("<archmage> Loaded (%s) %s%s (%dms)", item.Mapping, fd.paths, supplement, elapsed))
	item.Ready = true
	return nil
}

func compareLower(a, b string) int {
	return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
}

func readOverrideFile(cfg overrideConfig, name string) (string, []byte, error) {
	if cfg.fsys != nil {
		data, err := fs.ReadFile(cfg.fsys, name)
		return name, data, err
	}

	p := filepath.Join(cfg.root, name)
	data, err := os.ReadFile(p)
	return p, data, err
}

type overrideConfig struct {
	// fsys supplies the override files when set.
	fsys fs.FS
	// root is the directory holding the override files. It applies only when
	// fsys is nil: an fs.FS is already rooted at the directory it exposes, so
	// no extra prefix is needed.
	root string
}

type atlasOptions struct {
	Logger

	overrideConfigs []overrideConfig

	loadStrategy    func(iter.Seq2[string, *AtlasItem], AtlasItemLoadFunc) error
	cbAtlasModifier func(atlasJSON *AtlasJSON)

	whitelist []string
	blacklist []string

	variants map[string]string
}

func newAtlasOptions() *atlasOptions {
	return &atlasOptions{
		Logger:          &defaultLogger{},
		cbAtlasModifier: func(atlasJSON *AtlasJSON) {},
		variants:        make(map[string]string),
	}
}

func (opts *atlasOptions) shouldSkip(key string) (string, bool) {
	switch {
	case len(opts.whitelist) > 0:
		return "whitelist", !slices.Contains(opts.whitelist, key)
	case len(opts.blacklist) > 0 && slices.Contains(opts.blacklist, key):
		return "blacklist", true
	default:
		return "", false
	}
}

// Option configures the atlas loading behavior.
type Option func(*atlasOptions)

// WithLogger sets a custom logger for atlas loading operations.
func WithLogger(logger Logger) Option {
	return func(opts *atlasOptions) {
		opts.Logger = logger
	}
}

// WithAtlasModifier registers a callback to modify the atlas JSON data
// after it's loaded but before item processing takes place.
//
// Example:
//
//	archmage.LoadAtlas("atlas.json", "config", atlas,
//	    archmage.WithAtlasModifier(func(aj *archmage.AtlasJSON) {
//	        aj.Unique["item"] = "custom/item.json"
//	    }))
func WithAtlasModifier(cb func(atlasJSON *AtlasJSON)) Option {
	return func(opts *atlasOptions) {
		opts.cbAtlasModifier = cb
	}
}

// WithWhitelist restricts loading to only the specified items by their keys.
//
// Example:
//
//	archmage.LoadAtlas("atlas.json", "config", atlas,
//	    archmage.WithWhitelist([]string{"item", "hero", "skill"}))
func WithWhitelist(whitelist []string) Option {
	return func(opts *atlasOptions) {
		opts.whitelist = whitelist
	}
}

// WithBlacklist prevents loading of the specified items by their keys.
// If a whitelist is also specified, the blacklist is ignored.
//
// Example:
//
//	archmage.LoadAtlas("atlas.json", "config", atlas,
//	    archmage.WithBlacklist([]string{"hero", "skill"}))
func WithBlacklist(blacklist []string) Option {
	return func(opts *atlasOptions) {
		opts.blacklist = blacklist
	}
}

// WithVariant selects the variant to load for the item identified by key.
// A variant-mapped item that is not given a variant falls back to "/".
//
// Example:
//
//	archmage.LoadAtlas("atlas.json", "config", atlas,
//	    archmage.WithVariant("prop_floats", "x5"),
//	    archmage.WithVariant("game", "dev"))
func WithVariant(key, variant string) Option {
	return func(opts *atlasOptions) {
		opts.variants[key] = variant
	}
}

// WithOverrideRoot adds a directory to search for override JSON files
// that will be merged into loaded configurations.
//
// Example:
//
//	archmage.LoadAtlas("atlas.json", "config", atlas,
//	    archmage.WithOverrideRoot("new_feature_override"),
//	    archmage.WithOverrideRoot("local_override"))
func WithOverrideRoot(dir string) Option {
	return func(opts *atlasOptions) {
		opts.overrideConfigs = append(opts.overrideConfigs, overrideConfig{root: dir})
	}
}

// WithOverrideFS adds a filesystem to search for override JSON files
// that will be merged into loaded configurations.
//
// Example:
//
//	fsys := fstest.MapFS{
//	    "item.json": &fstest.MapFile{Data: []byte(`{"1":{"name":"Sword++"}}`)},
//	}
//	archmage.LoadAtlas("atlas.json", "config", atlas,
//	    archmage.WithOverrideFS(fsys))
func WithOverrideFS(fsys fs.FS) Option {
	return func(opts *atlasOptions) {
		opts.overrideConfigs = append(opts.overrideConfigs, overrideConfig{fsys: fsys})
	}
}

// AtlasItemLoadFunc is called to load each atlas item.
type AtlasItemLoadFunc func(ctx context.Context, key string, item *AtlasItem) error

// WithLoadStrategy replaces the default sequential execution with a custom
// implementation, allowing for parallel loading or other strategies.
//
// Example:
//
//	archmage.LoadAtlas("atlas.json", "config", atlas,
//	    archmage.WithLoadStrategy(func(all iter.Seq2[string, *archmage.AtlasItem], load archmage.AtlasItemLoadFunc) error {
//	        eg, ctx := errgroup.WithContext(context.Background())
//	        eg.SetLimit(10)
//	        for k, item := range all {
//	            eg.Go(func() error { return load(ctx, k, item) })
//	        }
//	        return eg.Wait()
//	    }))
func WithLoadStrategy(strategy func(all iter.Seq2[string, *AtlasItem], load AtlasItemLoadFunc) error) Option {
	return func(opts *atlasOptions) {
		opts.loadStrategy = strategy
	}
}
