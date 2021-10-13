package utils

import (
	"fmt"
	"reflect"
)

func ToInterfaceSlice(x interface{}) []interface{} {
	rx := reflect.ValueOf(x)
	if rx.Kind() != reflect.Slice {
		panic("ToInterfaceSlice must be received a slice type.")
	}

	r := make([]interface{}, rx.Len())
	for i := 0; i < rx.Len(); i++ {
		r[i] = rx.Index(i).Interface()
	}

	return r
}

func ToStringSlice(x interface{}) []string {
	rx := reflect.ValueOf(x)
	if rx.Kind() != reflect.Slice {
		panic("ToInterfaceSlice must be received a slice type.")
	}

	r := make([]string, rx.Len())
	for i := 0; i < rx.Len(); i++ {
		c := rx.Index(i).Interface().(string)
		r[i] = c
	}

	return r
}

func ToString(x interface{}) string {
	return fmt.Sprint(x)
}
