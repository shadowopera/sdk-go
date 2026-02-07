package internal

import (
	"context"
	"fmt"
	"iter"
	"runtime/debug"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"golang.org/x/sync/errgroup"
	"golang.org/x/text/language"
	"shadop.dev/pkg/sdk-go/archmage"
	"shadop.dev/pkg/sdk-go/archmage/internal/conf"
)

func TestAtlas_Basic(t *testing.T) {
	en := language.English
	cn := language.Chinese
	i10n := archmage.NewI18n(en)
	if err := i10n.MergeL10nFile("testdata/l10n.json", en); err != nil {
		t.Fatal(err)
	}
	if err := i10n.MergeL10nFile("testdata/l10n.cn.json", cn); err != nil {
		t.Fatal(err)
	}
	conf.GetI18n = func() *archmage.I18n {
		return i10n
	}

	var err error
	atlas := conf.NewConfigAtlas()
	err = archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas)
	if err != nil {
		t.Fatal(err)
	}
	err = archmage.DumpAtlas(atlas, "golden/basic")
	if err != nil {
		t.Fatal(err)
	}

	if atlas.GameCfg.XL10n.MustGetText(en) != "it is a good day" {
		t.Fatalf("unexpected l10n en value: %s", atlas.GameCfg.XL10n.MustGetText(en))
	}
	if atlas.GameCfg.XL10n.MustGetText(cn) != "今儿天气真好" {
		t.Fatalf("unexpected l10n cn value: %s", atlas.GameCfg.XL10n.MustGetText(cn))
	}

	itemEntry, ok := atlas.ItemTable[20]
	if !ok {
		t.Fatalf("item 20 not found")
	}
	if itemEntry.ID != 20 {
		t.Fatalf("unexpected item ID: %d", itemEntry.ID)
	}

	if atlas.CharacterArray[0].Race.Ref == nil {
		t.Fatalf("expected Race.Ref to be bound, got nil")
	}
	if atlas.CharacterArray[1].Runes[0].Ref == nil {
		t.Fatalf("expected Runes[0].Ref to be bound, got nil")
	}
	if atlas.GameCfg.XRef.Ref == nil {
		t.Fatalf("expected GameCfg.XRef.Ref to be bound, got nil")
	}
	if atlas.RaceTable["Dwarf"].Referrer2.Ref == nil {
		t.Fatalf("expected RaceTable['Dwarf'].Referrer2.Ref to be bound, got nil")
	}
	if atlas.RefTable[3].B.Ref == nil {
		t.Fatalf("expected RefTable[3].B.Ref to be bound, got nil")
	}
	if atlas.Matrix2Table["key1"]["key2"][0][0].Ref == nil {
		t.Fatalf("expected Matrix2Table['key1']['key2'][0][0].Ref to be bound, got nil")
	}
	if len(atlas.VtItemXTable) != 16 {
		t.Fatalf("expected len(vtItemXTable) = 16, got %d", len(atlas.VtItemXTable))
	}
}

