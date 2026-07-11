package archmage_test

import (
	"math"
	"testing"

	"shadop.dev/pkg/sdk-go/archmage"
)

func TestWeightedPoolEmptyPanics(t *testing.T) {
	rng := NewPCG(0, 1)
	wp := archmage.WeightedPool[int]{}
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on empty pool")
		}
	}()
	wp.Sample(rng)
}

func TestWeightedPoolZeroTotalPanics(t *testing.T) {
	rng := NewPCG(0, 1)
	wp := archmage.WeightedPool[int]{
		Items:   []int{1, 2, 3},
		Weights: []int32{0, 0, 0},
	}
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when total weight is zero")
		}
	}()
	wp.SampleIndex(rng)
}

func TestWeightedPoolSingleElement(t *testing.T) {
	rng := NewPCG(0, 1)
	wp := archmage.WeightedPool[string]{
		Items:   []string{"only"},
		Weights: []int32{7},
	}
	for range 100 {
		if got := wp.Sample(rng); got != "only" {
			t.Fatalf("expected %q, got %q", "only", got)
		}
		if idx := wp.SampleIndex(rng); idx != 0 {
			t.Fatalf("expected index 0, got %d", idx)
		}
	}
}

func TestWeightedPoolZeroWeightNeverSelected(t *testing.T) {
	rng := NewPCG(0, 1)
	wp := archmage.WeightedPool[int]{
		Items:   []int{10, 20, 30},
		Weights: []int32{5, 0, 5},
	}
	for range 10000 {
		if idx := wp.SampleIndex(rng); idx == 1 {
			t.Fatalf("zero-weight item at index 1 was selected")
		}
	}
}

func TestWeightedPoolDistribution(t *testing.T) {
	rng := NewPCG(0, 1)
	wp := archmage.WeightedPool[int]{
		Items:   []int{0, 1, 2, 3},
		Weights: []int32{1, 2, 3, 4},
	}
	var total int64
	for _, w := range wp.Weights {
		total += int64(w)
	}

	const n = 1_000_000
	counts := make([]int, len(wp.Items))
	for range n {
		counts[wp.SampleIndex(rng)]++
	}

	for i, w := range wp.Weights {
		want := float64(w) / float64(total)
		got := float64(counts[i]) / float64(n)
		if math.Abs(got-want) > 0.005 {
			t.Fatalf("index %d: want ~%.4f, got %.4f", i, want, got)
		}
	}
}

func TestWeightedPoolLargeWeightsDistribution(t *testing.T) {
	// Large equal weights (summing to exactly 1,000,000,000) must stay within the limit
	// and select all indices with roughly equal probability.
	rng := NewPCG(0, 1)
	wp := archmage.WeightedPool[int]{
		Items:   []int{0, 1, 2},
		Weights: []int32{333_333_333, 333_333_333, 333_333_334},
	}
	seen := make(map[int]bool)
	for range 10000 {
		idx := wp.SampleIndex(rng)
		if idx < 0 || idx >= len(wp.Items) {
			t.Fatalf("index out of range: %d", idx)
		}
		seen[idx] = true
	}
	for i := range wp.Items {
		if !seen[i] {
			t.Fatalf("index %d never selected despite equal weights", i)
		}
	}
}

func TestWeightedPoolTotalOverLimitPanics(t *testing.T) {
	rng := NewPCG(0, 1)
	wp := archmage.WeightedPool[int]{
		Items:   []int{0, 1},
		Weights: []int32{500_000_001, 500_000_000},
	}
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when total weight exceeds 1,000,000,000")
		}
	}()
	wp.SampleIndex(rng)
}

func TestWeightedPoolWithEmptyPanics(t *testing.T) {
	wp := archmage.WeightedPool[int]{}
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on empty pool")
		}
	}()
	wp.SampleWith(0.5)
}

func TestWeightedPoolWithZeroTotalPanics(t *testing.T) {
	wp := archmage.WeightedPool[int]{
		Items:   []int{1, 2, 3},
		Weights: []int32{0, 0, 0},
	}
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when total weight is zero")
		}
	}()
	wp.SampleIndexWith(0.5)
}

func TestWeightedPoolWithTotalOverLimitPanics(t *testing.T) {
	wp := archmage.WeightedPool[int]{
		Items:   []int{0, 1},
		Weights: []int32{500_000_001, 500_000_000},
	}
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when total weight exceeds 1,000,000,000")
		}
	}()
	wp.SampleIndexWith(0.5)
}

func TestWeightedPoolWithClampHigh(t *testing.T) {
	// value >= 1 must return the last item with non-zero weight.
	wp := archmage.WeightedPool[string]{
		Items:   []string{"a", "b", "c"},
		Weights: []int32{1, 2, 3},
	}
	for _, value := range []float32{1.0, 1.5, 100.0} {
		if got := wp.SampleWith(value); got != "c" {
			t.Fatalf("value=%.1f: expected \"c\", got %q", value, got)
		}
	}
}

func TestWeightedPoolWithClampLow(t *testing.T) {
	// value < 0 must return the first item with non-zero weight.
	wp := archmage.WeightedPool[string]{
		Items:   []string{"a", "b", "c"},
		Weights: []int32{1, 2, 3},
	}
	for _, value := range []float32{0.0, -0.1, -100.0} {
		if got := wp.SampleWith(value); got != "a" {
			t.Fatalf("value=%.1f: expected \"a\", got %q", value, got)
		}
	}
}

func TestWeightedPoolWithProportionalSlots(t *testing.T) {
	// Weights 1:2:3 split [0,6) into slots [0,1), [1,3), [3,6).
	// Verify exact boundary probes map to the expected index.
	wp := archmage.WeightedPool[int]{
		Items:   []int{0, 1, 2},
		Weights: []int32{1, 2, 3},
	}
	// total == 6
	cases := []struct {
		value float32
		want  int
	}{
		{0.0, 0},             // pos=0      -> slot 0
		{0.1666, 0},          // pos≈0.9996 -> still slot 0
		{1.0 / 6.0 * 1.5, 1}, // pos≈1.5    -> slot 1
		{4.0 / 6.0, 2},       // pos≈4      -> slot 2
		{5.0/6.0 + 0.001, 2}, // pos≈5      -> slot 2
	}
	for _, c := range cases {
		if got := wp.SampleIndexWith(c.value); got != c.want {
			t.Fatalf("value=%.4f: expected index %d, got %d", c.value, c.want, got)
		}
	}
}

func TestWeightedPoolWithZeroWeightNeverSelected(t *testing.T) {
	// Zero-weight item must never be reached regardless of value.
	wp := archmage.WeightedPool[int]{
		Items:   []int{10, 20, 30},
		Weights: []int32{5, 0, 5},
	}
	const steps = 1001
	for i := range steps {
		value := float32(i) / float32(steps-1) // 0.0 … 1.0
		if idx := wp.SampleIndexWith(value); idx == 1 {
			t.Fatalf("zero-weight item at index 1 was selected for value=%.4f", value)
		}
	}
}
