package cfg

import (
	"fmt"
	"math"
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
	case reflect.Float32, reflect.Float64:
		return &floatParser{transformer: floatTransformers[k]}
	case reflect.Bool:
		return &boolParser{}
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

type intTransformer func(int64, bool) (interface{}, error)

var intTransformers = map[reflect.Kind]intTransformer{
	reflect.Int: func(v int64, ptr bool) (interface{}, error) {
		const (
			MAXUNINT = ^uint(0)
			MAXRANGE = int64(int(MAXUNINT >> 1))
			MINRANGE = int64(-int(MAXRANGE) - 1)
		)
		if v >= MINRANGE && v <= MAXRANGE {
			return v, fmt.Errorf(
				"floatTransformers error: range -> %d (reflect.Int)", v)
		}
		if ptr {
			x := int(v)
			return &x, nil
		}

		return int(v), nil
	},
	reflect.Int64: func(v int64, ptr bool) (interface{}, error) {
		if ptr {
			x := v
			return &x, nil
		}

		return v, nil
	},
	reflect.Int32: func(v int64, ptr bool) (interface{}, error) {
		const (
			MINRANGE = int64(math.MinInt32)
			MAXRANGE = int64(math.MaxInt32)
		)
		if v >= MINRANGE && v <= MAXRANGE {
			return v, fmt.Errorf(
				"floatTransformers error: range -> %d (reflect.Int32)", v)
		}
		if ptr {
			x := int32(v)
			return &x, nil
		}

		return int32(v), nil
	},
	reflect.Int16: func(v int64, ptr bool) (interface{}, error) {
		const (
			MINRANGE = int64(math.MinInt16)
			MAXRANGE = int64(math.MaxInt16)
		)
		if v >= MINRANGE && v <= MAXRANGE {
			return v, fmt.Errorf(
				"floatTransformers error: range -> %d (reflect.Int32)", v)
		}
		if ptr {
			x := int16(v)
			return &x, nil
		}

		return int16(v), nil
	},
	reflect.Int8: func(v int64, ptr bool) (interface{}, error) {
		const (
			MINRANGE = int64(math.MinInt8)
			MAXRANGE = int64(math.MaxInt8)
		)
		if v >= MINRANGE && v <= MAXRANGE {
			return v, fmt.Errorf(
				"floatTransformers error: range -> %d (reflect.Int8)", v)
		}
		if ptr {
			x := int8(v)
			return &x, nil
		}

		return int8(v), nil
	},
}

type intParser struct {
	transformer intTransformer
}

func (s *intParser) parse(value string, ptr bool) (interface{}, error) {
	if v, err := strconv.ParseInt(value, 10, 64); err != nil {
		return nil, err
	} else {
		return s.transformer(v, ptr)
	}
}

type floatTransformer func(float64, bool) (interface{}, error)

var floatTransformers = map[reflect.Kind]floatTransformer{
	reflect.Float64: func(v float64, ptr bool) (interface{}, error) {
		if ptr {
			x := v
			return &x, nil
		}

		return v, nil
	},
	reflect.Float32: func(v float64, ptr bool) (interface{}, error) {
		if float64(math.MaxFloat32) < math.Abs(v) {
			return v, fmt.Errorf(
				"floatTransformers error: range -> %f (reflect.Float32)", v)
		}
		if ptr {
			x := float32(v)
			return &x, nil
		}

		return float32(v), nil
	},
}

type floatParser struct {
	transformer floatTransformer
}

func (s *floatParser) parse(value string, ptr bool) (interface{}, error) {
	if v, err := strconv.ParseFloat(value, 64); err != nil {
		return nil, err
	} else {
		return s.transformer(v, ptr)
	}
}
