package archmage

import (
	"errors"
	"time"
)

var (
	_errInvalidDurationShardsType   = errors.New("invalid duration shards type")
	_errInvalidDurationShardsLength = errors.New("invalid duration shards length")
	_errInvalidDurationShardsFormat = errors.New("invalid duration shards format")
)

func ParseDurationShards(v any) (time.Duration, error) {
	if v == nil {
		return 0, nil
	}

	switch x := v.(type) {
	case *[2]int64:
		return parseDurationShardsImpl(x[:])
	case *[3]int64:
		return parseDurationShardsImpl(x[:])
	case []int64:
		return parseDurationShardsImpl(x)
	default:
		return 0, _errInvalidDurationShardsType
	}
}

func parseDurationShardsImpl(a []int64) (time.Duration, error) {
	switch len(a) {
	case 0:
		return 0, nil

	case 2:
		switch a[0] {
		case 0:
			return time.Duration(a[1] * 1e9), nil
		case 1:
			return time.Duration(a[1] * 1e6), nil
		case 2:
			return time.Duration(a[1] * 1e3), nil
		case 3:
			return time.Duration(a[1]), nil
		default:
			return 0, _errInvalidDurationShardsFormat
		}

	case 3:
		switch a[0] {
		case 4:
			return time.Duration(a[1]*1e9) + time.Duration(a[2]), nil
		default:
			return 0, _errInvalidDurationShardsFormat
		}

	default:
		return 0, _errInvalidDurationShardsLength
	}
}

func ShardDuration(d time.Duration) any {
	if d == 0 {
		return nil
	}

	v := int64(d)
	switch {
	case v%1e9 == 0:
		return &[2]int64{0, v / 1e9}
	case v%1e6 == 0:
		return &[2]int64{1, v / 1e6}
	case v%1e3 == 0:
		return &[2]int64{2, v / 1e3}
	case v/1e9 == 0:
		return &[2]int64{3, v}
	default:
		return &[3]int64{4, v / 1e9, v % 1e9}
	}
}
