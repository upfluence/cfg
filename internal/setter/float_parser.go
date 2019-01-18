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

func (fp *floatParser) parse(value string, ptr bool) (interface{}, error) {
	var v, err = strconv.ParseFloat(value, 64)

	if err != nil {
		return nil, err
	}

	return fp.transformer(v, ptr)
}

func floatTransformerFactory(t reflect.Kind) floatTransformer {
	return func(v float64, ptr bool) (interface{}, error) {
		var fun = floatFuncs(t)
		var ret, err = fun(v, ptr)

		if err != nil {
			return nil, err
		}

		return ret, nil
	}
}

func floatFuncs(kind reflect.Kind) func(float64, bool) (interface{}, error) {
	switch kind {
	case reflect.Float32:
		return func(v float64, b bool) (interface{}, error) {
			if float64(math.MaxFloat32) < math.Abs(v) {
				return nil, &ErrInvalidRange{kind.String(), v}
			}

			var val = float32(v)

			if b {
				return &val, nil
			}

			return val, nil
		}
	case reflect.Float64:
		return func(v float64, b bool) (interface{}, error) {
			if b {
				return &v, nil
			}

			return v, nil
		}
	default:
		return func(v float64, b bool) (interface{}, error) {
			return nil, &ErrKindTypeNotImplemented{kind}
		}
	}
}
