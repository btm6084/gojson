package gojson

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func UnmarshalJSON(raw []byte, v interface{}) (err error) {
	defer PanicRecovery(&err)

	if len(raw) == 0 {
		return fmt.Errorf("empty json value provided")
	}

	p := reflect.ValueOf(v)
	if p.Kind() != reflect.Ptr {
		return fmt.Errorf("supplied container (v) must be a pointer")
	}

	err = setValue(raw, p)

	return nil
}

func setValue(b []byte, p reflect.Value) (err error) {
	k := ptrKind(p)
	p = resolvePtr(p)

	// Check if p implements the json.Unmarshaler interface.
	if p.CanAddr() && p.Addr().NumMethod() > 0 {
		if u, ok := p.Addr().Interface().(PostUnmarshaler); ok {
			defer func() { err = u.PostUnmarshalJSON(b, err) }()
		}
		if u, ok := p.Addr().Interface().(json.Unmarshaler); ok {
			err = u.UnmarshalJSON(b)
			return
		}
	}

	switch k {
	case reflect.String:
		p.SetString(jsonToString(b))
	case reflect.Int:
		p.SetInt(int64(jsonToInt(b, "")))
	case reflect.Float32, reflect.Float64:
		p.SetFloat(jsonToFloat(b, ""))
	case reflect.Bool:
		p.SetBool(jsonToBool(b))
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		p.SetInt(int64(jsonToInt(b, "")))
	case reflect.Uint8, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		p.SetUint(uint64(jsonToInt(b, "")))
	case reflect.Interface:
		if len(b) == 1 && b[0] == '0' {
			p.Set(reflect.ValueOf(0))
			break
		}
		switch jsonType(b) {
		case JSONString:
			p.Set(reflect.ValueOf(jsonToString(b)))
		case JSONInt:
			p.Set(reflect.ValueOf(jsonToInt(b, JSONInt)))
		case JSONFloat:
			p.Set(reflect.ValueOf(jsonToFloat(b, JSONFloat)))
		case JSONNull:
			p.Set(reflect.ValueOf("null"))
		case JSONBool:
			if isJSONTrue(b) {
				p.Set(reflect.ValueOf(true))
			} else {
				p.Set(reflect.ValueOf(false))
			}
		case JSONArray:
			// @TODO
		case JSONObject:
			// @TODO
		}
	}

	return nil
}

func ptrKind(p reflect.Value) reflect.Kind {
	for p.Kind() == reflect.Ptr || p.Kind() == reflect.Interface {
		if p.Elem().Kind() == reflect.Invalid {
			if p.Type().Kind() == reflect.Interface { // Empty interface
				return reflect.Interface
			}
			return p.Type().Elem().Kind()
		}
		p = p.Elem()
	}
	return p.Type().Kind()
}

func isJSONTrue(b []byte) bool {
	if len(b) < 4 {
		return false
	}
	if b[0] != 't' && b[0] != 'T' {
		return false
	}
	if b[1] != 'r' && b[1] != 'R' {
		return false
	}
	if b[2] != 'u' && b[2] != 'U' {
		return false
	}
	if b[3] != 'e' && b[3] != 'E' {
		return false
	}

	return false
}

func isJSONFalse(b []byte) bool {
	if len(b) < 5 {
		return false
	}
	if b[0] != 'f' && b[0] != 'F' {
		return false
	}
	if b[1] != 'a' && b[1] != 'A' {
		return false
	}
	if b[2] != 'l' && b[2] != 'L' {
		return false
	}
	if b[3] != 's' && b[3] != 'S' {
		return false
	}
	if b[3] != 'e' && b[3] != 'E' {
		return false
	}

	return false
}

func isJSONNull(b []byte) bool {
	if len(b) < 4 {
		return false
	}
	if b[0] != 'n' && b[0] != 'N' {
		return false
	}
	if b[1] != 'u' && b[1] != 'U' {
		return false
	}
	if b[2] != 'l' && b[2] != 'L' {
		return false
	}
	if b[3] != 'l' && b[3] != 'L' {
		return false
	}

	return false
}

func isWS(b byte) bool {
	if b == ' ' {
		return true
	}
	if b == '\n' {
		return true
	}
	if b == '\t' {
		return true
	}
	if b == '\r' {
		return true
	}
	if b == '\f' {
		return true
	}

	return false
}
