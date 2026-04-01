package archmage_test

import (
	"encoding/json/v2"
	"strconv"
	"testing"

	"shadop.dev/pkg/sdk-go/archmage"
)

func TestXRef(t *testing.T) {
	str := "foo"
	dataset := []int{0, -1, 1, 42, 1000000}
	for _, v := range dataset {
		ref1 := archmage.XRef[int, string]{CfgID: v, Ref: &str}
		data, err := json.Marshal(ref1)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		if _, err = strconv.ParseInt(string(data), 0, 64); err != nil {
			t.Fatalf("expected marshaled data to be integer, got %s", string(data))
		}

		ref2 := archmage.XRef[int, string]{CfgID: -999, Ref: &str}
		err = json.Unmarshal(data, &ref2)
		if err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}
		if ref2.CfgID != v {
			t.Fatalf("expected CfgID %v, got %v", v, ref2.CfgID)
		}
		if ref2.Ref != nil {
			t.Fatalf("expected Ref to empty string, got %v", ref2.Ref)
		}
	}
}
