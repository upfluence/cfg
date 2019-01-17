package setter

import (
	"reflect"
)

type Setter interface {
	Set(string, interface{}) error
}

type parser interface {
	parse(string, bool) (interface{}, error)
}

type ParserSetter struct {
	Field  reflect.StructField
	Parser parser
}

func (s *ParserSetter) Set(value string, target interface{}) error {
	var t = IndirectedValue(reflect.ValueOf(target)).FieldByName(s.Field.Name)

	v, err := s.Parser.parse(value, t.Type().Kind() == reflect.Ptr)

	if err != nil {
		return err
	}

	t.Set(reflect.ValueOf(v))

	return nil
}
