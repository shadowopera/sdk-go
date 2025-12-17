package internal

import (
	"context"
	"fmt"
	"iter"
	"runtime/debug"
	"slices"
	"strings"
	"testing"

	"github.com/shadowopera/sdk-go/archmage"
	"github.com/shadowopera/sdk-go/archmage/internal/conf"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/language"
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
		return line == "WRN <archmage> cannot find $.multiple['prop_floats']['/'] in testdata/atlas.json"
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

func TestAtlas_WithAtlasModifier(t *testing.T) {
	atlasModifier := func(atlasJSON *archmage.AtlasJSON) {
		atlasJSON.Multiple["prop_floats"]["/"] = atlasJSON.Multiple["prop_floats"]["x5"]
		delete(atlasJSON.Single, "character")
		delete(atlasJSON.Single, "matrix2")
		delete(atlasJSON.Multiple, "game")
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
		return line == "WRN <archmage> cannot find $.single['character'] in testdata/atlas.json"
	}) {
		t.Fatalf("expected warning log not found")
	}
	if !slices.ContainsFunc(logger.Lines, func(line string) bool {
		return line == "WRN <archmage> cannot find $.multiple['game'] in testdata/atlas.json"
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

func TestAtlas_WithOverridesRoot(t *testing.T) {
	logger := newScavenger()
	opts := []archmage.Option{
		archmage.WithLogger(logger),
		archmage.WithOverridesRoot("override/1"),
		archmage.WithOverridesRoot("override/2"),
	}

	atlas := conf.NewConfigAtlas()
	err := archmage.LoadAtlas("testdata/atlas.json", "testdata", atlas, opts...)
	if err != nil {
		t.Fatal(err)
	}
	err = archmage.DumpAtlas(atlas, "golden/overrides_root")
	if err != nil {
		t.Fatal(err)
	}
}
