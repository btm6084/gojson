package gojson

import (
	"strconv"
	"unsafe"
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
// 3) Read until you find an invalid number byte, and return the type of number found.
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
	b = findString(b)

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

// func jsonToIface(raw []byte) interface{} {
// 	if len(raw) == 0 {
// 		return nil
// 	}

// 	a := 0

// 	for i := 0; i < len(raw); i++ {
// 		if isWS(raw[i]) {
// 			a++
// 			continue
// 		}

// 		break
// 	}

// 	if len(raw) == 0 {
// 		return nil
// 	}

// 	switch raw[a] {
// 	case '{':
// 		// @todo
// 	case '[':
// 		// @todo
// 	case '"':
// 		a := jsonToString(raw)
// 		return a
// 		// return jsonToString(raw)
// 	case '-', '1', '2', '3', '4', '5', '6', '7', '8', '9':
// 		b, t := findNumber(raw)
// 		if t == JSONInt {
// 			return jsonToInt(b, t)
// 		}

// 		if t == JSONFloat {
// 			return jsonToFloat(b, t)
// 		}

// 		return jsonToString(raw)

// 	case '0':
// 		return 0
// 	case 't', 'T':
// 		if isJSONTrue(raw) {
// 			return true
// 		}
// 	case 'f', 'F':
// 		if isJSONFalse(raw) {
// 			return false
// 		}
// 	case 'n', 'N':
// 		if isJSONNull(raw) {
// 			return false
// 		}
// 	}

// 	return jsonToString(raw)
// }

func countSliceMembers2(raw []byte) int {
	a, b := trimWS(raw, true)

	_ = b

	if len(raw[a:]) < 3 {
		// 3 Bytes: Open Bracket, Close Bracket, and 1 member.
		return 0
	}

	if raw[a] != '[' || raw[b] != ']' {
		// Not a slice.
		return 0
	}

	a++ // Consume open bracket.
	b-- // Consume close bracket.

	raw = raw[a : b+1]
	count := 0

MEMBERS:
	for {
		raw = raw[firstNonWSByte(raw):]
		switch raw[0] {
		case '-', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			term := firstTerminator(raw)
			count++
			if term == len(raw)-1 {
				break MEMBERS
			}

			raw = raw[term+1:]
		default:
			break MEMBERS
		}
	}

	return count
}

func firstTerminator(raw []byte) int {
	for i := 0; i < len(raw); i++ {
		if raw[i] == ',' {
			return i
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
