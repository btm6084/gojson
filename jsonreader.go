package gojson

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

// JSONReader Provides utility functions for manipulating json structures.
type JSONReader struct {
	// Keys holds the list of top-level keys
	Keys []string

	// rawData is the initial byte string provided to NewJSONReader.
	rawData []byte

	// Type is the JSONType of the top-level data.
	Type string

	// parsed is the set of child nodes populated by the parser.
	parsed map[string]parsed

	// Empty is true if parsing failed or no data was supplied.
	Empty bool

	// StrictStandards directs the extraction functions to be strict with type
	// casting and extractions where applicable.
	StrictStandards bool
}

// NewJSONReader creates a new JSONReader object, which parses the rawData input and provides
// access to various accessor functions useful for working with JSONData.
//
// Behavior is undefined when a JSONReader is created via means other than NewJSONReader.
func NewJSONReader(rawData []byte) (reader *JSONReader, err error) {
	defer PanicRecovery(&err)

	if len(rawData) == 0 {
		return &JSONReader{Empty: true}, fmt.Errorf("No JSON Provided")
	}

	// We make a copy of rawData so that the backing array is completely incapsulated
	// by the reader, so that the user can't change the backing array later.
	reader = &JSONReader{}
	reader.rawData = make([]byte, len(rawData))
	copy(reader.rawData, rawData)

	reader.parse()

	if len(reader.parsed) == 0 {
		reader.Empty = true
		reader.rawData = nil
		return reader, err
	}

	return reader, err
}

// KeyExists returns true if a given key exists in the parsed json.
func (jr *JSONReader) KeyExists(key string) bool {
	keys := strings.Split(key, `.`)

	p := jr.parsed

	for _, k := range keys {
		if c, isset := p[k]; isset {
			p = c.children
			continue
		}
		return false
	}

	return true
}

/**
 * Nesting Functions
 */

// Get retrieves a nested object and returns a JSONReader with the root containing the contents of the delved key.
func (jr *JSONReader) Get(key string) *JSONReader {
	p := jr.getChildByKey(key)
	if p == nil {
		return &JSONReader{Empty: true}
	}

	var r JSONReader

	switch p.dtype {
	case JSONArray, JSONObject:
		r = JSONReader{rawData: p.bytes, parsed: p.children, Type: p.dtype, Keys: p.keys}
	default:
		r = JSONReader{rawData: p.bytes, parsed: map[string]parsed{"0": *p}, Type: p.dtype, Keys: []string{"0"}}
	}

	return &r
}

// GetCollection extracts a nested JSONArray and returns a slice of JSONReader, with one JSONReader for each
// element in the JSONArray.
func (jr *JSONReader) GetCollection(key string) []JSONReader {
	p := jr.getChildByKey(key)
	if p == nil {
		return []JSONReader(nil)
	}

	if len(p.keys) == 0 {
		slice := make([]JSONReader, 1)
		slice[0] = JSONReader{rawData: p.bytes, parsed: map[string]parsed{"0": *p}, Type: p.dtype, Keys: []string{"0"}}
		return slice
	}

	slice := make([]JSONReader, len(p.keys))
	count := 0
	for _, k := range p.keys {
		v := p.children[k]
		switch v.dtype {
		case JSONArray, JSONObject:
			slice[count] = JSONReader{rawData: v.bytes, parsed: v.children, Type: v.dtype, Keys: v.keys}
		default:
			slice[count] = JSONReader{rawData: v.bytes, parsed: map[string]parsed{"0": v}, Type: v.dtype, Keys: []string{"0"}}
		}
		count++
	}

	return slice
}

/**
 * String Functions
 */

// GetString retrieves a given key as a string, if it exists.
func (jr *JSONReader) GetString(key string) string {
	b, t, _ := jr.getDataByKey(key)
	if b == nil {
		return ""
	}
	return toString(b, t, jr.StrictStandards)
}

// ToString returns the top-level JSON as a string.
func (jr *JSONReader) ToString() string {
	return toString(jr.rawData, jr.Type, jr.StrictStandards)
}

