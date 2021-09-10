package gojson

import (
	"encoding/json"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/davecgh/go-spew/spew"
)

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

	return true
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
	if b[4] != 'e' && b[4] != 'E' {
		return false
	}

	return true
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

	return true
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

// Find a number:
// 1) Trim all leading whitespace.
// 2) If it's a quoted string, ignore the opening quote.
// 3) Read until you find an invalid number byte or a terminal character, and return the found type.
func findNumber(raw []byte) ([]byte, string) {
	a := 0
	// Here we're trimming whitespace and finding an opening quote if it exists.
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

	raw = raw[a:]

	if len(raw) == 0 {
		return nil, JSONInvalid
	}

	if len(raw) == 1 && raw[0] == '0' {
		return raw, JSONInt
	}

	// It's not a valid number unless it begins with minus, or 1-9
	end := 0
	switch raw[0] {
	case '-', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		end++
	default:
		return nil, JSONInvalid
	}

	e := false
	eSign := false
	period := false
SEARCH:
	for i := end; i < len(raw); i++ {
		b := raw[i]
		switch b {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			end++
			continue
		case '"', ',', ']', '}':
			break SEARCH
		case '-', '+':
			if !e || eSign {
				return nil, JSONInvalid
			}
			eSign = true
			end++
		case 'e', 'E':
			if e {
				return nil, JSONInvalid
			}
			e = true
			end++
		case '.':
			if period {
				return nil, JSONInvalid
			}
			period = true
			end++
		default:
			return nil, JSONInvalid
		}
	}

	if period || e {
		return raw[:end], JSONFloat
	}

	return raw[:end], JSONInt
}

// findString trims leading and trailing whitepsace, and removes the leading
// and trailing double quote if it exists. Non-validating.
func findString(raw []byte) ([]byte, bool) {
	a := 0
	b := len(raw)

	open := false
	close := false

	for i := 0; i < len(raw); i++ {
		if isWS(raw[i]) {
			a++
			continue
		}

		if raw[i] == '"' {
			a++
			open = true
			break
		}

		break
	}

	raw = raw[a:]

	for i := 0; i < len(raw); i++ {
		if raw[i] == '\\' {
			if i >= len(raw)-1 {
				continue
			}

			if raw[i+1] == '"' {
				i++ // consume the escaped quote
			}
		}

		if raw[i] == '"' {
			b = i
			close = true
			break
		}
	}

	return raw[:b], open && close
}

func jsonToInt(b []byte, t string) int {
	if isJSONTrue(b) {
		return 1
	}

	if t == "" {
		b, t = findNumber(b)
	}

	if t == JSONInvalid {
		return 0
	}

	if t == JSONInt {
		i, err := strconv.ParseInt(*(*string)(unsafe.Pointer(&b)), 10, 64)
		if err == nil {
			return int(i)
		}
	}

	f, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&b)), 64)
	if err != nil {
		return 0
	}
	return int(f)
}

func jsonToFloat(b []byte, t string) float64 {
	if isJSONTrue(b) {
		return 1.0
	}

	if t == "" {
		b, t = findNumber(b)
	}

	if t == JSONInvalid {
		return 0
	}

	if t == JSONInt {
		i, err := strconv.ParseInt(*(*string)(unsafe.Pointer(&b)), 10, 64)
		if err == nil {
			return float64(i)
		}
	}

	f, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&b)), 64)
	if err != nil {
		return 0
	}
	return f
}

func jsonToBool(b []byte) bool {
	b, _ = findString(b)

	if isJSONTrue(b) {
		return true
	}

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
func jsonToString(raw []byte) string {
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

	final := out[:end]
	return *(*string)(unsafe.Pointer(&final))
}

// infer the json type from the first non-whitespace character.
func jsonType(raw []byte) string {
	if len(raw) == 0 {
		return JSONInvalid
	}

	a := 0

	for i := 0; i < len(raw); i++ {
		if isWS(raw[i]) {
			a++
			continue
		}

		break
	}

	if len(raw) == 0 {
		return JSONInvalid
	}

	if len(raw[a:]) == 1 && raw[a] == '0' {
		return JSONInt
	}

	switch raw[a] {
	case '{':
		return JSONObject
	case '[':
		return JSONArray
	case '"':
		return JSONString
	case '-', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		_, t := findNumber(raw)
		return t
	case 't', 'T':
		if isJSONTrue(raw[a:]) {
			return JSONBool
		}

		return JSONInvalid
	case 'f', 'F':
		if isJSONFalse(raw[a:]) {
			return JSONBool
		}

		return JSONInvalid
	case 'n', 'N':
		if isJSONNull(raw[a:]) {
			return JSONNull
		}

		return JSONInvalid
	}

	return JSONInvalid
}

// Find the position after the first comma or end of line.
func firstTerminator(raw []byte) int {
	for i := 0; i < len(raw); i++ {
		if raw[i] == ',' {
			return i + 1
		}
	}

	return len(raw) - 1
}

func firstNonWSByte(raw []byte) int {
	for i := 0; i < len(raw); i++ {
		if isWS(raw[i]) {
			continue
		}

		return i
	}

	return 0
}

func trimWS(raw []byte, unquote bool) (int, int) {
	a := 0
	b := len(raw) - 1

	for i := 0; i < len(raw); i++ {
		if isWS(raw[i]) {
			a++
			continue
		}

		if unquote && raw[i] == '"' {
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

		if unquote && raw[i] == '"' {
			b--
			break
		}

		break
	}

	return a, b
}

func unmarshalSlice(b []byte, p reflect.Value) (err error) {
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

	t := jsonType(b)
	childType := p.Type().Elem().Kind()

	// ByteSlices are exceptionally hard to extract byte-by-byte given the difficulty
	// of finding the correct position in the RawData, so we circumvent that problem by
	// short-circuiting and treating it as if the whole array were an elemental type.
	if childType == reflect.Uint8 {
		if t == JSONString {
			b, _ = findString(b)
		}
		p.Set(reflect.ValueOf(b))
		return nil
	}

	// Count the member elements so that we can know how big to size our slice.
	length := countMembers(b, t)

	if length < 1 {
		return nil
	}

	spew.Dump(length)

	return nil
}
