package internal

import (
	"bytes"
	"encoding/json/v2"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"shadop.dev/pkg/sdk-go/archmage"
)

var (
	_porting = flag.Bool("porting", false, "compare golden files across language SDKs")
)

type portingLanguage struct {
	name      string
	goldenDir string
}

// portingLanguages lists all language SDKs whose golden files should be compared
// against the Go golden files. Add new entries here as more SDKs are ported.
var portingLanguages = []portingLanguage{
	{name: "cs", goldenDir: "../../sdk-cs/tests/golden"},
}

func TestPortingGoldenFiles(t *testing.T) {
	if !*_porting {
		t.Skip("skipping porting test; run with -porting to enable")
	}

	for _, lang := range portingLanguages {
		comparePortingGoldenRoot(t, "golden", lang.goldenDir, lang.name)
	}
}

func comparePortingGoldenRoot(t *testing.T, goRoot, langRoot string, langName string) {
	goSubs, err := readDirNames(goRoot, true)
	if err != nil {
		t.Fatalf("read go golden root %s: %v", goRoot, err)
	}
	langSubs, err := readDirNames(langRoot, true)
	if err != nil {
		t.Fatalf("read lang golden root %s: %v", langRoot, err)
	}

	checkNameSets(t, "subdirectory", goSubs, langSubs)

	for _, name := range goSubs {
		goPath := filepath.Join(goRoot, name)
		langPath := filepath.Join(langRoot, name)
		comparePortingGoldenSubdir(t, goPath, langPath, langName)
	}

	switch langName {
	case "cs":
		goPath := filepath.Join(goRoot, "custom_loader")
		langPath := filepath.Join(langRoot, "custom_async_loader")
		comparePortingGoldenSubdir(t, goPath, langPath, langName)
	}
}

func comparePortingGoldenSubdir(t *testing.T, goSubdir, langSubdir string, langName string) {
	goFiles, err := readDirNames(goSubdir, false)
	if err != nil {
		t.Fatalf("read go golden subdir %s: %v", goSubdir, err)
	}
	langFiles, err := readDirNames(langSubdir, false)
	if err != nil {
		t.Fatalf("read lang golden subdir %s: %v", langSubdir, err)
	}

	checkNameSets(t, "file", goFiles, langFiles)

	for _, name := range goFiles {
		goPath := filepath.Join(goSubdir, name)
		langPath := filepath.Join(langSubdir, name)
		comparePortingGoldenFile(t, goPath, langPath, langName)
	}
}

// comparePortingGoldenFile compares one golden file pair.
// The Go file is read as-is (it is already canonical). The language file is
// unmarshalled into a generic value and re-canonicalized with
// archmage.Canonicalize before comparison, so formatting differences are
// normalized away.
//
// If the byte comparison fails, a looser semantic comparison is performed
// where JSON null is considered equal to an empty string "". Only when that
// also fails is an error reported.
func comparePortingGoldenFile(t *testing.T, goFile, langFile string, langName string) {
	goData, err := os.ReadFile(goFile)
	if err != nil {
		t.Fatalf("read go golden %s: %v", goFile, err)
	}

	langRaw, err := os.ReadFile(langFile)
	if err != nil {
		t.Fatalf("read lang golden %s: %v", langFile, err)
	}
	var langObj any
	if err := json.Unmarshal(langRaw, &langObj); err != nil {
		t.Fatalf("unmarshal lang golden %s: %v", langFile, err)
	}
	langData, err := archmage.Canonicalize(langObj)
	if err != nil {
		t.Fatalf("canonicalize lang golden %s: %v", langFile, err)
	}

	if bytes.Equal(goData, langData) {
		fmt.Printf("[v] %s\n", strings.TrimLeft(langFile, "./\\"))
		return
	}

	// Bytes differ; fall back to a loose semantic comparison that treats
	// JSON null as equal to an empty string "".
	var goObj any
	if err := json.Unmarshal(goData, &goObj); err != nil {
		t.Fatalf("unmarshal go golden %s: %v", goFile, err)
	}
	if jsonEqualLoose(goObj, langObj) {
		fmt.Printf("[v] %s\n", strings.TrimLeft(langFile, "./\\"))
		return
	}

	t.Fatalf("golden mismatch. file: %s\nwant (go):\n%s\n\n%s\n\ngot (%s):\n%s",
		goFile, goData, strings.Repeat("=", 60), langName, langData)
}

// jsonEqualLoose recursively compares two JSON-unmarshalled values with the
// following extra equivalences beyond strict equality:
//   - nil equals any zero value (nil, [], {})
//   - two strings that both parse as RFC3339 timestamps are equal when they
//     represent the same instant in UTC
func jsonEqualLoose(a, b any) bool {
	// nil equals any zero value.
	if a == nil && isJSONZero(b) {
		return true
	}
	if b == nil && isJSONZero(a) {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch av := a.(type) {
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case float64:
		bv, ok := b.(float64)
		return ok && av == bv
	case string:
		bv, ok := b.(string)
		if !ok {
			return false
		}
		if av == bv {
			return true
		}
		return false
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !jsonEqualLoose(av[i], bv[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, va := range av {
			vb, exists := bv[k]
			if !exists {
				return false
			}
			if !jsonEqualLoose(va, vb) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}

// isJSONZero reports whether v is a zero value for its JSON-unmarshalled type:
// nil, [], or {}.
func isJSONZero(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case []any:
		return len(val) == 0
	case map[string]any:
		return len(val) == 0
	}
	return false
}

func readDirNames(dir string, dirsOnly bool) ([]string, error) {
	a, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range a {
		if e.IsDir() == dirsOnly {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func checkNameSets(t *testing.T, kind string, goNames, langNames []string) {
	goSet := buildSet(goNames)
	langSet := buildSet(langNames)
	for name := range goSet {
		if !langSet[name] {
			t.Errorf("%s %q exists in go golden but not in lang golden", kind, name)
		}
	}
	for name := range langSet {
		if !goSet[name] && name != "custom_async_loader" {
			t.Errorf("%s %q exists in lang golden but not in go golden", kind, name)
		}
	}
}

func buildSet(names []string) map[string]bool {
	s := make(map[string]bool, len(names))
	for _, n := range names {
		s[n] = true
	}
	return s
}