// GetStringSlice retrieves a given key as a string slice, if it exists.
func (jr *JSONReader) GetStringSlice(key string) []string {
	p := jr.getChildByKey(key)
	if p == nil {
		return nil
	}

	iface := make([]string, 0)

	switch p.dtype {
	case JSONInt, JSONFloat, JSONBool, JSONString:
		iface = append(iface, toString(p.bytes, p.dtype, jr.StrictStandards))
	case JSONArray, JSONObject:
		for _, k := range p.keys {
			iface = append(iface, toString(p.children[k].bytes, p.children[k].dtype, jr.StrictStandards))
		}
	default:
		iface = append(iface, "")
	}

	return iface
}

// ToStringSlice returns all top-level data as a string slice.
func (jr *JSONReader) ToStringSlice() []string {
	return jr.GetStringSlice("")
}

// ToMapStringString returns all top-level data as map of string onto string.
func (jr *JSONReader) ToMapStringString() map[string]string {
	p := jr.getChildByKey("")
	iface := make(map[string]string)

	switch p.dtype {
	case JSONInt, JSONFloat, JSONBool, JSONString:
		iface["0"] = toString(p.bytes, p.dtype, jr.StrictStandards)
	case JSONArray, JSONObject:
		for _, k := range p.keys {
			iface[k] = toString(p.children[k].bytes, p.children[k].dtype, jr.StrictStandards)
		}
	}

	return iface
}

func toString(b []byte, t string, strict bool) string {
	if len(b) == 0 {
		return ""
	}

	if strict {
		if !IsJSONString(b) {
			panic(fmt.Errorf("invalid escape sequence in segment '%s'", truncate(b, 50)))
		}
	}

	if t == JSONNull {
		return ""
	}

	return manualUnescapeString(b)
}

// manualUnescapeString unquotes a quoted string, and replaces any escaped quotes with plain quotes.
func manualUnescapeString(b []byte) string {
	if len(b) < 2 {
		return marshalerDecode(b)
	}

	if b[0] != '"' || b[len(b)-1] != '"' {
		return marshalerDecode(b)
	}

	return marshalerDecode(b[1 : len(b)-1])

}

// Revert the HTML escaping for printable characters the encode/json Marshal performs if necessary.
// see: https://golang.org/pkg/encoding/json/#HTMLEscape
func marshalerDecode(b []byte) string {
	escapes := map[byte]byte{
		'\\': '\\',
		'"':  '"',
		'/':  '/',
		'b':  '\b',
		'f':  '\f',
		'n':  '\n',
		'r':  '\r',
		't':  '\t',
	}

	alloc := false
	for i := 0; i < len(b); i++ {
		if b[i] == '\\' {
			alloc = true
			break
		}
	}

	if !alloc {
		return string(b)
	}

	out := make([]byte, len(b))
	outLen := 0

	for i := 0; i < len(b); i++ {
		if b[i] != '\\' {
			out[outLen] = b[i]
			outLen++
			continue
		}

		// End of String
		if i+1 >= len(b) {
			out[outLen] = b[i]
			outLen++
			continue
		}

		if c, ok := escapes[b[i+1]]; ok {
			out[outLen] = c
			outLen++
			i++ // Skip past the consumed escape
			continue
		}

		if b[i+1] == 'u' {
			if i+5 >= len(b) {
				out[outLen] = b[i]
				outLen++
				continue
			}

			r, err := getUnicodeValue(b[i : i+6])
			if err != nil {
				out[outLen] = b[i]
				outLen++
				continue
			}

			length := 6

			// https://unicodebook.readthedocs.io/unicode_encodings.html#utf-16-surrogate-pairs
			// Unicode Surrogate Pair hex {D800-DBFF},{DC00-DFFF} dec {55296-56319},{56320-57343}
			if i+11 < len(b) && (r >= 55296 && r <= 56319) {
				r2, err := getUnicodeValue(b[i+6 : i+12])
				if err != nil {
					break
				}

				if r2 >= 56320 && r2 <= 57343 {
					length = 12
					r = ((r - 0xD800) * 0x400) + (r2 - 0xDC00) + 0x10000
				}
			}

			for _, bn := range []byte(string(rune(r))) {
				out[outLen] = bn
				outLen++
			}
			i += length - 1 // -1 to account for the incoming i++ following the continue
			continue
		}

	}

	return string(out[:outLen])
}

