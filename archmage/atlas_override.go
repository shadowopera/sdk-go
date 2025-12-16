package archmage

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"reflect"
	"strings"
)

func ApplyMapOverride[K comparable, V any, T ~map[K]V](base *T, data []byte) (*T, error) {
	var ovr T
	err := json.Unmarshal(data, &ovr)
	if err != nil {
		return nil, err
	}

	m := *base
	if m == nil {
		if ovr != nil {
			m = make(T)
		}
	}

	for k, v := range ovr {
		m[k] = v
	}

	return &m, nil
}

type OverridablePtr[T any] interface {
	*T
	Overridable
}

func ApplyMapValueOverride[K comparable, E any, V OverridablePtr[E], T ~map[K]V](base *T, data []byte) (*T, error) {
	var ovr map[K]jsontext.Value
	err := json.Unmarshal(data, &ovr)
	if err != nil {
		return nil, err
	}

	m := *base
	if m == nil {
		if ovr != nil {
			m = make(T)
		}
	}

	for k, d := range ovr {
		if overridable, ok := m[k]; ok && overridable != nil {
			r, err := overridable.ApplyOverride(d)
			if err != nil {
				return nil, err
			}
			m[k] = r.(V)
		} else {
			var tmp E
			err = json.Unmarshal(d, &tmp)
			if err != nil {
				return nil, err
			}
			m[k] = &tmp
		}
	}

	return &m, nil
}

func ApplyArrayOverride[E any, T ~[]E](_ *T, data []byte) (*T, error) {
	var ovr *T
	err := json.Unmarshal(data, &ovr)
	if err != nil {
		return nil, err
	}
	return ovr, nil
}

func BuildJSONKeyToFieldIndexMap[T any](fields map[string]int8) map[string]int {
	var obj T
	x := reflect.ValueOf(obj)
	if x.Kind() != reflect.Struct {
		panic("unreachable")
	}

	typ := x.Type()
	m := make(map[string]int)
	for i := range typ.NumField() {
		f := typ.Field(i)
		t := f.Tag.Get("json")
		if t == "" {
			continue
		}
		k := t
		p := strings.Index(t, ",")
		if p >= 0 {
			k = strings.TrimSpace(t[:p])
		}
		if fields[k] != 0 {
			m[k] = i
		}
	}

	return m
}

func ApplyStructOverride[T any](obj *T, data []byte, typeName string, fields map[string]int8, fieldIndexMap map[string]int) (*T, error) {
	var ovr map[string]jsontext.Value
	err := json.Unmarshal(data, &ovr)
	if err != nil {
		return nil, err
	}

	if obj == nil {
		if ovr != nil {
			var x T
			obj = &x
		}
	}

	x := reflect.ValueOf(obj).Elem()
	for k, d := range ovr {
		fx := fields[k]
		if fx == 0 {
			return nil, fmt.Errorf("%s: unknown object field name %q in override data", typeName, k)
		}
		index, ok := fieldIndexMap[k]
		if !ok {
			continue
		}
		field := x.Field(index)
		switch fx {
		case 1:
			err = json.Unmarshal(d, field.Addr().Interface())
		case 2:
			field.SetZero()
			err = json.Unmarshal(d, field.Addr().Interface())
		case 3:
			var r Overridable
			r, err = field.Addr().Interface().(Overridable).ApplyOverride(d)
			field.Set(reflect.ValueOf(r))
		default:
			panic("unreachable")
		}
		if err != nil {
			return nil, fmt.Errorf("%s: failed to apply override data to field %q: %w", typeName, k, err)
		}
	}

	return obj, nil
}
