package setters

import (
	"reflect"
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
