package gojson

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	extractTestArrays = [][]byte{
		[]byte(`[]`),
		[]byte(`[1, 2, 3, 4]`),
		[]byte(`["a", "b", "c", "d"]`),
		[]byte(`[true, null]`),
		[]byte(`[1.1, 2.2, 3.3, 4.4]`),
		[]byte(`[{"a": 1, "b": 2, "c": 3}]`),
		[]byte(`[["a", 1], ["b", 2], ["c", 3]]`),
		[]byte(`[1, 2.2, "c", true, null, ["a"], {"a": 3}]`),
		[]byte(`["this array of strings has an embedded array[]"]`),
		[]byte(`["this array of strings has an embedded closing bracket ]"]`),
		[]byte(`["this array of strings has an embedded opening bracket ["]`),
		[]byte(`["this array of strings has both, but out of order ] ["]`),
	}

	extractTestObjects = [][]byte{
		[]byte(`{}`),
		[]byte(`{"a": 1, "b": 2, "c": 3}`),
		[]byte(`{"a": [["a", 1], ["b", 2], ["c", 3]], "b": {"b": 3}, "c": null}`),
		[]byte(`{"key_0": "this object has an embedded object{}"}`),
		[]byte(`{"key_1": "this object has an embedded closing bracket }"}`),
		[]byte(`{"key_2": "this object has an embedded opening bracket {"}`),
		[]byte(`{"key_3": "this object has both, but out of order } {"}`),
	}

	extractTestConstants = [][]byte{
		[]byte("true"),
		[]byte("false"),
	}

	extractTestNumbers = [][]byte{
		[]byte("0"),
		[]byte("8"),
		[]byte("17"),
		[]byte("17e83"),
		[]byte("17e-83"),
		[]byte("-19"),
		[]byte("-19e42"),
		[]byte("-19e-42"),
		[]byte("0.0"),
		[]byte("22.025"),
		[]byte("22.025e98"),
		[]byte("-28.7592"),
		[]byte("-28.7592e3221"),
	}

	extractTestStrings = [][]byte{
		[]byte(`"This is a good string"`),
		[]byte("\f\"true\"\n\r"),
		[]byte("\f\"true\n\"\n\r"),
		[]byte(`""`),
		[]byte(`"\n\t\r" `),
		[]byte("\"\n\t\r\""),
		[]byte(`"\"" `),
		[]byte(`"a"`),
	}

	extractTestStringsExpected = [][]byte{
		[]byte(`"This is a good string"`),
		[]byte(`"true"`),
		[]byte("\"true\n\""),
		[]byte(`""`),
		[]byte(`"\n\t\r"`),
		[]byte("\"\n\t\r\""),
		[]byte(`"\""`),
		[]byte(`"a"`),
	}
)

func TestExtractReader(t *testing.T) {
	t.Run("Extract Failure", func(t *testing.T) {
		data := []byte(`This is not json`)
		r, err := ExtractReader(data, "")
		assert.Nil(t, r)
		assert.Equal(t, "invalid character 'T' at position '0' in segment 'This is not json'", err.Error())
	})

	data := []byte(`{"a": "This is string", "b": 123, "c": -17e-83, "d": true, "e": false, "f": null, "g": [1, "st", false], "h": {"1": "ob", "2": 17, "3": [false, true]}}`)

	testCases := []struct {
		label    string
		key      string
		expected string
	}{
		{label: "Key Root", key: "", expected: `{"a": "This is string", "b": 123, "c": -17e-83, "d": true, "e": false, "f": null, "g": [1, "st", false], "h": {"1": "ob", "2": 17, "3": [false, true]}}`},
		{label: "Key A", key: "a", expected: `This is string`},
		{label: "Key B", key: "b", expected: `123`},
		{label: "Key C", key: "c", expected: `-17e-83`},
		{label: "Key D", key: "d", expected: `true`},
		{label: "Key E", key: "e", expected: `false`},
		{label: "Key F", key: "f", expected: ``},
		{label: "Key G", key: "g", expected: `[1, "st", false]`},
		{label: "Key H", key: "h", expected: `{"1": "ob", "2": 17, "3": [false, true]}`},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := ExtractReader(data, "")
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, r.GetString(tc.key))
		})
	}
}