// b should match \u[A-z0-9]{4}.
func getUnicodeValue(b []byte) (int64, error) {
	if len(b) < 6 {
		return 0, errors.New("No Unicode Value")
	}

	if b[0] != '\\' || b[1] != 'u' {
		return 0, errors.New("No Unicode Value")
	}

	return strconv.ParseInt(string(b[2:6]), 16, 32)
}

/**
 * Boolean Functions
 */

// GetBool retrieves a given key as boolean, if it exists.
func (jr *JSONReader) GetBool(key string) bool {
	b, t, _ := jr.getDataByKey(key)
	if b == nil {
		return false
	}
	return toBool(b, t, jr.StrictStandards)
}

// ToBool returns the top-level JSON into an integer.
func (jr *JSONReader) ToBool() bool {
	return toBool(jr.rawData, jr.Type, jr.StrictStandards)
}

// GetBoolSlice retrieves a given key as a bool slice, if it exists.
func (jr *JSONReader) GetBoolSlice(key string) []bool {
	p := jr.getChildByKey(key)
	if p == nil {
		return nil
	}

	iface := make([]bool, 0)

	switch p.dtype {
	case JSONInt, JSONFloat, JSONBool, JSONString:
		iface = append(iface, toBool(p.bytes, p.dtype, jr.StrictStandards))
	case JSONArray, JSONObject:
		for _, k := range p.keys {
			iface = append(iface, toBool(p.children[k].bytes, p.children[k].dtype, jr.StrictStandards))
		}
	default:
		iface = append(iface, false)
	}

	return iface
}

// ToBoolSlice returns all top-level data as a bool slice.
func (jr *JSONReader) ToBoolSlice() []bool {
	return jr.GetBoolSlice("")
}

// ToMapStringBool returns all top-level data as map of string onto bool.
func (jr *JSONReader) ToMapStringBool() map[string]bool {
	p := jr.getChildByKey("")
	iface := make(map[string]bool)

	switch p.dtype {
	case JSONInt, JSONFloat, JSONBool, JSONString:
		iface["0"] = toBool(p.bytes, p.dtype, jr.StrictStandards)
	case JSONArray, JSONObject:
		for _, k := range p.keys {
			iface[k] = toBool(p.children[k].bytes, p.children[k].dtype, jr.StrictStandards)
		}
	}

	return iface
}

// Cast the given byte array to bool based on its JSON type.
func toBool(b []byte, t string, strict bool) bool {
	switch t {
	case JSONBool:
		return IsJSONTrue(b)
	case JSONInt:
		return !(len(b) == 1 && b[0] == '0')
	case JSONString:
		if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
			b = b[1 : len(b)-1]
		}

		b, err := strconv.ParseBool(*(*string)(unsafe.Pointer(&b)))
		if err != nil {
			return false
		}
		return b
	case JSONFloat:
		i, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&b)), 64)
		if err != nil {
			if strict {
				panic(err)
			}
			return false
		}
		return i != 0
	default:
		return false
	}
}

/**
 * Integer Functions
 */

// GetInt retrieves a given key as an int, if it exists.
func (jr *JSONReader) GetInt(key string) int {
	b, t, _ := jr.getDataByKey(key)
	if b == nil {
		return 0
	}
	return toInt(b, t, jr.StrictStandards)
}

// ToInt returns the top-level JSON into an integer.
func (jr *JSONReader) ToInt() int {
	return toInt(jr.rawData, jr.Type, jr.StrictStandards)
}

