package archmage_test

import (
	"encoding/json/v2"
	"math/rand/v2"
	"testing"
	"time"

	"shadop.dev/pkg/sdk-go/archmage"
)

func NewPCG(seeds ...uint64) *rand.Rand {
	switch len(seeds) {
	case 0:
		return rand.New(rand.NewPCG(rand.Uint64(), uint64(time.Now().UnixMilli())))
	case 1:
		return rand.New(rand.NewPCG(seeds[0], 0))
	case 2:
		return rand.New(rand.NewPCG(seeds[0], seeds[1]))
	default:
		panic("NewPCG: too many seeds")
	}
}

func TestVec2(t *testing.T) {
	rnd := NewPCG(0, 1)
	genNum := func() int {
		w := rnd.IntN(100)
		if w < 35 {
			return 0
		} else {
			return rnd.Int()
		}
	}

	for range 1000 {
		v1 := archmage.MakeVec2(genNum(), genNum())
		data, err := json.Marshal(v1)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		v2 := archmage.MakeVec2(genNum(), genNum())
		err = json.Unmarshal(data, &v2)
		if err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}
		if v2 != v1 {
			t.Fatalf("expected %v, got %v", v1, v2)
		}
	}

	vec := archmage.MakeVec2(1, 2)
	if err := json.Unmarshal([]byte("null"), &vec); err != nil {
		t.Fatalf("json.Unmarshal null failed: %v", err)
	} else if vec != archmage.MakeVec2(0, 0) {
		t.Fatalf("expected zero vec, got %v", vec)
	}
}

func TestVec3(t *testing.T) {
	rnd := NewPCG(0, 1)
	genNum := func() float64 {
		w := rnd.IntN(100)
		if w < 35 {
			return 0
		} else {
			return rnd.Float64()
		}
	}

	for range 1000 {
		v1 := archmage.MakeVec3(genNum(), genNum(), genNum())
		data, err := json.Marshal(v1)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		v2 := archmage.MakeVec3(genNum(), genNum(), genNum())
		err = json.Unmarshal(data, &v2)
		if err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}
		if v2 != v1 {
			t.Fatalf("expected %v, got %v", v1, v2)
		}
	}

	vec := archmage.MakeVec3(1, 2, 3)
	if err := json.Unmarshal([]byte("null"), &vec); err != nil {
		t.Fatalf("json.Unmarshal null failed: %v", err)
	} else if vec != archmage.MakeVec3(0, 0, 0) {
		t.Fatalf("expected zero vec, got %v", vec)
	}
}

func TestVec4(t *testing.T) {
	rnd := NewPCG(0, 1)
	genNum := func() uint32 {
		w := rnd.IntN(100)
		if w < 35 {
			return 0
		} else {
			return rnd.Uint32()
		}
	}

	for range 1000 {
		v1 := archmage.MakeVec4(genNum(), genNum(), genNum(), genNum())
		data, err := json.Marshal(v1)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		v2 := archmage.MakeVec4(genNum(), genNum(), genNum(), genNum())
		err = json.Unmarshal(data, &v2)
		if err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}
		if v2 != v1 {
			t.Fatalf("expected %v, got %v", v1, v2)
		}
	}

	vec := archmage.MakeVec4(1, 2, 3, 4)
	if err := json.Unmarshal([]byte("null"), &vec); err != nil {
		t.Fatalf("json.Unmarshal null failed: %v", err)
	} else if vec != archmage.MakeVec4(0, 0, 0, 0) {
		t.Fatalf("expected zero vec, got %v", vec)
	}
}
