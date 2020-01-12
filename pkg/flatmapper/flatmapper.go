package flatmapper

import "reflect"

// MapToStruct maps a flat map (e.g. no struct fields) to a flat struct.
func MapToStruct(tagKey string, srcStringMap interface{}, dst interface{}) interface{} {
	v := reflect.ValueOf(dst).Elem()
	t := v.Type()
	sv := reflect.ValueOf(srcStringMap)
	st := sv.Type()
	if st.Kind() != reflect.Map {
		panic("kind of srcStringMap must be map")
	}
	if st.Key().Kind() != reflect.String {
		panic("kind srcStringMap's key must be string")
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag, ok := f.Tag.Lookup(tagKey)
		if !ok {
			continue
		}
		va := sv.MapIndex(reflect.ValueOf(tag))
		if !va.IsValid() {
			continue
		}
		// interfaces need another deference
		if va.Kind() == reflect.Interface {
			va = va.Elem()
		}
		v.Field(i).Set(va)
	}
	return dst
}
