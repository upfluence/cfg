package cfg

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type setterFactory interface {
	buildSetter(reflect.StructField) setter
}

type defaultSetterFactory struct{}

func indirectedValue(v reflect.Value) reflect.Value {
	if v.Type().Kind() == reflect.Ptr {
		return v.Elem()
	}

	return v
}

func indirectedType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}

	return t
}

func (dsf *defaultSetterFactory) buildBasicParser(t reflect.Type) (parser, bool) {
	var (
		k = t.Kind()

		ptr bool
	)

	if k == reflect.Ptr {
		k = t.Elem().Kind()
		ptr = true
	}

	switch k {
	case reflect.String:
		return &stringParser{}, ptr
	case reflect.Int, reflect.Int64:
		return &intParser{transformer: intTransformers[k]}, ptr
	case reflect.Bool:
		return &boolParser{}, ptr
	}

	return nil, false
}

func (dsf *defaultSetterFactory) buildParser(t reflect.Type) parser {
	k := t.Kind()

	switch k {
	case reflect.Slice:
		p, ptr := dsf.buildBasicParser(t.Elem())

		if p == nil {
			return nil
		}

		return &sliceParser{p: p, t: t, ptr: ptr}
	case reflect.Map:
		vp, vptr := dsf.buildBasicParser(t.Elem())

		if vp == nil {
			return nil
		}

		kp, kptr := dsf.buildBasicParser(t.Key())

		if kp == nil {
			return nil
		}

		return &mapParser{t: t, vp: vp, vptr: vptr, kp: kp, kptr: kptr}
	}

	p, _ := dsf.buildBasicParser(t)

	return p
}

func (factory *defaultSetterFactory) buildSetter(f reflect.StructField) setter {
	if p := factory.buildParser(indirectedType(f.Type)); p != nil {
		return &parserSetter{field: f, parser: p}
	}

	return nil
}

type setter interface {
	set(string, interface{}) error
}

type ErrSetterNotImplemented struct {
	field reflect.StructField
}

func (e *ErrSetterNotImplemented) Error() string {
	return fmt.Sprintf("cfg: Setter not implemented for type %v", e.field.Type)
}

type boolParser struct{}

type ErrNotBoolValue struct {
	value string
}

func (e *ErrNotBoolValue) Error() string {
	return fmt.Sprintf("cfg: Can't parse %q in a bool value", e.value)
}

func (s *boolParser) parse(value string, ptr bool) (interface{}, error) {
	var v bool

	switch strings.TrimSpace(value) {
	case "t", "1", "true":
		v = true
	case "f", "0", "false":
	default:
		return nil, &ErrNotBoolValue{value: value}
	}

	if ptr {
		return &v, nil
	}

	return v, nil
}

type parserSetter struct {
	field  reflect.StructField
	parser parser
}

func (s *parserSetter) set(value string, target interface{}) error {
	var t = indirectedValue(reflect.ValueOf(target)).FieldByName(s.field.Name)

	v, err := s.parser.parse(value, t.Type().Kind() == reflect.Ptr)

	if err != nil {
		return err
	}

	t.Set(reflect.ValueOf(v))

	return nil
}

type parser interface {
	parse(string, bool) (interface{}, error)
}

type mapParser struct {
	t reflect.Type

	vp, kp parser

	vptr, kptr bool
}

func (mp *mapParser) parse(v string, ptr bool) (interface{}, error) {
	args := strings.Split(v, ",")
	res := reflect.MakeMap(mp.t)

	for _, arg := range args {
		vs := strings.SplitN(arg, "=", 2)

		if len(vs) != 2 {
			continue
		}

		k, err := mp.kp.parse(vs[0], mp.kptr)

		if err != nil {
			return nil, err
		}

		v, err := mp.vp.parse(vs[1], mp.vptr)

		if err != nil {
			return nil, err
		}

		res.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
	}

	return res.Interface(), nil

}

type sliceParser struct {
	t reflect.Type

	p   parser
	ptr bool
}

func (sp *sliceParser) parse(v string, ptr bool) (interface{}, error) {
	args := strings.Split(v, ",")
	res := reflect.MakeSlice(sp.t, 0, len(args))

	for _, arg := range args {
		v, err := sp.p.parse(arg, sp.ptr)

		if err != nil {
			return nil, err
		}

		res = reflect.Append(res, reflect.ValueOf(v))
	}

	return res.Interface(), nil
}

type stringParser struct{}

func (*stringParser) parse(v string, ptr bool) (interface{}, error) {
	if ptr {
		x := v
		return &x, nil
	}

	return v, nil
}

type intTransformer func(int64, bool) interface{}

var intTransformers = map[reflect.Kind]intTransformer{
	reflect.Int: func(v int64, ptr bool) interface{} {
		if ptr {
			x := int(v)
			return &x
		}

		return int(v)
	},
	reflect.Int64: func(v int64, ptr bool) interface{} {
		if ptr {
			x := v
			return &x
		}

		return v
	},
}

type intParser struct {
	transformer intTransformer
}

func (s *intParser) parse(value string, ptr bool) (interface{}, error) {
	var v, err = strconv.ParseInt(value, 10, 0)

	if err != nil {
		return nil, err
	}

	return s.transformer(v, ptr), nil
}
