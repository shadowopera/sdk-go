package archmage

import (
	"math/rand/v2"
)

// WeightedPool holds items alongside their selection weights in two parallel slices
// of equal length. Sample and SampleIndex draw an item at random with probability
// proportional to its weight.
type WeightedPool[T any] struct {
	// Items are the candidate values, one per weight.
	Items []T `json:"items"`
	// Weights are the non-negative selection weights, parallel to Items.
	Weights []int32 `json:"weights"`
}

// Len returns the number of items in the pool.
func (wp *WeightedPool[T]) Len() int {
	return len(wp.Items)
}

// Sample returns a randomly selected item, weighted by Weights.
// It panics if the pool is empty or the total weight is zero.
func (wp *WeightedPool[T]) Sample(rng *rand.Rand) T {
	return wp.Items[wp.SampleIndex(rng)]
}

// SampleIndex returns the index of a randomly selected item, weighted by Weights.
// It panics if the pool is empty or the total weight is zero.
func (wp *WeightedPool[T]) SampleIndex(rng *rand.Rand) int {
	if len(wp.Items) == 0 {
		panic("<archmage> WeightedPool.SampleIndex: empty pool")
	}

	var total int64
	for _, w := range wp.Weights {
		total += int64(w)
	}
	if total == 0 {
		panic("<archmage> WeightedPool.SampleIndex: total weight is zero")
	}

	r := rng.Int64N(total)
	var acc int64
	for i, w := range wp.Weights {
		acc += int64(w)
		if acc > r {
			return i
		}
	}

	panic("unreachable")
}
