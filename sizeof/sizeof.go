package sizeof

import "reflect"

func StructSize(ptr interface{}) int64 {
	refVal := reflect.ValueOf(ptr)
	if (refVal.Kind() == reflect.Ptr || refVal.Kind() == reflect.Interface) && refVal.IsNil() {
		return 0
	}
	c := &calc{ptrs: map[uintptr]struct{}{}}
	return c.size(refVal)
}

type calc struct {
	ptrs map[uintptr]struct{}
}

func (c *calc) size(v reflect.Value) (size int64) {
	switch v.Kind() {
	case reflect.Invalid:
		return
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			size += c.size(v.Field(i))
		}
	case reflect.Ptr:
		arrd := v.Pointer()
		if _, ok := c.ptrs[arrd]; ok {
			return
		}
		c.ptrs[arrd] = struct{}{}

		size = int64(v.Type().Size())
		if v.IsNil() {
			return
		}

		size += c.size(v.Elem())
	case reflect.Slice:
		size = int64(v.Type().Size())
		if v.IsNil() {
			return
		}

		for i := 0; i < v.Len(); i++ {
			size += c.size(v.Index(i))
		}

		reserved := v.Cap() - v.Len()
		size += int64(v.Type().Elem().Size()) * int64(reserved)
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			size += c.size(v.Index(i))
		}
	case reflect.Map:
		size = int64(v.Type().Size())
		if v.IsNil() {
			return
		}

		for _, key := range v.MapKeys() {
			size += c.size(key)

			val := v.MapIndex(key)
			size += c.size(val)
		}
	case reflect.String:
		size += int64(v.Type().Size())
		size += int64(v.Len())
	case reflect.Interface:
		if v.IsNil() || !v.Elem().IsValid() {
			return
		}
		size += int64(v.Type().Size())
		size += c.size(v.Elem())
	default:
		size = int64(v.Type().Size())
		return
	}

	return
}