// GetIntSlice retrieves a given key as a int slice, if it exists.
func (jr *JSONReader) GetIntSlice(key string) []int {
	p := jr.getChildByKey(key)
	if p == nil {
		return nil
	}

	iface := make([]int, 0)

	switch p.dtype {
	case JSONInt, JSONFloat, JSONBool, JSONString:
		iface = append(iface, toInt(p.bytes, p.dtype, jr.StrictStandards))
	case JSONArray, JSONObject:
		for _, k := range p.keys {
			iface = append(iface, toInt(p.children[k].bytes, p.children[k].dtype, jr.StrictStandards))
		}
	default:
		iface = append(iface, 0)
	}

	return iface
}

// ToIntSlice returns all top-level data as a int slice.
func (jr *JSONReader) ToIntSlice() []int {
	return jr.GetIntSlice("")
}

// ToMapStringInt returns all top-level data as map of string onto int.
func (jr *JSONReader) ToMapStringInt() map[string]int {
	p := jr.getChildByKey("")
	iface := make(map[string]int)

	switch p.dtype {
	case JSONInt, JSONFloat, JSONBool, JSONString:
		iface["0"] = toInt(p.bytes, p.dtype, jr.StrictStandards)
	case JSONArray, JSONObject:
		for _, k := range p.keys {
			iface[k] = toInt(p.children[k].bytes, p.children[k].dtype, jr.StrictStandards)
		}
	}

	return iface
}

// Cast the given byte array to int based on its JSON type.
func toInt(b []byte, t string, strict bool) int {
	switch t {
	case JSONNull, JSONObject, JSONArray:
		return 0
	case JSONBool:
		if IsJSONTrue(b) {
			return 1
		}
		return 0
	case JSONString:
		b = trimString(b)
		t = GetJSONType(b, 0)
		if t != JSONString {
			return toInt(b, t, strict)
		}
	case JSONFloat:
		i, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&b)), 64)
		if err != nil {
			if strict {
				panic(err)
			}
			return 0
		}
		return int(i)
	}

	i, err := strconv.ParseInt(*(*string)(unsafe.Pointer(&b)), 10, 64)
	if err != nil {
		if strict {
			panic(err)
		}
		return 0
	}
	return int(i)
}

/**
 * Float Functions
 */

// GetFloat retrieves a given key as float64, if it exists.
func (jr *JSONReader) GetFloat(key string) float64 {
	b, t, _ := jr.getDataByKey(key)
	if b == nil {
		return 0
	}
	return toFloat(b, t, jr.StrictStandards)
}

// ToFloat returns the top-level JSON into a float64.
func (jr *JSONReader) ToFloat() float64 {
	return toFloat(jr.rawData, jr.Type, jr.StrictStandards)
}

// GetFloatSlice retrieves a given key as a float64 slice, if it exists.
func (jr *JSONReader) GetFloatSlice(key string) []float64 {
	p := jr.getChildByKey(key)
	if p == nil {
		return nil
	}

	iface := make([]float64, 0)

	switch p.dtype {
	case JSONInt, JSONFloat, JSONBool, JSONString:
		iface = append(iface, toFloat(p.bytes, p.dtype, jr.StrictStandards))
	case JSONArray, JSONObject:
		for _, k := range p.keys {
			iface = append(iface, toFloat(p.children[k].bytes, p.children[k].dtype, jr.StrictStandards))
		}
	default:
		iface = append(iface, 0)
	}

	return iface
}

// ToFloatSlice returns all top-level data as a float64 slice.
func (jr *JSONReader) ToFloatSlice() []float64 {
	return jr.GetFloatSlice("")
}

// ToMapStringFloat returns all top-level data as map of string onto float64.
func (jr *JSONReader) ToMapStringFloat() map[string]float64 {
	p := jr.getChildByKey("")
	iface := make(map[string]float64)

	switch p.dtype {
	case JSONInt, JSONFloat, JSONBool, JSONString:
		iface["0"] = toFloat(p.bytes, p.dtype, jr.StrictStandards)
	case JSONArray, JSONObject:
		for _, k := range p.keys {
			iface[k] = toFloat(p.children[k].bytes, p.children[k].dtype, jr.StrictStandards)
		}
	}

	return iface
}

