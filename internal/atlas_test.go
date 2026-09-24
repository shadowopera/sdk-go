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
	"shadop.dev/pkg/sdk-go/internal/enums"
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
	conf.GetPreferredLanguage = func() language.Tag {
		return language.Chinese
	}

	var err error
	atlas := conf.NewConfigAtlas()
	err = archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas,
		archmage.WithBlacklist([]string{"balance"}),
	)
	if err != nil {
		t.Fatal(err)
	}
	checkUpdateGoldenFiles(t, atlas, "golden/basic")
	if n := checkXRefs(t, atlas); n == 0 {
		t.Fatalf("expected bound XRefs, got none")
	}

	if text, err := atlas.GameCfg.Title.GetText(en); err != nil || text != "Legends of Avalon" {
		t.Fatalf("unexpected l10n en value: %s", text)
	}
	if text := atlas.GameCfg.Title.Text(); text != "阿瓦隆传说" {
		t.Fatalf("unexpected l10n cn value: %s", text)
	}
	if text := atlas.HeroTable[1].Name.Text(); text != "Arthur Pendragon" {
		t.Fatalf("unexpected l10n fallback value: %s", text)
	}
	if text := atlas.RaceTable["Elf"].Birthplace.Text(); text != "Silverwood" {
		t.Fatalf("unexpected l10n fallback value: %s", text)
	}
	if text, err := atlas.HeroTable[4].Name.GetText(cn); err != nil || text != "" {
		t.Fatalf("unexpected blank l10n value: %q, %v", text, err)
	}
	if text := atlas.HeroTable[4].Name.Text(); text != "" {
		t.Fatalf("unexpected blank l10n value: %q", text)
	}

	if key := enums.HeroClassWarrior.L10nKey(); key != "enum::HeroClass.Warrior" {
		t.Fatalf("unexpected HeroClassWarrior.L10nKey: %s", key)
	}
	if key := enums.HeroClassRanger.L10nKey(); key != "" {
		t.Fatalf("expected empty HeroClassRanger.L10nKey, got %s", key)
	}
	if text := i10n.Text(enums.HeroClassWarrior.L10nKey(), cn); text != "战士" {
		t.Fatalf("unexpected HeroClassWarrior cn text: %s", text)
	}

	itemEntry, ok := atlas.ItemTable[101]
	if !ok {
		t.Fatalf("item 101 not found")
	}
	if itemEntry.ID != 101 {
		t.Fatalf("unexpected item ID: %d", itemEntry.ID)
	}

	conf.GetConfigAtlas = func() *conf.ConfigAtlas {
		return atlas
	}
	if conf.ItemCfgID(101).Cfg() != itemEntry {
		t.Fatalf("conf.GetConfigAtlas does not work correctly")
	}
	if conf.RaceCfgID("Elf").Cfg() != atlas.RaceTable["Elf"] {
		t.Fatalf("conf.GetConfigAtlas does not work correctly")
	}

	if atlas.DataVersion == nil {
		t.Fatalf("expected DataVersion to be non-nil")
	}
	if atlas.DataVersion.Semver != "v1.0.0" {
		t.Fatalf(`expected atlas.DataVersion.Semver to be "v1.0.0", got "%s"`, atlas.DataVersion.Semver)
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
		archmage.WithBlacklist([]string{"balance"}),
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
		atlasJSON.Variant["balance"]["/"] = atlasJSON.Variant["balance"]["hard"]
		delete(atlasJSON.Unique, "chapter")
		delete(atlasJSON.Unique, "route")
		delete(atlasJSON.Variant, "game")
	}

	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithAtlasModifier(atlasModifier),
		archmage.WithBlacklist([]string{"chapter", "route", "game"}),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	checkUpdateGoldenFiles(t, atlas, "golden/atlas_modifier")
}

func TestAtlas_WithWhitelist(t *testing.T) {
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithWhitelist([]string{"hero", "item", "Race", "skill"}),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	checkUpdateGoldenFiles(t, atlas, "golden/whitelist")
}

