package archmage_test

import (
	"encoding/json/v2"
	"strings"
	"testing"

	"shadop.dev/pkg/sdk-go/archmage"
)

func TestRGBA(t *testing.T) {
	type Trial struct {
		subject  string
		input    string
		expected archmage.RGBA
		expStr   string
		expErr   string
	}

	dataset := []Trial{
		{
			subject:  "empty string resets to zero",
			input:    "",
			expected: archmage.RGBA{},
			expStr:   "#00000000",
		},
		{
			subject:  "#RRGGBB parsed as fully opaque",
			input:    "#FF6600",
			expected: archmage.RGBA{R: 0xFF, G: 0x66, B: 0x00, A: 0xFF},
			expStr:   "#FF6600",
		},
		{
			subject:  "#RRGGBBAA parsed",
			input:    "#FF660080",
			expected: archmage.RGBA{R: 0xFF, G: 0x66, B: 0x00, A: 0x80},
			expStr:   "#FF660080",
		},
		{
			subject:  "lowercase hex parsed",
			input:    "#aabbccdd",
			expected: archmage.RGBA{R: 0xAA, G: 0xBB, B: 0xCC, A: 0xDD},
			expStr:   "#AABBCCDD",
		},
		{
			subject:  "zero color",
			input:    "#00000000",
			expected: archmage.RGBA{},
			expStr:   "#00000000",
		},
		{
			subject: "missing # prefix",
			input:   "FF6600",
			expErr:  "invalid RGBA",
		},
		{
			subject: "wrong length",
			input:   "#FF66",
			expErr:  "invalid RGBA",
		},
		{
			subject: "invalid hex digit",
			input:   "#GGBBCCDD",
			expErr:  "invalid RGBA",
		},
	}

	for _, tt := range dataset {
		t.Run(tt.subject, func(t *testing.T) {
			c, err := archmage.ParseRGBA(tt.input)
			if tt.expErr != "" {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.expErr) {
					t.Fatalf("expected error containing %q, got %v", tt.expErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, c)
			}
			if s := c.String(); s != tt.expStr {
				t.Fatalf("expected String() %q, got %q", tt.expStr, s)
			}

			// round-trip via JSON
			data, err := json.Marshal(c)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}
			want := `"` + tt.expStr + `"`
			switch tt.subject {
			case "empty string resets to zero", "zero color":
				want = `""`
			}
			if string(data) != want {
				t.Fatalf("expected JSON %s, got %s", want, data)
			}

			var c2 archmage.RGBA
			if err = json.Unmarshal(data, &c2); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}
			if c2 != tt.expected {
				t.Fatalf("expected unmarshaled %v, got %v", tt.expected, c2)
			}
		})
	}

	var rgba archmage.RGBA
	rgba.R = 255
	if err := json.Unmarshal([]byte("null"), &rgba); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if rgba != (archmage.RGBA{}) {
		t.Fatalf("expected %v, got %v", rgba, rgba)
	}

	if err := json.Unmarshal([]byte("#12345678"), &rgba); err == nil {
		t.Fatalf("expected unmarshal error, got %v", err)
	}
	if err := json.Unmarshal([]byte(`"#FFF"`), &rgba); err == nil {
		t.Fatalf("expected unmarshal error, got %v", err)
	}
	if err := json.Unmarshal([]byte("255"), &rgba); err == nil || !strings.Contains(err.Error(), "RGBA: invalid JSON token kind") {
		t.Fatalf(`expected error containing "RGBA: invalid JSON token kind", got %v`, err)
	}
}