// Cast the given byte array to float64 based on its JSON type.
func toFloat(b []byte, t string, strict bool) float64 {
	switch t {
	case JSONNull, JSONObject, JSONArray:
		return 0.0
	case JSONBool:
		if IsJSONTrue(b) {
			return 1.0
		}
		return 0.0
	default:
		if t == JSONString {
			b = trimString(b)
		}

		i, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&b)), 64)
		if err != nil {
			if strict {
				panic(err)
			}
			return 0.0
		}
		return i
	}
}

/**
 * Byte Slice Functions
 */

// GetByteSlice returns the given key and all child elements as a byte array.
func (jr *JSONReader) GetByteSlice(key string) []byte {
	b, _, _ := jr.getDataByKey(key)
	if b == nil {
		return []byte(nil)
	}
	return b
}

// ToByteSlice returns all top-level data as a byte slice.
func (jr *JSONReader) ToByteSlice() []byte {
	b := jr.rawData
	if jr.Type == JSONString && len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	return b
}

// GetByteSlices retrieves a given key as a slice of byte slices, if it exists.
func (jr *JSONReader) GetByteSlices(key string) [][]byte {
	p := jr.getChildByKey(key)
	if p == nil {
		return nil
	}

	iface := make([][]byte, 0)

	switch p.dtype {
	case JSONString:
		iface = append(iface, p.bytes)
	case JSONArray, JSONObject:
		for _, k := range p.keys {
			iface = append(iface, p.children[k].bytes)
		}
	default:
		iface = append(iface, p.bytes)
	}

	return iface
}

// ToByteSlices returns all top-level data as a slice of byte slices.
func (jr *JSONReader) ToByteSlices() [][]byte {
	iface := make([][]byte, 0)

	switch jr.Type {
	case JSONString:
		b := jr.rawData[:]
		if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
			b = b[1 : len(b)-1]
		}

		iface = append(iface, b)
	case JSONArray, JSONObject:
		for _, k := range jr.Keys {
			iface = append(iface, jr.parsed[k].bytes)
		}
	default:
		iface = append(iface, jr.rawData)
	}

	return iface
}

// ToMapStringBytes returns all top-level data as map of string onto []byte.
func (jr *JSONReader) ToMapStringBytes() map[string][]byte {
	iface := make(map[string][]byte, 0)

	switch jr.Type {
	case JSONArray, JSONObject:
		for _, k := range jr.Keys {
			iface[k] = jr.parsed[k].bytes
		}
	case JSONString:
		b := jr.rawData[:]
		if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
			b = b[1 : len(b)-1]
		}

		iface["0"] = b
	default:
		iface["0"] = jr.rawData
	}

	return iface
}

/**
 * Empty Interface Functions
 */

// GetInterface returns the data given key as an interface{} based on the JSONType of the data.
func (jr *JSONReader) GetInterface(key string) interface{} {
	return jr.getIface(key)
}

// ToInterface returns the top-level JSON as an interface{} based on the JSONType of the data.
func (jr *JSONReader) ToInterface() interface{} {
	return jr.getIface("")
}

// GetInterfaceSlice returns the given key as an interface{} slice.
func (jr *JSONReader) GetInterfaceSlice(key string) []interface{} {
	b, t, _ := jr.getDataByKey(key)
	if b == nil {
		return nil
	}

	var slice []interface{}
	switch t {
	case JSONArray:
		return jr.getSlice(key)
	case JSONObject:
		o, keys := jr.getObject(key)
		for _, k := range keys {
			slice = append(slice, o[k])
		}
		return slice
	default:
		slice = []interface{}{jr.getIface(key)}
	}

	return slice
}

// ToInterfaceSlice returns all top-level data as an interface{} slice.
func (jr *JSONReader) ToInterfaceSlice() []interface{} {
	var slice []interface{}
	switch jr.Type {
	case JSONArray:
		return jr.getSlice("")
	case JSONObject:
		o, keys := jr.getObject("")
		for _, k := range keys {
			slice = append(slice, o[k])
		}
		return slice
	default:
		slice = []interface{}{toIface(jr.rawData, jr.Type, jr.StrictStandards)}
	}

	return slice
}

