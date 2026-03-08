package archmage

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
)

// Ref pairs a table entry ID with a typed pointer to the referenced table entry.
// Only RawValue is marshaled to JSON; Ref is populated by Atlas.BindRefs
// after all atlas items are loaded.
type Ref[V int | int32 | int64 | string, T any] struct {
	// RawValue is the ID of the referenced table entry.
	RawValue V
	// Ref is the resolved pointer, populated by Atlas.BindRefs.
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
