package setter

import (
	"math"
	"reflect"
	"strconv"
)

type floatTransformer func(float64, bool) (interface{}, error)

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

func floatTransformerFactory(t reflect.Kind) floatTransformer {
	return func(v float64, ptr bool) (interface{}, error) {
		fun := floatFuncs(t)

		if ret, err := fun(v); err != nil {
			return nil, err
		} else {
			if ptr {
				return &ret, nil
			}

			return ret, nil
		}
	}
}

func floatFuncs(kind reflect.Kind) func(float64) (interface{}, error) {
	switch kind {
	case reflect.Float32:
		return func(v float64) (interface{}, error) {
			if float64(math.MaxFloat32) < math.Abs(v) {
				return nil, &ErrInvalidRange{kind.String(), v}
			}

			return float32(v), nil
		}
	case reflect.Float64:
		return func(v float64) (interface{}, error) {
			return v, nil
		}
	default:
		return func(v float64) (interface{}, error) {
			return nil, &ErrKindTypeNotImplemented{kind.String()}
		}
	}
}