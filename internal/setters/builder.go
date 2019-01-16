package setters

import (
	"reflect"

	"github.com/upfluence/cfg/internal/setters/parsers"
)

type SetterFactory interface {
	buildSetter(reflect.StructField) Setter
}

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

type DefaultSetterFactory struct{}

func (factory *DefaultSetterFactory) buildSetter(f reflect.StructField) Setter {
	if p := factory.buildParser(indirectedFieldKind(f.Type)); p != nil {
		return &parserSetter{field: f, parser: p}
	}

	return nil
}

func (*DefaultSetterFactory) buildParser(k reflect.Kind) parsers.Parser {
	switch k {
	case reflect.String:
		return &parsers.StringParser{}
	case reflect.Int, reflect.Int64, reflect.Int32:
		return &parsers.IntParser{transformer: parsers.IntTransformerFactory(k)}
	case reflect.Float32, reflect.Float64:
		return &parsers.FloatParser{transformer: parsers.FloatTransformerFactory(k)}
	case reflect.Bool:
		return &parsers.BoolParser{}
	}

	return nil
}
