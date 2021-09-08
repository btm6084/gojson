package gojson

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"unsafe"
)

const (
	stateBegin uint = iota
	stateObject
	stateArray
	stateString
	stateNumber
	stateTrue
	stateFalse
	stateNull

	typeJSONObject uint8 = iota
	typeJSONArray
	typeJSONString
	typeJSONNumber
	typeJSONTrue
	typeJSONFalse
	typeJSONNull
	typeJSONErr
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
		p.SetString(toJSONString(b))
	case reflect.Int:
		p.SetInt(int64(jsonToInt(b)))
	case reflect.Float32, reflect.Float64:
		p.SetFloat(jsonToFloat(b))
	case reflect.Bool:
		p.SetBool(jsonToBool(b))
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		p.SetInt(int64(jsonToInt(b)))
	case reflect.Uint8, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		p.SetUint(uint64(jsonToInt(b)))
	}

	return nil
}

type jsonTree struct {
	begin int
	end   int
	dtype uint8
}

func jsonType(raw []byte) uint8 {
	for i := 0; i < len(raw); i++ {
		b := raw[i]
		switch {
		case b == '{':
			return typeJSONObject
		case b == '[':
			return typeJSONArray
		case b == '"':
			return typeJSONString
		case isJSONTrue(raw):
			return typeJSONTrue
		case isJSONFalse(raw):
			return typeJSONFalse
		case isJSONNull(raw):
			return typeJSONNull
		case !isWS(b):
			return typeJSONErr
		}
	}

	return typeJSONErr
}

func tokenize(raw []byte) (*jsonTree, error) {
	state := stateBegin
	for i := 0; i < len(raw); i++ {
		b := raw[i]
		switch state {
		case stateBegin:
			switch {
			case b == '{':
				state = stateObject
			case b == '[':
				state = stateArray
			case b == '"':
				state = stateString
			case isJSONTrue(raw):
				return &jsonTree{begin: i, end: i + 4, dtype: typeJSONTrue}, nil
			case isJSONFalse(raw):
				return &jsonTree{begin: i, end: i + 5, dtype: typeJSONFalse}, nil
			case isJSONNull(raw):
				return &jsonTree{begin: i, end: i + 4, dtype: typeJSONNull}, nil
			case !isWS(b):
				return nil, ErrMalformedJSON
			}
		case stateObject:
			return &jsonTree{begin: i, end: i + 4, dtype: typeJSONObject}, nil
		case stateArray:
			return &jsonTree{begin: i, end: i + 4, dtype: typeJSONArray}, nil
		case stateString:
			return &jsonTree{begin: i, end: i + 4, dtype: typeJSONString}, nil
		case stateNumber:
			return &jsonTree{begin: i, end: i + 4, dtype: typeJSONNumber}, nil
		}
	}

	return nil, nil
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

func findString(raw []byte) []byte {
	a := 0
	b := len(raw)

	for i := 0; i < len(raw); i++ {
		if isWS(raw[i]) {
			a++
			continue
		}

		if raw[i] == '"' {
			a++
			break
		}

		break
	}

	for i := len(raw) - 1; i >= 0; i-- {
		if isWS(raw[i]) {
			b--
			continue
		}

		if raw[i] == '"' {
			b--
			break
		}

		break
	}

	return raw[a:b]
}

func jsonToInt(b []byte) int {
	if isJSONTrue(b) {
		return 1
	}

	b = findString(b)

	i, err := strconv.ParseInt(*(*string)(unsafe.Pointer(&b)), 10, 64)
	if err != nil {
		log.Println(err)
		return 0
	}
	return int(i)
}

func jsonToFloat(b []byte) float64 {
	if isJSONTrue(b) {
		return 1.0
	}

	b = findString(b)

	i, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&b)), 64)
	if err != nil {
		log.Println(err)
		return 0
	}
	return i
}

func jsonToBool(b []byte) bool {
	if isJSONTrue(b) {
		return true
	}

	b = findString(b)

	if len(b) == 1 && b[0] == '0' {
		return false
	}

	out, err := strconv.ParseBool(*(*string)(unsafe.Pointer(&b)))
	if err != nil {
		return false
	}
	return out
}

// Turn a quoted string into a non-quoted string, and fix any escape sequences.
func toJSONString(raw []byte) string {
	start := 0

	if len(raw) < 2 {
		return string(raw)
	}

	// find the first non-whitespace character
	quotedString := false
	for i := 0; i < len(raw); i++ {
		b := raw[i]

		if isWS(b) {
			start++
			continue
		}

		if b == '"' {
			start++
			quotedString = true
		}

		break
	}

	raw = raw[start:]

	out := make([]byte, len(raw))

	end := 0
	for i := 0; i < len(raw); i++ {
		b := raw[i]

		if b != '\\' {
			if quotedString && b == '"' {
				break
			}
			out[end] = b
			end++
			continue
		}

		// We're done if i is the last character
		if i == len(raw)-1 {
			break
		}

		c := raw[i+1]

		if c == 'u' {
			if i+5 >= len(raw) {
				out[end] = raw[i]
				end++
				continue
			}

			piece := raw[i+2 : i+6]
			r, err := strconv.ParseInt(*(*string)(unsafe.Pointer(&piece)), 16, 32)
			if err != nil {
				out[end] = raw[i]
				end++
				continue
			}

			length := 6

			// https://unicodebook.readthedocs.io/unicode_encodings.html#utf-16-surrogate-pairs
			// Unicode Surrogate Pair hex {D800-DBFF},{DC00-DFFF} dec {55296-56319},{56320-57343}
			if i+11 < len(raw) && (r >= 55296 && r <= 56319) {
				piece := raw[i+8 : i+12]
				r2, err := strconv.ParseInt(*(*string)(unsafe.Pointer(&piece)), 16, 32)
				if err != nil {
					break
				}

				if r2 >= 56320 && r2 <= 57343 {
					length = 12
					r = ((r - 0xD800) * 0x400) + (r2 - 0xDC00) + 0x10000
				}
			}

			for _, bn := range []byte(string(rune(r))) {
				out[end] = bn
				end++
			}
			i += length - 1 // -1 to account for the incoming i++ following the continue
			continue
		}

		switch c {
		case '"':
			out[end] = '"'
		case 'n':
			out[end] = '\n'
		case 't':
			out[end] = '\t'
		case '\\':
			out[end] = '\\'
		case '/':
			out[end] = '/'
		case 'r':
			out[end] = '\r'
		case 'b':
			out[end] = '\b'
		case 'f':
			out[end] = '\f'
		}

		end++
		i++
	}

	return string(out[:end])
}
