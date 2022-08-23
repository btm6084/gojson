package gojson

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyMembers(t *testing.T) {
	patch := []byte(`{
		"intPatch": 0,
		"emptyStringPatch": "",
		"emptyArrayPatch": [],
		"emptyObjectPatch": {},
		"objectPatch": {
			"nestedEmptyObject": {},
			"hello": false
		}
	}`)

	base := []byte(`{
		"intBase": 1,
		"emptyStringBase":"",
		"emptyArrayBase": [],
		"emptyObjectBase": {},
	}`)

	out, err := MergeJSON(base, patch)
	assert.Nil(t, err)
	assert.True(t, IsJSON(out), string(out))
	assert.JSONEq(t, `{"intBase":1,"emptyStringBase":"","emptyArrayBase":[],"emptyObjectBase":{},"intPatch":0,"emptyStringPatch":"","emptyArrayPatch":[],"emptyObjectPatch":{},"objectPatch":{"nestedEmptyObject":{},"hello":false}}`, string(out), string(out))

}

func TestMerge(t *testing.T) {
	testCases := []struct {
		label    string
		base     string
		patch    string
		expected string
	}{
		{base: `{"a":"b"}`, patch: `{"a":"c"}`, expected: `{"a":"c"}`, label: "string->string"},
		{base: `{"a":"b"}`, patch: `{"a":true}`, expected: `{"a":true}`, label: "string->bool"},
		{base: `{"a":"b"}`, patch: `{"a":421}`, expected: `{"a":421}`, label: "string->int"},
		{base: `{"a":"b"}`, patch: `{"a":3.1415926}`, expected: `{"a":3.1415926}`, label: "string->float"},
		{base: `{"a":"b"}`, patch: `{"a":[1, 2, 3]}`, expected: `{"a":[1, 2, 3]}`, label: "string->array"},
		{base: `{"a":"b"}`, patch: `{"a":{"b": "c"}}`, expected: `{"a":{"b": "c"}}`, label: "string->object"},

		{base: `{"a":14}`, patch: `{"a":17}`, expected: `{"a":17}`, label: "int->int"},
		{base: `{"a":14}`, patch: `{"a":true}`, expected: `{"a":true}`, label: "int->bool"},
		{base: `{"a":14}`, patch: `{"a":"421"}`, expected: `{"a":"421"}`, label: "int->string"},
		{base: `{"a":14}`, patch: `{"a":3.1415926}`, expected: `{"a":3.1415926}`, label: "int->float"},
		{base: `{"a":14}`, patch: `{"a":[1, 2, 3]}`, expected: `{"a":[1, 2, 3]}`, label: "int->array"},
		{base: `{"a":14}`, patch: `{"a":{"b": "c"}}`, expected: `{"a":{"b": "c"}}`, label: "int->object"},

		{base: `{"a":true}`, patch: `{"a":"b"}`, expected: `{"a":"b"}`, label: "bool->string"},
		{base: `{"a":421}`, patch: `{"a":"b"}`, expected: `{"a":"b"}`, label: "int->string"},
		{base: `{"a":3.1415926}`, patch: `{"a":"b"}`, expected: `{"a":"b"}`, label: "float->string"},
		{base: `{"a":[1, 2, 3]}`, patch: `{"a":"b"}`, expected: `{"a":"b"}`, label: "array->string"},
		{base: `{"a":{"b": "c"}}`, patch: `{"a":"b"}`, expected: `{"a":"b"}`, label: "object->string"},

		{base: `{"a":"c"}`, patch: `{"a":42}`, expected: `{"a":42}`, label: "int->int"},
		{base: `{"a":true}`, patch: `{"a":42}`, expected: `{"a":42}`, label: "bool->int"},
		{base: `{"a":421}`, patch: `{"a":42}`, expected: `{"a":42}`, label: "int->int"},
		{base: `{"a":3.1415926}`, patch: `{"a":42}`, expected: `{"a":42}`, label: "float->int"},
		{base: `{"a":[1, 2, 3]}`, patch: `{"a":42}`, expected: `{"a":42}`, label: "array->int"},
		{base: `{"a":{"b": "c"}}`, patch: `{"a":42}`, expected: `{"a":42}`, label: "object->int"},

		{base: `{"a":{"b": "c"}}`, patch: `{"a":{"b": "d", "e": 2.718}}`, expected: `{"a":{"b":"d","e":2.718}}`, label: "object->object"},
		{base: `{"a":{"b": "d", "e": [3.14]}}`, patch: `{"a":{"b": "d", "e": [2.718]}}`, expected: `{"a":{"b":"d","e":[2.718]}}`, label: "object->object"},
		{base: `{"a":{"b": "d", "e": {"pi": 3.14}}}`, patch: `{"a":{"b": "d", "e": [2.718]}}`, expected: `{"a":{"b":"d","e":[2.718]}}`, label: "object->object"},

		{base: `{"a":{"b": "d", "e": {"pi": 3.14}}, "b": 14, "c": [123]}`, patch: `{"a":"override"}`, expected: `{"a":"override","b":14,"c":[123]}`, label: "override object"},
		{base: `{"a":{"b": "d", "e": {"pi": 3.14}}, "b": 14, "c": [123]}`, patch: `{"c":[456]}`, expected: `{"a":{"b":"d","e":{"pi":3.14}},"b":14,"c":[456]}`, label: "override arrays"},
	}

	for n, tc := range testCases {
		t.Run("Merge "+strconv.Itoa(n), func(t *testing.T) {
			out, err := MergeJSON([]byte(tc.base), []byte(tc.patch))
			assert.Nil(t, err)

			assert.JSONEq(t, tc.expected, string(out), tc.label)
		})
	}

	errorTestCases := []struct {
		label string
		base  string
		patch string
	}{
		{base: `"a"`, patch: `{"a":"b"}`, label: "string"},
		{base: `14`, patch: `{"a":"b"}`, label: "int"},
		{base: `3.14`, patch: `{"a":"b"}`, label: "float"},
		{base: `false`, patch: `{"a":"b"}`, label: "false"},
		{base: `true`, patch: `{"a":"b"}`, label: "true"},
		{base: `null`, patch: `{"a":"b"}`, label: "null"},
		{base: `[1, 2, 3]`, patch: `{"a":"b"}`, label: "array"},

		{base: `{"a":"b"}`, patch: `"a"`, label: "string"},
		{base: `{"a":"b"}`, patch: `14`, label: "int"},
		{base: `{"a":"b"}`, patch: `3.14`, label: "float"},
		{base: `{"a":"b"}`, patch: `false`, label: "false"},
		{base: `{"a":"b"}`, patch: `true`, label: "true"},
		{base: `{"a":"b"}`, patch: `null`, label: "null"},
		{base: `{"a":"b"}`, patch: `[1, 2, 3]`, label: "array"},
	}

	for n, tc := range errorTestCases {
		t.Run("Error "+strconv.Itoa(n), func(t *testing.T) {
			out, err := MergeJSON([]byte(tc.base), []byte(tc.patch))
			assert.Nil(t, out)
			assert.Equal(t, ErrMergeType, err)
		})
	}
}