func TestExtractString(t *testing.T) {

	t.Run("ToString Changes things.", func(t *testing.T) {
		data := []byte(`["Kongming, \"Sleeping Dragon\""]`)
		s, err := ExtractString(data, "0")

		assert.Nil(t, err)
		assert.Equal(t, `Kongming, "Sleeping Dragon"`, s)
		assert.Equal(t, []byte(`["Kongming, \"Sleeping Dragon\""]`), data, "Original Array should not be modified!")

	})

	t.Run("Extract Failure", func(t *testing.T) {
		data := []byte(`This is not json`)
		s, err := ExtractString(data, "")
		assert.Equal(t, "", s)
		assert.Equal(t, "invalid character 'T' at position '0' in segment 'This is not json'", err.Error())
	})

	data := []byte(`{"a": "This is string", "b": 123, "c": -17e-83, "d": true, "e": false, "f": null, "g": [1, "st", false], "h": {"1": "ob", "2": 17, "3": [false, true]}}`)

	testCases := []struct {
		label    string
		key      string
		expected string
	}{
		{label: "Key Root", key: "", expected: `{"a": "This is string", "b": 123, "c": -17e-83, "d": true, "e": false, "f": null, "g": [1, "st", false], "h": {"1": "ob", "2": 17, "3": [false, true]}}`},
		{label: "Key A", key: "a", expected: `This is string`},
		{label: "Key B", key: "b", expected: `123`},
		{label: "Key C", key: "c", expected: `-17e-83`},
		{label: "Key D", key: "d", expected: `true`},
		{label: "Key E", key: "e", expected: `false`},
		{label: "Key F", key: "f", expected: ``},
		{label: "Key G", key: "g", expected: `[1, "st", false]`},
		{label: "Key H", key: "h", expected: `{"1": "ob", "2": 17, "3": [false, true]}`},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			v, err := ExtractString(data, tc.key)
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, v)
		})
	}
}

func TestExtractInt(t *testing.T) {
	t.Run("Extract Failure", func(t *testing.T) {
		data := []byte(`This is not json`)
		i, err := ExtractInt(data, "")
		assert.Equal(t, 0, i)
		assert.Equal(t, "invalid character 'T' at position '0' in segment 'This is not json'", err.Error())
	})

	data := []byte(`{"a": "This is string", "b": 123, "c": 19.23, "d": true, "e": false, "f": null, "g": [1, "st", false], "h": {"1": "ob", "2": 17, "3": [false, true]}}`)

	testCases := []struct {
		label    string
		key      string
		expected int
	}{
		{label: "Key Root", key: "", expected: 0},
		{label: "Key A", key: "a", expected: 0},
		{label: "Key B", key: "b", expected: 123},
		{label: "Key C", key: "c", expected: 19},
		{label: "Key D", key: "d", expected: 1},
		{label: "Key E", key: "e", expected: 0},
		{label: "Key F", key: "f", expected: 0},
		{label: "Key G", key: "g", expected: 0},
		{label: "Key H", key: "h", expected: 0},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			v, err := ExtractInt(data, tc.key)
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, v)
		})
	}
}

func TestExtractFloat(t *testing.T) {
	t.Run("Extract Failure", func(t *testing.T) {
		data := []byte(`This is not json`)
		f, err := ExtractFloat(data, "")
		assert.Equal(t, 0.0, f)
		assert.Equal(t, "invalid character 'T' at position '0' in segment 'This is not json'", err.Error())
	})

	data := []byte(`{"a": "This is string", "b": 123, "c": 19.23, "d": true, "e": false, "f": null, "g": [1, "st", false], "h": {"1": "ob", "2": 17, "3": [false, true]}}`)

	testCases := []struct {
		label    string
		key      string
		expected float64
	}{
		{label: "Key Root", key: "", expected: 0},
		{label: "Key A", key: "a", expected: 0},
		{label: "Key B", key: "b", expected: 123.0},
		{label: "Key C", key: "c", expected: 19.23},
		{label: "Key D", key: "d", expected: 1.0},
		{label: "Key E", key: "e", expected: 0},
		{label: "Key F", key: "f", expected: 0},
		{label: "Key G", key: "g", expected: 0},
		{label: "Key H", key: "h", expected: 0},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			v, err := ExtractFloat(data, tc.key)
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, v)
		})
	}
}