func TestAtlas_WithAtlasModifier(t *testing.T) {
	atlasModifier := func(atlasJSON *archmage.AtlasJSON) {
		atlasJSON.Single["prop_floats"]["/"] = atlasJSON.Single["prop_floats"]["x5"]
		delete(atlasJSON.Unique, "character")
		delete(atlasJSON.Unique, "matrix2")
		delete(atlasJSON.Single, "game")
	}
	notFound := func(key string, atlasItem *archmage.AtlasItem) error {
		switch key {
		case "character":
		case "matrix2":
		case "game":
		default:
			panic("unreachable")
		}
		return nil
	}

	logger := newScavenger()
	opts := []archmage.Option{
		archmage.WithLogger(logger),
		archmage.WithAtlasModifier(atlasModifier),
		archmage.WithNotFoundCallback(notFound),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	err = archmage.DumpAtlas(atlas, "golden/atlas_modifier")
	if err != nil {
		t.Fatal(err)
	}

	if !slices.ContainsFunc(logger.Lines, func(line string) bool {
		return line == "WRN <archmage> cannot find $.unique['character'] in testdata/atlas.json"
	}) {
		t.Fatalf("expected warning log not found")
	}
	if !slices.ContainsFunc(logger.Lines, func(line string) bool {
		return line == "WRN <archmage> cannot find $.single['game']['/'] in testdata/atlas.json"
	}) {
		t.Fatalf("expected warning log not found")
	}

	var cnt int
	for _, line := range logger.Lines {
		if strings.HasPrefix(line, "WRN") {
			cnt++
		}
	}
	if cnt != 3 {
		t.Fatalf("expected 3 warning log, got %d", cnt)
	}
}

func TestAtlas_WithWhitelist(t *testing.T) {
	logger := newScavenger()
	opts := []archmage.Option{
		archmage.WithLogger(logger),
		archmage.WithWhitelist([]string{"Item", "game", "prop_floats", "weapon-rune"}),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	err = archmage.DumpAtlas(atlas, "golden/whitelist")
	if err != nil {
		t.Fatal(err)
	}

	var cnt int
	for _, line := range logger.Lines {
		if strings.HasPrefix(line, "WRN") {
			cnt++
		}
	}
	if cnt != 1 {
		t.Fatalf("expected 1 warning log, got %d", cnt)
	}
}

func TestAtlas_WithWhitelist_Error(t *testing.T) {
	logger := newScavenger()
	opts := []archmage.Option{
		archmage.WithLogger(logger),
		archmage.WithWhitelist([]string{"Item", "prop_float"}),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.HasPrefix(err.Error(), `<archmage> atlas whitelist: unknown item "prop_float"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAtlas_WithBlacklist(t *testing.T) {
	logger := newScavenger()
	opts := []archmage.Option{
		archmage.WithLogger(logger),
		archmage.WithBlacklist([]string{"game", "prop_floats", "character"}),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	err = archmage.DumpAtlas(atlas, "golden/blacklist")
	if err != nil {
		t.Fatal(err)
	}

	var cnt int
	for _, line := range logger.Lines {
		if strings.HasPrefix(line, "WRN") {
			cnt++
		}
	}
	if cnt != 0 {
		t.Fatalf("expected 1 warning log, got %d", cnt)
	}
}

func TestAtlas_WithBlacklist_Error(t *testing.T) {
	logger := newScavenger()
	opts := []archmage.Option{
		archmage.WithLogger(logger),
		archmage.WithBlacklist([]string{"gm"}),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.HasPrefix(err.Error(), `<archmage> atlas blacklist: unknown item "gm"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAtlas_WithOverrideRoot(t *testing.T) {
	logger := newScavenger()
	opts := []archmage.Option{
		archmage.WithLogger(logger),
		archmage.WithOverrideRoot("testdata_override/1"),
		archmage.WithOverrideRoot("testdata_override/2"),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	err = archmage.DumpAtlas(atlas, "golden/override_root")
	if err != nil {
		t.Fatal(err)
	}
}

func TestAtlas_WithOverrideRoot_Error1(t *testing.T) {
	logger := newScavenger()
	opts := []archmage.Option{
		archmage.WithLogger(logger),
		archmage.WithOverrideRoot("testdata_override/9"),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.HasPrefix(err.Error(), `<archmage> invalid override root directory "testdata_override/9"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAtlas_WithOverrideRoot_Error2(t *testing.T) {
	logger := newScavenger()
	opts := []archmage.Option{
		archmage.WithLogger(logger),
		archmage.WithOverrideRoot("testdata_override/1/game.json"),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.HasPrefix(err.Error(), `<archmage> override root "testdata_override/1/game.json" is not a directory`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAtlas_WithOverrideFS(t *testing.T) {
	fsys := fstest.MapFS{}
	fsys["game.json"] = &fstest.MapFile{
		Data: []byte(`{"x-string":"foo bar","x-map":{"7":"xxx","9":"rab"}}`),
	}
	fsys["clutter/magic.json"] = &fstest.MapFile{
		Data: []byte(`{"200":{"name":"Power Word: Shield"}}`),
	}

	logger := newScavenger()
	opts := []archmage.Option{
		archmage.WithLogger(logger),
		archmage.WithWhitelist([]string{"game", "Magic", "weapon-rune"}),
		archmage.WithOverrideRoot("testdata_override/2"),
		archmage.WithOverrideFS(fsys),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	err = archmage.DumpAtlas(atlas, "golden/override_fs")
	if err != nil {
		t.Fatal(err)
	}
}

func TestAtlas_WithOverrideRootAndFS(t *testing.T) {
	fsys := fstest.MapFS{}
	fsys["vtbl/weapon-sword.json"] = &fstest.MapFile{
		Data: []byte(`{"1000":{"name":"Dragonfang Blade","price":1200}}`),
	}
	fsys["vtbl/weapon-staff.json"] = &fstest.MapFile{
		Data: []byte(`{"1201":{"price":2050,"dps":2}}`),
	}

	logger := newScavenger()
	opts := []archmage.Option{
		archmage.WithLogger(logger),
		archmage.WithOverrideRoot("testdata_override/1"),
		archmage.WithOverrideRoot("testdata_override/2"),
		archmage.WithOverrideFS(fsys),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	err = archmage.DumpAtlas(atlas, "golden/override_root_and_fs")
	if err != nil {
		t.Fatal(err)
	}
}

func TestAtlas_WithCustomLoader(t *testing.T) {
	customLoader := func(all iter.Seq2[string, *archmage.AtlasItem], load archmage.AtlasItemLoadFunc) error {
		eg, ctx := errgroup.WithContext(context.Background())
		eg.SetLimit(10)
		for k, item := range all {
			eg.Go(func() (err error) {
				defer func() {
					if r := recover(); r != nil {
						if e, ok := r.(error); ok {
							err = e
						} else {
							err = fmt.Errorf("<archmage> panic: %+v. stack:\n%s", r, debug.Stack())
						}
					}
				}()
				return load(ctx, k, item)
			})
		}
		return eg.Wait()
	}
	logger := newScavenger()
	opts := []archmage.Option{
		archmage.WithLogger(logger),
		archmage.WithCustomLoader(customLoader),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	err = archmage.DumpAtlas(atlas, "golden/custom_loader")
	if err != nil {
		t.Fatal(err)
	}

	if !slices.ContainsFunc(logger.Lines, func(line string) bool {
		return line == "WRN <archmage> cannot find $.single['prop_floats']['/'] in testdata/atlas.json"
	}) {
		t.Fatalf("expected warning log not found")
	}

	var cnt int
	for _, line := range logger.Lines {
		if strings.HasPrefix(line, "WRN") {
			cnt++
		}
	}
	if cnt != 1 {
		t.Fatalf("expected 1 warning log, got %d", cnt)
	}
}

func TestAtlas_NotFoundCallback(t *testing.T) {
	atlas := conf.NewConfigAtlas()
	notFound := func(key string, atlasItem *archmage.AtlasItem) error {
		switch key {
		case "prop_floats":
			atlas.PropFloatsCfg.C = 100
			atlasItem.Ready = true
		default:
			panic("unreachable")
		}
		return nil
	}

	logger := newScavenger()
	opts := []archmage.Option{
		archmage.WithLogger(logger),
		archmage.WithNotFoundCallback(notFound),
	}

	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	err = archmage.DumpAtlas(atlas, "golden/not_found")
	if err != nil {
		t.Fatal(err)
	}

	var cnt int
	for _, line := range logger.Lines {
		if strings.HasPrefix(line, "WRN") {
			cnt++
		}
	}
	if cnt != 0 {
		t.Fatalf("expected 0 warning log, got %d", cnt)
	}
}
