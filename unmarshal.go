package gojson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cast"
)

type result struct {
	Pos   int
	Key   string
	Value []byte
	Type  string
}

// PostUnmarshaler is the interface implemented by types
// that can perform actions after Unmarshal / UnmarshalJSON
// has finished. This allows you to inspect the results and
// make corrections / adjustments / error checking after the
// unmarshaler has finished its work. An example use case is
// checking whether a slice is nil after unmarshal.
//
// Errors returned from PostUnmarshalJSON executed by the Unmarshal
// function (rather than called explicitly) will be ignored, UNLESS
// that error is included as the parameter to a panic. A panic
// inside PostUnmarshalJSON will be returned as the error return
// value from Unmarshal's PanicRecovery functionality.
type PostUnmarshaler interface {
	PostUnmarshalJSON([]byte, error) error
}

// UnmarshalStrict takes a json format byte string and extracts it into the given container using
// strict standards for type association.
func UnmarshalStrict(raw []byte, v interface{}) (err error) {
	u := unmarshaler{StrictStandards: true}
	return u.unmarshal(raw, v)
}

// Unmarshal takes a json format byte string and extracts it into the given container.
func Unmarshal(raw []byte, v interface{}) (err error) {
	u := unmarshaler{StrictStandards: false}
	return u.unmarshal(raw, v)
}

type unmarshaler struct {
	StrictStandards bool
}

func (u *unmarshaler) unmarshal(raw []byte, v interface{}) (err error) {
	defer PanicRecovery(&err)

	raw = trim(raw)

	if len(raw) == 0 {
		return fmt.Errorf("empty json value provided")
	}

	p := reflect.ValueOf(v)
	if p.Kind() != reflect.Ptr {
		return fmt.Errorf("supplied container (v) must be a pointer")
	}

	p = resolvePtr(p)
	if !p.CanSet() {
		return fmt.Errorf("unsettable value provided to Unmarshal")
	}

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

	t := GetJSONType(raw, 0)

	if t == JSONInvalid {
		err = ErrMalformedJSON
		return
	}

	switch p.Kind() {
	case reflect.Map:
		err = u.unmarshalMap(raw, t, p)
		return err
	case reflect.Slice:
		err = u.unmarshalSlice(raw, t, p)
		return err
	case reflect.Struct:
		err = u.unmarshalStruct(raw, t, p)
		return err
	case reflect.Interface:
		v := reflect.ValueOf(toIface(raw, t, u.StrictStandards))
		if v.IsValid() {
			p.Set(v)
		}
	default:
		err = u.setValue(raw, t, p)
		if err != nil {
			return err
		}
	}

	return err
}

// Extract the byte string into a slice container.
func (u *unmarshaler) unmarshalSlice(b []byte, t string, p reflect.Value) (err error) {
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

	if t == JSONNull {
		return nil
	}

	if u.StrictStandards && t != JSONArray {
		err = fmt.Errorf("strict standards: attempt to unmarshal JSON value with type '%s' into slice", t)
		return
	}

	childType := p.Type().Elem().Kind()

	// ByteSlices are exceptionally hard to extract byte-by-byte given the difficulty
	// of finding the correct position in the RawData, so we circumvent that problem by
	// short-circuiting and treating it as if the whole array were an elemental type.
	if childType == reflect.Uint8 {
		if t == JSONString && len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
			b = b[1 : len(b)-1]
		}
		p.Set(reflect.ValueOf(b))
		return nil
	}

	// Count the member elements so that we can know how big to size our slice.
	length := countMembers(b, t)

	if length < 1 {
		return nil
	}

	slice := reflect.MakeSlice(p.Type(), length, length)

	// Switch on the child type
	start := 1
	i := 0
	for start < len(b) {
		var v []byte
		var err error
		var pos int
		var vt string

		switch t {
		case JSONObject:
			v, _, vt, pos, err = extractObjectMember(b, start)
			if err != nil {
				return err
			}

			start = findTerminator(b, pos)
			if pos >= len(b) || start < 0 {
				return fmt.Errorf("expected value terminator ('}', ']' or ',') at position '%d' in segment '%s'", pos, truncate(b, 50))
			}
		case JSONArray:
			v, vt, pos, err = extractValue(b, start)
			if err != nil {
				return err
			}

			start = findTerminator(b, pos)
			if pos >= len(b) || start < 0 {
				return fmt.Errorf("expected value terminator ('}', ']' or ',') at position '%d' in segment '%s'", pos, truncate(b, 50))
			}
		default:
			v, vt, pos, err = extractValue(b, 0)
			if err != nil {
				return err
			}

			start = pos
		}

		if err != nil {
			return err
		}

		sliceMember := slice.Index(i)
		child := resolvePtr(sliceMember)

		switch child.Kind() {
		case reflect.Map:
			err = u.unmarshalMap(v, vt, child)
			if err != nil {
				return err
			}
		case reflect.Slice:
			err = u.unmarshalSlice(v, vt, child)
			if err != nil {
				return err
			}
		case reflect.Struct:
			err = u.unmarshalStruct(v, vt, child)
			if err != nil {
				return err
			}
		case reflect.Interface:
			if v := reflect.ValueOf(toIface(v, vt, u.StrictStandards)); v.IsValid() {
				child.Set(v)
			} else {
				child.Set(reflect.New(p.Type().Elem()).Elem())
			}
		default:
			err = u.setValue(v, vt, child)
			if err != nil {
				return err
			}
		}

		i++
	}

	p.Set(slice)
	return err
}