// GetMapStringInterface retrieves a given key as a map of string onto interface{}, if said key exists.
func (jr *JSONReader) GetMapStringInterface(key string) map[string]interface{} {
	b, t, _ := jr.getDataByKey(key)
	if b == nil {
		return nil
	}

	var slice map[string]interface{}
	switch t {
	case JSONArray:
		slice = make(map[string]interface{})
		for k, v := range jr.getSlice(key) {
			slice[strconv.Itoa(k)] = v
		}
		return slice
	case JSONObject:
		o, _ := jr.getObject(key)
		return o
	default:
		slice = make(map[string]interface{})
		slice["0"] = toIface(b, t, jr.StrictStandards)
	}

	return slice
}

// ToMapStringInterface retrieves a given key as a map of string onto interface{}, if said key exists.
func (jr *JSONReader) ToMapStringInterface() map[string]interface{} {
	var slice map[string]interface{}
	switch jr.Type {
	case JSONArray:
		slice = make(map[string]interface{})
		for k, v := range jr.getSlice("") {
			slice[strconv.Itoa(k)] = v
		}
		return slice
	case JSONObject:
		o, _ := jr.getObject("")
		return o
	default:
		slice = make(map[string]interface{})
		slice["0"] = jr.getIface("")
	}

	return slice
}

// Retrieve the data for a given key and return it as an interface{} based on its JSON type.
func (jr *JSONReader) getIface(key string) interface{} {
	p := jr.getChildByKey(key)
	if p == nil {
		return interface{}(nil)
	}

	switch p.dtype {
	case JSONInt:
		return toInt(p.bytes, p.dtype, jr.StrictStandards)
	case JSONFloat:
		return toFloat(p.bytes, p.dtype, jr.StrictStandards)
	case JSONBool:
		return toBool(p.bytes, p.dtype, jr.StrictStandards)
	case JSONString:
		return toString(p.bytes, p.dtype, jr.StrictStandards)
	case JSONObject:
		o, _ := jr.getObject(key)
		return o
	case JSONArray:
		return jr.getSlice(key)
	default:
		return interface{}(nil)
	}
}

// Retrieve the data for a given key and return it as a map[string]interface{} based on its JSON type.
// Also returns the set of keys in the orer they appear, so that ordering can be preserved.
func (jr *JSONReader) getObject(key string) (map[string]interface{}, []string) {
	p := jr.getChildByKey(key)
	if p == nil {
		return nil, nil
	}

	if p.dtype != JSONObject {
		return nil, nil
	}

	iface := make(map[string]interface{})

	for _, k := range p.keys {
		v := p.children[k]

		switch v.dtype {
		case JSONInt:
			iface[k] = toInt(v.bytes, v.dtype, jr.StrictStandards)
		case JSONFloat:
			iface[k] = toFloat(v.bytes, v.dtype, jr.StrictStandards)
		case JSONBool:
			iface[k] = toBool(v.bytes, v.dtype, jr.StrictStandards)
		case JSONString:
			iface[k] = toString(v.bytes, v.dtype, jr.StrictStandards)
		case JSONObject:
			iface[k], _ = jr.Get(key).getObject(k)
		case JSONArray:
			iface[k] = jr.Get(key).getSlice(k)
		default:
			iface[k] = nil
		}
	}

	return iface, p.keys
}

// Retrieve the data for a given key and return it as an interface{} slice based on its JSON type.
func (jr *JSONReader) getSlice(key string) []interface{} {
	p := jr.getChildByKey(key)
	if p == nil {
		return nil
	}

	if p.dtype != JSONArray {
		return nil
	}

	iface := make([]interface{}, 0)

	for _, k := range p.keys {
		v := p.children[k]

		switch v.dtype {
		case JSONInt:
			iface = append(iface, toInt(v.bytes, v.dtype, jr.StrictStandards))
		case JSONFloat:
			iface = append(iface, toFloat(v.bytes, v.dtype, jr.StrictStandards))
		case JSONBool:
			iface = append(iface, toBool(v.bytes, v.dtype, jr.StrictStandards))
		case JSONString:
			iface = append(iface, toString(v.bytes, v.dtype, jr.StrictStandards))
		case JSONObject:
			o, _ := jr.Get(key).getObject(k)
			iface = append(iface, o)
		case JSONArray:
			iface = append(iface, jr.Get(key).getSlice(k))
		default:
			iface = append(iface, nil)
		}
	}

	return iface
}

