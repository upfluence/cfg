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

func indirectedFieldKind(t reflect.Type) reflect.Kind {
	return indirectedType(t).Kind()
}

func (*defaultSetterFactory) buildParser(k reflect.Kind) parser {
	switch k {
	case reflect.String:
		return &stringParser{}
	case reflect.Int, reflect.Int64, reflect.Int32:
		return &intParser{transformer: intTransformers[k]}
	case reflect.Struct:
		return &structParser{transformer: structTransformers[k]}
	case reflect.Bool:
		return &boolParser{}
		//case reflect.Float32, reflect.Float64:
		//	return &floatParser{transformer: floatTransformers[k]}
	}

	return nil
}

func (factory *defaultSetterFactory) buildSetter(f reflect.StructField) setter {
	if p := factory.buildParser(indirectedFieldKind(f.Type)); p != nil {
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

	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "T", "1", "TRUE":
		v = true
	case "F", "0", "FALSE":
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
	reflect.Int32: func(v int64, ptr bool) interface{} {
		if ptr {
			x := int32(v)
			return &x
		}

		return int32(v)
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

type floatTransformer func(int64, bool) interface{}

var floatTransformers = map[reflect.Kind]floatTransformer{
	reflect.Float64: func(v int64, ptr bool) interface{} {
		if ptr {
			x := float64(v)
			return &x
		}

		return int(v)
	},
	reflect.Float32: func(v int64, ptr bool) interface{} {
		if ptr {
			x := v
			return &x
		}

		return v
	},
}

type floatParser struct {
	transformer intTransformer
}

//func (s *floatParser) parse(value string, ptr bool) (interface{}, error) {
//	var v, err = strconv.ParseFloat(value, )
//
//	if err != nil {
//		return nil, err
//	}
//
//	return s.transformer(v, ptr), nil
//}

var structTransformers = map[reflect.Kind]structTransformer{
	reflect.Struct: func(v interface{}, ptr bool) interface{} {
		return v
	},
}

type structTransformer func(interface{}, bool) interface{}

type structParser struct {
	transformer structTransformer
}

func (s *structParser) parse(value string, ptr bool) (interface{}, error) {
	return s.transformer(value, ptr), nil
}
