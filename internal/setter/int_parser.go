package setter

import (
	"math"
	"reflect"
	"strconv"
)

type intTransformer func(int64, bool) (interface{}, error)

type intParser struct {
	transformer intTransformer
}

func (s *intParser) parse(value string, ptr bool) (interface{}, error) {
	if v, err := strconv.ParseInt(value, 10, 64); err != nil {
		return nil, err
	} else {
		return s.transformer(v, ptr)
	}
}

func intTransformerFactory(t reflect.Kind) intTransformer {
	return func(v int64, ptr bool) (interface{}, error) {
		var fun = intFuncs(t)

		if ret, err := fun(v, ptr); err != nil {
			return nil, err
		} else {
			return ret, nil
		}
	}
}

func intFuncs(kind reflect.Kind) func(int64, bool) (interface{}, error) {
	switch kind {
	case reflect.Int:
		return func(v int64, b bool) (interface{}, error) {
			const (
				maxUint  = ^uint(0)
				maxRange = int64(int(maxUint >> 1))
				minRange = int64(-int(maxRange) - 1)
			)

			if v > maxRange || v < minRange {
				return nil, &ErrInvalidRange{kind.String(), v}
			}

			var val = int(v)
			if b {
				return &val, nil
			}

			return val, nil
		}
	case reflect.Int8:
		return func(v int64, b bool) (interface{}, error) {
			const (
				minRange = int64(math.MinInt8)
				maxRange = int64(math.MaxInt8)
			)

			if v > maxRange || v < minRange {
				return nil, &ErrInvalidRange{kind.String(), v}
			}

			var val = int8(v)
			if b {
				return &val, nil
			}

			return val, nil
		}
	case reflect.Int16:
		return func(v int64, b bool) (interface{}, error) {
			const (
				minRange = int64(math.MinInt16)
				maxRange = int64(math.MaxInt16)
			)

			if v > maxRange || v < minRange {
				return nil, &ErrInvalidRange{kind.String(), v}
			}

			var val = int16(v)
			if b {
				return &val, nil
			}

			return val, nil
		}
	case reflect.Int32:
		return func(v int64, b bool) (interface{}, error) {
			const (
				minRange = int64(math.MinInt32)
				maxRange = int64(math.MaxInt32)
			)

			if v > maxRange || v < minRange {
				return nil, &ErrInvalidRange{kind.String(), v}
			}

			var val = int32(v)
			if b {
				return &val, nil
			}

			return val, nil
		}
	case reflect.Int64:
		return func(v int64, b bool) (interface{}, error) {
			if b {
				return &v, nil
			}

			return v, nil
		}
	default:
		return func(v int64, b bool) (interface{}, error) {
			return nil, &ErrKindTypeNotImplemented{kind.String()}
		}
	}

}
