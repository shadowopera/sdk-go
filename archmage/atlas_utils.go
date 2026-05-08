package archmage

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DumpAtlas writes all loaded atlas items to JSON files in outputDir.
// Each item is written to a separate file named <key>.json.
func DumpAtlas(atlas Atlas, outputDir string, opts ...json.Options) error {
	for k, item := range atlas.AtlasItems() {
		if item.Ready {
			err := dumpAtlasItem(item, k, outputDir, opts)
			if err != nil {
				return fmt.Errorf("<archmage> failed to dump atlas item %q | %w", k, err)
			}
		}
	}

	return nil
}

func dumpAtlasItem(item *AtlasItem, k string, outputDir string, opts []json.Options) error {
	data, err := Canonicalize(item.Cfg, opts...)
	if err != nil {
		return err
	}
	p := filepath.Join(outputDir, k+".json")
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(p, data, 0644); err != nil {
		return err
	}
	return nil
}

// Canonicalize marshals obj to indented, canonical JSON.
// It applies archmage's standard marshal options (nil maps/slices as null,
// zero time.Time and zero Vec pointers as null), then sorts object keys and
// indents the output. Additional opts are appended after the defaults.
// Returns the formatted JSON bytes or an error.
func Canonicalize(obj any, opts ...json.Options) ([]byte, error) {
	data, err := json.Marshal(obj, getMarshalOptions(opts)...)
	if err != nil {
		return nil, err
	}
	jsonValue := jsontext.Value(data)
	if err := jsonValue.Canonicalize(jsontext.CanonicalizeRawInts(false)); err != nil {
		return nil, err
	}
	if err := jsonValue.Indent(); err != nil {
		return nil, err
	}
	data = append(jsonValue, '\n')
	return data, nil
}

func getMarshalOptions(opts []json.Options) []json.Options {
	return append([]json.Options{
		json.FormatNilMapAsNull(true),
		json.FormatNilSliceAsNull(true),
		json.WithMarshalers(json.JoinMarshalers(
			json.MarshalToFunc[time.Time](func(enc *jsontext.Encoder, t time.Time) error {
				if t.IsZero() {
					return enc.WriteToken(jsontext.Null)
				}
				return json.SkipFunc
			}),
		)),
	}, opts...)
}
