package archmage

import (
	"cmp"
	"context"
	"encoding/json/v2"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

type Atlas interface {
	AtlasItems() map[string]*AtlasItem
	BindRefs()
	SaveOpts(opts any)
	LoadOpts() any
}

type Overridable interface {
	ApplyOverride(data []byte) error
}

type AtlasItem struct {
	Cfg   Overridable
	Arity string
	File  string
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

func LoadAtlasFS(fsys fs.FS, atlasFile string, cfgRoot string, out Atlas, opts ...Option) error {
	atlasOpts := newAtlasOptions()
	atlasOpts.readFile = func(name string) ([]byte, error) {
		return fs.ReadFile(fsys, name)
	}
	for _, opt := range opts {
		opt(atlasOpts)
	}
	return loadAtlasImpl(atlasFile, cfgRoot, out, atlasOpts)
}

func loadAtlasImpl(atlasFile string, cfgRoot string, out Atlas, opts *atlasOptions) (err error) {
	defer func() {
		if err != nil {
			return
		}
		out.BindRefs()
		out.SaveOpts(&atlasOptions{
			whitelist: opts.whitelist,
			blacklist: opts.blacklist,
		})
	}()

	atlasData, err := opts.readFile(atlasFile)
	if err != nil {
		return err
	}

	var atlasJSON AtlasJSON
	err = json.Unmarshal(atlasData, &atlasJSON)
	if err != nil {
		return err
	}

	if opts.maxConcurrency <= 1 {
		items := out.AtlasItems()
		sortedKeys := slices.SortedFunc(maps.Keys(items), compareLower)
		for _, k := range sortedKeys {
			if cause, yes := opts.shouldIgnore(k); yes {
				opts.Infof("<archmage> skipping atlas item: %s. cause: %s", k, cause)
				continue
			}
			if err = loadItem(context.Background(), k, items[k], atlasJSON, atlasFile, cfgRoot, opts); err != nil {
				return err
			}
		}
		return nil
	}

	eg, ctx := errgroup.WithContext(context.Background())
	eg.SetLimit(opts.maxConcurrency)
	for k, item := range out.AtlasItems() {
		if cause, yes := opts.shouldIgnore(k); yes {
			opts.Infof("<archmage> skipping atlas item: %s. cause: %s", k, cause)
			continue
		}
		eg.Go(func() error {
			return loadItem(ctx, k, item, atlasJSON, atlasFile, cfgRoot, opts)
		})
	}
	return eg.Wait()
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
		for _, dir := range opts.overwriteRoots {
			ovr := filepath.Join(dir, file)
			if _, err := os.Stat(ovr); err == nil {
				var ovrData []byte
				ovrData, err = opts.readFile(ovr)
				if err != nil {
					return err
				}
				overrideFiles = append(overrideFiles, ovr)
				overrides = append(overrides, ovrData)
			}
		}
		return nil
	}

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
			opts.Warnf("<archmage> cannot find $.single['%s'] in %s", key, atlasFile)
			if err = opts.cbNotFound(key, item); err != nil {
				return err
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
				opts.Warnf("<archmage> cannot find $.multiple['%s']['/'] in %s", key, atlasFile)
				if err = opts.cbNotFound(key, item); err != nil {
					return err
				}
				return nil
			}
		} else {
			opts.Warnf("<archmage> cannot find $.multiple['%s'] in %s", key, atlasFile)
			if err = opts.cbNotFound(key, item); err != nil {
				return err
			}
			return nil
		}
	default:
		panic("unsupported arity: " + item.Arity)
	}

	err = json.Unmarshal(data, item.Cfg)
	if err != nil {
		return err
	}
	for i, d := range overrides {
		err = item.Cfg.ApplyOverride(d)
		if err != nil {
			return fmt.Errorf("applying override %s failed: %w", overrideFiles[i], err)
		}
	}

	akCfg, ok := item.Cfg.(interface{ ApplyKeys() })
	if ok {
		akCfg.ApplyKeys()
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
	return nil
}

func compareLower(a, b string) int {
	return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
}

type atlasOptions struct {
	Logger
	maxConcurrency int

	overwriteRoots []string

	cbNotFound func(name string, atlasItem *AtlasItem) error

	whitelist []string
	blacklist []string

	readFile func(name string) ([]byte, error)
}

func newAtlasOptions() *atlasOptions {
	return &atlasOptions{
		Logger: &defaultLogger{},
		cbNotFound: func(string, *AtlasItem) error {
			return nil
		},
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

func WithMaxConcurrency(n int) Option {
	return func(opts *atlasOptions) {
		opts.maxConcurrency = n
	}
}

func WithNotFoundCallback(cb func(name string, atlasItem *AtlasItem) error) Option {
	return func(opts *atlasOptions) {
		opts.cbNotFound = cb
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

func WithOverridesRoot(dir string) Option {
	return func(opts *atlasOptions) {
		opts.overwriteRoots = append(opts.overwriteRoots, dir)
	}
}
