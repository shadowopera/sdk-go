package archmage_test

import (
	"testing"
	"time"

	"shadop.dev/pkg/sdk-go/archmage"
)

func newDurationRange(lo, hi time.Duration) archmage.MinMax[archmage.Duration] {
	return archmage.MinMax[archmage.Duration]{
		Min: archmage.Duration{Duration: lo},
		Max: archmage.Duration{Duration: hi},
	}
}

func TestMinMax_SampleDuration(t *testing.T) {
	const ms = time.Millisecond

	type Trial struct {
		subject string
		minMax  archmage.MinMax[archmage.Duration]
	}

	dataset := []Trial{
		{
			subject: "zero value",
			minMax:  archmage.MinMax[archmage.Duration]{},
		},
		{
			subject: "single point",
			minMax:  newDurationRange(5*ms, 5*ms),
		},
		{
			subject: "smallest span",
			minMax:  newDurationRange(1*ms, 2*ms),
		},
		{
			subject: "negative bounds",
			minMax:  newDurationRange(-10*ms, 10*ms),
		},
		{
			subject: "coarser units",
			minMax:  newDurationRange(3*time.Second, 2*time.Minute),
		},
		{
			subject: "largest span",
			minMax:  newDurationRange(0, 1e9*ms),
		},
	}

	rng := NewPCG(20260827)
	for _, tt := range dataset {
		t.Run(tt.subject, func(t *testing.T) {
			for range 10000 {
				v := tt.minMax.Sample(rng)
				if v.Duration < tt.minMax.Min.Duration || v.Duration > tt.minMax.Max.Duration {
					t.Fatalf("expected sample within [%v, %v], got %v", tt.minMax.Min, tt.minMax.Max, v)
				}
				if v.Duration%ms != 0 {
					t.Fatalf("expected sample to be a whole number of milliseconds, got %v", v)
				}
			}
		})
	}

	t.Run("both ends drawable", func(t *testing.T) {
		minMax := newDurationRange(1*ms, 3*ms)
		seen := make(map[time.Duration]bool)
		for range 10000 {
			seen[minMax.Sample(rng).Duration] = true
		}
		for _, expected := range []time.Duration{1 * ms, 2 * ms, 3 * ms} {
			if !seen[expected] {
				t.Fatalf("expected %v to be drawn at least once", expected)
			}
		}
		if len(seen) != 3 {
			t.Fatalf("expected 3 distinct samples, got %d", len(seen))
		}
	})
}

func BenchmarkMinMax(b *testing.B) {
	rng := NewPCG()

	b.Run("NativeInt64", func(b *testing.B) {
		b.ReportAllocs()
		var result int64
		minMax := archmage.MinMax[int64]{
			Min: 1,
			Max: 10000,
		}
		for b.Loop() {
			result = minMax.Min + rng.Int64N(minMax.Max-minMax.Min+1)
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
			result = minMax.Sample(rng)
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
			result = minMax.Min + rng.Float32()*(minMax.Max-minMax.Min)
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
			result = minMax.Sample(rng)
		}
		_ = result
	})

	b.Run("SampleDuration", func(b *testing.B) {
		b.ReportAllocs()
		var result archmage.Duration
		minMax := newDurationRange(time.Millisecond, 10000*time.Millisecond)
		for b.Loop() {
			result = minMax.Sample(rng)
		}
		_ = result
	})
}
