package archmage

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
)

var (
	_ json.MarshalerTo     = (*Vec2[int])(nil)
	_ json.UnmarshalerFrom = (*Vec2[int])(nil)
	_ json.MarshalerTo     = (*Vec3[int])(nil)
	_ json.UnmarshalerFrom = (*Vec3[int])(nil)
	_ json.MarshalerTo     = (*Vec4[int])(nil)
	_ json.UnmarshalerFrom = (*Vec4[int])(nil)
)

// Vec2 represents a 2D vector with comparable values.
// It marshals to JSON as a two-element array.
type Vec2[T comparable] struct {
	X, Y T
}

// MakeVec2 creates a Vec2 from x and y.
func MakeVec2[T comparable](x, y T) Vec2[T] {
	return Vec2[T]{X: x, Y: y}
}

// MarshalJSONTo encodes the vector as a JSON array [x, y].
func (v *Vec2[T]) MarshalJSONTo(enc *jsontext.Encoder) error {
	err := enc.WriteToken(jsontext.BeginArray)
	if err != nil {
		return err
	}

	if err = json.MarshalEncode(enc, v.X); err != nil {
		return err
	}
	if err = json.MarshalEncode(enc, v.Y); err != nil {
		return err
	}

	return enc.WriteToken(jsontext.EndArray)
}

// UnmarshalJSONFrom decodes a JSON array [x, y] or null into the vector.
func (v *Vec2[T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}

	switch tok.Kind() {
	case 'n':
		var zero T
		v.X = zero
		v.Y = zero
		return nil
	case '[':
	default:
		return fmt.Errorf("archmage.Vec2: invalid JSON token kind %q, expected 'n' or '['", tok.Kind())
	}

	if err = json.UnmarshalDecode(dec, &v.X); err != nil {
		return err
	}
	if err = json.UnmarshalDecode(dec, &v.Y); err != nil {
		return err
	}

	if tok, err = dec.ReadToken(); err != nil {
		return err
	} else if tok.Kind() != ']' {
		return fmt.Errorf("archmage.Vec2: invalid JSON token kind %q, expected ']'", tok.Kind())
	}

	return nil
}

// Vec3 represents a 3D vector with comparable values.
// It marshals to JSON as a three-element array.
type Vec3[T comparable] struct {
	X, Y, Z T
}

// MakeVec3 creates a Vec3 from x, y, and z.
func MakeVec3[T comparable](x, y, z T) Vec3[T] {
	return Vec3[T]{X: x, Y: y, Z: z}
}

// MarshalJSONTo encodes the vector as a JSON array [x, y, z].
func (v *Vec3[T]) MarshalJSONTo(enc *jsontext.Encoder) error {
	err := enc.WriteToken(jsontext.BeginArray)
	if err != nil {
		return err
	}

	if err = json.MarshalEncode(enc, v.X); err != nil {
		return err
	}
	if err = json.MarshalEncode(enc, v.Y); err != nil {
		return err
	}
	if err = json.MarshalEncode(enc, v.Z); err != nil {
		return err
	}

	return enc.WriteToken(jsontext.EndArray)
}

// UnmarshalJSONFrom decodes a JSON array [x, y, z] or null into the vector.
func (v *Vec3[T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}

	switch tok.Kind() {
	case 'n':
		var zero T
		v.X = zero
		v.Y = zero
		v.Z = zero
		return nil
	case '[':
	default:
		return fmt.Errorf("archmage.Vec3: invalid JSON token kind %q, expected 'n' or '['", tok.Kind())
	}

	if err = json.UnmarshalDecode(dec, &v.X); err != nil {
		return err
	}
	if err = json.UnmarshalDecode(dec, &v.Y); err != nil {
		return err
	}
	if err = json.UnmarshalDecode(dec, &v.Z); err != nil {
		return err
	}

	if tok, err = dec.ReadToken(); err != nil {
		return err
	} else if tok.Kind() != ']' {
		return fmt.Errorf("archmage.Vec3: invalid JSON token kind %q, expected ']'", tok.Kind())
	}

	return nil
}

// Vec4 represents a 4D vector with comparable values.
// It marshals to JSON as a four-element array.
type Vec4[T comparable] struct {
	X, Y, Z, W T
}

// MakeVec4 creates a Vec4 from x, y, z, and w.
func MakeVec4[T comparable](x, y, z, w T) Vec4[T] {
	return Vec4[T]{X: x, Y: y, Z: z, W: w}
}

// MarshalJSONTo encodes the vector as a JSON array [x, y, z, w].
func (v *Vec4[T]) MarshalJSONTo(enc *jsontext.Encoder) error {
	err := enc.WriteToken(jsontext.BeginArray)
	if err != nil {
		return err
	}

	if err = json.MarshalEncode(enc, v.X); err != nil {
		return err
	}
	if err = json.MarshalEncode(enc, v.Y); err != nil {
		return err
	}
	if err = json.MarshalEncode(enc, v.Z); err != nil {
		return err
	}
	if err = json.MarshalEncode(enc, v.W); err != nil {
		return err
	}

	return enc.WriteToken(jsontext.EndArray)
}

// UnmarshalJSONFrom decodes a JSON array [x, y, z, w] or null into the vector.
func (v *Vec4[T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}

	switch tok.Kind() {
	case 'n':
		var zero T
		v.X = zero
		v.Y = zero
		v.Z = zero
		v.W = zero
		return nil
	case '[':
	default:
		return fmt.Errorf("archmage.Vec4: invalid JSON token kind %q, expected 'n' or '['", tok.Kind())
	}

	if err = json.UnmarshalDecode(dec, &v.X); err != nil {
		return err
	}
	if err = json.UnmarshalDecode(dec, &v.Y); err != nil {
		return err
	}
	if err = json.UnmarshalDecode(dec, &v.Z); err != nil {
		return err
	}
	if err = json.UnmarshalDecode(dec, &v.W); err != nil {
		return err
	}

	if tok, err = dec.ReadToken(); err != nil {
		return err
	} else if tok.Kind() != ']' {
		return fmt.Errorf("archmage.Vec4: invalid JSON token kind %q, expected ']'", tok.Kind())
	}

	return nil
}
