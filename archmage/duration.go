package archmage

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"time"
)

var (
	_errInvalidDurationShardsType   = errors.New("invalid duration shards type")
	_errInvalidDurationShardsLength = errors.New("invalid duration shards length")
	_errInvalidDurationShardsFormat = errors.New("invalid duration shards format")
)

var (
	_ json.MarshalerTo     = (*Duration)(nil)
	_ json.UnmarshalerFrom = (*Duration)(nil)
)

// Duration wraps time.Duration with custom JSON marshaling.
// It serializes to/from an array format optimized for different time units.
type Duration struct {
	time.Duration
}

// MarshalJSONTo encodes Duration as a compact JSON integer array,
// or null if the duration is zero.
func (d *Duration) MarshalJSONTo(enc *jsontext.Encoder) error {
	var a []int64
	r := ShardDuration(d.Duration)
	switch x := r.(type) {
	case nil:
		return enc.WriteToken(jsontext.Null)
	case *[2]int64:
		a = (*x)[:]
	case *[3]int64:
		a = (*x)[:]
	default:
		panic("unreachable")
	}

	err := enc.WriteToken(jsontext.BeginArray)
	if err != nil {
		return err
	}
	for _, v := range a {
		if err = enc.WriteToken(jsontext.Int(v)); err != nil {
			return err
		}
	}

	return enc.WriteToken(jsontext.EndArray)
}

// UnmarshalJSONFrom decodes a JSON array or null into Duration.
func (d *Duration) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}

	switch tok.Kind() {
	case 'n':
		d.Duration = 0
		return nil
	case '[':
	default:
		return fmt.Errorf("archmage.Duration: invalid JSON token kind %q, expected 'n' or '['", tok.Kind())
	}

	var pitch [3]int64
	var a = pitch[:0]
	for {
		tok, err = dec.ReadToken()
		if err != nil {
			return err
		}
		switch tok.Kind() {
		case '0':
			a = append(a, tok.Int())
			continue
		case ']':
		default:
			return fmt.Errorf("archmage.Duration: invalid JSON token kind %q in array, expected '0' or ']'", tok.Kind())
		}
		break
	}

	r, err := parseDurationShardsImpl(a)
	if err != nil {
		return err
	}

	d.Duration = r
	return nil
}

// ParseDurationShards converts a duration shard array to time.Duration.
// Accepts *[2]int64, *[3]int64, []int64, or nil.
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

// ShardDuration converts time.Duration to an optimized shard array.
// Returns nil for zero duration, otherwise *[2]int64 or *[3]int64.
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
