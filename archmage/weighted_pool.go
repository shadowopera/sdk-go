package archmage

import (
	"math/rand/v2"
)

// WeightedPool holds items alongside their selection weights in two parallel slices
// of equal length. Sample and SampleIndex draw an item at random with probability
// proportional to its weight. SampleWith and SampleIndexWith map a
// caller-supplied value to an item deterministically according to the weights.
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
	idx := wp.SampleIndex(rng)
	return wp.Items[idx]
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
	if total > 1_000_000_000 {
		panic("<archmage> WeightedPool.SampleIndex: total weight exceeds 1,000,000,000")
	}

	value := rng.Int32N(int32(total))
	var acc int32
	for i, w := range wp.Weights {
		acc += w
		if acc > value {
			return i
		}
	}

	panic("unreachable")
}

// SampleWith maps the value to an item deterministically according to the weights.
// value < 0 returns the first item with non-zero weight; value >= 1 returns the last.
// It panics if the pool is empty or the total weight is zero.
func (wp *WeightedPool[T]) SampleWith(value float32) T {
	idx := wp.SampleIndexWith(value)
	return wp.Items[idx]
}

// SampleIndexWith maps the value to an item index deterministically according to the weights.
// value < 0 returns the first item index with non-zero weight; value >= 1 returns the last.
// It panics if the pool is empty or the total weight is zero.
func (wp *WeightedPool[T]) SampleIndexWith(value float32) int {
	if len(wp.Items) == 0 {
		panic("<archmage> WeightedPool.SampleIndexWith: empty pool")
	}

	var total int64
	for _, w := range wp.Weights {
		total += int64(w)
	}
	if total == 0 {
		panic("<archmage> WeightedPool.SampleIndexWith: total weight is zero")
	}
	if total > 1_000_000_000 {
		panic("<archmage> WeightedPool.SampleIndexWith: total weight exceeds 1,000,000,000")
	}

	pos1 := int64(float64(value) * float64(total))
	pos2 := int32(max(0, min(total-1, pos1)))

	var acc int32
	for i, w := range wp.Weights {
		acc += w
		if acc > pos2 {
			return i
		}
	}

	panic("<archmage> WeightedPool.SampleIndexWith: unreachable")
}
