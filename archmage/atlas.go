package archmage

import (
	"cmp"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"reflect"
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

func loadItem(ctx context.Context, k string, item *AtlasItem,
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
		if f, ok := atlasJSON.Single[k]; ok {
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
			opts.Warnf("<archmage> cannot find $.single['%s'] in %s", k, atlasFile)
			if err = opts.cbNotFound(k, item); err != nil {
				return err
			}
			return nil
		}
	case "multiple":
		if m, ok := atlasJSON.Multiple[k]; ok {
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
				for _, v := range atlasJSON.Multiple[k] {
					dir, file := path.Split(v)
					if x1 := path.Ext(file); x1 != "" && x1 != file {
						base1 := file[:len(file)-len(x1)]
						if x2 := path.Ext(base1); x2 != "" && x2 != base1 {
							base2 := base1[:len(base1)-len(x2)]
							newFile := base2 + x1
							item.File = path.Join(dir, newFile)
							break
						}
					}
				}
				opts.Warnf("<archmage> cannot find $.multiple['%s']['/'] in %s", k, atlasFile)
				if err = opts.cbNotFound(k, item); err != nil {
					return err
				}
				return nil
			}
		} else {
			opts.Warnf("<archmage> cannot find $.multiple['%s'] in %s", k, atlasFile)
			if err = opts.cbNotFound(k, item); err != nil {
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

func ApplyMapOverride[K comparable, V any, T map[K]V](base *T, data []byte) error {
	var ovr T
	err := json.Unmarshal(data, &ovr)
	if err != nil {
		return err
	}

	m := *base
	if m == nil {
		if ovr != nil {
			m = make(map[K]V)
			*base = m
		}
	}

	for k, v := range ovr {
		m[k] = v
	}
	return nil
}

func BuildJSONKeyToFieldIndexMap[T any](fields map[string]int8) map[string]int {
	var obj T
	x := reflect.ValueOf(obj)
	if x.Kind() != reflect.Struct {
		return nil
	}

	typ := x.Type()
	m := make(map[string]int)
	for i := range typ.NumField() {
		f := typ.Field(i)
		t := f.Tag.Get("json")
		if t == "" {
			continue
		}
		k := t
		p := strings.Index(t, ",")
		if p >= 0 {
			k = strings.TrimSpace(t[:p])
		}
		if fields[k] != 0 {
			m[k] = i
		}
	}

	return m
}

func ApplyStructOverride[T any](base *T, data []byte, typeName string, fields map[string]int8, fieldIndexMap map[string]int) error {
	var tmp map[string]jsontext.Value
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	x := reflect.ValueOf(base).Elem()
	for k, d := range tmp {
		if fields[k] == 0 {
			return fmt.Errorf("%s: unknown object field name %q in override data", typeName, k)
		}
		index, ok := fieldIndexMap[k]
		if !ok {
			continue
		}
		field := x.Field(index)
		switch fields[k] {
		case 1:
			err = json.Unmarshal(d, field.Addr().Interface())
		case 2:
			field.SetZero()
			err = json.Unmarshal(d, field.Addr().Interface())
		case 3:
			err = field.Addr().Interface().(Overridable).ApplyOverride(d)
		default:
			panic("unreachable")
		}
		if err != nil {
			return fmt.Errorf("%s: failed to apply override data to field %q: %w", typeName, k, err)
		}
	}

	return nil
}
