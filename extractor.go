package gojson

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

// Extract a specific key from a given JSON string. Returns value, type, and error.
//
// Key Path Values:
// String, Number, and Constant not in a surrounding array or object will always
//     return the root, regardless of the key given.
// Arrays will index from 0.
// Objects will will index as they are keyed.
// Empty key will always return whatever is at the root.
// Path is a period separated list of keys.
//
// Examples:
// data := []byte(`{
// 	"active": 0,
// 	"url": "http://www.example.com",
// 	"metadata": {
// 		"keywords": [
// 			"example",
// 			"sample",
// 			"illustration",
// 		],
// 		"error_code": 0
// 	}
// }`)
// Extract(data, "active") returns 0, "int", nil
// Extract(data, "url") returns "http://www.example.com", "string", nil
// Extract(data, "metadata.keywords.1") returns []byte(`"sample"`), "string", nil
// Extract(data, "metadata.error_code") returns []byte{'0'}, "int", nil
// Extract(data, "metadata.keywords.17") returns []byte(nil), "", "requested key 'metadata.keywords.17' doesn't exist
//
// On return, a copy is made of the extracted data. This allows it to be modified without changing the original JSON.
func Extract(search []byte, path string) ([]byte, string, error) {
	if len(search) == 0 {
		return nil, "", ErrEmpty
	}

	// If the key is empty, return the root.
	if path == "" {
		b, t, _, err := extractValue(search, 0)
		var retVal []byte
		if b != nil {
			retVal = make([]byte, len(b))
			copy(retVal, b)
		}
		return retVal, t, err
	}

	// Find the primary JSON type
	switch t := GetJSONType(search, 0); t {
	case JSONString, JSONFloat, JSONInt, JSONBool, JSONNull:
		if path != "" {
			return nil, "", fmt.Errorf("key path provided '%s' is invalid for JSON type '%s'", path, t)
		}

		b, t, _, err := extractValue(search, 0)
		var retVal []byte
		if b != nil {
			retVal = make([]byte, len(b))
			copy(retVal, b)
		}
		return retVal, t, err
	case JSONObject, JSONArray:
		b, t, _, err := extractKeyPath(search, path)
		var retVal []byte
		if b != nil {
			retVal = make([]byte, len(b))
			copy(retVal, b)
		}
		return retVal, t, err
	}

	return nil, "", fmt.Errorf("requested key path '%s' doesn't exist or json is malformed", path)
}

// ExtractReader performs an Extract on the given JSON path. The resulting value
// is returned in the form of a JSONReader primed with the value returned from Extract.
func ExtractReader(search []byte, path string) (*JSONReader, error) {
	b, _, err := Extract(search, path)
	if err != nil {
		return nil, err
	}

	return NewJSONReader(b)
}

// ExtractString performs an Extract on the given JSON path. The resulting value
// is returned in the form of a string.
func ExtractString(search []byte, path string) (string, error) {
	b, t, err := Extract(search, path)
	if err != nil {
		return "", err
	}
	return toString(b, t, false), nil
}

// ExtractInt performs an Extract on the given JSON path. The resulting value
// is returned in the form of an int.
func ExtractInt(search []byte, path string) (int, error) {
	b, t, err := Extract(search, path)
	if err != nil {
		return 0, err
	}
	return toInt(b, t, false), nil
}

// ExtractFloat performs an Extract on the given JSON path. The resulting value
// is returned in the form of a float64.
func ExtractFloat(search []byte, path string) (float64, error) {
	b, t, err := Extract(search, path)
	if err != nil {
		return 0, err
	}
	return toFloat(b, t, false), nil
}

// ExtractBool performs an Extract on the given JSON path. The resulting value
// is returned in the form of a bool. Numeric values are considered true if they
// are non-zero. String values are evaluated using strings.ParseBool. null is always false.
func ExtractBool(search []byte, path string) (bool, error) {
	b, t, err := Extract(search, path)
	if err != nil {
		return false, err
	}
	return toBool(b, t, false), nil
}

