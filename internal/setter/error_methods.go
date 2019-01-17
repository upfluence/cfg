package setter

import (
	"fmt"
	"reflect"
)

type ErrNotBoolValue struct {
	value string
}

type ErrSetterNotImplemented struct {
	field reflect.StructField
}

type ErrInvalidRange struct {
	kind  string
	value interface{}
}

type ErrKindTypeNotImplemented struct {
	kind string
}

func (e *ErrNotBoolValue) Error() string {
	return fmt.Sprintf("cfg: Can't parse %q in a bool value", e.value)
}

func (e *ErrSetterNotImplemented) Error() string {
	return fmt.Sprintf("cfg: Setter not implemented for type %v", e.field.Type)
}

func (e *ErrInvalidRange) Error() string {
	return fmt.Sprintf(
		"cfg: Range overextended for type %s with value %v",
		e.kind,
		e.value,
	)
}

func (e *ErrKindTypeNotImplemented) Error() string {
	return fmt.Sprintf(
		"INTERNAL ERROR: %s transformer not implemented",
		e.kind,
	)
}
