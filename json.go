package gojson

import (
	"bytes"
)

const (
	// JSONNull denotes JSON value 'null'
	JSONNull = "null"

	// JSONBool denotes JSON values 'true' or 'false'
	JSONBool = "bool"

	// JSONInt denotes a JSON integer
	JSONInt = "int"

	// JSONFloat denotes a JSON floating point value
	JSONFloat = "float"

	// JSONString denotes a JSON string
	JSONString = "string"

	// JSONArray denotes a JSON array
	JSONArray = "array"

	// JSONObject denotes a JSON object
	JSONObject = "object"

	// JSONInvalid denotes a value that is not valid JSON
	JSONInvalid = ""
)

// IsJSON performs validation on a JSON string.
//
// JSON = null
//     or true or false
//     or JSONNumber
//     or JSONString
//     or JSONObject
//     or JSONArray
func IsJSON(b []byte) bool {
	b = trim(b)
	if len(b) <= 0 {
		return false
	}

	return IsJSONTrue(b) || IsJSONFalse(b) || IsJSONNull(b) ||
		IsJSONNumber(b) || IsJSONString(b) ||
		IsJSONObject(b) || IsJSONArray(b)
}

// IsJSONNull returns true if the byte array is a JSON null value.
func IsJSONNull(b []byte) bool {
	b = trim(b)
	if len(b) != 4 {
		return false
	}

	return (b[0] == 'n' || b[0] == 'N') &&
		(b[1] == 'u' || b[1] == 'U') &&
		(b[2] == 'l' || b[2] == 'L') &&
		(b[3] == 'l' || b[3] == 'L')
}

// IsJSONTrue returns true if the byte array is a JSON true value.
func IsJSONTrue(b []byte) bool {
	b = trim(b)
	if len(b) != 4 {
		return false
	}

	return (b[0] == 't' || b[0] == 'T') &&
		(b[1] == 'r' || b[1] == 'R') &&
		(b[2] == 'u' || b[2] == 'U') &&
		(b[3] == 'e' || b[3] == 'E')
}

// IsJSONFalse returns true if the byte array is a JSON false value.
func IsJSONFalse(b []byte) bool {
	b = trim(b)
	if len(b) != 5 {
		return false
	}

	return (b[0] == 'f' || b[0] == 'F') &&
		(b[1] == 'a' || b[1] == 'A') &&
		(b[2] == 'l' || b[2] == 'L') &&
		(b[3] == 's' || b[3] == 'S') &&
		(b[4] == 'e' || b[4] == 'E')
}

// IsJSONNumber validates a string as a JSON Number.
//
// JSONNumber = - PositiveNumber
//           or PositiveNumber
// PositiveNumber = DecimalNumber
//               or DecimalNumber . Digits
//               or DecimalNumber . Digits ExponentPart
//               or DecimalNumber ExponentPart
// DecimalNumber = 0
//              or OneToNine Digits
// ExponentPart = e Exponent
//             or E Exponent
// Exponent = Digits
//         or + Digits
//         or - Digits
// Digits = Digit
//       or Digits Digit
// Digit = 0 through 9
// OneToNine = 1 through 9
func IsJSONNumber(b []byte) bool {
	b = trim(b)
	if len(b) == 0 {
		return false
	}

	if b[0] == '-' {
		return isPositiveNumber(b[1:])
	}

	return isPositiveNumber(b[:])
}

