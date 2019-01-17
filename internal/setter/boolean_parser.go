package setter

import "strings"

type boolParser struct{}

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
