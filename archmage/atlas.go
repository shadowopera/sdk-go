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
	// MappingUnique indicates one-to-one mapping between key and file.
	MappingUnique = "unique"
	// MappingSingle indicates a key maps to a single variant file.
	MappingSingle = "single"
	// MappingMultiple indicates a key maps to multiple files.
	MappingMultiple = "multiple"
)

// Atlas represents a collection of configuration items that can be loaded
// from JSON files and bound to references after loading.
type Atlas interface {
	SetVersionInfo(map[string]any)
	AtlasItems() map[string]*AtlasItem
	BindRefs()
}

// AtlasItem represents a single configuration item in the atlas.
// The Cfg field holds the unmarshaled configuration data, and Ready
// indicates whether the item has been successfully loaded.
type AtlasItem struct {
	Cfg     any
	Mapping string
	Key     string
	Ready   bool
}

// AtlasJSON defines the structure of the atlas index file.
// It maps configuration keys to their corresponding file paths using three
// different mapping strategies: unique, single, and multiple.
type AtlasJSON struct {
	VCS      map[string]any               `json:"vcs"`
	Unique   map[string]string            `json:"unique"`
	Single   map[string]map[string]string `json:"single"`
	Multiple map[string][]string          `json:"multiple"`
}

func (atlas *AtlasJSON) pickSingle(key string) (string, bool) {
	m, ok := atlas.Single[key]
	if ok {
		f, ok := m["/"]
		if ok {
			return f, ok
		}
	}
	return "", false
}