func TestExtractBool(t *testing.T) {
	t.Run("Extract Failure", func(t *testing.T) {
		data := []byte(`This is not json`)
		b, err := ExtractBool(data, "")
		assert.Equal(t, false, b)
		assert.Equal(t, "invalid character 'T' at position '0' in segment 'This is not json'", err.Error())
	})

	data := []byte(`{"a": "This is string", "b": 123, "c": 19.23, "d": true, "e": false, "f": null, "g": [1, "st", false], "h": {"1": "ob", "2": 17, "3": [false, true]}}`)

	testCases := []struct {
		label    string
		key      string
		expected bool
	}{
		{label: "Key Root", key: "", expected: false},
		{label: "Key A", key: "a", expected: false},
		{label: "Key B", key: "b", expected: true},
		{label: "Key C", key: "c", expected: true},
		{label: "Key D", key: "d", expected: true},
		{label: "Key E", key: "e", expected: false},
		{label: "Key F", key: "f", expected: false},
		{label: "Key G", key: "g", expected: false},
		{label: "Key H", key: "h", expected: false},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			v, err := ExtractBool(data, tc.key)
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, v)
		})
	}
}

func TestExtractInterface(t *testing.T) {
	t.Run("Extract Failure", func(t *testing.T) {
		data := []byte(`This is not json`)
		i, dt, err := ExtractInterface(data, "")
		assert.Equal(t, interface{}(nil), i)
		assert.Equal(t, "", dt)
		assert.Equal(t, "invalid character 'T' at position '0' in segment 'This is not json'", err.Error())
	})

	data := []byte(`{"a": "This is string", "b": 123, "c": 19.23, "d": true, "e": false, "f": null, "g": [1, "st", false], "h": {"1": "ob", "2": 17, "3": [false, true]}}`)

	testCases := []struct {
		label    string
		key      string
		expected interface{}
		dType    string
	}{
		// {label: "Key Root", key: "", expected: false},
		{label: "Key A", key: "a", dType: "string", expected: "This is string"},
		{label: "Key B", key: "b", dType: "int", expected: 123},
		{label: "Key C", key: "c", dType: JSONFloat, expected: 19.23},
		{label: "Key D", key: "d", dType: "bool", expected: true},
		{label: "Key E", key: "e", dType: "bool", expected: false},
		{label: "Key F", key: "f", dType: "null", expected: interface{}(nil)},
		{label: "Key G", key: "g", dType: "array", expected: []interface{}{1, "st", false}},
		{label: "Key H", key: "h", dType: "object", expected: map[string]interface{}{"1": "ob", "2": 17, "3": []interface{}{false, true}}},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			v, dt, err := ExtractInterface(data, tc.key)
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, v)
			assert.Equal(t, tc.dType, dt)
		})
	}
}

func TestExtract(t *testing.T) {
	t.Run("Empty Input", func(t *testing.T) {
		v, dt, err := Extract([]byte{}, "things.stuff")
		assert.Equal(t, []byte(nil), v)
		assert.Equal(t, "", dt)
		assert.Equal(t, ErrEmpty, err)
	})

	t.Run("Invalid Key", func(t *testing.T) {
		v, dt, err := Extract([]byte(largeJSONTestBlob), "things.stuff")
		assert.Equal(t, []byte(nil), v)
		assert.Equal(t, "", dt)
		assert.Equal(t, "requested key 'things.stuff' doesn't exist", err.Error())
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		v, dt, err := Extract([]byte(`http://not.valid.json/`), "things.stuff")
		assert.Equal(t, []byte(nil), v)
		assert.Equal(t, "", dt)
		assert.Equal(t, "requested key path 'things.stuff' doesn't exist or json is malformed", err.Error())
	})

	t.Run("Extract Empty Array", func(t *testing.T) {
		val, dt, err := Extract([]byte(`{"a": [], "b": {}}`), "a")
		assert.Nil(t, err)
		assert.Equal(t, []byte(`[]`), val)
		assert.Equal(t, "array", dt)
	})

	t.Run("Extract Empty Object", func(t *testing.T) {
		val, dt, err := Extract([]byte(`{"a": [], "b": {}}`), "b")
		assert.Nil(t, err)
		assert.Equal(t, []byte(`{}`), val)
		assert.Equal(t, "object", dt)
	})

	t.Run("Extract From Invalid JSON", func(t *testing.T) {
		v, dt, err := Extract([]byte(`"p_0":{"pfId":"p_0","creationDate":1423681272350,"lastUpdated":1423681272350,"default_portfolio":false,"positions":[]}`), "some_key")
		assert.Nil(t, v)
		assert.Equal(t, "", dt)
		assert.Equal(t, "key path provided 'some_key' is invalid for JSON type 'string'", err.Error())
	})
}