func TestAtlas_WithWhitelist_Error(t *testing.T) {
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithWhitelist([]string{"item", "balanc"}),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.HasPrefix(err.Error(), `<archmage> atlas whitelist: unknown item "balanc"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAtlas_WithBlacklist(t *testing.T) {
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithBlacklist([]string{"balance", "game", "chapter"}),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	checkUpdateGoldenFiles(t, atlas, "golden/blacklist")
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

func TestAtlas_WithVariant(t *testing.T) {
	for _, variant := range []string{"easy", "hard"} {
		t.Run(variant, func(t *testing.T) {
			opts := []archmage.Option{
				archmage.WithLogger(newScavenger()),
				archmage.WithWhitelist([]string{"balance"}),
				archmage.WithVariant("balance", variant),
			}

			atlas := conf.NewConfigAtlas()
			err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
			if err != nil {
				t.Fatal(err)
			}
			checkUpdateGoldenFiles(t, atlas, "golden/variant_"+variant)

			if v := atlas.AtlasItems()["balance"].Variant; v != variant {
				t.Fatalf(`expected Variant to be %q, got %q`, variant, v)
			}
		})
	}
}

func TestAtlas_WithVariant_Default(t *testing.T) {
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithWhitelist([]string{"game", "hero", "item", "Race", "skill"}),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}

	items := atlas.AtlasItems()
	if v := items["game"].Variant; v != "/" {
		t.Fatalf(`expected game Variant to be "/", got %q`, v)
	}
	if v := items["hero"].Variant; v != "" {
		t.Fatalf(`expected hero Variant to be empty, got %q`, v)
	}
	if v := items["skill"].Variant; v != "" {
		t.Fatalf(`expected skill Variant to be empty, got %q`, v)
	}
}

func TestAtlas_WithVariant_LastWins(t *testing.T) {
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithWhitelist([]string{"balance"}),
		archmage.WithVariant("balance", "easy"),
		archmage.WithVariant("balance", "hard"),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	checkUpdateGoldenFiles(t, atlas, "golden/variant_hard")
}

func TestAtlas_WithVariant_UnknownItem(t *testing.T) {
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithBlacklist([]string{"balance"}),
		archmage.WithVariant("balanc", "hard"),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.HasPrefix(err.Error(), `<archmage> atlas variant: unknown item "balanc"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAtlas_WithVariant_EmptyVariant(t *testing.T) {
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithWhitelist([]string{"balance"}),
		archmage.WithVariant("balance", ""),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.HasPrefix(err.Error(), `<archmage> atlas variant: empty variant for item "balance"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAtlas_WithVariant_NotFound(t *testing.T) {
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithWhitelist([]string{"balance"}),
		archmage.WithVariant("balance", "medium"),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != `<archmage> failed to load atlas item "balance". atlasFile: testdata/atlas.json, cfgRoot: testdata | `+
		`could not find $.variant['balance']['medium'] in testdata/atlas.json` {
		t.Fatalf("unexpected error, got %s", err)
	}
}

func TestAtlas_WithVariant_SkippedItem(t *testing.T) {
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithBlacklist([]string{"balance"}),
		archmage.WithVariant("balance", "medium"),
		archmage.WithVariant("hero", "hard"),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}

	items := atlas.AtlasItems()
	if items["balance"].Ready {
		t.Fatal("expected balance to be skipped")
	}
	if v := items["hero"].Variant; v != "" {
		t.Fatalf(`expected hero Variant to be empty, got %q`, v)
	}
}

func TestAtlas_WithOverrideRoot(t *testing.T) {
	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithBlacklist([]string{"balance"}),
		archmage.WithOverrideRoot("override/1"),
		archmage.WithOverrideRoot("override/2"),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	checkUpdateGoldenFiles(t, atlas, "golden/override_root")
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
		Data: []byte(`{"bgm":"audio/night.ogg","levelRewards":{"40":104},"motd":{"item0":"Hello"}}`),
	}
	fsys["item.json"] = &fstest.MapFile{
		Data: []byte(`{"105":{"name":"Aegis of Camelot","tags":["shield"]}}`),
	}

	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithWhitelist([]string{"game", "hero", "item", "Race", "skill"}),
		archmage.WithOverrideRoot("override/2"),
		archmage.WithOverrideFS(fsys),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	checkUpdateGoldenFiles(t, atlas, "golden/override_fs")
}

func TestAtlas_WithOverrideRootAndFS(t *testing.T) {
	fsys := fstest.MapFS{}
	fsys["vtbl/skill-magic.json"] = &fstest.MapFile{
		Data: []byte(`{"heal":{"mana":25,"cooldown":[0,8]}}`),
	}
	fsys["vtbl/skill-passive.json"] = &fstest.MapFile{
		Data: []byte(`{"aura":{"radius":7}}`),
	}

	opts := []archmage.Option{
		archmage.WithLogger(newScavenger()),
		archmage.WithBlacklist([]string{"balance"}),
		archmage.WithOverrideRoot("override/1"),
		archmage.WithOverrideRoot("override/2"),
		archmage.WithOverrideFS(fsys),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	checkUpdateGoldenFiles(t, atlas, "golden/override_root_and_fs")
}

func TestAtlas_WithLoadStrategy(t *testing.T) {
	loadStrategy := func(all iter.Seq2[string, *archmage.AtlasItem], load archmage.AtlasItemLoadFunc) error {
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
		archmage.WithBlacklist([]string{"balance"}),
		archmage.WithLoadStrategy(loadStrategy),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	checkUpdateGoldenFiles(t, atlas, "golden/custom_loader")
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
	if err.Error() != `<archmage> failed to load atlas item "balance". atlasFile: testdata/atlas.json, cfgRoot: testdata | `+
		`could not find $.variant['balance']['/'] in testdata/atlas.json` {
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
		atlasJSON.Unique["item"] = "nonexistent/item.json"
	}
	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas,
		archmage.WithLogger(newScavenger()),
		archmage.WithAtlasModifier(atlasModifier),
		archmage.WithBlacklist([]string{"balance"}),
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAtlas_ContextCancellation(t *testing.T) {
	loadStrategy := func(all iter.Seq2[string, *archmage.AtlasItem], load archmage.AtlasItemLoadFunc) error {
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
		archmage.WithLoadStrategy(loadStrategy),
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
	fsys["item.json"] = &fstest.MapFile{
		Data: []byte(`{invalid json}`),
	}
	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas,
		archmage.WithLogger(newScavenger()),
		archmage.WithWhitelist([]string{"item"}),
		archmage.WithOverrideFS(fsys),
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.HasPrefix(err.Error(), `<archmage> failed to load atlas item "item". atlasFile: testdata/atlas.json, cfgRoot: testdata | `+
		`failed to apply override "item.json" | jsontext: invalid character`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
