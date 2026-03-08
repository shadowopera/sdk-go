package archmage

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"os"
	"path/filepath"
	"time"
)

// DumpAtlas writes all loaded atlas items to JSON files in outputDir.
// Each item is written to a separate file named <key>.json.
func DumpAtlas(atlas Atlas, outputDir string, opts ...json.Options) error {
	for k, item := range atlas.AtlasItems() {
		if item.Ready {
			data, err := Canonicalize(item.Cfg, opts...)
			if err != nil {
				return err
			}
			p := filepath.Join(outputDir, k+".json")
			if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(p, data, 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

// Canonicalize marshals obj to indented, canonical JSON.
// It applies archmage's standard marshal options (nil maps/slices as null,
// zero time.Time and zero Vec pointers as null), then sorts object keys and
// indents the output. Additional opts are appended after the defaults.
// Returns the formatted JSON bytes or an error.
func Canonicalize(obj any, opts ...json.Options) ([]byte, error) {
	data, err := json.Marshal(obj, getMarshalOptions(opts)...)
	if err != nil {
		return nil, err
	}
	jsonValue := jsontext.Value(data)
	if err := jsonValue.Canonicalize(jsontext.CanonicalizeRawInts(false)); err != nil {
		return nil, err
	}
	if err := jsonValue.Indent(); err != nil {
		return nil, err
	}
	data = append(jsonValue, '\n')
	return data, nil
}

func getMarshalOptions(opts []json.Options) []json.Options {
	return append([]json.Options{
		json.FormatNilMapAsNull(true),
		json.FormatNilSliceAsNull(true),
		json.WithMarshalers(json.JoinMarshalers(
			json.MarshalToFunc[time.Time](func(enc *jsontext.Encoder, t time.Time) error {
				if t.IsZero() {
					return enc.WriteToken(jsontext.Null)
				}
				return json.SkipFunc
			}),
			json.MarshalToFunc[*Vec2[int]](zeroVecNullMarshalTo2),
			json.MarshalToFunc[*Vec2[int8]](zeroVecNullMarshalTo2),
			json.MarshalToFunc[*Vec2[int16]](zeroVecNullMarshalTo2),
			json.MarshalToFunc[*Vec2[int32]](zeroVecNullMarshalTo2),
			json.MarshalToFunc[*Vec2[int64]](zeroVecNullMarshalTo2),
			json.MarshalToFunc[*Vec2[uint]](zeroVecNullMarshalTo2),
			json.MarshalToFunc[*Vec2[uint8]](zeroVecNullMarshalTo2),
			json.MarshalToFunc[*Vec2[uint16]](zeroVecNullMarshalTo2),
			json.MarshalToFunc[*Vec2[uint32]](zeroVecNullMarshalTo2),
			json.MarshalToFunc[*Vec2[uint64]](zeroVecNullMarshalTo2),
			json.MarshalToFunc[*Vec2[float32]](zeroVecNullMarshalTo2),
			json.MarshalToFunc[*Vec2[float64]](zeroVecNullMarshalTo2),
			json.MarshalToFunc[*Vec3[int]](zeroVecNullMarshalTo3),
			json.MarshalToFunc[*Vec3[int8]](zeroVecNullMarshalTo3),
			json.MarshalToFunc[*Vec3[int16]](zeroVecNullMarshalTo3),
			json.MarshalToFunc[*Vec3[int32]](zeroVecNullMarshalTo3),
			json.MarshalToFunc[*Vec3[int64]](zeroVecNullMarshalTo3),
			json.MarshalToFunc[*Vec3[uint]](zeroVecNullMarshalTo3),
			json.MarshalToFunc[*Vec3[uint8]](zeroVecNullMarshalTo3),
			json.MarshalToFunc[*Vec3[uint16]](zeroVecNullMarshalTo3),
			json.MarshalToFunc[*Vec3[uint32]](zeroVecNullMarshalTo3),
			json.MarshalToFunc[*Vec3[uint64]](zeroVecNullMarshalTo3),
			json.MarshalToFunc[*Vec3[float32]](zeroVecNullMarshalTo3),
			json.MarshalToFunc[*Vec3[float64]](zeroVecNullMarshalTo3),
			json.MarshalToFunc[*Vec4[int]](zeroVecNullMarshalTo4),
			json.MarshalToFunc[*Vec4[int8]](zeroVecNullMarshalTo4),
			json.MarshalToFunc[*Vec4[int16]](zeroVecNullMarshalTo4),
			json.MarshalToFunc[*Vec4[int32]](zeroVecNullMarshalTo4),
			json.MarshalToFunc[*Vec4[int64]](zeroVecNullMarshalTo4),
			json.MarshalToFunc[*Vec4[uint]](zeroVecNullMarshalTo4),
			json.MarshalToFunc[*Vec4[uint8]](zeroVecNullMarshalTo4),
			json.MarshalToFunc[*Vec4[uint16]](zeroVecNullMarshalTo4),
			json.MarshalToFunc[*Vec4[uint32]](zeroVecNullMarshalTo4),
			json.MarshalToFunc[*Vec4[uint64]](zeroVecNullMarshalTo4),
			json.MarshalToFunc[*Vec4[float32]](zeroVecNullMarshalTo4),
			json.MarshalToFunc[*Vec4[float64]](zeroVecNullMarshalTo4),
		)),
	}, opts...)
}

func zeroVecNullMarshalTo2[T comparable](enc *jsontext.Encoder, v *Vec2[T]) error {
	var z T
	if v != nil && v.X == z && v.Y == z {
		return enc.WriteToken(jsontext.Null)
	}
	return json.SkipFunc
}

func zeroVecNullMarshalTo3[T comparable](enc *jsontext.Encoder, v *Vec3[T]) error {
	var z T
	if v != nil && v.X == z && v.Y == z && v.Z == z {
		return enc.WriteToken(jsontext.Null)
	}
	return json.SkipFunc
}

func zeroVecNullMarshalTo4[T comparable](enc *jsontext.Encoder, v *Vec4[T]) error {
	var z T
	if v != nil && v.X == z && v.Y == z && v.Z == z && v.W == z {
		return enc.WriteToken(jsontext.Null)
	}
	return json.SkipFunc
}
