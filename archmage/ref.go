package archmage

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
)

type Ref[V int | string, T any] struct {
	RawValue V
	Ref      T `json:"-"`
}

func (r *Ref[V, T]) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, r.RawValue)
}

func (r *Ref[V, T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var zero T
	r.Ref = zero
	return json.UnmarshalDecode(dec, &r.RawValue)
}
