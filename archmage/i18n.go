package archmage

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"

	"golang.org/x/text/language"
)

// I18n manages localized text translations across multiple languages.
// It supports loading translations from JSON files or raw data, and
// automatically falls back to a default language when translations are missing.
type I18n struct {
	fallback language.Tag
	texts    map[language.Tag]map[string]string
}

// NewI18n creates an I18n instance with the specified fallback language.
func NewI18n(fallback language.Tag) *I18n {
	return &I18n{
		fallback: fallback,
		texts:    make(map[language.Tag]map[string]string),
	}
}

// Fallback returns the fallback language tag.
func (i18n *I18n) Fallback() language.Tag {
	return i18n.fallback
}

// AllTexts returns all loaded translations.
func (i18n *I18n) AllTexts() map[language.Tag]map[string]string {
	return i18n.texts
}

// MergeTexts adds or updates translations for the specified language.
func (i18n *I18n) MergeTexts(texts map[string]string, lang language.Tag) {
	if _, ok := i18n.texts[lang]; !ok {
		i18n.texts[lang] = make(map[string]string)
	}
	store := i18n.texts[lang]
	for k, v := range texts {
		store[k] = v
	}
}

// MergeL10nData parses JSON translation data and merges it for the language.
func (i18n *I18n) MergeL10nData(data []byte, lang language.Tag) error {
	var tmp map[string]string
	if err := json.Unmarshal(data, &tmp); err != nil {
		return fmt.Errorf("<archmage> failed to merge l10n data: %w", err)
	}
	i18n.MergeTexts(tmp, lang)
	return nil
}

// MergeL10nFile reads a JSON translation file and merges it for the language.
func (i18n *I18n) MergeL10nFile(path string, lang language.Tag) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := i18n.MergeL10nData(data, lang); err != nil {
		return fmt.Errorf("<archmage> failed to merge l10n file %q | %w", path, errors.Unwrap(err))
	}

	return nil
}

// GetText returns the translation for key in the specified language.
// It falls back to the fallback language if the key isn't found there.
// It returns an error if the key is missing in both languages.
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

	return "", fmt.Errorf("<archmage> i18n: text not found. key: %q, lang: %s", key, lang.String())
}

// Text returns the translation for key in the specified language, with the
// same fallback behavior as GetText, or panics if not found.
func (i18n *I18n) Text(key string, lang language.Tag) string {
	x, err := i18n.GetText(key, lang)
	if err != nil {
		panic(err)
	}
	return x
}