// ExtractInterface performs an Extract on the given JSON path. The resulting value
// is returned in the form defined below.
// The returned type is the JSON type. These map as follows:
// json type -> interface{} type
//
// JSONInt -> int
// JSONFloat -> float64
// JSONString -> string
// JSONBool -> bool
// JSONNull -> interface{}(nil)
// JSONArray -> []interface{}
// JSONObject -> map[string]interface{}
func ExtractInterface(search []byte, path string) (interface{}, string, error) {
	b, t, err := Extract(search, path)
	if err != nil {
		return interface{}(nil), "", err
	}

	return toIface(b, t, false), t, nil
}

func extractValue(search []byte, start int) ([]byte, string, int, error) {
	start = ltrim(search, start)

	switch {
	case len(search) < 1, start < 0, start >= len(search):
		return nil, "", 0, ErrMalformedJSON
	case search[start] == '{': // Objects
		return extractObject(search, start)
	case search[start] == '[': // Arrays
		return extractArray(search, start)
	case search[start] == '"':
		return extractString(search, start)
	case isDigit(search[start]) || search[start] == '-':
		return extractNumber(search, start)
	case len(search[start:]) >= 4 && (IsJSONTrue(search[start:start+4]) || IsJSONNull(search[start:start+4])):
		return extractConstant(search, start)
	case len(search[start:]) >= 5 && IsJSONFalse(search[start:start+5]):
		return extractConstant(search, start)
	default:
		return nil, "", 0, fmt.Errorf("invalid character '%s' at position '%d' in segment '%s'", string(search[start]), start, search)
	}
}

func pathToKeys(path string) []string {
	var keys []string
	keyStart := 0
	hasEscape := false
	for i := keyStart; i < len(path); i++ {
		if path[i] == '.' {
			if i == 0 {
				keyStart++
				continue
			}

			if path[i-1] == '\\' {
				hasEscape = true
				continue
			}

			val := path[keyStart:i]
			if hasEscape {
				val = strings.ReplaceAll(val, "\\.", ".")
			}
			keys = append(keys, val)
			keyStart = i + 1
			hasEscape = false
		}

		if i == (len(path) - 1) {
			val := path[keyStart : i+1]
			if hasEscape {
				val = strings.ReplaceAll(val, "\\.", ".")
			}
			keys = append(keys, val)
			hasEscape = false

		}
	}

	return keys
}

// Given a JSON search space and a key path in the form key[,.keyN+1], return the value defined
// at that key.
// If your key contains a period which isn't nested, escape it with a backslash: a.b@example.com => a\.b@example\.com
func extractKeyPath(search []byte, path string) ([]byte, string, int, error) {
	found := false
	start := ltrim(search, 0)
	keys := pathToKeys(path)

	if len(keys) < 1 || len(path) < 1 {
		return nil, "", 0, fmt.Errorf("extractKeyPath: no keys to extract")
	}

	for _, k := range keys {
		start = ltrim(search, start)
		found = false

		switch GetJSONType(search[start:], 0) {
		case JSONObject:
			// Move past opening bracket
			start++

			for start <= len(search)-1 {
				key, pos, err := extractKey(search, start)
				if err != nil {
					return nil, "", 0, err
				}

				if k == *(*string)(unsafe.Pointer(&key)) {
					start = pos
					found = true
					break
				}

				// If this is not our key, move the cursor past the value, so we can process the next key
				_, _, pos, err = extractValue(search, pos)
				if err != nil {
					return nil, "", 0, err
				}

				start = findTerminator(search, pos)
			}
		case JSONArray:
			// Non-numeric keys are invalid
			if !isDecimalNumber([]byte(k)) {
				break
			}

			// Move past opening bracket
			start++

			idx, _ := strconv.Atoi(k)

			// Move past values until we find the right index or encounter an error
			for i := 0; i < idx; i++ {
				_, _, pos, err := extractValue(search, start)
				if err != nil {
					return nil, "", 0, err
				}

				start = findTerminator(search, pos)
			}

			found = true
		}
	}

	if found {
		return extractValue(search, start)
	}

	return nil, "", 0, fmt.Errorf("requested key '%s' doesn't exist", path)
}

