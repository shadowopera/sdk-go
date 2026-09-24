package internal

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"shadop.dev/pkg/sdk-go/archmage"
)

var (
	_archmagePkgPath = reflect.TypeFor[archmage.AtlasItem]().PkgPath()
)

// checkXRefs walks every XRef reachable from atlas and checks that its binding
// matches its CfgID: a zero CfgID has a nil Ref, and a non-zero CfgID has a Ref
// whose ID equals the CfgID. It returns the number of bound refs.
func checkXRefs(t *testing.T, atlas archmage.Atlas) int {
	t.Helper()
	w := &xrefWalker{t: t, visited: make(map[uintptr]bool)}
	w.walk(reflect.ValueOf(atlas), "atlas")
	return w.bound
}

type xrefWalker struct {
	t       *testing.T
	visited map[uintptr]bool
	bound   int
}

func (w *xrefWalker) walk(v reflect.Value, path string) {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() || w.visited[v.Pointer()] {
			return
		}
		w.visited[v.Pointer()] = true
		w.walk(v.Elem(), path)
	case reflect.Interface:
		if !v.IsNil() {
			w.walk(v.Elem(), path)
		}
	case reflect.Struct:
		if isXRef(v.Type()) {
			w.check(v, path)
			return
		}
		for i := range v.NumField() {
			if f := v.Type().Field(i); f.IsExported() {
				w.walk(v.Field(i), path+"."+f.Name)
			}
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			w.walk(v.Index(i), fmt.Sprintf("%s[%d]", path, i))
		}
	case reflect.Map:
		for iter := v.MapRange(); iter.Next(); {
			w.walk(iter.Value(), fmt.Sprintf("%s[%v]", path, iter.Key()))
		}
	default:
	}
}

func (w *xrefWalker) check(v reflect.Value, path string) {
	w.t.Helper()
	cfgID := v.FieldByName("CfgID")
	ref := v.FieldByName("Ref")
	switch {
	case cfgID.IsZero():
		if !ref.IsNil() {
			w.t.Errorf("%s: zero CfgID but Ref is bound", path)
		}
	case ref.IsNil():
		w.t.Errorf("%s: CfgID %v but Ref is nil", path, cfgID)
	default:
		if id := ref.Elem().FieldByName("ID"); !id.Equal(cfgID) {
			w.t.Errorf("%s: CfgID %v but Ref.ID is %v", path, cfgID, id)
		}
		w.bound++
	}
}

func isXRef(typ reflect.Type) bool {
	return typ.PkgPath() == _archmagePkgPath && strings.HasPrefix(typ.Name(), "XRef[")
}
