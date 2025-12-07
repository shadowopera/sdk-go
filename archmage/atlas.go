package archmage

import (
	"encoding/json/v2"
	"os"
	"path/filepath"
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

	var data []byte
	var p string
	for k, item := range out.AtlasItems() {
		data = nil
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
				continue
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
					atlOpts.Infof("cannot find $.multiple['%s']['/'] in %s", k, atlasFile)
					if err = atlOpts.cbNotFound(k, item); err != nil {
						return err
					}
					continue
				}
			} else {
				atlOpts.Warnf("cannot find $.multiple['%s'] in %s", k, atlasFile)
				if err = atlOpts.cbNotFound(k, item); err != nil {
					return err
				}
				continue
			}
		default:
			panic("unsupported arity: " + item.Arity)
		}

		err = json.Unmarshal(data, item.Cfg)
		if err != nil {
			return err
		}

		atlOpts.Infof("successfully loaded %s", p)
	}

	return nil
}

type atlasOptions struct {
	Logger
	cbNotFound func(name string, atlasItem *AtlasItem) error
}

type Option func(*atlasOptions)

func WithLogger(logger Logger) Option {
	return func(opts *atlasOptions) {
		opts.Logger = logger
	}
}

func WithNotFoundCallback(cb func(name string, atlasItem *AtlasItem) error) Option {
	return func(opts *atlasOptions) {
		opts.cbNotFound = cb
	}
}