// PositiveNumber = DecimalNumber
//               or DecimalNumber . Digits
//               or DecimalNumber . Digits ExponentPart
//               or DecimalNumber ExponentPart
func isPositiveNumber(b []byte) bool {
	b = trim(b)
	l := len(b)

	if l == 0 {
		return false
	}

	periods := bytes.Count(b, period)
	if periods > 1 {
		return false
	}

	idx := 0

	E := bytes.Count(b, exponentE)
	if E > 1 {
		return false
	}
	if E == 1 {
		idx = bytes.Index(b, exponentE)
	}

	exponents := bytes.Count(b, exponent)
	if exponents > 1 {
		return false
	}
	if exponents == 1 {
		idx = bytes.Index(b, exponent)
	}

	exponents += E

	switch true {
	case periods == 1 && exponents == 0: // DecimalNumber . Digits
		per := bytes.Index(b, period)
		return isDecimalNumber(b[:per]) && isExponent(b[per+1:])
	case periods == 0 && exponents == 1: // DecimalNumber ExponentPart
		return isDecimalNumber(b[:idx]) && isExponent(b[idx+1:])
	case periods == 1 && exponents == 1: // DecimalNumber . Digits ExponentPart
		per := bytes.Index(b, period)
		return isDecimalNumber(b[:per]) && isDigits(b[per+1:idx]) && isExponent(b[idx+1:])
	default: // DecimalNumber
		return isDecimalNumber(b)
	}
}

// DecimalNumber = 0
//              or OneToNine Digits
func isDecimalNumber(b []byte) bool {
	b = trim(b)
	l := len(b)

	if l == 0 {
		return false
	}

	if l > 1 && b[0] == '0' {
		return false
	}

	return isDigits(b)
}

// Exponent = Digits
//         or + Digits
//         or - Digits
func isExponent(b []byte) bool {
	b = trim(b)
	l := len(b)

	if l == 0 {
		return false
	}

	if l == 1 {
		return isDigit(b[0])
	}

	if b[0] == '+' || b[0] == '-' {
		return isDigits(b[1:])
	}

	return isDigits(b)
}

// Digits = Digit
//       or Digits Digit
func isDigits(b []byte) bool {
	b = trim(b)
	l := len(b)

	if l == 0 {
		return false
	}

	if l == 1 {
		return isDigit(b[0])
	}

	for i := 0; i < l; i++ {
		if !isDigit(b[i]) {
			return false
		}
	}

	return true
}

// Digit = 0 through 9
func isDigit(b byte) bool {
	return b == '0' || b == '1' || b == '2' || b == '3' || b == '4' || b == '5' || b == '6' || b == '7' || b == '8' || b == '9'
}

// OneToNine = 1 through 9
func isOneToNine(b byte) bool {
	return b == '1' || b == '2' || b == '3' || b == '4' || b == '5' || b == '6' || b == '7' || b == '8' || b == '9'
}

// IsJSONString validates a string as a JSON String.
//
// JSONString = ""
//           or " StringCharacters "
func IsJSONString(b []byte) bool {
	b = trim(b)
	if len(b) == 0 {
		return false
	}

	if len(b) == 2 && b[0] == '"' && b[1] == '"' {
		return true
	}

	// No quotes after trim
	if b[0] != '"' || b[len(b)-1] != '"' {
		return false
	}

	if len(b) <= 1 {
		return isStringCharacters(b)
	}

	return isStringCharacters(b[1 : len(b)-1])
}

// StringCharacters = StringCharacter
//                 or StringCharacters StringCharacter
func isStringCharacters(b []byte) bool {
	if len(b) == 0 {
		return false
	}

	s := 0
	for s < len(b) {
		// Escape Sequence
		if b[s] == '\\' {
			if s+6 <= len(b) && isEscapeSequence(b[s:s+6]) {
				s += 6
				continue
			}

			if s+2 <= len(b) && isEscapeSequence(b[s:s+2]) {
				s += 2
				continue
			}

			return false
		}

		if !isStringCharacter(b[s : s+1]) {
			return false
		}

		s++
	}

	return true
}

// StringCharacter = any character
//                   except " or \ or U+0000 through U+001F
//                or EscapeSequence
func isStringCharacter(b []byte) bool {
	if len(b) == 1 {
		return (b[0] != '"' && b[0] != '\\' && !(b[0] >= '\u0000' && b[0] <= '\u001F'))
	}

	return isEscapeSequence(b)
}

