package archmage

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
)

// Ref stores a raw value and a resolved reference.
// It marshals only RawValue to JSON.
type Ref[V int | int32 | int64 | string, T any] struct {
	// RawValue is the original unresolved value.
	RawValue V
	// Ref is the resolved reference, not included in JSON.
	Ref *T `json:"-"`
}

// MarshalJSONTo encodes RawValue to the JSON encoder.
func (r *Ref[V, T]) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, r.RawValue)
}

// UnmarshalJSONFrom decodes RawValue from the JSON decoder and resets Ref.
func (r *Ref[V, T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	err := json.UnmarshalDecode(dec, &r.RawValue)
	if err == nil {
		r.Ref = nil
	}
	return err
}
