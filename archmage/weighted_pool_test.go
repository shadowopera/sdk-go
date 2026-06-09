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

func TestWeightedPoolInt32SumOverflow(t *testing.T) {
	// Sum of weights exceeds the int32 range; the int64 accumulator must not overflow.
	rng := NewPCG(0, 1)
	const big = math.MaxInt32
	wp := archmage.WeightedPool[int]{
		Items:   []int{0, 1, 2},
		Weights: []int32{big, big, big},
	}
	seen := make(map[int]bool)
	for range 100000 {
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
