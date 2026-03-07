package internal

import (
	"bytes"
	"encoding/json/v2"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"shadop.dev/pkg/sdk-go/archmage"
)

var (
	save = flag.Bool("save", false, "update golden files")
)

// checkSaveAtlas serializes all ready atlas items and either writes them to
// disk (when -save is set) or compares them byte-for-byte with existing
// golden files.
func checkSaveAtlas(t *testing.T, atlas archmage.Atlas, goldenDir string) {
	t.Helper()
	for k, item := range atlas.AtlasItems() {
		if item.Ready {
			data, err := json.Marshal(item.Cfg, archmage.BuildMarshalOptions()...)
			if err != nil {
				t.Fatalf("marshal %s: %v", k, err)
			}
			data = append(data, '\n')
			p := filepath.Join(goldenDir, k+".json")
			if *save {
				if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
					t.Fatalf("mkdir %s: %v", filepath.Dir(p), err)
				}
				if err := os.WriteFile(p, data, 0644); err != nil {
					t.Fatalf("write golden %s: %v", p, err)
				}
				t.Logf("updated golden %s", p)
			} else {
				want, err := os.ReadFile(p)
				if err != nil {
					t.Fatalf("read golden %s: %v (run with -save to create it)", p, err)
				}
				if !bytes.Equal(data, want) {
					t.Errorf("golden mismatch for %s\ngot:\n%s\nwant:\n%s", k, data, want)
				}
			}
		}
	}
}
