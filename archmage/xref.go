package archmage

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
)

// XRef pairs a table entry ID with a typed pointer to the referenced table entry.
// Only CfgID is marshaled to JSON; Ref is populated by Atlas.BindRefs
// after all atlas items are loaded.
type XRef[V comparable, T any] struct {
	// CfgID is the ID of the referenced table entry.
	CfgID V
	// Ref is the resolved pointer, populated by Atlas.BindRefs.
	Ref *T `json:"-"`
}

// MarshalJSONTo encodes CfgID to the JSON encoder.
func (r *XRef[V, T]) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, r.CfgID)
}

// UnmarshalJSONFrom decodes CfgID from the JSON decoder and resets Ref.
func (r *XRef[V, T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	err := json.UnmarshalDecode(dec, &r.CfgID)
	if err == nil {
		r.Ref = nil
	}
	return err
}
