package archmage_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/shadowopera/sdk-go/archmage"
	"golang.org/x/text/language"
)

func TestNewI18n(t *testing.T) {
	langs := []language.Tag{language.English, language.Chinese, language.Und}
	for _, lang := range langs {
		t.Run(lang.String(), func(t *testing.T) {
			i18n := archmage.NewI18n(lang)
			if i18n.Fallback() != lang {
				t.Fatalf("expected fallback %v, got %v", lang, i18n.Fallback())
			}
			if i18n.AllTexts() == nil {
				t.Fatalf("expected non-nil texts map")
			}
			if len(i18n.AllTexts()) != 0 {
				t.Fatalf("expected empty texts, got %v", i18n.AllTexts())
			}
		})
	}
}

func TestI18n_MergeTexts(t *testing.T) {
	type Trial struct {
		subject      string
		initialTexts map[string]string
		mergeTexts   map[string]string
		lang         language.Tag
		expected     map[language.Tag]map[string]string
	}

	dataset := []Trial{
		{
			subject:      "merge into empty",
			initialTexts: nil,
			mergeTexts: map[string]string{
				"hello": "Hello",
				"world": "World",
			},
			lang: language.English,
			expected: map[language.Tag]map[string]string{
				language.English: {
					"hello": "Hello",
					"world": "World",
				},
			},
		},
		{
			subject: "merge with existing keys",
			initialTexts: map[string]string{
				"hello": "Hi",
				"foo":   "Bar",
			},
			mergeTexts: map[string]string{
				"hello": "Hello",
				"world": "World",
			},
			lang: language.English,
			expected: map[language.Tag]map[string]string{
				language.English: {
					"hello": "Hello", // overwritten
					"foo":   "Bar",
					"world": "World",
				},
			},
		},
		{
			subject: "merge empty map",
			initialTexts: map[string]string{
				"hello": "Hello",
			},
			mergeTexts: map[string]string{},
			lang:       language.English,
			expected: map[language.Tag]map[string]string{
				language.English: {
					"hello": "Hello",
				},
			},
		},
		{
			subject:      "merge empty map into empty",
			initialTexts: nil,
			mergeTexts:   nil,
			lang:         language.English,
			expected: map[language.Tag]map[string]string{
				language.English: {},
			},
		},
		{
			subject: "merge different languages",
			initialTexts: map[string]string{
				"hello": "Hello",
			},
			mergeTexts: map[string]string{
				"hello": "你好",
			},
			lang: language.Chinese,
			expected: map[language.Tag]map[string]string{
				language.English: {
					"hello": "Hello",
				},
				language.Chinese: {
					"hello": "你好",
				},
			},
		},
	}

	for _, tt := range dataset {
		t.Run(tt.subject, func(t *testing.T) {
			i18n := archmage.NewI18n(language.English)
			if tt.initialTexts != nil {
				i18n.MergeTexts(tt.initialTexts, language.English)
			}
			i18n.MergeTexts(tt.mergeTexts, tt.lang)
			if !reflect.DeepEqual(tt.expected, i18n.AllTexts()) {
				t.Fatalf("expected texts %v, got %v", tt.expected, i18n.AllTexts())
			}
		})
	}
}

