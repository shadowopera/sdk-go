package archmage

import (
	"math/rand/v2"
)

// WeightedPool holds items alongside their selection weights in two parallel slices
// of equal length. Sample and SampleIndex draw an item at random with probability
// proportional to its weight. SampleWithNoise and SampleIndexWithNoise map a
// caller-supplied noise value to an item deterministically according to the weights.
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

// SampleWithNoise maps the noise value to an item deterministically according to the weights.
// noise < 0 returns the first item with non-zero weight; noise >= 1 returns the last.
// It panics if the pool is empty or the total weight is zero.
func (wp *WeightedPool[T]) SampleWithNoise(noise float32) T {
	idx := wp.SampleIndexWithNoise(noise)
	return wp.Items[idx]
}

// SampleIndexWithNoise maps the noise value to an item index deterministically according to the weights.
// noise < 0 returns the first item index with non-zero weight; noise >= 1 returns the last.
// It panics if the pool is empty or the total weight is zero.
func (wp *WeightedPool[T]) SampleIndexWithNoise(noise float32) int {
	if len(wp.Items) == 0 {
		panic("<archmage> WeightedPool.SampleIndexWithNoise: empty pool")
	}

	var total int64
	for _, w := range wp.Weights {
		total += int64(w)
	}
	if total == 0 {
		panic("<archmage> WeightedPool.SampleIndexWithNoise: total weight is zero")
	}
	if total > 1_000_000_000 {
		panic("<archmage> WeightedPool.SampleIndexWithNoise: total weight exceeds 1,000,000,000")
	}

	value := int64(float64(noise) * float64(total))
	if value >= total {
		value = total - 1
	} else if value < 0 {
		value = 0
	}

	var acc int32
	for i, w := range wp.Weights {
		acc += w
		if acc > int32(value) {
			return i
		}
	}

	panic("<archmage> WeightedPool.SampleIndexWithNoise: unreachable")
}
