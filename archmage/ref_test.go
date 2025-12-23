package archmage_test

import (
	"encoding/json/v2"
	"strconv"
	"testing"

	"shadop.dev/pkg/sdk-go/archmage"
)

func TestRef(t *testing.T) {
	dataset := []int{0, -1, 1, 42, 1000000}
	for _, v := range dataset {
		ref1 := archmage.Ref[int, string]{RawValue: v, Ref: "foo"}
		data, err := json.Marshal(ref1)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		if _, err = strconv.ParseInt(string(data), 0, 64); err != nil {
			t.Fatalf("expected marshaled data to be integer, got %s", string(data))
		}

		ref2 := archmage.Ref[int, string]{RawValue: -999, Ref: "bar"}
		err = json.Unmarshal(data, &ref2)
		if err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}
		if ref2.RawValue != v {
			t.Fatalf("expected RawValue %v, got %v", v, ref2.RawValue)
		}
		if ref2.Ref != "" {
			t.Fatalf("expected Ref to empty string, got %v", ref2.Ref)
		}
	}
}