// Extract a key from a JSONObject.
func extractKey(search []byte, start int) ([]byte, int, error) {
	// Find the key
	k, _, pos, err := extractString(search, start)
	if err != nil {
		return nil, 0, fmt.Errorf("expected object key at position %d in segment '%s'", start, truncate(search, 50))
	}

	// Advance past the key
	start = pos
	found := false
	for !found && start <= len(search)-1 {
		switch {
		case search[start] == ':':
			start++
			found = true
		case isWhitespace(search[start]):
			start++
		default:
			return nil, 0, fmt.Errorf("invalid character '%s' as position %d (expecting ':' following object key)", string(search[start]), start)
		}
	}

	if len(k) >= 2 && k[0] == '"' && k[len(k)-1] == '"' {
		k = k[1 : len(k)-1]
	}

	return k, start, err
}

// Extract a string from a starting position.
// The returned string does not include leading or trailing double quote.
// Those should be accounted for when moving position afterword, as they won't
// be included in a len() calculation.
// Start is expected to point to the opening double quote.
func extractString(search []byte, start int) ([]byte, string, int, error) {
	start = ltrim(search, start)

	if start >= len(search) {
		return nil, "", 0, fmt.Errorf("expected string not found")
	}

	if search[start] != '"' {
		return nil, "", 0, fmt.Errorf(`invalid character '%s' as position %d (expecting '"' for open string)`, string(search[start]), start)
	}

	start++
	end := start
	for end <= len(search)-1 {
		b := search[end]
		if b == '\\' {
			end += 2
			continue
		}

		if b == '"' {
			return search[start-1 : end+1], JSONString, end + 1, nil
		}

		end++
	}

	return nil, "", 0, fmt.Errorf("expected string not found")
}

// Extract a number from a starting position.
// Start is expected to be pointing at the first byte in the numeric value.
func extractNumber(search []byte, start int) ([]byte, string, int, error) {
	start = ltrim(search, start)
	end := start
	found := false

	// Single Digit Case
	if len(search) == 1 {
		end++
		found = true
	}

	for !found && end <= len(search)-1 {
		if isTermByte(search[end]) {
			break
		}
		end++
	}

	if IsJSONNumber(search[start:end]) {
		return trim(search[start:end]), extractNumberType(search[start:end]), end, nil
	}

	return nil, "", 0, fmt.Errorf("expected number not found")
}

func extractNumberType(number []byte) string {
	if bytes.Count(number, period) == 1 || bytes.Count(number, exponent) == 1 || bytes.Count(number, exponentE) == 1 {
		return JSONFloat
	}

	return JSONInt
}

// Extract a constant from a starting position.
// Start is expected to be pointing at the first byte of the constant.
func extractConstant(search []byte, start int) ([]byte, string, int, error) {
	start = ltrim(search, start)

	switch {
	case search[start] == 't' || search[start] == 'T':
		if IsJSONTrue(search[start : start+4]) {
			return search[start : start+4], JSONBool, start + 4, nil
		}
	case search[start] == 'f' || search[start] == 'F':
		if IsJSONFalse(search[start : start+5]) {
			return search[start : start+5], JSONBool, start + 5, nil
		}
	case search[start] == 'n' || search[start] == 'N':
		if IsJSONNull(search[start : start+4]) {
			return search[start : start+4], JSONNull, start + 4, nil
		}
	}

	return nil, "", 0, fmt.Errorf("expected constant not found")
}

// Extract an object from a starting position.
// The returned object does include leading or trailing curly brackets.
// Those should be accounted for when moving position afterword, as they will
// be included in any len() calculation.
// Start is expected to point to the opening curly brace.
func extractObject(search []byte, start int) ([]byte, string, int, error) {
	return extractComplexItem(search, start, '{', '}', JSONObject)
}

// Extract an array from a starting position.
// The returned array does include leading or trailing brackets.
// Those should be accounted for when moving position afterword, as they will
// be included in any len() calculation.
// Start is expected to point to the opening bracket.
func extractArray(search []byte, start int) ([]byte, string, int, error) {
	return extractComplexItem(search, start, '[', ']', JSONArray)
}