// Extract the byte string into a map container.
func (u *unmarshaler) unmarshalMap(b []byte, t string, p reflect.Value) (err error) {
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

	if t == JSONNull {
		return nil
	}

	if u.StrictStandards && t != JSONObject {
		err = fmt.Errorf("strict standards: attempt to unmarshal JSON value with type '%s' into map", t)
		return
	}

	childType := p.Type().Elem().Kind()

	// ByteSlices are exceptionally hard to extract byte-by-byte given the difficulty
	// of finding the correct position in the RawData, so we circumvent that problem by
	// short-circuiting and treating it as if the whole array were an elemental type.
	if childType == reflect.Uint8 {
		if t == JSONString && len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
			b = b[1 : len(b)-1]
		}

		newMap := reflect.MakeMap(p.Type())

		for k, v := range b {
			newMap.SetMapIndex(reflect.ValueOf(strconv.Itoa(k)), reflect.ValueOf(v))
		}

		p.Set(newMap)
		return nil
	}

	switch {
	case t == JSONObject && IsEmptyObject(b):
		return nil
	case t == JSONArray && IsEmptyArray(b):
		return nil
	}

	newMap := reflect.MakeMap(p.Type())

	// Switch on the child type
	start := 1
	i := 0
	for start < len(b) {
		var v []byte
		var err error
		var pos int
		var vt, k string

		switch t {
		case JSONObject:
			v, k, vt, pos, err = extractObjectMember(b, start)
			if err != nil {
				return err
			}

			start = findTerminator(b, pos)
			if pos >= len(b) || start < 0 {
				return fmt.Errorf("expected value terminator ('}', ']' or ',') at position '%d' in segment '%s'", pos, truncate(b, 50))
			}
		case JSONArray:
			v, vt, pos, err = extractValue(b, start)
			if err != nil {
				return err
			}

			start = findTerminator(b, pos)
			if pos >= len(b) || start < 0 {
				return fmt.Errorf("expected value terminator ('}', ']' or ',') at position '%d' in segment '%s'", pos, truncate(b, 50))
			}
			k = cast.ToString(i)
		default:
			v, vt, pos, err = extractValue(b, 0)
			if err != nil {
				return err
			}

			start = pos
			k = cast.ToString(i)
		}

		key := reflect.ValueOf(k)
		mapElement := reflect.New(p.Type().Elem()).Elem()
		child := resolvePtr(mapElement)

		switch child.Kind() {
		case reflect.Map:
			err = u.unmarshalMap(v, vt, child)
			if err != nil {
				return err
			}
			newMap.SetMapIndex(key, mapElement)

		case reflect.Slice:
			err = u.unmarshalSlice(v, vt, child)
			if err != nil {
				return err
			}
			newMap.SetMapIndex(key, mapElement)
		case reflect.Struct:
			err = u.unmarshalStruct(v, vt, child)
			if err != nil {
				return err
			}
			newMap.SetMapIndex(key, mapElement)
		case reflect.Interface:
			if v := reflect.ValueOf(toIface(v, vt, u.StrictStandards)); v.IsValid() {
				newMap.SetMapIndex(key, v)
			} else {
				newMap.SetMapIndex(key, mapElement)
			}
		default:
			err = u.setValue(v, vt, child)
			if err != nil {
				return err
			}
			newMap.SetMapIndex(key, mapElement)
		}

		i++
	}

	p.Set(newMap)
	return nil
}

