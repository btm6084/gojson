package gojson

import (
	"errors"
	"fmt"
	"strconv"
	"unsafe"
)

type parsed struct {
	keys     []string
	key      string
	bytes    []byte
	dtype    string
	children map[string]parsed
}

var (
	// ErrEmpty is returned when no input is provided.
	ErrEmpty = errors.New("empty input value")

	// ErrMalformedJSON is returned when input failed to parse.
	ErrMalformedJSON = errors.New("malformed json provided")

	period    = []byte{'.'}
	exponent  = []byte{'e'}
	exponentE = []byte{'E'}
)

// parse turns the stored rawData into an indexable list of values
func (jr *JSONReader) parse() error {
	if len(jr.rawData) == 0 {
		return ErrEmpty
	}

	jr.rawData = trim(jr.rawData)

	p, _ := jr.parseValue(0)
	if p.dtype == "" {
		return ErrMalformedJSON
	}

	jr.Type = p.dtype

	if p.dtype == JSONArray || p.dtype == JSONObject {
		jr.Keys = p.keys
		jr.parsed = p.children
		return nil
	}

	jr.Keys = []string{"0"}
	jr.parsed = map[string]parsed{
		"0": p,
	}

	return nil
}

func (jr *JSONReader) parseKey(start int) ([]byte, int) {
	start = ltrim(jr.rawData, start)

	if start < 0 || start >= len(jr.rawData) || jr.rawData[start] != '"' {
		return nil, -1
	}

	start++
	end := start
	keyStart := start
	keyEnd := -1

	escape := false
	for found := false; !found && end <= len(jr.rawData)-1; end++ {
		switch {
		case escape:
			escape = false
		case jr.rawData[end] == '\\' && !escape:
			escape = true
		case jr.rawData[end] == '"' && !escape: // Found an ending string
			keyEnd = end
			found = true
		}
	}

	// Advance past the key
	found := false
	for !found && end <= len(jr.rawData)-1 {
		switch {
		case jr.rawData[end] == ':':
			end++
			found = true
		case isWhitespace(jr.rawData[end]):
			end++
		default:
			return nil, -1
		}
	}

	return jr.rawData[keyStart:keyEnd], end
}

// ParseKeyValue assumes we start at the beginning of a string.
func (jr *JSONReader) parseKeyValue(current int) (parsed, int) {
	var key []byte
	key, current = jr.parseKey(current)
	if current < 0 {
		return parsed{}, -1
	}

	p, current := jr.parseValue(current)
	if current < 0 {
		return parsed{}, -1
	}

	p.key = *(*string)(unsafe.Pointer(&key))
	return p, current
}

func (jr *JSONReader) parseValue(current int) (parsed, int) {
	var p parsed
	current = ltrim(jr.rawData, current)

	switch GetJSONType(jr.rawData, current) {
	case JSONFloat, JSONInt:
		p, current = jr.parseNumber(current)
	case JSONBool, JSONNull:
		p, current = jr.parseConst(current)
	case JSONString:
		p, current = jr.parseString(current)
	case JSONObject:
		p, current = jr.parseObject(current)
	case JSONArray:
		p, current = jr.parseArray(current)
	default:
		return p, -1
	}

	// Consume the comma, ], or } following the value.
	// Anything not-those-three and non-whitespace causes an error.
	initial := current
	current = findTerminator(jr.rawData, current)
	if current < 0 {
		panic(fmt.Errorf("expected ',', ']', or '}' at position %d", initial))
	}

	// Don't consume the ending ] or }, as they're not part of the value
	if current > 0 && (jr.rawData[current-1] == ']' || jr.rawData[current-1] == '}') {
		current--
	}

	return p, current
}

