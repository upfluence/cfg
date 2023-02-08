package cfg

import (
	"fmt"
	"reflect"

	"github.com/upfluence/cfg/provider"
)

type ProvidingError struct {
	Err error

	Key      string
	Field    reflect.StructField
	Provider provider.Provider
}

func (pe *ProvidingError) Unwrap() error { return pe.Err }

func (pe *ProvidingError) Error() string {
	return fmt.Sprintf(
		"cant provide value for %s.%s(%q, %q): %s",
		pe.Field.Type.Name(),
		pe.Field.Name,
		pe.Key,
		pe.Provider.StructTag(),
		pe.Err.Error(),
	)
}

type SettingError struct {
	Err error

	Value    string
	Key      string
	Field    reflect.StructField
	Provider provider.Provider
}

func (se *SettingError) Unwrap() error { return se.Err }

func (se *SettingError) Error() string {
	return fmt.Sprintf(
		"cant set value for %s.%s(%q, %q, %q): %s",
		se.Field.Type.Name(),
		se.Field.Name,
		se.Key,
		se.Provider.StructTag(),
		se.Value,
		se.Err.Error(),
	)
}
