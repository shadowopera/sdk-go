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
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type Atlas interface {
	AtlasItems() map[string]*AtlasItem
	BindRefs()
}

type AtlasItem struct {
	Cfg   any
	Arity string
	File  string
	Ready bool
}

type AtlasJSON struct {
	Single   map[string]string            `json:"single"`
	Multiple map[string]map[string]string `json:"multiple"`
}

func LoadAtlas(atlasFile string, cfgRoot string, out Atlas, opts ...Option) error {
	atlasOpts := newAtlasOptions()
	atlasOpts.readFile = func(name string) ([]byte, error) {
		return os.ReadFile(name)
	}
	for _, opt := range opts {
		opt(atlasOpts)
	}
	return loadAtlasImpl(atlasFile, cfgRoot, out, atlasOpts)
}

func loadAtlasImpl(atlasFile string, cfgRoot string, out Atlas, opts *atlasOptions) error {
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

	items := out.AtlasItems()
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

	sortedKeys := slices.SortedFunc(maps.Keys(items), compareLower)
	filtered := slices.DeleteFunc(sortedKeys, func(k string) bool {
		cause, yes := opts.shouldIgnore(k)
		if yes {
			opts.Infof("<archmage> skipping atlas item: %s. cause: %s", k, cause)
		}
		return yes
	})
	seq := func(yield func(string, *AtlasItem) bool) {
		for _, k := range filtered {
			if !yield(k, items[k]) {
				break
			}
		}
	}
	loadWrapper := func(ctx context.Context, key string, item *AtlasItem) error {
		return loadItem(ctx, key, item, atlasJSON, atlasFile, cfgRoot, opts)
	}

	if opts.customLoader == nil {
		for key, item := range seq {
			err = loadWrapper(context.Background(), key, item)
			if err != nil {
				return err
			}
		}
	} else {
		err = opts.customLoader(seq, loadWrapper)
		if err != nil {
			return err
		}
	}

	out.BindRefs()
	return nil
}

func loadItem(ctx context.Context, key string, item *AtlasItem,
	atlasJSON AtlasJSON, atlasFile string, cfgRoot string, opts *atlasOptions,
) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	var overrideFiles []string
	var overrides [][]byte
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
			file, ovrData, err := readOverrideFile(cfg, file)
			if err != nil {
				return err
			}
			overrideFiles = append(overrideFiles, file)
			overrides = append(overrides, ovrData)
		}
		return nil
	}

	item.File = key
	start := time.Now()
	var err error
	var data []byte
	var p string
	switch item.Arity {
	case "single":
		if f, ok := atlasJSON.Single[key]; ok {
			item.File = f
			p = filepath.Join(cfgRoot, f)
			data, err = opts.readFile(p)
			if err != nil {
				return err
			}
			if err = readOverrides(f); err != nil {
				return err
			}
		} else {
			if err = opts.cbNotFound(key, item); err != nil {
				return err
			}
			if !item.Ready {
				opts.Warnf("<archmage> cannot find $.single['%s'] in %s", key, atlasFile)
			}
			return nil
		}
	case "multiple":
		if m, ok := atlasJSON.Multiple[key]; ok {
			if f, ok := m["/"]; ok {
				item.File = f
				p = filepath.Join(cfgRoot, f)
				data, err = opts.readFile(p)
				if err != nil {
					return err
				}
				if err = readOverrides(f); err != nil {
					return err
				}
			} else {
				multiple := atlasJSON.Multiple[key]
				sortedKeys := slices.SortedFunc(maps.Keys(multiple), compareLower)
				for _, k := range sortedKeys {
					v := multiple[k]
					dir, file := path.Split(v)
					item.File = file
					if x1 := path.Ext(file); x1 != "" && x1 != file {
						base1 := file[:len(file)-len(x1)]
						if x2 := path.Ext(base1); x2 != "" && x2 != base1 {
							base2 := base1[:len(base1)-len(x2)]
							file2 := base2 + x1
							item.File = path.Join(dir, file2)
							break
						}
					}
				}
				if err = opts.cbNotFound(key, item); err != nil {
					return err
				}
				if !item.Ready {
					opts.Warnf("<archmage> cannot find $.multiple['%s']['/'] in %s", key, atlasFile)
				}
				return nil
			}
		} else {
			if err = opts.cbNotFound(key, item); err != nil {
				return err
			}
			if !item.Ready {
				opts.Warnf("<archmage> cannot find $.multiple['%s'] in %s", key, atlasFile)
			}
			return nil
		}
	default:
		panic("<archmage> unsupported arity: " + item.Arity)
	}

	if !item.Ready {
		err = json.Unmarshal(data, item.Cfg)
		if err != nil {
			return fmt.Errorf("<archmage> invalid %q | %w", item.File, err)
		}
		for i, data := range overrides {
			err = json.Unmarshal(data, item.Cfg)
			if err != nil {
				return fmt.Errorf("<archmage> applying override %s failed | %w", overrideFiles[i], err)
			}
		}
	}

	applier, ok := item.Cfg.(interface{ ApplyKeys() })
	if ok {
		applier.ApplyKeys()
	}

	var supplement string
	switch len(overrides) {
	case 0:
	case 1:
		supplement = " with 1 override"
	default:
		supplement = fmt.Sprintf(" with %d overrides", len(overrides))
	}
	opts.Infof("<archmage> loaded %s%s (%dms)", p, supplement, time.Since(start).Milliseconds())
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
	} else {
		p := filepath.Join(cfg.root, name)
		data, err := os.ReadFile(p)
		return p, data, err
	}
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

func (opts *atlasOptions) shouldIgnore(key string) (string, bool) {
	switch {
	case opts.whitelist != nil:
		return "whitelist", !slices.Contains(opts.whitelist, key)
	case opts.blacklist != nil && slices.Contains(opts.blacklist, key):
		return "blacklist", true
	default:
		return "", false
	}
}

type Option func(*atlasOptions)

func WithLogger(logger Logger) Option {
	return func(opts *atlasOptions) {
		opts.Logger = logger
	}
}

type AtlasItemLoadFunc func(ctx context.Context, key string, item *AtlasItem) error

func WithCustomLoader(loader func(all iter.Seq2[string, *AtlasItem], load AtlasItemLoadFunc) error) Option {
	return func(opts *atlasOptions) {
		opts.customLoader = loader
	}
}

func WithNotFoundCallback(cb func(key string, atlasItem *AtlasItem) error) Option {
	return func(opts *atlasOptions) {
		opts.cbNotFound = cb
	}
}

func WithAtlasModifier(cb func(atlasJSON *AtlasJSON)) Option {
	return func(opts *atlasOptions) {
		opts.cbAtlasModifier = cb
	}
}

func WithWhitelist(whitelist []string) Option {
	return func(opts *atlasOptions) {
		opts.whitelist = whitelist
	}
}

func WithBlacklist(blacklist []string) Option {
	return func(opts *atlasOptions) {
		opts.blacklist = blacklist
	}
}

func WithOverrideRoot(dir string) Option {
	return func(opts *atlasOptions) {
		opts.overrideConfigs = append(opts.overrideConfigs, overrideConfig{root: dir})
	}
}

func WithOverrideFS(fsys fs.FS) Option {
	return func(opts *atlasOptions) {
		opts.overrideConfigs = append(opts.overrideConfigs, overrideConfig{fsys: fsys})
	}
}
