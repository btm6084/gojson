package gojson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/spf13/cast"
)

var (
	closers = map[string]byte{JSONObject: '}', JSONArray: ']'}
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
func findNumber(raw []byte) ([]byte, int, string) {
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
		return nil, 0, JSONInvalid
	}

	if len(raw) == 1 && raw[0] == '0' {
		return raw, 1, JSONInt
	}

	// It's not a valid number unless it begins with minus, or 1-9; or 0.
	end := 0
	switch raw[0] {
	case '-', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		end++
	case '0':
		if len(raw) > 2 && raw[1] == '.' { // If a number starts with 0, it must be followed by a .
			end++
			break
		}
		return nil, 0, JSONInvalid
	default:
		return nil, 0, JSONInvalid
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
				return nil, 0, JSONInvalid
			}
			eSign = true
			end++
		case 'e', 'E':
			if e {
				return nil, 0, JSONInvalid
			}
			e = true
			end++
		case '.':
			if period {
				return nil, 0, JSONInvalid
			}
			period = true
			end++
		case ' ':
			break SEARCH
		default:
			return nil, 0, JSONInvalid
		}
	}

	if period || e {
		return raw[:end], (a + end), JSONFloat
	}

	return raw[:end], a + end, JSONInt
}

// findString trims leading and trailing whitepsace, and removes the leading
// and trailing double quote if it exists.
func findBounded(raw []byte, open, close byte) ([]byte, int, error) {
	a := 0
	b := len(raw)
	lenRaw := len(raw)

	numOpen := 0
	for i := 0; i < len(raw); i++ {
		if isWS(raw[i]) {
			a++
			continue
		}

		if raw[i] == open {
			a++
			numOpen++
			break
		}

		return nil, 0, fmt.Errorf("expected string at position %d in segment '%s'", i, truncate(raw, 50))
	}

	found := false
	for i := a; i < lenRaw; i++ {
		if raw[i] == '\\' {
			if i >= len(raw)-1 {
				continue
			}

			if raw[i+1] == close {
				i++ // consume the escaped quote
				continue
			}
		}

		if raw[i] == close {
			numOpen--

			if numOpen == 0 {
				b = i
				found = true
				break
			}
		}
	}

	if !found {
		return nil, 0, fmt.Errorf("expected string to terminate in segment '%s'", truncate(raw, 50))
	}

	// Keep the open and close inside it.
	if a > 0 {
		a = a - 1
	}

	if b < len(raw) {
		b = b + 1
	}

	return raw[a:b], b, nil
}

func findString(raw []byte) ([]byte, int, error) {
	return findBounded(raw, '"', '"')
}

func findArray(raw []byte) ([]byte, int, error) {
	return findBounded(raw, '[', ']')
}

func findObject(raw []byte) ([]byte, int, error) {
	return findBounded(raw, '{', '}')
}

