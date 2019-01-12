package setters

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type (
	setter interface {
		set(string, interface{}) error
	}
	parser interface {
		parse(string, bool) (interface{}, error)
	}
	parserSetter struct {
		field  reflect.StructField
		parser parser
	}
)

func (s *parserSetter) set(value string, target interface{}) error {
	var t = indirectedValue(reflect.ValueOf(target)).FieldByName(s.field.Name)

	v, err := s.parser.parse(value, t.Type().Kind() == reflect.Ptr)

	if err != nil {
		return err
	}

	t.Set(reflect.ValueOf(v))

	return nil
}

type (
	ErrNotBoolValue struct {
		value string
	}
	ErrSetterNotImplemented struct {
		field reflect.StructField
	}
)

func (e *ErrNotBoolValue) Error() string {
	return fmt.Sprintf("cfg: Can't parse %q in a bool value", e.value)
}

func (e *ErrSetterNotImplemented) Error() string {
	return fmt.Sprintf("cfg: Setter not implemented for type %v", e.field.Type)
}

type (
	boolParser   struct{}
	stringParser struct{}

	intParser struct {
		transformer intTransformer
	}
	floatParser struct {
		transformer floatTransformer
	}
)

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

func (*stringParser) parse(v string, ptr bool) (interface{}, error) {
	if ptr {
		x := v
		return &x, nil
	}

	return v, nil
}

func (s *intParser) parse(value string, ptr bool) (interface{}, error) {
	if v, err := strconv.ParseInt(value, 10, 64); err != nil {
		return nil, err
	} else {
		return s.transformer(v, ptr)
	}
}

func (s *floatParser) parse(value string, ptr bool) (interface{}, error) {
	if v, err := strconv.ParseFloat(value, 64); err != nil {
		return nil, err
	} else {
		return s.transformer(v, ptr)
	}
}