func TestExtractValue(t *testing.T) {
	t.Run("Empty Byte Set", func(t *testing.T) {
		val, dt, pos, err := extractValue([]byte{}, 0)
		assert.Equal(t, []byte(nil), val)
		assert.Equal(t, "", dt)
		assert.Equal(t, 0, pos)
		assert.Equal(t, "malformed json provided", err.Error())
	})

	t.Run("Start Negative", func(t *testing.T) {
		val, dt, pos, err := extractValue([]byte{'"', 'a', 'b', 'c', '"'}, -1)
		assert.Equal(t, []byte(nil), val)
		assert.Equal(t, "", dt)
		assert.Equal(t, 0, pos)
		assert.Equal(t, "malformed json provided", err.Error())
	})

	// Test Extract with more complex keys.
	data := []byte(`{"a": [["a", 1], ["b", 2], ["c", 3]], "b": {"b": 4}, "c": null, "d": true}`)

	testCases := []struct {
		label        string
		key          string
		expectedType string
		expectedVal  string
	}{
		{label: "Key Root", key: "", expectedType: "object", expectedVal: `{"a": [["a", 1], ["b", 2], ["c", 3]], "b": {"b": 4}, "c": null, "d": true}`},
		{label: "Key A", key: "a", expectedType: "array", expectedVal: `[["a", 1], ["b", 2], ["c", 3]]`},
		{label: "Key B", key: "b", expectedType: "object", expectedVal: `{"b": 4}`},
		{label: "Key C", key: "c", expectedType: "null", expectedVal: `null`},
		{label: "Key D", key: "d", expectedType: "bool", expectedVal: `true`},
		{label: "Key A.0", key: "a.0", expectedType: "array", expectedVal: `["a", 1]`},
		{label: "Key A.0.0", key: "a.0.0", expectedType: "string", expectedVal: `"a"`},
		{label: "Key A.0.1", key: "a.0.1", expectedType: "int", expectedVal: `1`},
		{label: "Key A.1", key: "a.1", expectedType: "array", expectedVal: `["b", 2]`},
		{label: "Key A.1.0", key: "a.1.0", expectedType: "string", expectedVal: `"b"`},
		{label: "Key A.1.1", key: "a.1.1", expectedType: "int", expectedVal: `2`},
		{label: "Key A.2", key: "a.2", expectedType: "array", expectedVal: `["c", 3]`},
		{label: "Key A.2.0", key: "a.2.0", expectedType: "string", expectedVal: `"c"`},
		{label: "Key A.2.1", key: "a.2.1", expectedType: "int", expectedVal: `3`},
		{label: "Key B.B", key: "b.b", expectedType: "int", expectedVal: `4`},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			v, dt, err := Extract(data, tc.key)
			assert.Nil(t, err)
			assert.Equal(t, []byte(tc.expectedVal), v)
			assert.Equal(t, tc.expectedType, dt)
		})
	}

	// Test Extract when the JSON has a key and a valu with the same name.
	data = []byte(`{"blank": null, "assets": [ { "small_image_height": "360", "icon": "video", "video": "expected value" } ] }`)

	t.Run("KeyAndValueCollision", func(t *testing.T) {
		v, dt, err := Extract(data, "assets.0.video")
		assert.Nil(t, err)
		assert.Equal(t, []byte(`"expected value"`), v)
		assert.Equal(t, "string", dt)
	})

	// Test Extract with arrays.
	for k, i := range extractTestArrays {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			v, dt, err := Extract(i, "")

			assert.Nil(t, err)
			assert.Equal(t, i, v)
			assert.Equal(t, "array", dt)
		})
	}

	// Test Extract with objects.
	for k, i := range extractTestObjects {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			v, dt, err := Extract(i, "")

			assert.Nil(t, err)
			assert.Equal(t, i, v)
			assert.Equal(t, "object", dt)
		})
	}

	// Test Extract with Constants, empty key.
	for k, i := range parseTestConstants {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			v, dt, err := Extract(i, "")

			assert.Nil(t, err)
			assert.Equal(t, i, v)
			assert.Equal(t, "bool", dt)
		})
	}

	// Test Extract with Constants, zero key.
	for k, i := range parseTestConstants {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			v, dt, err := Extract(i, "")

			assert.Nil(t, err)
			assert.Equal(t, i, v)
			assert.Equal(t, "bool", dt)
		})
	}

	t.Run("Extract Null, Empty Key", func(t *testing.T) {
		v, _, err := Extract([]byte(`null`), "")

		assert.Nil(t, err)
		assert.Equal(t, []byte("null"), v)
	})

	t.Run("Extract Null, Zero Key", func(t *testing.T) {
		v, _, err := Extract([]byte(`null`), "")

		assert.Nil(t, err)
		assert.Equal(t, []byte("null"), v)
	})

	// Test Extract with Numbers, empty key
	for k, i := range extractTestNumbers {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			v, dt, err := Extract(i, "")

			assert.Nil(t, err)
			assert.Equal(t, i, v)
			assert.True(t, stringInArray(dt, []string{"int", JSONFloat}))
		})
	}

	// Test Extract with Numbers, zero key
	for k, i := range extractTestNumbers {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			v, dt, err := Extract(i, "")

			assert.Nil(t, err)
			assert.Equal(t, i, v)
			assert.True(t, stringInArray(dt, []string{"int", JSONFloat}))
		})
	}

	// Test Extract with Strings, empty key
	for k, i := range extractTestStrings {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			v, dt, err := Extract(i, "")

			assert.Nil(t, err)
			assert.Equal(t, extractTestStringsExpected[k], v)
			assert.Equal(t, "string", dt)
		})
	}

	// Test Extract with Strings, zero key
	for k, i := range extractTestStrings {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			v, dt, err := Extract(i, "")

			assert.Nil(t, err)
			assert.Equal(t, extractTestStringsExpected[k], v)
			assert.Equal(t, "string", dt)
		})
	}
}