// LoadAtlas loads configuration items from an atlas index file and populates
// the provided Atlas with data from JSON files in cfgRoot. It applies any
// override configurations before calling BindRefs on the atlas.
func LoadAtlas(atlasFile string, cfgRoot string, atlas Atlas, opts ...Option) error {
	atlasOpts := newAtlasOptions()
	atlasOpts.readFile = func(name string) ([]byte, error) {
		return os.ReadFile(name)
	}
	atlasOpts.customLoader = func(seq iter.Seq2[string, *AtlasItem], itemLoadFunc AtlasItemLoadFunc) error {
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

	return loadAtlasImpl(atlasFile, cfgRoot, atlas, atlasOpts)
}

func loadAtlasImpl(atlasFile string, cfgRoot string, atlas Atlas, opts *atlasOptions) error {
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

	atlasData, err := opts.readFile(atlasFile)
	if err != nil {
		return err
	}

	var atlasJSON AtlasJSON
	err = json.Unmarshal(atlasData, &atlasJSON)
	if err != nil {
		return fmt.Errorf("<archmage> invalid %q | %w", atlasFile, err)
	}

	opts.cbAtlasModifier(&atlasJSON)
	atlas.SetVersionInfo(atlasJSON.VCS)

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

	keys := slices.SortedFunc(maps.Keys(items), compareLower)
	filtered := slices.DeleteFunc(keys, func(k string) bool {
		cause, yes := opts.shouldSkip(k)
		if yes {
			opts.Infof("<archmage> skipping atlas item: %s. cause: %s", k, cause)
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
	err = opts.customLoader(filteredItemSeq, func(ctx context.Context, key string, item *AtlasItem) error {
		return loadItem(ctx, key, item, &atlasJSON, atlasFile, cfgRoot, opts)
	})
	if err != nil {
		return err
	}

	atlas.BindRefs()
	return nil
}

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
				return err
			}
			fd.overrideFiles = append(fd.overrideFiles, ovrFile)
			fd.overrides = append(fd.overrides, ovrData)
		}
		return nil
	}

	item.Key = key
	start := time.Now()

	var files []string
	var notFoundHint string
	switch item.Mapping {
	case MappingUnique:
		if f, ok := atlasJSON.Unique[key]; ok {
			files = []string{f}
		}
		notFoundHint = fmt.Sprintf("$.unique['%s']", key)
	case MappingSingle:
		if f, ok := atlasJSON.pickSingle(key); ok {
			files = []string{f}
		}
		notFoundHint = fmt.Sprintf("$.single['%s']['/']", key)
	case MappingMultiple:
		files = atlasJSON.Multiple[key]
		notFoundHint = fmt.Sprintf("$.multiple['%s']", key)
	default:
		return fmt.Errorf("<archmage> unsupported mapping: %s", item.Mapping)
	}

	if len(files) == 0 {
		if err := opts.cbNotFound(key, item); err != nil {
			return err
		}
		if !item.Ready {
			opts.Warnf("<archmage> cannot find %s in %s", notFoundHint, atlasFile)
		}
		return nil
	}

	for i, f := range files {
		filePath := filepath.Join(cfgRoot, f)
		fileData, err := opts.readFile(filePath)
		if err != nil {
			return err
		}
		err = json.Unmarshal(fileData, item.Cfg)
		if err != nil {
			return fmt.Errorf("<archmage> invalid %q | %w", f, err)
		}
		if err = readOverrides(f); err != nil {
			return err
		}
		if i > 0 {
			fd.paths += ", "
		}
		fd.paths += filePath
	}

	for i, data := range fd.overrides {
		err := json.Unmarshal(data, item.Cfg)
		if err != nil {
			return fmt.Errorf("<archmage> applying override %s failed | %w", fd.overrideFiles[i], err)
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
	opts.Infof("<archmage> loaded (%s) %s%s (%dms)", item.Mapping, fd.paths, supplement, elapsed)
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
	fsys fs.FS
	root string
}

type atlasOptions struct {
	Logger

	overrideConfigs []overrideConfig

	customLoader    func(iter.Seq2[string, *AtlasItem], AtlasItemLoadFunc) error
	cbAtlasModifier func(atlasJSON *AtlasJSON)
	cbNotFound      func(key string, atlasItem *AtlasItem) error

	whitelist []string
	blacklist []string

	readFile func(name string) ([]byte, error)
}

func newAtlasOptions() *atlasOptions {
	return &atlasOptions{
		Logger:          &defaultLogger{},
		cbAtlasModifier: func(atlasJSON *AtlasJSON) {},
		cbNotFound:      func(string, *AtlasItem) error { return nil },
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
// after it's loaded but before processing items.
//
// Example:
//
//	archmage.LoadAtlas("atlas.json", "config", atlas,
//	    archmage.WithAtlasModifier(func(aj *archmage.AtlasJSON) {
//	        aj.Single["game"]["/"] = aj.Single["game"]["dev"]
//	    }))
func WithAtlasModifier(cb func(atlasJSON *AtlasJSON)) Option {
	return func(opts *atlasOptions) {
		opts.cbAtlasModifier = cb
	}
}

// WithWhitelist restricts loading to only the specified item keys.
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

// WithBlacklist prevents loading of the specified item keys.
// If a whitelist is also specified, the blacklist will be ignored.
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

// WithCustomLoader replaces the default sequential loader with a custom
// implementation, allowing for parallel or specialized loading strategies.
//
// Example:
//
//	archmage.LoadAtlas("atlas.json", "config", atlas,
//	    archmage.WithCustomLoader(func(all iter.Seq2[string, *archmage.AtlasItem], load archmage.AtlasItemLoadFunc) error {
//	        eg, ctx := errgroup.WithContext(context.Background())
//	        eg.SetLimit(10)
//	        for k, item := range all {
//	            eg.Go(func() error { return load(ctx, k, item) })
//	        }
//	        return eg.Wait()
//	    }))
func WithCustomLoader(loader func(all iter.Seq2[string, *AtlasItem], load AtlasItemLoadFunc) error) Option {
	return func(opts *atlasOptions) {
		opts.customLoader = loader
	}
}

// WithNotFoundCallback registers a callback invoked when a configuration
// file isn't found for an item key.
// The callback can set item.Ready to suppress the not-found warning.
//
// Example:
//
//	archmage.LoadAtlas("atlas.json", "config", atlas,
//	    archmage.WithNotFoundCallback(func(key string, item *archmage.AtlasItem) error {
//	        if key == "special_gift" {
//	            atlas.SpecialGiftTable = readyMadeSpecialGiftTable
//	            item.Ready = true
//	        }
//	        return nil
//	    }))
func WithNotFoundCallback(cb func(key string, atlasItem *AtlasItem) error) Option {
	return func(opts *atlasOptions) {
		opts.cbNotFound = cb
	}
}
