package archmage

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"reflect"
	"strings"
)

func ApplyMapOverride[K comparable, V any, T map[K]V](base *T, data []byte) error {
	var ovr T
	err := json.Unmarshal(data, &ovr)
	if err != nil {
		return err
	}

	m := *base
	if m == nil {
		if ovr != nil {
			m = make(map[K]V)
			*base = m
		}
	}

	for k, v := range ovr {
		m[k] = v
	}

	return nil
}

type OverridablePtr[T any] interface {
	*T
	Overridable
}

func ApplyMapValueOverride[K comparable, E any, V OverridablePtr[E], T map[K]V](base *T, data []byte) error {
	var ovr map[K]jsontext.Value
	err := json.Unmarshal(data, &ovr)
	if err != nil {
		return err
	}

	m := *base
	if m == nil {
		if ovr != nil {
			m = make(map[K]V)
			*base = m
		}
	}

	for k, d := range ovr {
		if overridable, ok := m[k]; ok && overridable != nil {
			err = overridable.ApplyOverride(d)
			if err != nil {
				return nil
			}
		} else {
			var tmp E
			err = json.Unmarshal(d, &tmp)
			if err != nil {
				return nil
			}
			m[k] = &tmp
		}
	}

	return nil
}

func BuildJSONKeyToFieldIndexMap[T any](fields map[string]int8) map[string]int {
	var obj T
	x := reflect.ValueOf(obj)
	if x.Kind() != reflect.Struct {
		return nil
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

func ApplyStructOverride[T any](base *T, data []byte, typeName string, fields map[string]int8, fieldIndexMap map[string]int) error {
	var tmp map[string]jsontext.Value
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	x := reflect.ValueOf(base).Elem()
	for k, d := range tmp {
		if fields[k] == 0 {
			return fmt.Errorf("%s: unknown object field name %q in override data", typeName, k)
		}
		index, ok := fieldIndexMap[k]
		if !ok {
			continue
		}
		field := x.Field(index)
		switch fields[k] {
		case 1:
			err = json.Unmarshal(d, field.Addr().Interface())
		case 2:
			field.SetZero()
			err = json.Unmarshal(d, field.Addr().Interface())
		case 3:
			err = field.Addr().Interface().(Overridable).ApplyOverride(d)
		default:
			panic("unreachable")
		}
		if err != nil {
			return fmt.Errorf("%s: failed to apply override data to field %q: %w", typeName, k, err)
		}
	}

	return nil
}
