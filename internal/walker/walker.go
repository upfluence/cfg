package walker

import (
	"reflect"
	"unicode"

	"github.com/upfluence/errors"
)

var (
	SkipStruct            = errors.New("skip struct")
	ErrShouldBeAStructPtr = errors.New("input should be a pointer")
)

type Field struct {
	Field reflect.StructField
	Value reflect.Value

	Ancestor *Field
}

type WalkFunc func(*Field) error

func Walk(in any, fn WalkFunc) error {
	return walkValue(in, fn, nil)
}

func walkValue(in any, fn WalkFunc, ancestor *Field) error {
	if p, ok := in.(Prefixed); ok {
		return walkPrefixed(p, fn, ancestor)
	}

	return walkStruct(in, fn, ancestor)
}

func walkPrefixed(p Prefixed, fn WalkFunc, ancestor *Field) error {
	extra := p.WalkAncestor()

	if extra == nil {
		return walkValue(p.WalkValue(), fn, ancestor)
	}

	// Clone the extra chain and graft it onto the incoming ancestor
	// so that the original chain is not mutated.
	clone := &Field{Field: extra.Field}
	tip := clone

	for cur := extra.Ancestor; cur != nil; cur = cur.Ancestor {
		tip.Ancestor = &Field{Field: cur.Field}
		tip = tip.Ancestor
	}

	tip.Ancestor = ancestor

	return walkValue(p.WalkValue(), fn, clone)
}

func walkStruct(in any, fn WalkFunc, ancestor *Field) error {
	if in == nil {
		return ErrShouldBeAStructPtr
	}

	inv := reflect.ValueOf(in)

	if inv.Type().Kind() != reflect.Ptr {
		return ErrShouldBeAStructPtr
	}

	if inv.Type().Elem().Kind() != reflect.Struct {
		return ErrShouldBeAStructPtr
	}

	if inv.IsNil() {
		return ErrShouldBeAStructPtr
	}

	return walk(inv, fn, ancestor)
}

func indirectedType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}

	return t
}

func indirectedValue(v reflect.Value) reflect.Value {
	if v.Type().Kind() == reflect.Ptr {
		return v.Elem()
	}

	return v
}

func addressValue(v reflect.Value) reflect.Value {
	if v.Type().Kind() == reflect.Ptr {
		return v
	}

	return v.Addr()
}

func walkField(nv reflect.Value, fn WalkFunc, f *Field) error {
	if nv.CanInterface() {
		if p, ok := nv.Interface().(Prefixed); ok {
			return walkPrefixed(p, fn, f)
		}
	}

	return walk(nv, fn, f)
}

func walk(v reflect.Value, fn WalkFunc, a *Field) error {
	vit := indirectedType(v.Type())

	for i := 0; i < vit.NumField(); i++ {
		sf := vit.Field(i)
		f := Field{
			Field:    sf,
			Value:    addressValue(v),
			Ancestor: a,
		}

		nv := indirectedValue(v).FieldByName(sf.Name)

		if sf.Type.Kind() != reflect.Ptr {
			nv = nv.Addr()
		} else if !nv.CanSet() {
			continue
		}

		if unicode.IsUpper(rune(sf.Name[0])) {
			switch err := fn(&f); err {
			case SkipStruct:
				continue
			case nil:
			default:
				return err
			}
		}

		if indirectedType(sf.Type).Kind() != reflect.Struct {
			continue
		}

		wasNil := sf.Type.Kind() == reflect.Ptr && nv.IsNil()

		if wasNil {
			nv.Set(reflect.New(sf.Type.Elem()))
		}

		if err := walkField(nv, fn, &f); err != nil {
			return err
		}

		if wasNil && reflect.DeepEqual(
			nv.Elem().Interface(),
			reflect.New(sf.Type.Elem()).Elem().Interface(),
		) {
			nv.Set(reflect.Zero(sf.Type))
		}
	}

	return nil
}
