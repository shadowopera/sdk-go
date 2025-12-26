package archmage

import (
	"encoding/json/v2"
	"fmt"
	"os"

	"golang.org/x/text/language"
)

type I18n struct {
	fallback language.Tag
	texts    map[language.Tag]map[string]string
}

func NewI18n(fallback language.Tag) *I18n {
	return &I18n{
		fallback: fallback,
		texts:    make(map[language.Tag]map[string]string),
	}
}

func (i18n *I18n) Fallback() language.Tag {
	return i18n.fallback
}

func (i18n *I18n) AllTexts() map[language.Tag]map[string]string {
	return i18n.texts
}

func (i18n *I18n) MergeTexts(texts map[string]string, lang language.Tag) {
	if _, ok := i18n.texts[lang]; !ok {
		i18n.texts[lang] = make(map[string]string)
	}
	store := i18n.texts[lang]
	for k, v := range texts {
		store[k] = v
	}
}

func (i18n *I18n) MergeL10nData(data []byte, lang language.Tag) error {
	var tmp map[string]string
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	i18n.MergeTexts(tmp, lang)
	return nil
}

func (i18n *I18n) MergeL10nFile(path string, lang language.Tag) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return i18n.MergeL10nData(data, lang)
}

func (i18n *I18n) GetText(key string, lang language.Tag) (string, error) {
	if m, ok := i18n.texts[lang]; ok {
		if v, ok := m[key]; ok {
			return v, nil
		}
	}
	if m, ok := i18n.texts[i18n.fallback]; ok {
		if v, ok := m[key]; ok {
			return v, nil
		}
	}

	return "", fmt.Errorf("i18n: text not found. key: %q, lang: %s", key, lang.String())
}

func (i18n *I18n) MustGetText(key string, lang language.Tag) string {
	x, err := i18n.GetText(key, lang)
	if err != nil {
		panic(err)
	}
	return x
}
