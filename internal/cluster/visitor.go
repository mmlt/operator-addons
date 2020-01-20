package cluster

import (
	"reflect"
	"strconv"
	"strings"
)

type VisitorFn func([]string, string)

// Visit calls fn for each field in the object starting a root.
func Visit(root interface{}, fn VisitorFn) {
	visit([]string{}, reflect.ValueOf(root), fn)
}

func MapToEnv(root interface{}, prefix string) []string {
	var result []string
	Visit(root, func(p []string, v string) {
		result = append(result, strings.ToUpper(prefix+strings.Join(p, "_"))+"="+v)
	})
	return result
}

func visit(path []string, v reflect.Value, fn VisitorFn) {
	switch v.Kind() {
	case reflect.Invalid:
		//fmt.Printf("%s = invalid\n", path)
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			//display(fmt.Sprintf("%s[%d]", path, i), v.Index(i))
			visit(append(path, strconv.Itoa(i)), v.Index(i), fn)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			//fieldPath := fmt.Sprintf("%s.%s", path, v.Type().Field(i).Name)
			//display(fieldPath, v.Field(i))
			visit(append(path, v.Type().Field(i).Name), v.Field(i), fn)
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			//display(fmt.Sprintf("%s[%s]", path,
			//	formatAtom(key)), v.MapIndex(key))
			visit(append(path, formatAtom(key)), v.MapIndex(key), fn)
		}
	case reflect.Ptr:
		if !v.IsNil() {
			//display(fmt.Sprintf("(*%s)", path), v.Elem())
			visit(path, v.Elem(), fn)
		}
		//else {
		//	fmt.Printf("%s = nil\n", path)
		//}
	case reflect.Interface:
		if !v.IsNil() {
			//fmt.Printf("%s.type = %s\n", path, v.Elem().Type())
			//display(path+".value", v.Elem())
			visit(path, v.Elem(), fn)
		}
		//else {
		//	fmt.Printf("%s = nil\n", path)
		//}
	default: // basic types, channels, funcs
		//fmt.Printf("%s = %s\n", path, formatAtom(v))
		fn(path, formatAtom(v))
	}
}

// FormatAtom formats a value without inspecting its internal structure.
// It is a copy of the the function in gopl.io/ch11/format with the following modifications:
//	- case reflect.Interface is added.
//	- strings are printed without quotes.
func formatAtom(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Invalid:
		return "invalid"
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10)
	// ...floating-point and complex cases omitted for brevity...
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	case reflect.String:
		//return strconv.Quote(v.String())
		return v.String()
	case reflect.Chan, reflect.Func, reflect.Ptr,
		reflect.Slice, reflect.Map:
		return v.Type().String() + " 0x" +
			strconv.FormatUint(uint64(v.Pointer()), 16)
	case reflect.Interface:
		return v.Elem().String()
	default: // reflect.Array, reflect.Struct
		return v.Type().String() + " value"
	}
}