// EscapeSequence = \" or \/ or \\ or \b or \f or \n or \r or \t
//               or \u HexDigit HexDigit HexDigit HexDigit
func isEscapeSequence(b []byte) bool {
	b = trim(b)
	if len(b) == 0 {
		return false
	}

	if len(b) == 6 && b[0] == '\\' && b[1] == 'u' {
		return isHexDigit(b[2]) && isHexDigit(b[3]) && isHexDigit(b[4]) && isHexDigit(b[5])
	}

	if len(b) == 2 && b[0] == '\\' {
		return b[1] == '"' || b[1] == '/' || b[1] == '\\' ||
			b[1] == 'b' || b[1] == 'f' || b[1] == 'n' || b[1] == 'r' || b[1] == 't' ||
			b[1] == 'B' || b[1] == 'F' || b[1] == 'N' || b[1] == 'R' || b[1] == 'T'
	}

	return false
}

// HexDigit = 0 through 9
//         or A through F
//         or a through f
func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'A' && b <= 'F') || (b >= 'a' && b <= 'f')
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\n' || b == '\t' || b == '\r' || b == '\f'
}

// TermBytes are the characters the signify the end of a JSON value.
//
// TermByte = ','
//         or ']'
//         or '}'
func isTermByte(b byte) bool {
	return b == ',' || b == ']' || b == '}'
}

// IsJSONObject validates a string as a JSON Object.
//
// JSONObject = { }
//           or { Members }
func IsJSONObject(b []byte) bool {
	b = trim(b)
	if len(b) == 0 {
		return false
	}

	if IsEmptyObject(b) {
		return true
	}

	// No quotes after trim
	if b[0] != '{' || b[len(b)-1] != '}' {
		return false
	}

	return isMembers(b[1 : len(b)-1])
}

// Members = JSONString : JSON
//        or Members , JSONString : JSON
func isMembers(b []byte) bool {
	b = trim(b)
	if len(b) == 0 {
		return false
	}

	start := 0
	requiresValue := false
	for start < len(b) {
		_, pos, err := extractKey(b, start)
		if err != nil {
			return false
		}

		v, _, pos, err := extractValue(b, pos)
		if err != nil || !IsJSON(v) {
			return false
		}

		if pos == len(b) {
			return true
		}

		requiresValue = false
		start = ltrim(b, pos)

		// Post value must be a term byte, or the JSON is malformed.
		if start < len(b) && !isTermByte(b[start]) {
			return false
		}

		// If we see a comma, we require at least one more value to extract.
		if start < len(b) && b[start] == ',' {
			requiresValue = true
			start++
			continue
		}
	}

	return !requiresValue
}

// IsJSONArray validates a string as a JSON Array.
//
// JSONArray = [ ]
//          or [ ArrayElements ]
func IsJSONArray(b []byte) bool {
	b = trim(b)
	if len(b) == 0 {
		return false
	}

	if IsEmptyArray(b) {
		return true
	}

	// No quotes after trim
	if b[0] != '[' || b[len(b)-1] != ']' {
		return false
	}

	return isArrayElements(b[1 : len(b)-1])
}

// ArrayElements = JSON
//              or ArrayElements , JSON
func isArrayElements(b []byte) bool {
	b = trim(b)
	if len(b) == 0 {
		return false
	}

	start := 0
	requiresValue := false
	for start < len(b) {
		v, _, pos, err := extractValue(b, start)
		if err != nil || !IsJSON(v) {
			return false
		}

		if pos == len(b) {
			return true
		}

		requiresValue = false
		start = ltrim(b, pos)

		// Post value must be a term byte, or the JSON is malformed.
		if start < len(b) && !isTermByte(b[start]) {
			return false
		}

		// If we see a comma, we require at least one more value to extract.
		if start < len(b) && b[start] == ',' {
			requiresValue = true
			start++
		}
	}

	return !requiresValue
}

// Scan through from a starting position, and return the next non-Whitespace position.
func ltrim(search []byte, start int) int {
	if start < 0 {
		return start
	}

	for start < len(search) {
		switch {
		case isWhitespace(search[start]):
			start++
		default:
			return start
		}
	}

	return start
}

// Remove all leading and trailing whitespace.
func trim(search []byte) []byte {
	if len(search) <= 0 {
		return search
	}

	start := 0
	end := len(search)
	for start <= len(search)-1 && isWhitespace(search[start]) {
		start++
	}

	for end > 0 && isWhitespace(search[end-1]) {
		end--
	}

	if end == 0 {
		return []byte{}
	}

	return search[start:end]
}