func TestI18n_MergeL10nData(t *testing.T) {
	type Trial struct {
		subject  string
		data     []byte
		lang     language.Tag
		expErr   string
		expected map[string]string
	}

	dataset := []Trial{
		{
			subject:  "valid JSON",
			data:     []byte(`{"hello":"Hello","world":"World"}`),
			lang:     language.English,
			expErr:   "",
			expected: map[string]string{"hello": "Hello", "world": "World"},
		},
		{
			subject:  "empty JSON object",
			data:     []byte(`{}`),
			lang:     language.English,
			expErr:   "",
			expected: map[string]string{},
		},
		{
			subject:  "JSON with unicode",
			data:     []byte(`{"hello":"你好","world":"世界"}`),
			lang:     language.Chinese,
			expErr:   "",
			expected: map[string]string{"hello": "你好", "world": "世界"},
		},
		{
			subject:  "invalid JSON",
			data:     []byte(`{invalid json`),
			lang:     language.English,
			expErr:   "jsontext: invalid character",
			expected: nil,
		},
		{
			subject:  "malformed JSON",
			data:     []byte(`{"hello":}`),
			lang:     language.English,
			expErr:   "jsontext: invalid character",
			expected: nil,
		},
		{
			subject:  "null JSON",
			data:     []byte(`null`),
			lang:     language.English,
			expErr:   "",
			expected: map[string]string{},
		},
	}

	for _, tt := range dataset {
		t.Run(tt.subject, func(t *testing.T) {
			i18n := archmage.NewI18n(language.English)
			err := i18n.MergeL10nData(tt.data, tt.lang)
			if tt.expErr != "" {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.HasPrefix(err.Error(), tt.expErr) {
					t.Fatalf("expected error having prefix %q, got %v", tt.expErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if !reflect.DeepEqual(tt.expected, i18n.AllTexts()[tt.lang]) {
					t.Fatalf("expected texts %v, got %v", tt.expected, i18n.AllTexts()[tt.lang])
				}
			}
		})
	}
}

func TestI18n_MergeL10nFile(t *testing.T) {
	langs := []language.Tag{language.English, language.Chinese, language.Und}
	for _, lang := range langs {
		err := archmage.NewI18n(language.English).MergeL10nFile("/nonexistent/path/to/l10n.json", lang)
		if !os.IsNotExist(err) {
			t.Fatalf("expected file not found error for lang %v, got %v", lang, err)
		}
	}
}

func TestI18n_GetText(t *testing.T) {
	type Trial struct {
		subject    string
		setupTexts map[language.Tag]map[string]string
		fallback   language.Tag
		key        string
		lang       language.Tag
		expErr     string
		expected   string
	}

	dataset := []Trial{
		{
			subject: "get existing text in requested language",
			setupTexts: map[language.Tag]map[string]string{
				language.English: {"hello": "Hello"},
			},
			fallback: language.English,
			key:      "hello",
			lang:     language.English,
			expErr:   "",
			expected: "Hello",
		},
		{
			subject: "fallback to default language",
			setupTexts: map[language.Tag]map[string]string{
				language.English: {"hello": "Hello"},
			},
			fallback: language.English,
			key:      "hello",
			lang:     language.Chinese, // not available in Chinese
			expErr:   "",
			expected: "Hello", // falls back to English
		},
		{
			subject: "prefer requested language over fallback",
			setupTexts: map[language.Tag]map[string]string{
				language.English: {"hello": "Hello"},
				language.Chinese: {"hello": "你好"},
			},
			fallback: language.English,
			key:      "hello",
			lang:     language.Chinese,
			expErr:   "",
			expected: "你好",
		},
		{
			subject: "key not found in any language",
			setupTexts: map[language.Tag]map[string]string{
				language.English: {"hello": "Hello"},
			},
			fallback: language.English,
			key:      "goodbye",
			lang:     language.English,
			expErr:   "i18n: text not found",
			expected: "",
		},
		{
			subject: "key not found in requested or fallback language",
			setupTexts: map[language.Tag]map[string]string{
				language.Japanese: {"hello": "こんにちは"},
			},
			fallback: language.English,
			key:      "hello",
			lang:     language.Chinese,
			expErr:   "i18n: text not found",
			expected: "",
		},
		{
			subject:    "empty texts",
			setupTexts: map[language.Tag]map[string]string{},
			fallback:   language.English,
			key:        "hello",
			lang:       language.English,
			expErr:     "i18n: text not found",
			expected:   "",
		},
		{
			subject: "key exists in fallback but not requested language",
			setupTexts: map[language.Tag]map[string]string{
				language.English: {"welcome": "Welcome"},
				language.Spanish: {"hello": "Hola"},
			},
			fallback: language.English,
			key:      "welcome",
			lang:     language.Spanish,
			expErr:   "",
			expected: "Welcome",
		},
	}

	for _, tt := range dataset {
		t.Run(tt.subject, func(t *testing.T) {
			i18n := archmage.NewI18n(tt.fallback)
			for lang, texts := range tt.setupTexts {
				i18n.MergeTexts(texts, lang)
			}

			r, err := i18n.GetText(tt.key, tt.lang)
			if tt.expErr != "" {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.HasPrefix(err.Error(), tt.expErr) {
					t.Fatalf("expected error having prefix %q, got %v", tt.expErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if r != tt.expected {
					t.Fatalf("expected result %q, got %q", tt.expected, r)
				}
			}
		})
	}
}

func TestI18n_MustGetText(t *testing.T) {
	i18n := archmage.NewI18n(language.English)
	i18n.MergeTexts(map[string]string{"hello": "Hello"}, language.English)
	r1 := i18n.MustGetText("hello", language.English)
	if r1 != "Hello" {
		t.Fatalf("expected 'Hello', got %q", r1)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for missing key, but did not panic")
		}
	}()
	i18n.MustGetText("world", language.English)
}
