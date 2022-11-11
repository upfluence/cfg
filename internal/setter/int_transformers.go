// Do not edit. Generated from internal/setter/int_transformers.go.erb
package setter

import "reflect"

var intTransformers = map[reflect.Kind]intTransformer{
	reflect.Int64: func(v int64, ptr bool) interface{} {
		if ptr {
			x := v
			return &x
		}

		return v
	},

	reflect.Int: func(v int64, ptr bool) interface{} {
		if ptr {
			x := int(v)
			return &x
		}

		return int(v)
	},

	reflect.Int8: func(v int64, ptr bool) interface{} {
		if ptr {
			x := int8(v)
			return &x
		}

		return int8(v)
	},

	reflect.Int16: func(v int64, ptr bool) interface{} {
		if ptr {
			x := int16(v)
			return &x
		}

		return int16(v)
	},

	reflect.Int32: func(v int64, ptr bool) interface{} {
		if ptr {
			x := int32(v)
			return &x
		}

		return int32(v)
	},

	reflect.Uint: func(v int64, ptr bool) interface{} {
		if ptr {
			x := uint(v)
			return &x
		}

		return uint(v)
	},

	reflect.Uint8: func(v int64, ptr bool) interface{} {
		if ptr {
			x := uint8(v)
			return &x
		}

		return uint8(v)
	},

	reflect.Uint16: func(v int64, ptr bool) interface{} {
		if ptr {
			x := uint16(v)
			return &x
		}

		return uint16(v)
	},

	reflect.Uint32: func(v int64, ptr bool) interface{} {
		if ptr {
			x := uint32(v)
			return &x
		}

		return uint32(v)
	},

	reflect.Uint64: func(v int64, ptr bool) interface{} {
		if ptr {
			x := uint64(v)
			return &x
		}

		return uint64(v)
	},
}
