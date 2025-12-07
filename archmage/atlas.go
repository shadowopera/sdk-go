package archmage

import (
	"cmp"
	"context"
	"encoding/json/v2"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/sync/errgroup"
)

type Atlas interface {
	AtlasItems() map[string]*AtlasItem
	BindRefs()
}

type AtlasItem struct {
	Cfg   any
	Arity string
}

func LoadAtlas(atlasFile string, rootDir string, out Atlas, opts ...Option) error {
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
				p = filepath.Join(rootDir, f)
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
					p = filepath.Join(rootDir, f)
					data, err = os.ReadFile(p)
					if err != nil {
						return err
					}
				} else {
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
