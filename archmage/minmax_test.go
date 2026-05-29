package archmage_test

import (
	"testing"

	"shadop.dev/pkg/sdk-go/archmage"
)

func BenchmarkMinMax(b *testing.B) {
	rnd := NewPCG()

	b.Run("NativeInt64", func(b *testing.B) {
		b.ReportAllocs()
		var result int64
		minMax := archmage.MinMax[int64]{
			Min: 1,
			Max: 10000,
		}
		for b.Loop() {
			result = minMax.Min + rnd.Int64N(minMax.Max-minMax.Min+1)
		}
		_ = result
	})

	b.Run("SampleInt", func(b *testing.B) {
		b.ReportAllocs()
		var result int64
		minMax := archmage.MinMax[int64]{
			Min: 1,
			Max: 10000,
		}
		for b.Loop() {
			result = minMax.Sample(rnd)
		}
		_ = result
	})

	b.Run("NativeFloat32", func(b *testing.B) {
		b.ReportAllocs()
		var result float32
		minMax := archmage.MinMax[float32]{
			Min: 1,
			Max: 10000,
		}
		for b.Loop() {
			result = minMax.Min + rnd.Float32()*(minMax.Max-minMax.Min)
		}
		_ = result
	})

	b.Run("SampleFloat32", func(b *testing.B) {
		b.ReportAllocs()
		var result float32
		minMax := archmage.MinMax[float32]{
			Min: 1,
			Max: 10000,
		}
		for b.Loop() {
			result = minMax.Sample(rnd)
		}
		_ = result
	})
}