// Extract the byte string into a struct container.
func (u *unmarshaler) unmarshalStruct(b []byte, t string, p reflect.Value) (err error) {
	// Check if p implements the json.Unmarshaler interface.
	if p.CanAddr() && p.Addr().NumMethod() > 0 {
		if u, ok := p.Addr().Interface().(PostUnmarshaler); ok {
			defer func() { err = u.PostUnmarshalJSON(b, err) }()
		}
		if u, ok := p.Addr().Interface().(json.Unmarshaler); ok {
			return u.UnmarshalJSON(b)
		}
	}

	info := getStructInfo(p.Type())
	keys := info.Keys

	if t != JSONObject {
		if u.StrictStandards {
			err = fmt.Errorf("attempt to unmarshal JSON value with type '%s' into struct", t)
			return
		}

		if len(info.RequiredKeys) > 0 {
			err = fmt.Errorf("missing required keys '%s' for struct '%s'", strings.Join(info.RequiredKeys, ","), p.Type().Name())
			return
		}

		return nil
	}

	if IsEmptyObject(b) || IsEmptyArray(b) {
		if len(info.RequiredKeys) > 0 {
			err = fmt.Errorf("missing required keys '%s' for struct '%s'", strings.Join(info.RequiredKeys, ","), p.Type().Name())
			return
		}

		return nil
	}

	// Extract the Data
	start := 0
	if t == JSONArray || t == JSONObject {
		start = 1
	}

	required := make(map[string]bool, len(info.RequiredKeys))
	for _, k := range info.RequiredKeys {
		required[k] = false
	}

	count := len(keys)
	for start < len(b) && count > 0 {
		v, k, vt, pos, eErr := extractKeyValue(b, start)
		start = pos
		if eErr != nil {
			err = eErr
			return err
		}

		if _, isset := required[k]; isset {
			required[k] = true
		}

		if _, ok := keys[k]; !ok {
			continue
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

		if info.NonEmpty(k) && isZeroValue(v, vt) {
			return fmt.Errorf("nonempty key '%s' for struct '%s' has %s zero value", keys[k].Name, p.Type().Name(), vt)
		}

		switch f.Kind() {
		case reflect.Map:
			err = u.unmarshalMap(v, vt, f)
			if err != nil {
				return err
			}
		case reflect.Slice:
			err = u.unmarshalSlice(v, vt, f)
			if err != nil {
				return err
			}
		case reflect.Struct:
			err = u.unmarshalStruct(v, vt, f)
			if err != nil {
				return err
			}
		case reflect.Interface:
			v := reflect.ValueOf(toIface(v, vt, u.StrictStandards))
			if v.IsValid() {
				f.Set(v)
			}
		default:
			err = u.setValue(v, vt, f)
			if err != nil {
				return err
			}
		}

		count--
	}

	for _, k := range info.RequiredKeys {
		if !required[k] {
			err = fmt.Errorf("required key '%s' for struct '%s' was not found", k, p.Type().Name())
			return
		}
	}

	return nil
}

func isZeroValue(v []byte, t string) bool {
	switch t {
	case JSONBool:
		return IsJSONFalse(v)
	case JSONString:
		// Since strings require opening and closing quotes and we have passed the parser by now, empty string "" has len 2.
		return len(v) <= 2
	case JSONInt:
		return len(v) == 1 && v[0] == '0'
	case JSONFloat:
		return (len(v) == 1 && v[0] == '0') || (len(v) == 3 && v[0] == '0' && v[1] == '.' && v[2] == '0')
	case JSONArray:
		return IsEmptyArray(v)
	case JSONObject:
		return IsEmptyObject(v)
	}
	return false
}

// Resolve a pointer to a concrete Value. If necessary, memory will be allocated to
// store the object being pointed to.
func resolvePtr(p reflect.Value) reflect.Value {
	op := p

	for p.Kind() == reflect.Ptr || p.Kind() == reflect.Interface {
		if p.Kind() == reflect.Ptr && !p.Elem().CanAddr() {
			child := reflect.New(p.Type().Elem()).Elem()
			p.Set(child.Addr())
		}

		if !p.Elem().IsValid() {
			break
		}

		p = p.Elem()

		// Retain the last setable value. This usually comes into play when we have
		// an interface that represents a non-settable value. The end result is we will
		// perform the extraction as if we were an interface. This is in alignment with
		// the behavior of encoding/json.Unmarshal
		if p.CanSet() {
			op = p
		}
	}

	if !p.CanSet() {
		p = op
	}

	return p
}

// Store the given value into the container based on the JSON type of the value.
func (u *unmarshaler) setValue(b []byte, t string, p reflect.Value) (err error) {
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

	switch p.Kind() {
	// Common Types First
	case reflect.String:
		if u.StrictStandards && t != JSONString {
			panic(fmt.Errorf("strict standards error, expected string, got %s", t))
		}
		p.SetString(toString(b, t, u.StrictStandards))
		return nil
	case reflect.Int:
		if u.StrictStandards && t != JSONInt {
			panic(fmt.Errorf("strict standards error, expected int, got %s", t))
		}
		p.SetInt(int64(toInt(b, t, u.StrictStandards)))
		return nil
	case reflect.Float64, reflect.Float32:
		if u.StrictStandards && t != JSONFloat {
			panic(fmt.Errorf("strict standards error, expected float, got %s", t))
		}
		p.SetFloat(toFloat(b, t, u.StrictStandards))
		return nil
	case reflect.Bool:
		if u.StrictStandards && t != JSONBool {
			panic(fmt.Errorf("strict standards error, expected bool, got %s", t))
		}
		p.SetBool(toBool(b, t, u.StrictStandards))
		return nil

	// Less Common Types
	case reflect.Uint8, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if u.StrictStandards && t != JSONInt {
			panic(fmt.Errorf("strict standards error, expected int, got %s", t))
		}
		p.SetUint(uint64(toInt(b, t, u.StrictStandards)))
		return nil
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		if u.StrictStandards && t != JSONInt {
			panic(fmt.Errorf("strict standards error, expected int, got %s", t))
		}
		p.SetInt(int64(toInt(b, t, u.StrictStandards)))
		return nil

	default:
		// Invalid, Complex64, Complex128, Array, Chan, Func
		err = fmt.Errorf("Unmarshal: Invalid Container Type '%s'", p.Kind())
		return
	}
}

// For objects and arrays, parse the data and collect information about each member element for further processing.
func getNodeList(b []byte, t string) ([]result, error) {
	start := 0
	if t == JSONObject {
		start = 1
		if IsEmptyObject(b) {
			return make([]result, 0), nil
		}
	}
	if t == JSONArray {
		start = 1
		if IsEmptyArray(b) {
			return make([]result, 0), nil
		}
	}

	// Extract the slice values so we can transform them.
	nodes := make([]result, 20)
	i := 0

	for start < len(b) {
		var v []byte
		var err error
		var pos int
		var k, vt string

		switch t {
		case JSONObject:
			v, k, vt, pos, err = extractObjectMember(b, start)
			start = findTerminator(b, pos)
			if err != nil {
				return nil, err
			}
			if pos >= len(b) || start < 0 {
				return nil, fmt.Errorf("expected value terminator ('}', ']' or ',') at position '%d' in segment '%s'", pos, truncate(b, 50))
			}
		case JSONArray:
			k = strconv.Itoa(i)
			v, vt, pos, err = extractValue(b, start)
			start = findTerminator(b, pos)
			if err != nil {
				return nil, err
			}
			if pos >= len(b) || start < 0 {
				return nil, fmt.Errorf("expected value terminator ('}', ']' or ',') at position '%d' in segment '%s'", pos, truncate(b, 50))
			}
		default:
			k = strconv.Itoa(i)
			v, vt, pos, err = extractValue(b, start)
			start = findTerminator(b, pos)
		}
		if err != nil {
			return nil, err
		}

		// Expand the node list if we're out of space.
		if i >= len(nodes) {
			nodes = append(nodes[:i], append(make([]result, 1+len(nodes)*2), nodes[i:]...)...)
		}

		nodes[i] = result{Value: v, Key: k, Type: vt, Pos: pos}
		i++
	}

	return nodes[:i], nil
}
