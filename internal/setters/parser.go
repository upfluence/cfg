package setters

import (
	"reflect"

	"github.com/upfluence/cfg/internal/setters/parsers"
)

type Setter interface {
	set(string, interface{}) error
}

type parserSetter struct {
	field  reflect.StructField
	parser parsers.Parser
}

func (s *parserSetter) set(value string, target interface{}) error {
	var t = indirectedValue(reflect.ValueOf(target)).FieldByName(s.field.Name)

	v, err := s.parser.Parse(value, t.Type().Kind() == reflect.Ptr)

	if err != nil {
		return err
	}

	t.Set(reflect.ValueOf(v))

	return nil
}