func TestExtractKeyPath(t *testing.T) {
	t.Run("Object with Invalid Key", func(t *testing.T) {
		v, dt, k, err := extractKeyPath([]byte(`{key: "value"}`), "key")
		assert.Equal(t, []byte(nil), v)
		assert.Equal(t, "", dt)
		assert.Equal(t, 0, k)
		assert.Equal(t, `expected object key at position 1 in segment '{key: "value"}'`, err.Error())
	})

	t.Run("Array with invalid JSON", func(t *testing.T) {
		v, dt, k, err := extractKeyPath([]byte(`["key", is not json]`), "3")
		assert.Equal(t, []byte(nil), v)
		assert.Equal(t, "", dt)
		assert.Equal(t, 0, k)
		assert.Equal(t, `invalid character 'i' at position '8' in segment '["key", is not json]'`, err.Error())
	})

	t.Run("Array with Non-Numeric Key", func(t *testing.T) {
		v, dt, k, err := extractKeyPath([]byte(`["key"]`), "is not key")
		assert.Equal(t, []byte(nil), v)
		assert.Equal(t, "", dt)
		assert.Equal(t, 0, k)
		assert.Equal(t, `requested key 'is not key' doesn't exist`, err.Error())
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		v, dt, k, err := extractKeyPath([]byte(`{"is not key": is not json}`), "is key")
		assert.Equal(t, []byte(nil), v)
		assert.Equal(t, "", dt)
		assert.Equal(t, 0, k)
		assert.Equal(t, `invalid character 'i' at position '15' in segment '{"is not key": is not json}'`, err.Error())
	})

	t.Run("No Keys", func(t *testing.T) {
		v, dt, k, err := extractKeyPath([]byte{}, "")
		assert.Equal(t, []byte(nil), v)
		assert.Equal(t, "", dt)
		assert.Equal(t, 0, k)
		assert.Equal(t, `extractKeyPath: no keys to extract`, err.Error())
	})
}

func TestExtractKey(t *testing.T) {
	t.Run("Whitespace after Key", func(t *testing.T) {
		v, k, err := extractKey([]byte(`{"this is a good key"  : "some value"}`), 1)
		assert.Nil(t, err)
		assert.Equal(t, []byte(`this is a good key`), v)
		assert.Equal(t, 24, k)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		v, k, err := extractKey([]byte(`{"this is a good key" s : some value}`), 1)
		assert.Equal(t, []byte(nil), v)
		assert.Equal(t, 0, k)
		assert.Equal(t, `invalid character 's' as position 22 (expecting ':' following object key)`, err.Error())
	})
}

