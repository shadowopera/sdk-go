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

type Vec2[T any] struct {
	X, Y T
}

func MakeVec2[T any](x, y T) Vec2[T] {
	return Vec2[T]{X: x, Y: y}
}

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

type Vec3[T any] struct {
	X, Y, Z T
}

func MakeVec3[T any](x, y, z T) Vec3[T] {
	return Vec3[T]{X: x, Y: y, Z: z}
}

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

type Vec4[T any] struct {
	X, Y, Z, W T
}

func MakeVec4[T any](x, y, z, w T) Vec4[T] {
	return Vec4[T]{X: x, Y: y, Z: z, W: w}
}

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
