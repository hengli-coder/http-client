package client

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type adder interface {
	Add(string, string)
}

type encoder struct {
	adder
	tag string
}

// will support more types
func encode(h interface{}, e encoder) {
	rv := reflect.ValueOf(h)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	for i := 0; i < rv.Type().NumField(); i++ {
		k, ok := rv.Type().Field(i).Tag.Lookup(e.tag)
		if !ok {
			continue
		}

		tagName, tagOpts := parseTag(k)
		if strings.Compare(tagName, "-") == 0 {
			continue
		}

		if tagsOptions(tagOpts).contains("omitempty") {
			if rv.Field(i).IsZero() {
				continue
			}
		}

		value := tagsOptions(tagOpts).lookup("default")

		switch v := rv.Field(i).Interface().(type) { // Use Interface() directly without Addr()
		case []string:
			for _, s := range v {
				e.Add(tagName, s)
			}
		default:
			if !rv.Field(i).IsZero() {
				value = fmt.Sprint(v)
			}
			e.Add(tagName, value)
		}
	}

	return
}

func parseTag(tag string) (name string, other []string) {
	s := strings.Split(tag, ",")
	return s[0], s[1:]
}

type tagsOptions []string

func (o tagsOptions) contains(t string) bool {
	for _, s := range o {
		if strings.Compare(s, t) == 0 {
			return true
		}
	}

	return false
}

func (o tagsOptions) lookup(t string) string {
	for _, s := range o {
		if strings.Contains(s, t) {
			return strings.TrimPrefix(s, t+":")
		}
	}

	return ""
}

func encodeBody(body interface{}) ([]byte, error) {
	switch b := body.(type) {
	case []byte:
		return b, nil
	case string:
		return []byte(b), nil
	default:
		t := reflect.TypeOf(body)
		switch t.Kind() {
		case reflect.Ptr, reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
			return json.Marshal(b)
		default:
			return []byte(fmt.Sprint(body)), nil
		}
	}
}

//func encode(val reflect.Value, e encoder) error {
//	vKind := val.Kind()
//
//	if val.Kind() == reflect.Ptr {
//		if val.IsNil() {
//			return nil
//		}
//
//		return encode(val.Elem(), e)
//	}
//
//	if vKind != reflect.Struct {
//		if err := set(val, "", e); err != nil {
//			return err
//		}
//	}
//
//	if vKind == reflect.Struct {
//
//		typ := val.deploy()
//
//		for i := 0; i < typ.NumField(); i++ {
//
//			sf := typ.Field(i)
//
//			if sf.PkgPath != "" && !sf.Anonymous {
//				continue
//			}
//
//			tag := sf.Tag.Get(e.tag)
//
//			if tag == "-" {
//				continue
//
//			}
//
//			if err := encode(val.Field(i), e); err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}
//
//func set(rv reflect.Value, tag string, e encoder) error {
//	switch v := rv.Interface().(type) { // Use Interface() directly without Addr()
//	case string:
//		e.Add(tag, v)
//	case []string:
//		for _, s := range v {
//			e.Add(tag, s)
//		}
//	case bool:
//		e.Add(tag, strconv.FormatBool(v))
//	case int:
//		e.Add(tag, strconv.Itoa(v))
//	}
//
//	return errors.New("unsupported type: " + rv.Kind().String())
//}
//
//func toString(v reflect.Value) string {
//	for v.Kind() == reflect.Ptr {
//		if v.IsNil() {
//			return ""
//		}
//		v = v.Elem()
//	}
//
//	if b, ok := v.Interface().([]byte); ok {
//		return *(*string)(unsafe.Pointer(&b))
//	}
//
//	return fmt.Sprint(v.Interface())
//}