func (jr *JSONReader) parseString(start int) (parsed, int) {
	initial := start
	if start < 0 {
		jr.Empty = true
		panic(fmt.Errorf(`invalid starting position in parseString`))
	}

	if jr.rawData[start] != '"' {
		jr.Empty = true
		panic(fmt.Errorf(`expected '"', found '%s' at position %d`, string(jr.rawData[start]), start))
	}

	start++
	end := start
	escape := false

	for found := false; !found && end <= len(jr.rawData)-1; end++ {

		switch {
		case escape:
			escape = false
		case jr.rawData[end] == '\\' && !escape:
			escape = true
		case jr.rawData[end] == '"' && !escape: // Found an ending string
			return parsed{bytes: jr.rawData[start:end], dtype: JSONString}, end + 1
		}
	}

	jr.Empty = true
	panic(fmt.Errorf(`unterminated string at starting position %d`, initial))
}

func (jr *JSONReader) parseArray(current int) (parsed, int) {
	var p parsed
	arrStart := current

	// Consume the [
	current++

	index := 0
	value := current
	lastValid := current
	var cp parsed

	for ; value > 0; index++ {
		cp, value = jr.parseValue(current)
		if value < 0 {
			break
		}

		if p.children == nil {
			p.children = make(map[string]parsed)
		}

		sIndex := strconv.Itoa(index)
		p.children[sIndex] = cp
		p.keys = append(p.keys, sIndex)

		current = value
		lastValid = value
	}

	if lastValid == len(jr.rawData) {
		lastValid = len(jr.rawData) - 1
	}

	// Consume the ]
	current = ltrim(jr.rawData, current)
	if current >= len(jr.rawData) || jr.rawData[current] != ']' {
		jr.Empty = true
		panic(fmt.Errorf("expected ']', found '%s' at position %d", string(jr.rawData[lastValid]), lastValid))
	}

	current++
	p.bytes = jr.rawData[arrStart:current]
	p.dtype = JSONArray

	return p, current
}

func (jr *JSONReader) parseObject(current int) (parsed, int) {
	var p parsed
	objStart := current

	// Consume the {
	current++

	value := current
	lastValid := current
	var cp parsed

	for value > 0 {
		cp, value = jr.parseKeyValue(current)
		if value < 0 {
			break
		}

		if p.children == nil {
			p.children = make(map[string]parsed)
		}

		p.children[cp.key] = cp
		p.keys = append(p.keys, cp.key)

		current = value
		lastValid = value
	}

	// Consume the }
	current = ltrim(jr.rawData, current)
	if jr.rawData[current] != '}' {
		jr.Empty = true
		panic(fmt.Errorf("expected '}', found '%s' at position %d", string(jr.rawData[lastValid]), lastValid))
	}

	current++
	p.bytes = jr.rawData[objStart:current]
	p.dtype = JSONObject

	return p, current
}

func (jr *JSONReader) parseConst(start int) (parsed, int) {
	start = ltrim(jr.rawData, start)
	initial := start
	length := len(jr.rawData) - start

	if length >= 4 {
		if IsJSONTrue(jr.rawData[start : start+4]) {
			return parsed{bytes: jr.rawData[start : start+4], dtype: JSONBool}, start + 4
		}

		if IsJSONNull(jr.rawData[start : start+4]) {
			return parsed{bytes: jr.rawData[start : start+4], dtype: JSONNull}, start + 4
		}

		if length >= 5 && IsJSONFalse(jr.rawData[start:start+5]) {
			return parsed{bytes: jr.rawData[start : start+5], dtype: JSONBool}, start + 5
		}
	}

	jr.Empty = true
	panic(fmt.Errorf("expected const at position %d", initial))
}

func (jr *JSONReader) parseNumber(start int) (parsed, int) {
	start = ltrim(jr.rawData, start)
	initial := start
	end := start
	found := false

	// Single Digit Case
	if len(jr.rawData) == 1 {
		end++
		found = true
	}

	for !found && end <= len(jr.rawData)-1 {
		if isTermByte(jr.rawData[end]) {
			break
		}
		end++
	}

	if IsJSONNumber(jr.rawData[start:end]) {
		b := trim(jr.rawData[start:end])
		return parsed{bytes: b, dtype: extractNumberType(b)}, end
	}

	jr.Empty = true
	panic(fmt.Errorf("expected number at position %d, found '%s'", initial, jr.rawData[start:end]))
}