func jsonToInt(b []byte, t string) int {
	if isJSONTrue(b) {
		return 1
	}

	if t == "" {
		b, _, t = findNumber(b)
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
		b, _, t = findNumber(b)
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

func jsonToBool(raw []byte) bool {
	a, b := trimWS(raw, true)
	raw = raw[a:b]

	if isJSONTrue(raw) {
		return true
	}

	if len(raw) == 1 && raw[0] == '0' {
		return false
	}

	out, err := strconv.ParseBool(*(*string)(unsafe.Pointer(&raw)))
	if err != nil {
		return false
	}
	return out
}

func jsonToIface(raw []byte) interface{} {
	a, b := trimWS(raw, false)
	raw = raw[a:b]

	if len(raw) == 1 && raw[0] == '0' {
		return int(0)
	}

	switch jsonType(raw) {
	case JSONString:
		return jsonToString(raw)
	case JSONInt:
		return jsonToInt(raw, JSONInt)
	case JSONFloat:
		return jsonToFloat(raw, JSONFloat)
	case JSONNull:
		return nil
	case JSONBool:
		if isJSONTrue(raw) {
			return true
		} else {
			return false
		}
	case JSONObject:
		if IsEmptyObject(raw) {
			return nil
		}

		iface := make(map[string]interface{})
		raw = raw[1:] // Consume the opening brace.

		for {
			b, kb, n, err := getKeyValue(raw)
			if err != nil {
				panic(err)
			}

			iface[string(kb)] = jsonToIface(b)

			// Stop if we hit the end of the JSON
			raw = raw[n:] // Move past the consumed bytes

			a := afterNextWS(raw)
			if a == len(raw) {
				return fmt.Errorf("expected }, found EOL")
			}

			raw = raw[a:]
			switch raw[0] {
			case ',':
				raw = raw[1:] // Consume the comma
				continue
			case '}':
				return iface
			default:
				return fmt.Errorf("expected ',' or '}', found '%s'", string(raw[0]))
			}
		}
	case JSONArray: // @TODO: Do faster
		var iface []interface{}
		err := UnmarshalJSON(raw, &iface)
		if err != nil {
			panic(err)
		}

		return iface
	}

	return nil
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
	case '-', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
		_, _, t := findNumber(raw)
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

func trimWS(raw []byte, unquote bool) (int, int) {
	a := 0
	b := len(raw)

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

func unmarshalSlice(raw []byte, p reflect.Value) (err error) {
	// Check if p implements the json.Unmarshaler interface.
	if p.CanAddr() && p.Addr().NumMethod() > 0 {
		if u, ok := p.Addr().Interface().(PostUnmarshaler); ok {
			defer func() { err = u.PostUnmarshalJSON(raw, err) }()
		}
		if u, ok := p.Addr().Interface().(json.Unmarshaler); ok {
			err = u.UnmarshalJSON(raw)
			return
		}
	}

	a, b := trimWS(raw, false)
	raw = raw[a:b]

	t := jsonType(raw)

	childType := p.Type().Elem().Kind()

	// ByteSlices are exceptionally hard to extract byte-by-byte given the difficulty
	// of finding the correct position in the RawData, so we circumvent that problem by
	// short-circuiting and treating it as if the whole array were an elemental type.
	if childType == reflect.Uint8 {
		if t == JSONString {
			raw, _, err = findString(raw)
			if err != nil {
				return err
			}

			a, b := trimWS(raw, true)
			raw = raw[a:b]
		}
		p.Set(reflect.ValueOf(raw))
		return nil
	}

	length := countMembers(raw, t)
	if length < 1 {
		return nil
	}

	slice := reflect.MakeSlice(p.Type(), length, length)
	if t == JSONObject || t == JSONArray {
		raw = raw[1:] // Consume the opening bracket/brace.
	}

	for i := 0; i < length; i++ {
		var b []byte
		var n int
		var err error

		if t == JSONObject {
			b, _, n, err = getKeyValue(raw)
			if err != nil {
				return err
			}
		} else {
			b, n, err = findValue(raw)
			if err != nil {
				return err
			}
		}

		sliceMember := slice.Index(i)
		child := resolvePtr(sliceMember)

		switch child.Kind() {
		case reflect.Map:
			err = unmarshalMap(b, child)
			if err != nil {
				return err
			}
		case reflect.Slice:
			err = unmarshalSlice(b, child)
			if err != nil {
				return err
			}
		case reflect.Struct:
			err = unmarshalStruct(b, child)
			if err != nil {
				return err
			}
		case reflect.Interface:
			v := jsonToIface(b)
			if v != nil {
				child.Set(reflect.ValueOf(v))
			} else {
				child.Set(reflect.New(p.Type().Elem()).Elem())
			}
		default:
			setValue(b, child)
		}

		raw = raw[n:] // Move past the consumed bytes.
		if i == length-1 {
			break // We've finished extracting all the members.
		}

		// Next element must be a comma
		a := afterNextWS(raw)
		if a == len(raw) {
			return fmt.Errorf("expected ',', found EOL")
		}

		raw = raw[a:]

		if raw[0] != ',' {
			return fmt.Errorf("expected ',', found '%s' in segment '%s'", string(raw[0]), truncate(raw, 50))
		}

		// Consume the comma
		raw = raw[1:]
	}

	p.Set(slice)
	return nil
}

func unmarshalMap(raw []byte, p reflect.Value) (err error) {
	// Check if p implements the json.Unmarshaler interface.
	if p.CanAddr() && p.Addr().NumMethod() > 0 {
		if u, ok := p.Addr().Interface().(PostUnmarshaler); ok {
			defer func() { err = u.PostUnmarshalJSON(raw, err) }()
		}
		if u, ok := p.Addr().Interface().(json.Unmarshaler); ok {
			err = u.UnmarshalJSON(raw)
			return
		}
	}

	a, b := trimWS(raw, false)
	raw = raw[a:b]

	t := jsonType(raw)
	if t == JSONNull {
		return nil
	}

	childType := p.Type().Elem().Kind()

	// ByteSlices are exceptionally hard to extract byte-by-byte given the difficulty
	// of finding the correct position in the RawData, so we circumvent that problem by
	// short-circuiting and treating it as if the whole array were an elemental type.
	if childType == reflect.Uint8 {
		if t == JSONString {
			raw, _, err = findString(raw)
			if err != nil {
				return err
			}

			a, b := trimWS(raw, true)
			raw = raw[a:b]
		}
		newMap := reflect.MakeMap(p.Type())

		for k, v := range raw {
			newMap.SetMapIndex(reflect.ValueOf(strconv.Itoa(k)), reflect.ValueOf(v))
		}

		p.Set(newMap)
		return nil
	}

	switch {
	case t == JSONObject && IsEmptyObject(raw):
		return nil
	case t == JSONArray && IsEmptyArray(raw):
		return nil
	}

	newMap := reflect.MakeMap(p.Type())

	if t == JSONObject || t == JSONArray {
		raw = raw[1:] // Consume the opening bracket/brace.
	}

	i := 0
SEARCH:
	for {
		var k string
		var b, kb []byte
		var n int
		var err error

		if t == JSONObject {
			b, kb, n, err = getKeyValue(raw)
			if err != nil {
				return err
			}

			k = string(kb)
		} else {
			b, n, err = findValue(raw)
			if err != nil {
				return err
			}
			k = cast.ToString(i)
		}

		key := reflect.ValueOf(k)
		mapElement := reflect.New(p.Type().Elem()).Elem()
		child := resolvePtr(mapElement)

		switch child.Kind() {
		case reflect.Map:
			err := unmarshalMap(b, child)
			if err != nil {
				return err
			}
			newMap.SetMapIndex(key, mapElement)
		case reflect.Slice:
			err := unmarshalSlice(b, child)
			if err != nil {
				return err
			}
			newMap.SetMapIndex(key, mapElement)
		case reflect.Struct:
			err := unmarshalStruct(b, child)
			if err != nil {
				return err
			}
			newMap.SetMapIndex(key, mapElement)
		case reflect.Interface:
			v := jsonToIface(b)
			if v != nil {
				newMap.SetMapIndex(key, reflect.ValueOf(v))
			} else {
				newMap.SetMapIndex(key, mapElement)
			}
		default:
			setValue(b, child)
			newMap.SetMapIndex(key, mapElement)
		}

		i++

		raw = raw[n:] // Move past the consumed bytes
		a := afterNextWS(raw)

		// Stop if we hit the end of the JSON
		if a == len(raw) {
			if t == JSONArray || t == JSONObject {
				return fmt.Errorf("expected ',' or '%s', found EOL", string(closers[t]))
			}

			// Any other type, we hit the EOL, so all good.
			break SEARCH
		}

		raw = raw[a:]
		switch t {
		case JSONObject, JSONArray:
			if raw[0] == ',' {
				raw = raw[1:] // Consume the comma
				continue SEARCH
			}
			if raw[0] == closers[t] {
				break SEARCH
			}
			return fmt.Errorf("expected ',' or '%s', found '%s' in segment '%s'", string(closers[t]), string(raw[0]), truncate(raw, 50))
		default:
			// For all other types, anything other than member separator or EOL is malformed JSON.
			return ErrMalformedJSON
		}
	}

	p.Set(newMap)
	return nil
}

func unmarshalStruct(raw []byte, p reflect.Value) (err error) {
	// Check if p implements the json.Unmarshaler interface.
	if p.CanAddr() && p.Addr().NumMethod() > 0 {
		if u, ok := p.Addr().Interface().(PostUnmarshaler); ok {
			defer func() { err = u.PostUnmarshalJSON(raw, err) }()
		}
		if u, ok := p.Addr().Interface().(json.Unmarshaler); ok {
			return u.UnmarshalJSON(raw)
		}
	}

	t := jsonType(raw)
	if t != JSONObject {
		err = fmt.Errorf("attempt to unmarshal JSON value with type '%s' into struct", t)
		return
	}

	info := getStructInfo(p.Type())
	keys := info.Keys

	if IsEmptyObject(raw) || IsEmptyArray(raw) {
		if len(info.RequiredKeys) > 0 {
			err = fmt.Errorf("missing required keys '%s' for struct '%s'", strings.Join(info.RequiredKeys, ","), p.Type().Name())
			return
		}

		return nil
	}

	if t == JSONObject || t == JSONArray {
		raw = raw[1:] // Consume the opening bracket/brace.
	}

	required := make(map[string]bool, len(info.RequiredKeys))
	for _, k := range info.RequiredKeys {
		required[k] = false
	}

	// Tracking count keeps us from doing extra work if there's no where to put the remaining JSON anyway.
	// ie. If the JSON has 17 keys, but the struct has 3, we stop parsing once we've filled those three.
	count := len(keys)
SEARCH:
	for count > 0 {
		b, kb, n, gkvErr := getKeyValue(raw)
		if gkvErr != nil {
			return gkvErr
		}

		k := *(*string)(unsafe.Pointer(&kb))
		if _, isset := required[k]; isset {
			required[k] = true
		}

		// We didn't find the key, so proceed forward.
		if _, ok := keys[k]; !ok {
			// Stop if we hit the end of the JSON
			raw = raw[n:] // Move past the consumed bytes
			a := afterNextWS(raw)
			if a == len(raw) {
				return fmt.Errorf("expected }, found EOL")
			}

			raw = raw[a:]
			switch raw[0] {
			case ',':
				raw = raw[1:] // Consume the comma
				continue SEARCH
			case '}':
				break SEARCH
			default:
				return fmt.Errorf("expected ',' or '}', found '%s'", string(raw[0]))
			}
		}

		if info.NonEmpty(k) && isZeroValue(b, jsonType(b)) {
			return fmt.Errorf("nonempty key '%s' for struct '%s' has %s zero value", keys[k].Name, p.Type().Name(), jsonType(b))
		}

		// If we're dealing with an embeded struct, make sure we're expanding properly.
		var f reflect.Value
		if len(keys[k].Path) > 0 {
			f = p
			// Follow the path through the parent nodes until we hit the bottom.
			for _, i := range keys[k].Path {
				f = resolvePtr(f.Field(i))
			}
			f = resolvePtr(f.Field(keys[k].Index))
		} else {
			f = resolvePtr(p.Field(keys[k].Index))
		}

		switch f.Kind() {
		case reflect.Map:
			err = unmarshalMap(b, f)
			if err != nil {
				return
			}
		case reflect.Slice:
			err = unmarshalSlice(b, f)
			if err != nil {
				return
			}
		case reflect.Struct:
			err = unmarshalStruct(b, f)
			if err != nil {
				return
			}
		case reflect.Interface:
			v := jsonToIface(b)
			if v != nil {
				f.Set(reflect.ValueOf(v))
			}
		default:
			err = setValue(b, f)
			if err != nil {
				return
			}
		}

		count--

		// Stop if we hit the end of the JSON
		raw = raw[n:] // Move past the consumed bytes
		a := afterNextWS(raw)
		if a == len(raw) {
			return fmt.Errorf("expected }, found EOL")
		}

		raw = raw[a:]
		switch raw[0] {
		case ',':
			raw = raw[1:] // Consume the comma
			continue SEARCH
		case '}':
			break SEARCH
		default:
			return fmt.Errorf("expected ',' or '}', found '%s'", string(raw[0]))
		}

	}

	for _, k := range info.RequiredKeys {
		if !required[k] {
			err = fmt.Errorf("required key '%s' for struct '%s' was not found", k, p.Type().Name())
			return
		}
	}

	return nil
}

func afterNextWS(raw []byte) int {
	for i := 0; i < len(raw); i++ {
		if isWS(raw[i]) {
			continue
		}

		return i
	}

	return len(raw)
}

func getKeyValue(raw []byte) ([]byte, []byte, int, error) {
	consumed := 0
	kb, kn, err := findString(raw)
	if err != nil {
		return nil, nil, 0, err
	}

	consumed += kn
	raw = raw[consumed:]

	skip := afterNextWS(raw)
	if skip == len(raw) {
		return nil, nil, 0, fmt.Errorf("expected ':', got EOL")
	}

	consumed += skip
	raw = raw[skip:] // Consume any skipped whitespace

	if raw[0] != ':' {
		return nil, nil, 0, fmt.Errorf("expected ':', got %s", string(raw[0]))
	}

	raw = raw[1:]
	consumed++

	b, n, err := findValue(raw)
	if err != nil {
		return nil, nil, 0, err
	}
	consumed += n

	s, e := trimWS(kb, true)
	return b, kb[s:e], consumed, nil
}

// Find the next value up to a terminator
// returns: Bytes Found, Number of Bytes Consumed, Error State
func findValue(raw []byte) ([]byte, int, error) {
	trimmed := 0
	for i := 0; i < len(raw); i++ {
		if isWS(raw[i]) {
			trimmed++
			continue
		}
		break
	}

	raw = raw[trimmed:]

	if len(raw) == 0 {
		return nil, 0, ErrMalformedJSON
	}

	if len(raw) == 1 && raw[0] == 0 {
		return raw, 0, nil
	}

	switch raw[0] {
	case '{':
		b, n, err := findObject(raw)
		if err != nil {
			return nil, 0, err
		}

		return b, trimmed + n, nil
	case '[':
		b, n, err := findArray(raw)
		if err != nil {
			return nil, 0, err
		}

		return b, trimmed + n, nil
	case '"':
		b, n, err := findString(raw)
		if err != nil {
			return nil, 0, err
		}

		return b, trimmed + n, nil
	case '-', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
		b, n, t := findNumber(raw)
		if t == JSONInvalid {
			return nil, 0, fmt.Errorf("expected number in segment '%s'", truncate(raw, 50))
		}

		return b, trimmed + n, nil
	case 't', 'T':
		if len(raw) < 4 || !isJSONTrue(raw[:4]) {
			return nil, 0, fmt.Errorf("expected json `true` in segment '%s'", truncate(raw, 50))
		}
		return raw[:4], trimmed + 4, nil
	case 'f', 'F':
		if len(raw) < 5 || !isJSONFalse(raw[:5]) {
			return nil, 0, fmt.Errorf("expected json `false` in segment '%s'", truncate(raw, 50))
		}
		return raw[:5], trimmed + 5, nil
	case 'n', 'N':
		if len(raw) < 4 || !isJSONNull(raw[:4]) {
			return nil, 0, fmt.Errorf("expected json `null` in segment '%s'", truncate(raw, 50))
		}
		return raw[:4], trimmed + 4, nil
	}

	return nil, 0, ErrMalformedJSON
}
