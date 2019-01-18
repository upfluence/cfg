package setter

import (
	"reflect"
)

type SetterFactory interface {
	BuildSetter(reflect.StructField) Setter
}

func IndirectedValue(v reflect.Value) reflect.Value {
	if v.Type().Kind() == reflect.Ptr {
		return v.Elem()
	}

	return v
}

func IndirectedType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}

	return t
}

func IndirectedFieldKind(t reflect.Type) reflect.Kind {
	return IndirectedType(t).Kind()
}

type DefaultSetterFactory struct{}

func NewDefaultSetterFactory() *DefaultSetterFactory {
	return &DefaultSetterFactory{}
}

func (factory *DefaultSetterFactory) BuildSetter(f reflect.StructField) Setter {
	if p := factory.buildParser(IndirectedFieldKind(f.Type)); p != nil {
		return &ParserSetter{Field: f, Parser: p}
	}

	return nil
}

func (*DefaultSetterFactory) buildParser(k reflect.Kind) parser {
	switch k {
	case reflect.String:
		return &stringParser{}
	case reflect.Int, reflect.Int64, reflect.Int32:
		return &intParser{transformer: intTransformerFactory(k)}
	case reflect.Float32, reflect.Float64:
		return &floatParser{transformer: floatTransformerFactory(k)}
	case reflect.Bool:
		return &boolParser{}
	}

	return nil
}
