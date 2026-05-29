package archmage

import (
	"fmt"
	"math/rand/v2"
	"unsafe"
)

type MinMax[T comparable] struct {
	Min T `json:"min"`
	Max T `json:"max"`
}

func (mm *MinMax[T]) Sample(rnd *rand.Rand) T {
	switch any((*T)(nil)).(type) {
	case *int:
		xMin := *(*int)(unsafe.Pointer(&mm.Min))
		xMax := *(*int)(unsafe.Pointer(&mm.Max))
		v := xMin + rnd.IntN(xMax-xMin+1)
		return *(*T)(unsafe.Pointer(&v))

	case *int8:
		xMin := *(*int8)(unsafe.Pointer(&mm.Min))
		xMax := *(*int8)(unsafe.Pointer(&mm.Max))
		v := int8(int64(xMin) + rnd.Int64N(int64(xMax)-int64(xMin)+1))
		return *(*T)(unsafe.Pointer(&v))

	case *int16:
		xMin := *(*int16)(unsafe.Pointer(&mm.Min))
		xMax := *(*int16)(unsafe.Pointer(&mm.Max))
		v := int16(int64(xMin) + rnd.Int64N(int64(xMax)-int64(xMin)+1))
		return *(*T)(unsafe.Pointer(&v))

	case *int32:
		xMin := *(*int32)(unsafe.Pointer(&mm.Min))
		xMax := *(*int32)(unsafe.Pointer(&mm.Max))
		v := int32(int64(xMin) + rnd.Int64N(int64(xMax)-int64(xMin)+1))
		return *(*T)(unsafe.Pointer(&v))

	case *int64:
		xMin := *(*int64)(unsafe.Pointer(&mm.Min))
		xMax := *(*int64)(unsafe.Pointer(&mm.Max))
		v := xMin + rnd.Int64N(xMax-xMin+1)
		return *(*T)(unsafe.Pointer(&v))

	case *uint:
		xMin := *(*uint)(unsafe.Pointer(&mm.Min))
		xMax := *(*uint)(unsafe.Pointer(&mm.Max))
		v := xMin + rnd.UintN(xMax-xMin+1)
		return *(*T)(unsafe.Pointer(&v))

	case *uint8:
		xMin := *(*uint8)(unsafe.Pointer(&mm.Min))
		xMax := *(*uint8)(unsafe.Pointer(&mm.Max))
		v := uint8(uint(xMin) + rnd.UintN(uint(xMax)-uint(xMin)+1))
		return *(*T)(unsafe.Pointer(&v))

	case *uint16:
		xMin := *(*uint16)(unsafe.Pointer(&mm.Min))
		xMax := *(*uint16)(unsafe.Pointer(&mm.Max))
		v := uint16(uint(xMin) + rnd.UintN(uint(xMax)-uint(xMin)+1))
		return *(*T)(unsafe.Pointer(&v))

	case *uint32:
		xMin := *(*uint32)(unsafe.Pointer(&mm.Min))
		xMax := *(*uint32)(unsafe.Pointer(&mm.Max))
		v := uint32(uint(xMin) + rnd.UintN(uint(xMax)-uint(xMin)+1))
		return *(*T)(unsafe.Pointer(&v))

	case *uint64:
		xMin := *(*uint64)(unsafe.Pointer(&mm.Min))
		xMax := *(*uint64)(unsafe.Pointer(&mm.Max))
		v := xMin + rnd.Uint64N(xMax-xMin+1)
		return *(*T)(unsafe.Pointer(&v))

	case *float32:
		xMin := *(*float32)(unsafe.Pointer(&mm.Min))
		xMax := *(*float32)(unsafe.Pointer(&mm.Max))
		v := xMin + rnd.Float32()*(xMax-xMin)
		return *(*T)(unsafe.Pointer(&v))

	case *float64:
		xMin := *(*float64)(unsafe.Pointer(&mm.Min))
		xMax := *(*float64)(unsafe.Pointer(&mm.Max))
		v := xMin + rnd.Float64()*(xMax-xMin)
		return *(*T)(unsafe.Pointer(&v))

	default:
		panic("unsupported type: " + fmt.Sprintf("%T", mm.Min))
	}
}