// isBoolChar returns true if the byte could be the start of a
// JSONTrue or JSONFalse
func isBoolChar(b byte) bool {
	return b == 't' || b == 'T' || b == 'f' || b == 'F'
}

// isBoolChar returns true if the byte could be the start of a JSONNull
func isNullChar(b byte) bool {
	return b == 'n' || b == 'N'
}

// Move the pointer past the TermByte after a value extraction.
func findTerminator(search []byte, start int) int {
	if start < 0 {
		return start
	}

	start = ltrim(search, start)
	if start < len(search) {
		if isTermByte(search[start]) {
			return start + 1
		}

		return -1
	}

	return start
}

// Remove open and closing quotes from a JSONString
func trimString(b []byte) []byte {
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}

	return b
}

// GetJSONType inspects the given byte string and attempts to identify the primary JSON type.
// JSONString, JSONInt, JSONFloat, JSONBool, JSONNull, JSONObject, or JSONArray
// JSONInvalid ("") denotes a JSON error.
//
// Note that this is an early return guess. It's not guaranteed accurate if you send invalid JSON,
// as very little validation is performed for the sake of speed.
//
// e.g. GetJSONType([]byte(`"No End Quote`), 0) would return JSONString, NOT JSONInvalid,
// despite the fact that the given data is invalid JSON.
//
// If you must validate that you have valid JSON, call IsJSON with your byte string prior
// to calling GetJSONType. Or call GetJSONTypeStrict.
func GetJSONType(search []byte, start int) string {
	current := ltrim(search, start)

	switch {
	case current < 0 || len(search) < 1 || len(search) <= current:
		return JSONInvalid
	case search[current] == '{': // Objects
		return JSONObject
	case search[current] == '[': // Arrays
		return JSONArray
	case search[current] == '"':
		return JSONString
	case isDigit(search[current]) || search[current] == '-':
		_, t, _, err := extractNumber(search, start)
		if err != nil {
			return JSONInvalid
		}

		return t
	case isBoolChar(search[current]):
		if len(search) >= 4 && IsJSONTrue(search[current:current+4]) {
			return JSONBool
		}
		if len(search) >= 5 && IsJSONFalse(search[current:current+5]) {
			return JSONBool
		}
		return JSONInvalid
	case isNullChar(search[current]):
		if len(search) >= 4 && IsJSONNull(search[current:current+4]) {
			return JSONNull
		}
		return JSONInvalid

	default:
		return JSONInvalid
	}
}

// GetJSONTypeStrict validates the given byte string as JSON and returns the primary JSON type.
// JSONString, JSONInt, JSONFloat, JSONBool, JSONNull, JSONObject, or JSONArray
// JSONInvalid ("") denotes a JSON error.
//
// GetJSONTypeStrict WILL perform JSON Validation, and return JSONInvalid if that validation fails.
// This is slower than GetJSONType due to the extra validation involved.
func GetJSONTypeStrict(search []byte, start int) string {
	current := ltrim(search, start)

	switch {
	case current < 0 || len(search) < 1 || len(search) <= current:
		return JSONInvalid
	case !IsJSON(search[current:]):
		return JSONInvalid
	default:
		return GetJSONType(search, current)
	}
}

// IsEmptyArray returns true if the given JSON is an empty array.
func IsEmptyArray(b []byte) bool {
	return isEmptyComplexItem(b, '[', ']')
}

// IsEmptyObject returns true if the given JSON is an empty object.
func IsEmptyObject(b []byte) bool {
	return isEmptyComplexItem(b, '{', '}')
}

// Ignoring whitespace, attempts to determine whether the given byte string represents
// and empty JSONObject ({}) or empty JSONArray ([])
func isEmptyComplexItem(b []byte, open, close byte) bool {
	b = trim(b)
	current := 0
	if len(b) < 2 {
		return false
	}

	if b[current] != open {
		return false
	}

	current = ltrim(b, current+1)
	if b[current] != close || len(b[current:]) > 1 {
		return false
	}

	return true
}
