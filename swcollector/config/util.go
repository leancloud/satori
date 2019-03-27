package config

import (
	"encoding/json"
	"reflect"
)

func FillDefault(v interface{}) {
	iface := reflect.ValueOf(v)
	elem := iface.Elem()
	typ := elem.Type()
	for i := 0; i < elem.NumField(); i++ {
		f := elem.Field(i)
		ft := typ.Field(i)
		if v := ft.Tag.Get("default"); v != "" {
			var zero bool
			switch ft.Type.Kind() {
			case reflect.Chan, reflect.Func, reflect.Map:
				fallthrough
			case reflect.Ptr, reflect.Interface, reflect.Slice:
				zero = f.IsNil()
			default:
				zero = reflect.Zero(ft.Type).Interface() == f.Interface()
			}
			if zero {
				json.Unmarshal([]byte(v), f.Addr().Interface())
			}
		}
	}
}