func TestExtractStringNoTerminatingQuote(t *testing.T) {
	v, dt, k, err := extractString([]byte(`"this string isn't terminated`), 0)
	assert.Equal(t, []byte(nil), v)
	assert.Equal(t, "", dt)
	assert.Equal(t, 0, k)
	assert.Equal(t, `expected string not found`, err.Error())
}

func TestExtractConstantIsNotContant(t *testing.T) {
	v, dt, k, err := extractConstant([]byte(`-true`), 0)
	assert.Equal(t, []byte(nil), v)
	assert.Equal(t, "", dt)
	assert.Equal(t, 0, k)
	assert.Equal(t, `expected constant not found`, err.Error())
}

func TestExtractKeyValue(t *testing.T) {
	t.Run("Bad JSON Value", func(t *testing.T) {
		v, k, dt, pos, err := extractKeyValue([]byte(`{"key": bad value}`), 1)
		assert.Equal(t, []byte(nil), v)
		assert.Equal(t, "", k)
		assert.Equal(t, "", dt)
		assert.Equal(t, 0, pos)
		assert.Equal(t, `invalid character 'b' at position '8' in segment '{"key": bad value}' (expected object value)`, err.Error())
	})

	t.Run("Bad JSON Key", func(t *testing.T) {
		v, k, dt, pos, err := extractKeyValue([]byte(`{{not really a key, is it?}:""}`), 0)
		assert.Equal(t, []byte(nil), v)
		assert.Equal(t, "", k)
		assert.Equal(t, "", dt)
		assert.Equal(t, 0, pos)
		assert.Equal(t, `expected object key at position 0 in segment '{{not really a key, is it?}:""}'`, err.Error())
	})

	t.Run("No Object Separator", func(t *testing.T) {
		v, k, dt, pos, err := extractKeyValue([]byte(`{"a":"b" "c":"d"}`), 1)
		assert.Equal(t, []byte(`"b"`), v)
		assert.Equal(t, "a", k)
		assert.Equal(t, "string", dt)
		assert.Equal(t, -1, pos)
		assert.Equal(t, `expected object value terminator ('}', ']' or ',') at position '8' in segment '{"a":"b" "c":"d"}'`, err.Error())
	})
}

func TestExtractArrayValue(t *testing.T) {
	t.Run("Bad JSON Value", func(t *testing.T) {
		v, dt, pos, err := extractArrayValue([]byte(`[Is Not Actually JSON]`), 1)
		assert.Equal(t, []byte(nil), v)
		assert.Equal(t, "", dt)
		assert.Equal(t, 0, pos)
		assert.Equal(t, `invalid character 'I' at position '1' in segment '[Is Not Actually JSON]' (expected array value)`, err.Error())
	})

	t.Run("Extract First Value", func(t *testing.T) {
		v, dt, pos, err := extractArrayValue([]byte(`["a", ["b"]]`), 1)
		assert.Equal(t, []byte(`"a"`), v)
		assert.Equal(t, "string", dt)
		assert.Equal(t, 5, pos)
		assert.Nil(t, err)
	})

	t.Run("Extract Second Value", func(t *testing.T) {
		v, dt, pos, err := extractArrayValue([]byte(`["a", ["b"]]`), 5)
		assert.Equal(t, []byte(`["b"]`), v)
		assert.Equal(t, "array", dt)
		assert.Equal(t, 12, pos)
		assert.Nil(t, err)
	})

	t.Run("Missing Terminator", func(t *testing.T) {
		v, dt, pos, err := extractArrayValue([]byte(`["a" ["b"]]`), 1)
		assert.Equal(t, []byte(`"a"`), v)
		assert.Equal(t, "string", dt)
		assert.Equal(t, -1, pos)
		assert.Equal(t, `expected array value terminator ('}', ']' or ',') at position '4' in segment '["a" ["b"]]'`, err.Error())
	})
}

func TestExtractEscapedBackslash(t *testing.T) {
	data := `{"results":[{"keywords":"\\","canonical":"nbc-world_of_dance:srank_world_finale_front_row-hulu2"}]}`

	id, err := ExtractString([]byte(data), "results.0.canonical")

	assert.Nil(t, err)
	assert.Equal(t, `nbc-world_of_dance:srank_world_finale_front_row-hulu2`, id)
}
