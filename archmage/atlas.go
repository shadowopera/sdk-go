package archmage

import (
	"cmp"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
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
}

type AtlasItem struct {
	Cfg   any
	Arity string
	File  string
}

func LoadAtlas(atlasFile string, cfgRoot string, out Atlas, opts ...Option) error {
	var atlOpts atlasOptions
	atlOpts.Logger = &defaultLogger{}
	atlOpts.cbNotFound = func(string, *AtlasItem) error { return nil }
	for _, opt := range opts {
		opt(&atlOpts)
	}

	atlasData, err := os.ReadFile(atlasFile)
	if err != nil {
		return err
	}

	var atlasJSON struct {
		Single   map[string]string            `json:"single"`
		Multiple map[string]map[string]string `json:"multiple"`
	}

	err = json.Unmarshal(atlasData, &atlasJSON)
	if err != nil {
		return err
	}

	loadImpl := func(ctx context.Context, k string, item *AtlasItem) error {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		var err error
		var data []byte
		var p string
		switch item.Arity {
		case "single":
			if f, ok := atlasJSON.Single[k]; ok {
				item.File = f
				p = filepath.Join(cfgRoot, f)
				data, err = os.ReadFile(p)
				if err != nil {
					return err
				}
			} else {
				atlOpts.Warnf("cannot find $.single['%s'] in %s", k, atlasFile)
				if err = atlOpts.cbNotFound(k, item); err != nil {
					return err
				}
				return nil
			}
		case "multiple":
			if m, ok := atlasJSON.Multiple[k]; ok {
				if f, ok := m["/"]; ok {
					item.File = f
					p = filepath.Join(cfgRoot, f)
					data, err = os.ReadFile(p)
					if err != nil {
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
					atlOpts.Warnf("cannot find $.multiple['%s']['/'] in %s", k, atlasFile)
					if err = atlOpts.cbNotFound(k, item); err != nil {
						return err
					}
					return nil
				}
			} else {
				atlOpts.Warnf("cannot find $.multiple['%s'] in %s", k, atlasFile)
				if err = atlOpts.cbNotFound(k, item); err != nil {
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

		akCfg, ok := item.Cfg.(interface{ ApplyKeys() })
		if ok {
			akCfg.ApplyKeys()
		}

		atlOpts.Infof("successfully loaded %s", p)
		return nil
	}

	if atlOpts.maxConcurrency <= 1 {
		items := out.AtlasItems()
		sortedKeys := slices.SortedFunc(maps.Keys(items), compareLower)
		for _, k := range sortedKeys {
			if err = loadImpl(context.Background(), k, items[k]); err != nil {
				return err
			}
		}
		return nil
	}

	eg, ctx := errgroup.WithContext(context.Background())
	eg.SetLimit(atlOpts.maxConcurrency)
	for k, item := range out.AtlasItems() {
		eg.Go(func() error {
			return loadImpl(ctx, k, item)
		})
	}
	return eg.Wait()
}

func compareLower(a, b string) int {
	return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
}

type atlasOptions struct {
	Logger
	maxConcurrency int

	cbNotFound func(name string, atlasItem *AtlasItem) error
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

func DumpAtlas(atlas Atlas, outputDir string, opts ...json.Options) error {
	opts = append([]json.Options{
		jsontext.WithIndent("\t"),
		json.Deterministic(true),
		json.FormatNilMapAsNull(true),
		json.FormatNilSliceAsNull(true),
		json.WithMarshalers(json.MarshalToFunc[time.Time](func(enc *jsontext.Encoder, t time.Time) error {
			if t.IsZero() {
				return enc.WriteToken(jsontext.Null)
			}
			return json.SkipFunc
		})),
	}, opts...)

	for k, item := range atlas.AtlasItems() {
		data, err := json.Marshal(item.Cfg, opts...)
		if err != nil {
			return err
		}
		p := filepath.Join(outputDir, cmp.Or(item.File, k+".json"))
		if err = os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			return err
		}
		data = append(data, '\n')
		if err = os.WriteFile(p, data, 0644); err != nil {
			return err
		}
	}

	return nil
}