// Extract either an array or object from the provided JSON.
func extractComplexItem(search []byte, start int, open, close byte, dtype string) ([]byte, string, int, error) {
	start = ltrim(search, start)
	end := start
	depth := 0
	var err error

	for end <= len(search)-1 {
		b := search[end]

		switch {
		case b == '\\':
			end += 2
			continue
		case b == '"':
			_, _, end, err = extractString(search, end)
			if err != nil {
				return nil, "", 0, err
			}
			continue
		case b == close: // If Depth is one, we've found the end of our object.
			if depth == 1 {
				return search[start : end+1], dtype, end + 1, nil
			}
			depth--
			end++
			continue
		case b == open: // Sub-Object begin
			depth++
			end++
			continue
		default:
			end++
		}
	}

	return nil, "", 0, fmt.Errorf("expected %s not found in segment '%s'", dtype, truncate(search, 50))
}

// Extract a key/value pair from an object.
// Value, Key, Type, EndPosition, Error
// Start needs to be pointing at the opening quote (") (or whitespace) of the key in order to succeed.
func extractKeyValue(search []byte, start int) ([]byte, string, string, int, error) {
	key, start, err := extractKey(search, start)
	if err != nil {
		return nil, "", "", 0, err
	}

	v, t, start, err := extractValue(search, start)
	if err != nil {
		return nil, "", "", 0, errors.New(err.Error() + " (expected object value)")
	}

	var termErr error
	finalPos := findTerminator(search, start)
	if finalPos < 0 {
		termErr = fmt.Errorf("expected object value terminator ('}', ']' or ',') at position '%d' in segment '%s'", start, truncate(search, 50))
	}

	return v, *(*string)(unsafe.Pointer(&key)), t, finalPos, termErr
}

// Extract a key/value pair from an object, without consuming the terminator.
// Value, Key, Type, EndPosition, Error
// Start needs to be pointing at the opening quote (") (or whitespace) of the key in order to succeed.
func extractObjectMember(search []byte, start int) ([]byte, string, string, int, error) {
	key, start, err := extractKey(search, start)
	if err != nil {
		return nil, "", "", 0, err
	}

	v, t, start, err := extractValue(search, start)
	if err != nil {
		return nil, "", "", 0, errors.New(err.Error() + " (expected object value)")
	}

	return v, *(*string)(unsafe.Pointer(&key)), t, start, err
}

// Extract the next available value from an array.
// Start needs to be pointing at the opening item (or whitespace) of the next array value.
// If start points at the opening bracket of the parent array, extractArrayValue will fail.
func extractArrayValue(search []byte, start int) ([]byte, string, int, error) {
	v, t, start, err := extractValue(search, start)
	if err != nil {
		return nil, "", 0, errors.New(err.Error() + " (expected array value)")
	}

	var termErr error
	finalPos := findTerminator(search, start)
	if finalPos < 0 {
		termErr = fmt.Errorf("expected array value terminator ('}', ']' or ',') at position '%d' in segment '%s'", start, truncate(search, 50))
	}

	return v, t, finalPos, termErr
}

// countMembers assumes a full object or slice, complete with opening and closing brackets.
func countMembers(b []byte, t string) int {
	if IsEmptyArray(b) || IsEmptyObject(b) {
		return 0
	}

	switch t {
	case JSONObject:
		return countObjectMembers(b)
	case JSONArray:
		return countSliceMembers(b)
	case JSONNull:
		return -1
	default:
		return 1
	}
}

func countObjectMembers(b []byte) int {
	start := 1
	length := 0
	for start < len(b) {
		_, _, _, pos, err := extractObjectMember(b, start)
		if err != nil {
			panic(err)
		}

		start = findTerminator(b, pos)
		if pos >= len(b) || start < 0 {
			panic(fmt.Errorf("expected value terminator ('}', ']' or ',') at position '%d' in segment '%s'", pos, truncate(b, 50)))
		}

		length++
	}

	return length
}

func countSliceMembers(b []byte) int {
	start := 1
	length := 0
	for start < len(b) {
		_, _, pos, err := extractValue(b, start)
		if err != nil {
			panic(err)
		}

		start = findTerminator(b, pos)
		if pos >= len(b) || start < 0 {
			panic(fmt.Errorf("expected value terminator ('}', ']' or ',') at position '%d' in segment '%s'", pos, truncate(b, 50)))
		}

		length++
	}

	return length
}
