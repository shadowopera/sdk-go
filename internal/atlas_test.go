package internal

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"runtime/debug"
	"strings"
	"testing"
	"testing/fstest"

	"golang.org/x/sync/errgroup"
	"golang.org/x/text/language"
	"shadop.dev/pkg/sdk-go/archmage"
	"shadop.dev/pkg/sdk-go/internal/conf"
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
	err = archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas,
		archmage.WithBlacklist([]string{"prop_floats"}),
	)
	if err != nil {
		t.Fatal(err)
	}
	checkSaveAtlas(t, atlas, "golden/basic")

	if atlas.GameCfg.XL10n.Text(en) != "it is a good day" {
		t.Fatalf("unexpected l10n en value: %s", atlas.GameCfg.XL10n.Text(en))
	}
	if atlas.GameCfg.XL10n.Text(cn) != "今儿天气真好" {
		t.Fatalf("unexpected l10n cn value: %s", atlas.GameCfg.XL10n.Text(cn))
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
	if atlas.DataVersion != nil {
		t.Fatalf("expected DataVersion to be nil, got %v", atlas.DataVersion)
	}
	if conf.CodeVersion() == nil {
		t.Fatalf("expected CodeVersion to be non-nil")
	}
	if conf.CodeVersion().ShortID != "7f3a2b9" {
		t.Fatalf("expected CodeVersion.ShortID to be 7f3a2b9, got %v", conf.CodeVersion().ShortID)
	}
}

func TestAtlas_DataVersion(t *testing.T) {
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithBlacklist([]string{"prop_floats"}),
	}

	var err error
	atlas := conf.NewConfigAtlas()
	err = archmage.LoadAtlas("testdata/atlas_with_version.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}

	if atlas.DataVersion == nil {
		t.Fatalf("expected DataVersion to be not nil")
	}
	if atlas.DataVersion.Branch != "main" {
		t.Fatalf(`expected atlas.DataVersion.Branch to be "main", got "%s"`, atlas.DataVersion.Branch)
	}
}

func TestAtlas_WithAtlasModifier(t *testing.T) {
	atlasModifier := func(atlasJSON *archmage.AtlasJSON) {
		atlasJSON.Single["prop_floats"]["/"] = atlasJSON.Single["prop_floats"]["x5"]
		delete(atlasJSON.Unique, "character")
		delete(atlasJSON.Unique, "matrix2")
		delete(atlasJSON.Single, "game")
	}

	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithAtlasModifier(atlasModifier),
		archmage.WithBlacklist([]string{"character", "matrix2", "game"}),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	checkSaveAtlas(t, atlas, "golden/atlas_modifier")
}

func TestAtlas_WithWhitelist(t *testing.T) {
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithWhitelist([]string{"Item", "game", "weapon-rune"}),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	checkSaveAtlas(t, atlas, "golden/whitelist")
}

func TestAtlas_WithWhitelist_Error(t *testing.T) {
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
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
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithBlacklist([]string{"game", "prop_floats", "character"}),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	checkSaveAtlas(t, atlas, "golden/blacklist")
}

func TestAtlas_WithBlacklist_Error(t *testing.T) {
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
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
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithBlacklist([]string{"prop_floats"}),
		archmage.WithOverrideRoot("override/1"),
		archmage.WithOverrideRoot("override/2"),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	checkSaveAtlas(t, atlas, "golden/override_root")
}

func TestAtlas_WithOverrideRoot_Error1(t *testing.T) {
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithOverrideRoot("override/9"),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.HasPrefix(err.Error(), `<archmage> invalid override root directory "override/9"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAtlas_WithOverrideRoot_Error2(t *testing.T) {
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithOverrideRoot("override/1/game.json"),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.HasPrefix(err.Error(), `<archmage> override root "override/1/game.json" is not a directory`) {
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

	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithWhitelist([]string{"game", "Magic", "weapon-rune"}),
		archmage.WithOverrideRoot("override/2"),
		archmage.WithOverrideFS(fsys),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	checkSaveAtlas(t, atlas, "golden/override_fs")
}

func TestAtlas_WithOverrideRootAndFS(t *testing.T) {
	fsys := fstest.MapFS{}
	fsys["vtbl/weapon-sword.json"] = &fstest.MapFile{
		Data: []byte(`{"1000":{"name":"Dragonfang Blade","price":1200}}`),
	}
	fsys["vtbl/weapon-staff.json"] = &fstest.MapFile{
		Data: []byte(`{"1201":{"price":2050,"dps":2}}`),
	}

	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithBlacklist([]string{"prop_floats"}),
		archmage.WithOverrideRoot("override/1"),
		archmage.WithOverrideRoot("override/2"),
		archmage.WithOverrideFS(fsys),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	checkSaveAtlas(t, atlas, "golden/override_root_and_fs")
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
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithBlacklist([]string{"prop_floats"}),
		archmage.WithCustomLoader(customLoader),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	checkSaveAtlas(t, atlas, "golden/custom_loader")
}

func TestAtlas_NotFoundCallback(t *testing.T) {
	atlas := conf.NewConfigAtlas()
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
	}

	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != `<archmage> cannot find $.single['prop_floats']['/'] in testdata/atlas.json` {
		t.Fatalf("unexpected error, got %s", err)
	}
}

func TestAtlas_AtlasFileNotFound(t *testing.T) {
	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/nonexistent_atlas.json", "testdata", atlas)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAtlas_InvalidAtlasJSON(t *testing.T) {
	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas_invalid.json", "testdata", atlas)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.HasPrefix(err.Error(), `<archmage> invalid "testdata/atlas_invalid.json"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAtlas_ConfigFileNotFound(t *testing.T) {
	atlasModifier := func(atlasJSON *archmage.AtlasJSON) {
		atlasJSON.Unique["Item"] = "nonexistent/item.json"
	}
	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas,
		archmage.WithLogger(newScavenger()),
		archmage.WithAtlasModifier(atlasModifier),
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAtlas_ContextCancellation(t *testing.T) {
	customLoader := func(all iter.Seq2[string, *archmage.AtlasItem], load archmage.AtlasItemLoadFunc) error {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		for k, item := range all {
			if err := load(ctx, k, item); err != nil {
				return err
			}
		}
		return nil
	}
	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas,
		archmage.WithCustomLoader(customLoader),
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func TestAtlas_InvalidOverrideJSON(t *testing.T) {
	fsys := fstest.MapFS{}
	fsys["clutter/item.json"] = &fstest.MapFile{
		Data: []byte(`{invalid json}`),
	}
	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas,
		archmage.WithLogger(newScavenger()),
		archmage.WithWhitelist([]string{"Item"}),
		archmage.WithOverrideFS(fsys),
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.HasPrefix(err.Error(), `<archmage> applying override clutter/item.json failed`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