/**
 * Library Functions
 */

// Return the data at the associated key. Use empty string ("") to represent the root.
func (jr *JSONReader) getDataByKey(key string) ([]byte, string, []string) {
	if key == "" {
		return jr.rawData, jr.Type, jr.Keys
	}

	var p parsed
	isset := false
	search := jr.parsed

	a := 0
	for b := range key {
		if b == len(key)-1 {
			if p, isset = search[key[a:b+1]]; !isset {
				return nil, "", nil
			}
		}

		if key[b] == '.' {
			if p, isset = search[key[a:b]]; !isset {
				return nil, "", nil
			}

			search = p.children
			a = b + 1
		}
	}

	return p.bytes, p.dtype, p.keys
}

// Return the child node at the associated key. Use empty string ("") to represent the root.
func (jr *JSONReader) getChildByKey(key string) *parsed {

	if key == "" {
		return &parsed{bytes: jr.rawData, dtype: jr.Type, children: jr.parsed, keys: jr.Keys}
	}

	var p parsed
	isset := false
	search := jr.parsed

	a := 0
	for b := range key {
		if b == len(key)-1 {
			if p, isset = search[key[a:b+1]]; !isset {
				return nil
			}
		}

		if key[b] == '.' {
			if p, isset = search[key[a:b]]; !isset {
				return nil
			}

			search = p.children
			a = b + 1
		}
	}

	return &p
}

// Turn a byte string into the given interface type. Objects and Arrays are expensive.
func toIface(b []byte, t string, strict bool) interface{} {
	switch t {
	case JSONInt:
		return toInt(b, t, strict)
	case JSONFloat:
		return toFloat(b, t, strict)
	case JSONBool:
		return toBool(b, t, strict)
	case JSONString:
		return toString(b, t, strict)
	case JSONObject:
		iface := make(map[string]interface{})
		if IsEmptyObject(b) {
			return iface
		}

		expectsValue := true
		start := 1
		for start < len(b) {
			v, k, t, pos, err := extractObjectMember(b, start)
			start = findTerminator(b, pos)
			if err != nil {
				panic(err)
			}
			if pos >= len(b) || start < 0 {
				panic(fmt.Errorf("expected value terminator ('}', ']' or ',') at position '%d' in segment '%s'", pos, truncate(b, 50)))
			}

			expectsValue = false
			if b[start-1] == ',' {
				expectsValue = true
			}

			iface[k] = toIface(v, t, strict)
		}

		if expectsValue {
			panic(fmt.Errorf("expected array terminator '}' at position '%d' in segment '%s'", start-1, truncate(b, 50)))
		}

		return iface
	case JSONArray:
		iface := make([]interface{}, 0)
		if IsEmptyArray(b) {
			return iface
		}

		expectsValue := true
		start := 1
		for start < len(b) {
			v, t, pos, err := extractValue(b, start)
			start = findTerminator(b, pos)
			if err != nil {
				panic(err)
			}
			if pos >= len(b) {
				panic(fmt.Errorf("expected value terminator ('}', ']' or ',') at position '%d' in segment '%s'", pos, truncate(b, 50)))
			}

			expectsValue = false
			if b[start-1] == ',' {
				expectsValue = true
			}

			iface = append(iface, toIface(v, t, strict))
		}

		if expectsValue {
			panic(fmt.Errorf("expected array terminator ']' at position '%d' in segment '%s'", start-1, truncate(b, 50)))
		}

		return iface
	default:
		return nil
	}
}

// stringInArray returns whether the given string exists in the provided string slice.
func stringInArray(needle string, haystack []string) bool {
	for _, n := range haystack {
		if n == needle {
			return true
		}
	}

	return false
}
