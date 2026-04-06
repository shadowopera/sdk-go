package archmage

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
)

const (
	_hexUpper = "0123456789ABCDEF"
)

var (
	_ json.MarshalerTo     = (*RGBA)(nil)
	_ json.UnmarshalerFrom = (*RGBA)(nil)
)

// RGBA represents a color with red, green, blue, and alpha channels.
// It marshals to JSON as "#RRGGBBAA" and unmarshals from "#RRGGBB" or "#RRGGBBAA".
type RGBA struct {
	R, G, B, A uint8
}

// ParseRGBA parses a "#RRGGBB" or "#RRGGBBAA" hex string.
// An empty string returns the zero value.
func ParseRGBA(s string) (RGBA, error) {
	if s == "" {
		return RGBA{}, nil
	}
	if s[0] != '#' {
		return RGBA{}, fmt.Errorf("<archmage> invalid RGBA string %q", s)
	}

	var r, g, b, a uint8
	var ok bool
	switch len(s) {
	case 7: // #RRGGBB
		r, ok = unhexByte(s[1], s[2])
		if ok {
			g, ok = unhexByte(s[3], s[4])
			if ok {
				b, ok = unhexByte(s[5], s[6])
			}
		}
		a = 0xFF
	case 9: // #RRGGBBAA
		r, ok = unhexByte(s[1], s[2])
		if ok {
			g, ok = unhexByte(s[3], s[4])
			if ok {
				b, ok = unhexByte(s[5], s[6])
				if ok {
					a, ok = unhexByte(s[7], s[8])
				}
			}
		}
	}
	if !ok {
		return RGBA{}, fmt.Errorf("<archmage> invalid RGBA string %q", s)
	}

	return RGBA{R: r, G: g, B: b, A: a}, nil
}

func unhexNibble(b byte) (byte, bool) {
	switch {
	case b >= '0' && b <= '9':
		return b - '0', true
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10, true
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10, true
	}
	return 0, false
}

func unhexByte(hi, lo byte) (uint8, bool) {
	h, ok1 := unhexNibble(hi)
	l, ok2 := unhexNibble(lo)
	return h<<4 | l, ok1 && ok2
}

// String returns the color as "#RRGGBBAA".
func (c *RGBA) String() string {
	buf := [9]byte{
		'#',
		_hexUpper[c.R>>4], _hexUpper[c.R&0xF],
		_hexUpper[c.G>>4], _hexUpper[c.G&0xF],
		_hexUpper[c.B>>4], _hexUpper[c.B&0xF],
		_hexUpper[c.A>>4], _hexUpper[c.A&0xF],
	}
	return string(buf[:])
}

var (
	_quotedEmptyString = []byte(`""`)
)

// MarshalJSONTo encodes RGBA as a JSON string in "#RRGGBBAA" format or an empty string if zero.
func (c *RGBA) MarshalJSONTo(enc *jsontext.Encoder) error {
	if c.R == 0 && c.G == 0 && c.B == 0 && c.A == 0 {
		return enc.WriteValue(_quotedEmptyString)
	}

	buf := [11]byte{
		'"', '#',
		_hexUpper[c.R>>4], _hexUpper[c.R&0xF],
		_hexUpper[c.G>>4], _hexUpper[c.G&0xF],
		_hexUpper[c.B>>4], _hexUpper[c.B&0xF],
		_hexUpper[c.A>>4], _hexUpper[c.A&0xF],
		'"',
	}
	return enc.WriteValue(buf[:])
}

// UnmarshalJSONFrom decodes a JSON string or null into RGBA.
func (c *RGBA) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}

	switch tok.Kind() {
	case 'n':
		*c = RGBA{}
		return nil
	case '"':
		v, err := ParseRGBA(tok.String())
		if err != nil {
			return err
		}
		*c = v
		return nil
	default:
		return fmt.Errorf("<archmage> RGBA: invalid JSON token kind %q, expected string or null", tok.Kind())
	}
}
