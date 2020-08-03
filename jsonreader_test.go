package gojson

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewJSONReader(t *testing.T) {
	t.Run("Empty ByteString", func(t *testing.T) {
		r, err := NewJSONReader([]byte{})
		assert.True(t, r.Empty)
		assert.Equal(t, "No JSON Provided", err.Error())
	})

	t.Run("No Array Terminal Char", func(t *testing.T) {
		var r *JSONReader
		var err error
		defer func() {
			assert.True(t, r.Empty)
			assert.Equal(t, `expected ']', found '"' at position 14`, err.Error())
		}()
		defer PanicRecovery(&err)

		r, err = NewJSONReader([]byte(`["Invalid JSON"`))
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		r, err := NewJSONReader([]byte(`Invalid JSON`))
		assert.True(t, r.Empty)
		assert.Nil(t, err)
	})

	t.Run("Valid JSON", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)
		assert.False(t, r.Empty)
		assert.Len(t, r.Keys, 13)
	})
}

func TestKeyExists(t *testing.T) {
	t.Run("Label", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)
		assert.True(t, r.KeyExists("empty_string"))
		assert.True(t, r.KeyExists("string"))
		assert.True(t, r.KeyExists("int"))
		assert.True(t, r.KeyExists("bool"))
		assert.True(t, r.KeyExists("null"))
		assert.True(t, r.KeyExists("float"))
		assert.True(t, r.KeyExists("string_slice"))
		assert.True(t, r.KeyExists("bool_slice"))
		assert.True(t, r.KeyExists("int_slice"))
		assert.True(t, r.KeyExists("float_slice"))
		assert.True(t, r.KeyExists("object"))
		assert.True(t, r.KeyExists("objects"))
		assert.True(t, r.KeyExists("objects.2"))
		assert.True(t, r.KeyExists("complex"))
		assert.True(t, r.KeyExists("complex.0"))
		assert.True(t, r.KeyExists("complex.5.c"))
		assert.False(t, r.KeyExists("complex.5.c.d"))
	})
}

func TestGet(t *testing.T) {
	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		c := r.Get("Invalid Key")
		assert.True(t, c.Empty)
	})

	t.Run("Simple Type", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		c := r.Get("string")
		assert.False(t, c.Empty)
		assert.Len(t, c.Keys, 1)
	})

	t.Run("Complex Type", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		c := r.Get("objects")
		assert.False(t, c.Empty)
		assert.Len(t, c.Keys, 3)
	})

	t.Run("Simple Type Nested", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		c := r.Get("complex.5.c")
		assert.False(t, c.Empty)
		assert.Len(t, c.Keys, 1)
		assert.Equal(t, c.GetString(""), `d`)
	})

	t.Run("Complex Type Nested", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		c := r.Get("objects.1")
		assert.False(t, c.Empty)
		assert.Len(t, c.Keys, 2)
		assert.Equal(t, c.GetString(`i`), `j`)
		assert.Equal(t, c.GetString(`k`), `l`)
	})
}

func TestGetCollection(t *testing.T) {
	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		c := r.GetCollection("Invalid Key")
		assert.Equal(t, []JSONReader(nil), c)
	})

	t.Run("Simple Type", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		c := r.GetCollection("string")
		assert.Len(t, c, 1)
	})

	t.Run("Complex Type", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		c := r.GetCollection("complex")
		assert.Len(t, c, 7)
		assert.Equal(t, `a`, c[0].GetString(``))
		assert.Equal(t, 2, c[1].GetInt(``))
		assert.Equal(t, []byte(`null`), c[2].GetByteSlice(``))
		assert.Equal(t, false, c[3].GetBool(``))
		assert.Equal(t, 2.2, c[4].GetFloat(``))
		assert.Equal(t, `d`, c[5].Get(`c`).GetString(``))
	})
}

func TestGetString(t *testing.T) {
	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v := r.GetString("Invalid Key")
		assert.Equal(t, "", v)
	})

	t.Run("Escape Characters", func(t *testing.T) {
		b := []byte{'"', 'a', '\u0026', 'b', '\u003c', 'c', '\\', '"', '\u003e', '\\', '"', 'd', '"'}
		r, err := NewJSONReader(b)
		assert.Nil(t, err)

		assert.Equal(t, `a&b<c">"d`, r.GetString(""))
	})

	testCases := []struct {
		key string
		exp string
	}{
		{key: "empty_string", exp: ""},
		{key: "string", exp: `some string`},
		{key: "int", exp: `17`},
		{key: "bool", exp: `true`},
		{key: "null", exp: ``},
		{key: "float", exp: `22.83`},
		{key: "string_slice", exp: `[ "a", "b", "c", "d", "e", "t" ]`},
		{key: "bool_slice", exp: `[ true, false, true, false ]`},
		{key: "int_slice", exp: `[ -1, 0, 1, 2, 3, 4 ]`},
		{key: "float_slice", exp: `[ -1.1, 0.0, 1.1, 2.2, 3.3 ]`},
		{key: "object", exp: `{ "a": "b", "c": "d" }`},
		{key: "objects", exp: "[\n\t\t\t{ \"e\": \"f\", \"g\": \"h\" },\n\t\t\t{ \"i\": \"j\", \"k\": \"l\" },\n\t\t\t{ \"m\": \"n\", \"o\": \"t\" }\n\t\t]"},
		{key: "objects.2.o", exp: "t"},
		{key: "complex", exp: `[ "a", 2, null, false, 2.2, { "c": "d", "empty_string": "" }, [ "s" ] ]`},
		{key: "complex.5.c", exp: `d`},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			r, err := NewJSONReader(readerTestData)
			assert.Nil(t, err)

			v := r.GetString(tc.key)
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestToString(t *testing.T) {

	t.Run("ToString Examples", func(t *testing.T) {
		testCases := []struct {
			label string
			data  []byte
			exp   string
		}{
			{label: "EmptyString", data: tdEmptyString, exp: ""},
			{label: "String", data: tdString, exp: `some string`},
			{label: "Int", data: tdInt, exp: `17`},
			{label: "Bool", data: tdBool, exp: `true`},
			{label: "Null", data: tdNull, exp: ``},
			{label: "Float", data: tdFloat, exp: `22.83`},
			{label: "StringSlice", data: tdStringSlice, exp: `[ "a", "b", "c", "d", "e", "t" ]`},
			{label: "BoolSlice", data: tdBoolSlice, exp: `[ true, false, true, false ]`},
			{label: "IntSlice", data: tdIntSlice, exp: `[ -1, 0, 1, 2, 3, 4 ]`},
			{label: "FloatSlice", data: tdFloatSlice, exp: `[ -1.1, 0.0, 1.1, 2.2, 3.3 ]`},
			{label: "Object", data: tdObject, exp: `{ "a": "b", "c": "d" }`},
			{label: "Objects", data: tdObjects, exp: `[ { "e": "f", "g": "h" }, { "i": "j", "k": "l" }, { "m": "n", "o": "t" } ]`},
			{label: "Complex", data: tdComplex, exp: `[ "a", 2, null, false, 2.2, { "c": "d", "empty_string": "" }, [ "s" ] ]`},
		}

		for _, tc := range testCases {
			t.Run(tc.label, func(t *testing.T) {
				r, err := NewJSONReader(tc.data)
				assert.Nil(t, err)

				v := r.ToString()
				assert.Equal(t, tc.exp, v)
			})
		}
	})

	t.Run("Dealing With Unicode", func(t *testing.T) {
		testCases := []struct {
			data     string
			expected string
		}{
			{`"We\u2019ve been had"`, `We’ve been had`},
			{`"\u2018Hello there.\u2019"`, `‘Hello there.’`},
			{`"\u003cGeneral Kenobi\u003e"`, `<General Kenobi>`},
			{`"Shoots \u0026 Giggles"`, `Shoots & Giggles`},
			{`"Quoted String"`, `Quoted String`},
		}

		for i, tc := range testCases {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				r, err := NewJSONReader([]byte(tc.data))
				assert.Nil(t, err)

				v := r.ToString()
				assert.Equal(t, tc.expected, v)
			})
		}
	})

	t.Run("Pointer Safety", func(t *testing.T) {
		data := []byte(`"Yes I want the cheesy poofs"`)
		r, err := NewJSONReader(data)
		assert.Nil(t, err)

		s := r.ToString()
		assert.Equal(t, "Yes I want the cheesy poofs", s)

		// Change the data slice
		data[12] = 'N'
		data[13] = 'o'
		data[14] = ' '

		// Make sure s hasn't changed.
		assert.Equal(t, "Yes I want the cheesy poofs", s)
		assert.Equal(t, `"Yes I want No  cheesy poofs"`, string(data))
	})
}

func TestGetStringSlice(t *testing.T) {
	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v := r.GetStringSlice("Invalid Key")
		assert.Equal(t, []string(nil), v)
	})

	testCases := []struct {
		key string
		exp []string
	}{
		{key: "empty_string", exp: []string{""}},
		{key: "string", exp: []string{`some string`}},
		{key: "int", exp: []string{`17`}},
		{key: "bool", exp: []string{`true`}},
		{key: "null", exp: []string{``}},
		{key: "float", exp: []string{`22.83`}},
		{key: "string_slice", exp: []string{"a", "b", "c", "d", "e", "t"}},
		{key: "bool_slice", exp: []string{"true", "false", "true", "false"}},
		{key: "int_slice", exp: []string{"-1", "0", "1", "2", "3", "4"}},
		{key: "float_slice", exp: []string{"-1.1", "0.0", "1.1", "2.2", "3.3"}},
		{key: "object", exp: []string{"b", "d"}},
		{key: "objects", exp: []string{`{ "e": "f", "g": "h" }`, `{ "i": "j", "k": "l" }`, `{ "m": "n", "o": "t" }`}},
		{key: "objects.2.o", exp: []string{`t`}},
		{key: "complex", exp: []string{"a", "2", "", "false", "2.2", `{ "c": "d", "empty_string": "" }`, `[ "s" ]`}},
		{key: "complex.5.c", exp: []string{`d`}},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			r, err := NewJSONReader(readerTestData)
			assert.Nil(t, err)

			v := r.GetStringSlice(tc.key)
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestToStringSlice(t *testing.T) {
	testCases := []struct {
		label string
		data  []byte
		exp   []string
	}{
		{label: "EmptyString", data: tdEmptyString, exp: []string{""}},
		{label: "String", data: tdString, exp: []string{`some string`}},
		{label: "Int", data: tdInt, exp: []string{`17`}},
		{label: "Bool", data: tdBool, exp: []string{`true`}},
		{label: "Null", data: tdNull, exp: []string{``}},
		{label: "Float", data: tdFloat, exp: []string{`22.83`}},
		{label: "StringSlice", data: tdStringSlice, exp: []string{"a", "b", "c", "d", "e", "t"}},
		{label: "BoolSlice", data: tdBoolSlice, exp: []string{"true", "false", "true", "false"}},
		{label: "IntSlice", data: tdIntSlice, exp: []string{"-1", "0", "1", "2", "3", "4"}},
		{label: "FloatSlice", data: tdFloatSlice, exp: []string{"-1.1", "0.0", "1.1", "2.2", "3.3"}},
		{label: "Object", data: tdObject, exp: []string{"b", "d"}},
		{label: "Objects", data: tdObjects, exp: []string{`{ "e": "f", "g": "h" }`, `{ "i": "j", "k": "l" }`, `{ "m": "n", "o": "t" }`}},
		{label: "Complex", data: tdComplex, exp: []string{"a", "2", "", "false", "2.2", `{ "c": "d", "empty_string": "" }`, `[ "s" ]`}},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := NewJSONReader(tc.data)
			assert.Nil(t, err)

			v := r.ToStringSlice()
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestToMapStringString(t *testing.T) {
	testCases := []struct {
		label string
		data  []byte
		exp   map[string]string
	}{
		{label: "EmptyString", data: tdEmptyString, exp: map[string]string{"0": ""}},
		{label: "String", data: tdString, exp: map[string]string{"0": `some string`}},
		{label: "Int", data: tdInt, exp: map[string]string{"0": `17`}},
		{label: "Bool", data: tdBool, exp: map[string]string{"0": `true`}},
		{label: "Null", data: tdNull, exp: map[string]string{}},
		{label: "Float", data: tdFloat, exp: map[string]string{"0": `22.83`}},
		{label: "StringSlice", data: tdStringSlice, exp: map[string]string{"0": "a", "1": "b", "2": "c", "3": "d", "4": "e", "5": "t"}},
		{label: "BoolSlice", data: tdBoolSlice, exp: map[string]string{"0": "true", "1": "false", "2": "true", "3": "false"}},
		{label: "IntSlice", data: tdIntSlice, exp: map[string]string{"0": "-1", "1": "0", "2": "1", "3": "2", "4": "3", "5": "4"}},
		{label: "FloatSlice", data: tdFloatSlice, exp: map[string]string{"0": "-1.1", "1": "0.0", "2": "1.1", "3": "2.2", "4": "3.3"}},
		{label: "Object", data: tdObject, exp: map[string]string{"a": "b", "c": "d"}},
		{label: "Objects", data: tdObjects, exp: map[string]string{"0": `{ "e": "f", "g": "h" }`, "1": `{ "i": "j", "k": "l" }`, "2": `{ "m": "n", "o": "t" }`}},
		{label: "Complex", data: tdComplex, exp: map[string]string{"4": "2.2", "5": `{ "c": "d", "empty_string": "" }`, "6": `[ "s" ]`, "0": "a", "1": "2", "2": "", "3": "false"}},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := NewJSONReader(tc.data)
			assert.Nil(t, err)

			v := r.ToMapStringString()
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestGetBool(t *testing.T) {
	t.Run("Float Overflow", func(t *testing.T) {
		r, err := NewJSONReader([]byte(`1.0e12121212121212121212121212`))
		assert.Nil(t, err)

		assert.Equal(t, false, r.GetBool(""))
	})

	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v := r.GetBool("Invalid Key")
		assert.Equal(t, false, v)
	})

	testCases := []struct {
		key string
		exp bool
	}{
		{key: "empty_bool", exp: false},
		{key: "string", exp: false},
		{key: "int", exp: true},
		{key: "bool", exp: true},
		{key: "null", exp: false},
		{key: "float", exp: true},
		{key: "bool_slice", exp: false},
		{key: "bool_slice", exp: false},
		{key: "int_slice", exp: false},
		{key: "float_slice", exp: false},
		{key: "object", exp: false},
		{key: "objects", exp: false},
		{key: "objects.2.o", exp: true},
		{key: "complex", exp: false},
		{key: "complex.3", exp: false},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			r, err := NewJSONReader(readerTestData)
			assert.Nil(t, err)

			v := r.GetBool(tc.key)
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestToBool(t *testing.T) {
	testCases := []struct {
		label string
		data  []byte
		exp   bool
	}{
		{label: "EmptyString", data: tdEmptyString, exp: false},
		{label: "String", data: tdString, exp: false},
		{label: "Int", data: tdInt, exp: true},
		{label: "Bool", data: tdBool, exp: true},
		{label: "Null", data: tdNull, exp: false},
		{label: "Float", data: tdFloat, exp: true},
		{label: "StringSlice", data: tdStringSlice, exp: false},
		{label: "BoolSlice", data: tdBoolSlice, exp: false},
		{label: "IntSlice", data: tdIntSlice, exp: false},
		{label: "FloatSlice", data: tdFloatSlice, exp: false},
		{label: "Object", data: tdObject, exp: false},
		{label: "Objects", data: tdObjects, exp: false},
		{label: "Complex", data: tdComplex, exp: false},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := NewJSONReader(tc.data)
			assert.Nil(t, err)

			v := r.ToBool()
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestGetBoolSlice(t *testing.T) {
	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v := r.GetBoolSlice("Invalid Key")
		assert.Equal(t, []bool(nil), v)
	})

	testCases := []struct {
		key string
		exp []bool
	}{
		{key: "empty_bool", exp: []bool(nil)},
		{key: "string", exp: []bool{false}},
		{key: "int", exp: []bool{true}},
		{key: "bool", exp: []bool{true}},
		{key: "null", exp: []bool{false}},
		{key: "float", exp: []bool{true}},
		{key: "string_slice", exp: []bool{false, false, false, false, false, true}},
		{key: "bool_slice", exp: []bool{true, false, true, false}},
		{key: "int_slice", exp: []bool{true, false, true, true, true, true}},
		{key: "float_slice", exp: []bool{true, false, true, true, true}},
		{key: "object", exp: []bool{false, false}},
		{key: "objects", exp: []bool{false, false, false}},
		{key: "objects.2.o", exp: []bool{true}},
		{key: "complex", exp: []bool{false, true, false, false, true, false, false}},
		{key: "complex.3", exp: []bool{false}},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			r, err := NewJSONReader(readerTestData)
			assert.Nil(t, err)

			v := r.GetBoolSlice(tc.key)
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestToBoolSlice(t *testing.T) {
	testCases := []struct {
		label string
		data  []byte
		exp   []bool
	}{
		{label: "EmptyString", data: tdEmptyString, exp: []bool{false}},
		{label: "String", data: tdString, exp: []bool{false}},
		{label: "Int", data: tdInt, exp: []bool{true}},
		{label: "Bool", data: tdBool, exp: []bool{true}},
		{label: "Null", data: tdNull, exp: []bool{false}},
		{label: "Float", data: tdFloat, exp: []bool{true}},
		{label: "StringSlice", data: tdStringSlice, exp: []bool{false, false, false, false, false, true}},
		{label: "BoolSlice", data: tdBoolSlice, exp: []bool{true, false, true, false}},
		{label: "IntSlice", data: tdIntSlice, exp: []bool{true, false, true, true, true, true}},
		{label: "FloatSlice", data: tdFloatSlice, exp: []bool{true, false, true, true, true}},
		{label: "Object", data: tdObject, exp: []bool{false, false}},
		{label: "Objects", data: tdObjects, exp: []bool{false, false, false}},
		{label: "Complex", data: tdComplex, exp: []bool{false, true, false, false, true, false, false}},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := NewJSONReader(tc.data)
			assert.Nil(t, err)

			v := r.ToBoolSlice()
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestToMapStringBool(t *testing.T) {
	testCases := []struct {
		label string
		data  []byte
		exp   map[string]bool
	}{
		{label: "EmptyString", data: tdEmptyString, exp: map[string]bool{"0": false}},
		{label: "String", data: tdString, exp: map[string]bool{"0": false}},
		{label: "Int", data: tdInt, exp: map[string]bool{"0": true}},
		{label: "Bool", data: tdBool, exp: map[string]bool{"0": true}},
		{label: "Null", data: tdNull, exp: map[string]bool{}},
		{label: "Float", data: tdFloat, exp: map[string]bool{"0": true}},
		{label: "StringSlice", data: tdStringSlice, exp: map[string]bool{"0": false, "1": false, "2": false, "3": false, "4": false, "5": true}},
		{label: "BoolSlice", data: tdBoolSlice, exp: map[string]bool{"0": true, "1": false, "2": true, "3": false}},
		{label: "IntSlice", data: tdIntSlice, exp: map[string]bool{"0": true, "1": false, "2": true, "3": true, "4": true, "5": true}},
		{label: "FloatSlice", data: tdFloatSlice, exp: map[string]bool{"0": true, "1": false, "2": true, "3": true, "4": true}},
		{label: "Object", data: tdObject, exp: map[string]bool{"a": false, "c": false}},
		{label: "Objects", data: tdObjects, exp: map[string]bool{"0": false, "1": false, "2": false}},
		{label: "Complex", data: tdComplex, exp: map[string]bool{"0": false, "1": true, "2": false, "3": false, "4": true, "5": false, "6": false}},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := NewJSONReader(tc.data)
			assert.Nil(t, err)

			v := r.ToMapStringBool()
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestGetInt(t *testing.T) {
	t.Run("Float Overflow", func(t *testing.T) {
		r, err := NewJSONReader([]byte(`1.0e12121212121212121212121212`))
		assert.Nil(t, err)

		assert.Equal(t, 0, r.GetInt(""))
	})

	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v := r.GetInt("Invalid Key")
		assert.Equal(t, 0, v)
	})

	testCases := []struct {
		key string
		exp int
	}{
		{key: "empty_int", exp: 0},
		{key: "string", exp: 0},
		{key: "bool", exp: 1},
		{key: "int", exp: 17},
		{key: "null", exp: 0},
		{key: "float", exp: 22},
		{key: "int_slice", exp: 0},
		{key: "int_slice", exp: 0},
		{key: "int_slice", exp: 0},
		{key: "float_slice", exp: 0},
		{key: "object", exp: 0},
		{key: "objects", exp: 0},
		{key: "objects.2.o", exp: 0},
		{key: "complex", exp: 0},
		{key: "complex.1", exp: 2},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			r, err := NewJSONReader(readerTestData)
			assert.Nil(t, err)

			v := r.GetInt(tc.key)
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestToInt(t *testing.T) {
	testCases := []struct {
		label string
		data  []byte
		exp   int
	}{
		{label: "EmptyString", data: tdEmptyString, exp: 0},
		{label: "String", data: tdString, exp: 0},
		{label: "Bool", data: tdBool, exp: 1},
		{label: "Int", data: tdInt, exp: 17},
		{label: "Null", data: tdNull, exp: 0},
		{label: "Float", data: tdFloat, exp: 22},
		{label: "StringSlice", data: tdStringSlice, exp: 0},
		{label: "BoolSlice", data: tdBoolSlice, exp: 0},
		{label: "IntSlice", data: tdIntSlice, exp: 0},
		{label: "FloatSlice", data: tdFloatSlice, exp: 0},
		{label: "Object", data: tdObject, exp: 0},
		{label: "Objects", data: tdObjects, exp: 0},
		{label: "Complex", data: tdComplex, exp: 0},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := NewJSONReader(tc.data)
			assert.Nil(t, err)

			v := r.ToInt()
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestGetIntSlice(t *testing.T) {
	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v := r.GetIntSlice("Invalid Key")
		assert.Equal(t, []int(nil), v)
	})

	testCases := []struct {
		key string
		exp []int
	}{
		{key: "empty_int", exp: []int(nil)},
		{key: "string", exp: []int{0}},
		{key: "bool", exp: []int{1}},
		{key: "int", exp: []int{17}},
		{key: "null", exp: []int{0}},
		{key: "float", exp: []int{22}},
		{key: "string_slice", exp: []int{0, 0, 0, 0, 0, 0}},
		{key: "bool_slice", exp: []int{1, 0, 1, 0}},
		{key: "int_slice", exp: []int{-1, 0, 1, 2, 3, 4}},
		{key: "float_slice", exp: []int{-1, 0, 1, 2, 3}},
		{key: "object", exp: []int{0, 0}},
		{key: "objects", exp: []int{0, 0, 0}},
		{key: "objects.2.o", exp: []int{0}},
		{key: "complex", exp: []int{0, 2, 0, 0, 2, 0, 0}},
		{key: "complex.1", exp: []int{2}},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			r, err := NewJSONReader(readerTestData)
			assert.Nil(t, err)

			v := r.GetIntSlice(tc.key)
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestToIntSlice(t *testing.T) {
	testCases := []struct {
		label string
		data  []byte
		exp   []int
	}{
		{label: "EmptyString", data: tdEmptyString, exp: []int{0}},
		{label: "String", data: tdString, exp: []int{0}},
		{label: "Bool", data: tdBool, exp: []int{1}},
		{label: "Int", data: tdInt, exp: []int{17}},
		{label: "Null", data: tdNull, exp: []int{0}},
		{label: "Float", data: tdFloat, exp: []int{22}},
		{label: "StringSlice", data: tdStringSlice, exp: []int{0, 0, 0, 0, 0, 0}},
		{label: "BoolSlice", data: tdBoolSlice, exp: []int{1, 0, 1, 0}},
		{label: "IntSlice", data: tdIntSlice, exp: []int{-1, 0, 1, 2, 3, 4}},
		{label: "FloatSlice", data: tdFloatSlice, exp: []int{-1, 0, 1, 2, 3}},
		{label: "Object", data: tdObject, exp: []int{0, 0}},
		{label: "Objects", data: tdObjects, exp: []int{0, 0, 0}},
		{label: "Complex", data: tdComplex, exp: []int{0, 2, 0, 0, 2, 0, 0}},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := NewJSONReader(tc.data)
			assert.Nil(t, err)

			v := r.ToIntSlice()
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestToMapStringInt(t *testing.T) {
	testCases := []struct {
		label string
		data  []byte
		exp   map[string]int
	}{
		{label: "EmptyString", data: tdEmptyString, exp: map[string]int{"0": 0}},
		{label: "String", data: tdString, exp: map[string]int{"0": 0}},
		{label: "Bool", data: tdBool, exp: map[string]int{"0": 1}},
		{label: "Int", data: tdInt, exp: map[string]int{"0": 17}},
		{label: "Null", data: tdNull, exp: map[string]int{}},
		{label: "Float", data: tdFloat, exp: map[string]int{"0": 22}},
		{label: "StringSlice", data: tdStringSlice, exp: map[string]int{"0": 0, "1": 0, "2": 0, "3": 0, "4": 0, "5": 0}},
		{label: "BoolSlice", data: tdBoolSlice, exp: map[string]int{"0": 1, "1": 0, "2": 1, "3": 0}},
		{label: "IntSlice", data: tdIntSlice, exp: map[string]int{"0": -1, "1": 0, "2": 1, "3": 2, "4": 3, "5": 4}},
		{label: "FloatSlice", data: tdFloatSlice, exp: map[string]int{"0": -1, "1": 0, "2": 1, "3": 2, "4": 3}},
		{label: "Object", data: tdObject, exp: map[string]int{"a": 0, "c": 0}},
		{label: "Objects", data: tdObjects, exp: map[string]int{"0": 0, "1": 0, "2": 0}},
		{label: "Complex", data: tdComplex, exp: map[string]int{"0": 0, "1": 2, "2": 0, "3": 0, "4": 2, "5": 0, "6": 0}},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := NewJSONReader(tc.data)
			assert.Nil(t, err)

			v := r.ToMapStringInt()
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestGetFloat(t *testing.T) {
	t.Run("Float Overflow", func(t *testing.T) {
		r, err := NewJSONReader([]byte(`1.0e12121212121212121212121212`))
		assert.Nil(t, err)

		assert.Equal(t, 0.0, r.GetFloat(""))
	})

	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v := r.GetFloat("Invalid Key")
		assert.Equal(t, 0.0, v)
	})

	testCases := []struct {
		key string
		exp float64
	}{
		{key: "empty_float", exp: 0.0},
		{key: "string", exp: 0.0},
		{key: "bool", exp: 1},
		{key: "int", exp: 17.0},
		{key: "null", exp: 0.0},
		{key: "float", exp: 22.83},
		{key: "int_slice", exp: 0.0},
		{key: "float_slice", exp: 0.0},
		{key: "object", exp: 0.0},
		{key: "objects", exp: 0.0},
		{key: "objects.2.o", exp: 0.0},
		{key: "complex", exp: 0.0},
		{key: "complex.1", exp: 2.0},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			r, err := NewJSONReader(readerTestData)
			assert.Nil(t, err)

			v := r.GetFloat(tc.key)
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestToFloat(t *testing.T) {
	testCases := []struct {
		label string
		data  []byte
		exp   float64
	}{
		{label: "EmptyString", data: tdEmptyString, exp: 0.0},
		{label: "String", data: tdString, exp: 0.0},
		{label: "Bool", data: tdBool, exp: 1.0},
		{label: "Int", data: tdInt, exp: 17.0},
		{label: "Null", data: tdNull, exp: 0.0},
		{label: "Float", data: tdFloat, exp: 22.83},
		{label: "StringSlice", data: tdStringSlice, exp: 0.0},
		{label: "BoolSlice", data: tdBoolSlice, exp: 0.0},
		{label: "IntSlice", data: tdIntSlice, exp: 0.0},
		{label: "FloatSlice", data: tdFloatSlice, exp: 0.0},
		{label: "Object", data: tdObject, exp: 0.0},
		{label: "Objects", data: tdObjects, exp: 0.0},
		{label: "Complex", data: tdComplex, exp: 0.0},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := NewJSONReader(tc.data)
			assert.Nil(t, err)

			v := r.ToFloat()
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestGetFloatSlice(t *testing.T) {
	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v := r.GetFloatSlice("Invalid Key")
		assert.Equal(t, []float64(nil), v)
	})

	testCases := []struct {
		key string
		exp []float64
	}{
		{key: "empty_float", exp: []float64(nil)},
		{key: "string", exp: []float64{0.0}},
		{key: "bool", exp: []float64{1.0}},
		{key: "int", exp: []float64{17.0}},
		{key: "null", exp: []float64{0.0}},
		{key: "float", exp: []float64{22.83}},
		{key: "string_slice", exp: []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0}},
		{key: "bool_slice", exp: []float64{1.0, 0.0, 1.0, 0.0}},
		{key: "int_slice", exp: []float64{-1.0, 0.0, 1.0, 2.0, 3.0, 4.0}},
		{key: "float_slice", exp: []float64{-1.1, 0.0, 1.1, 2.2, 3.3}},
		{key: "object", exp: []float64{0.0, 0.0}},
		{key: "objects", exp: []float64{0.0, 0.0, 0.0}},
		{key: "objects.2.o", exp: []float64{0.0}},
		{key: "complex", exp: []float64{0.0, 2.0, 0.0, 0.0, 2.2, 0.0, 0.0}},
		{key: "complex.1", exp: []float64{2.0}},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			r, err := NewJSONReader(readerTestData)
			assert.Nil(t, err)

			v := r.GetFloatSlice(tc.key)
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestToFloatSlice(t *testing.T) {
	testCases := []struct {
		label string
		data  []byte
		exp   []float64
	}{
		{label: "EmptyString", data: tdEmptyString, exp: []float64{0.0}},
		{label: "String", data: tdString, exp: []float64{0.0}},
		{label: "Bool", data: tdBool, exp: []float64{1.0}},
		{label: "Int", data: tdInt, exp: []float64{17.0}},
		{label: "Null", data: tdNull, exp: []float64{0.0}},
		{label: "Float", data: tdFloat, exp: []float64{22.83}},
		{label: "StringSlice", data: tdStringSlice, exp: []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0}},
		{label: "BoolSlice", data: tdBoolSlice, exp: []float64{1.0, 0.0, 1.0, 0.0}},
		{label: "IntSlice", data: tdIntSlice, exp: []float64{-1.0, 0.0, 1.0, 2.0, 3.0, 4.0}},
		{label: "FloatSlice", data: tdFloatSlice, exp: []float64{-1.1, 0.0, 1.1, 2.2, 3.3}},
		{label: "Object", data: tdObject, exp: []float64{0.0, 0.0}},
		{label: "Objects", data: tdObjects, exp: []float64{0.0, 0.0, 0.0}},
		{label: "Complex", data: tdComplex, exp: []float64{0.0, 2.0, 0.0, 0.0, 2.2, 0.0, 0.0}},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := NewJSONReader(tc.data)
			assert.Nil(t, err)

			v := r.ToFloatSlice()
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestToMapStringFloat(t *testing.T) {
	testCases := []struct {
		label string
		data  []byte
		exp   map[string]float64
	}{
		{label: "EmptyString", data: tdEmptyString, exp: map[string]float64{"0": 0.0}},
		{label: "String", data: tdString, exp: map[string]float64{"0": 0.0}},
		{label: "Bool", data: tdBool, exp: map[string]float64{"0": 1.0}},
		{label: "Int", data: tdInt, exp: map[string]float64{"0": 17.0}},
		{label: "Null", data: tdNull, exp: map[string]float64{}},
		{label: "Float", data: tdFloat, exp: map[string]float64{"0": 22.83}},
		{label: "StringSlice", data: tdStringSlice, exp: map[string]float64{"0": 0.0, "1": 0.0, "2": 0.0, "3": 0.0, "4": 0.0, "5": 0.0}},
		{label: "BoolSlice", data: tdBoolSlice, exp: map[string]float64{"0": 1.0, "1": 0.0, "2": 1.0, "3": 0.0}},
		{label: "IntSlice", data: tdIntSlice, exp: map[string]float64{"0": -1.0, "1": 0.0, "2": 1.0, "3": 2.0, "4": 3.0, "5": 4.0}},
		{label: "FloatSlice", data: tdFloatSlice, exp: map[string]float64{"0": -1.1, "1": 0.0, "2": 1.1, "3": 2.2, "4": 3.3}},
		{label: "Object", data: tdObject, exp: map[string]float64{"a": 0.0, "c": 0.0}},
		{label: "Objects", data: tdObjects, exp: map[string]float64{"0": 0.0, "1": 0.0, "2": 0.0}},
		{label: "Complex", data: tdComplex, exp: map[string]float64{"0": 0.0, "1": 2.0, "2": 0.0, "3": 0.0, "4": 2.2, "5": 0.0, "6": 0.0}},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := NewJSONReader(tc.data)
			assert.Nil(t, err)

			v := r.ToMapStringFloat()
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestGetByteSlice(t *testing.T) {
	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v := r.GetByteSlice("Invalid Key")
		assert.Equal(t, []byte(nil), v)
	})

	testCases := []struct {
		key string
		exp string
	}{
		{key: "empty_string", exp: ""},
		{key: "string", exp: `some string`},
		{key: "int", exp: `17`},
		{key: "bool", exp: `true`},
		{key: "null", exp: `null`},
		{key: "float", exp: `22.83`},
		{key: "string_slice", exp: `[ "a", "b", "c", "d", "e", "t" ]`},
		{key: "bool_slice", exp: `[ true, false, true, false ]`},
		{key: "int_slice", exp: `[ -1, 0, 1, 2, 3, 4 ]`},
		{key: "float_slice", exp: `[ -1.1, 0.0, 1.1, 2.2, 3.3 ]`},
		{key: "object", exp: `{ "a": "b", "c": "d" }`},
		{key: "objects", exp: "[\n\t\t\t{ \"e\": \"f\", \"g\": \"h\" },\n\t\t\t{ \"i\": \"j\", \"k\": \"l\" },\n\t\t\t{ \"m\": \"n\", \"o\": \"t\" }\n\t\t]"},
		{key: "objects.2.o", exp: "t"},
		{key: "complex", exp: `[ "a", 2, null, false, 2.2, { "c": "d", "empty_string": "" }, [ "s" ] ]`},
		{key: "complex.5.c", exp: `d`},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			r, err := NewJSONReader(readerTestData)
			assert.Nil(t, err)

			v := r.GetByteSlice(tc.key)
			assert.Equal(t, []byte(tc.exp), v)
		})
	}
}

func TestToByteSlice(t *testing.T) {
	testCases := []struct {
		label string
		data  []byte
		exp   string
	}{
		{label: "EmptyString", data: tdEmptyString, exp: ""},
		{label: "String", data: tdString, exp: `some string`},
		{label: "Int", data: tdInt, exp: `17`},
		{label: "Bool", data: tdBool, exp: `true`},
		{label: "Null", data: tdNull, exp: `null`},
		{label: "Float", data: tdFloat, exp: `22.83`},
		{label: "StringSlice", data: tdStringSlice, exp: `[ "a", "b", "c", "d", "e", "t" ]`},
		{label: "BoolSlice", data: tdBoolSlice, exp: `[ true, false, true, false ]`},
		{label: "IntSlice", data: tdIntSlice, exp: `[ -1, 0, 1, 2, 3, 4 ]`},
		{label: "FloatSlice", data: tdFloatSlice, exp: `[ -1.1, 0.0, 1.1, 2.2, 3.3 ]`},
		{label: "Object", data: tdObject, exp: `{ "a": "b", "c": "d" }`},
		{label: "Objects", data: tdObjects, exp: `[ { "e": "f", "g": "h" }, { "i": "j", "k": "l" }, { "m": "n", "o": "t" } ]`},
		{label: "Complex", data: tdComplex, exp: `[ "a", 2, null, false, 2.2, { "c": "d", "empty_string": "" }, [ "s" ] ]`},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := NewJSONReader(tc.data)
			assert.Nil(t, err)

			v := r.ToByteSlice()
			assert.Equal(t, []byte(tc.exp), v)
		})
	}
}

func TestGetByteSlices(t *testing.T) {
	// Alias just to enhance readability of the test cases
	type bslices = [][]byte

	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v := r.GetByteSlices("Invalid Key")
		assert.Equal(t, bslices(nil), v)
	})

	testCases := []struct {
		key string
		exp bslices
	}{
		{key: "empty_string", exp: bslices{[]byte(``)}},
		{key: "string", exp: bslices{[]byte(`some string`)}},
		{key: "int", exp: bslices{[]byte(`17`)}},
		{key: "bool", exp: bslices{[]byte(`true`)}},
		{key: "null", exp: bslices{[]byte(`null`)}},
		{key: "float", exp: bslices{[]byte(`22.83`)}},
		{key: "string_slice", exp: bslices{[]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e"), []byte("t")}},
		{key: "bool_slice", exp: bslices{[]byte(`true`), []byte(`false`), []byte(`true`), []byte(`false`)}},
		{key: "int_slice", exp: bslices{[]byte(`-1`), []byte(`0`), []byte(`1`), []byte(`2`), []byte(`3`), []byte(`4`)}},
		{key: "float_slice", exp: bslices{[]byte(`-1.1`), []byte(`0.0`), []byte(`1.1`), []byte(`2.2`), []byte(`3.3`)}},
		{key: "object", exp: bslices{[]byte("b"), []byte("d")}},
		{key: "objects", exp: bslices{[]byte(`{ "e": "f", "g": "h" }`), []byte(`{ "i": "j", "k": "l" }`), []byte(`{ "m": "n", "o": "t" }`)}},
		{key: "objects.2.o", exp: bslices{[]byte(`t`)}},
		{key: "complex", exp: bslices{[]byte("a"), []byte("2"), []byte("null"), []byte("false"), []byte("2.2"), []byte(`{ "c": "d", "empty_string": "" }`), []byte(`[ "s" ]`)}},
		{key: "complex.5.c", exp: bslices{[]byte(`d`)}},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			r, err := NewJSONReader(readerTestData)
			assert.Nil(t, err)

			v := r.GetByteSlices(tc.key)
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestToByteSlices(t *testing.T) {
	// Alias just to enhance readability of the test cases
	type bslices = [][]byte

	testCases := []struct {
		label string
		data  []byte
		exp   bslices
	}{
		{label: "EmptyString", data: tdEmptyString, exp: bslices{[]byte(``)}},
		{label: "String", data: tdString, exp: bslices{[]byte(`some string`)}},
		{label: "Int", data: tdInt, exp: bslices{[]byte(`17`)}},
		{label: "Bool", data: tdBool, exp: bslices{[]byte(`true`)}},
		{label: "Null", data: tdNull, exp: bslices{[]byte(`null`)}},
		{label: "Float", data: tdFloat, exp: bslices{[]byte(`22.83`)}},
		{label: "StringSlice", data: tdStringSlice, exp: bslices{[]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e"), []byte("t")}},
		{label: "BoolSlice", data: tdBoolSlice, exp: bslices{[]byte(`true`), []byte(`false`), []byte(`true`), []byte(`false`)}},
		{label: "IntSlice", data: tdIntSlice, exp: bslices{[]byte(`-1`), []byte(`0`), []byte(`1`), []byte(`2`), []byte(`3`), []byte(`4`)}},
		{label: "FloatSlice", data: tdFloatSlice, exp: bslices{[]byte(`-1.1`), []byte(`0.0`), []byte(`1.1`), []byte(`2.2`), []byte(`3.3`)}},
		{label: "Object", data: tdObject, exp: bslices{[]byte("b"), []byte("d")}},
		{label: "Objects", data: tdObjects, exp: bslices{[]byte(`{ "e": "f", "g": "h" }`), []byte(`{ "i": "j", "k": "l" }`), []byte(`{ "m": "n", "o": "t" }`)}},
		{label: "Complex", data: tdComplex, exp: bslices{[]byte("a"), []byte("2"), []byte("null"), []byte("false"), []byte("2.2"), []byte(`{ "c": "d", "empty_string": "" }`), []byte(`[ "s" ]`)}},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := NewJSONReader(tc.data)
			assert.Nil(t, err)

			v := r.ToByteSlices()
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestToMapStringBytes(t *testing.T) {
	testCases := []struct {
		label string
		data  []byte
		exp   map[string][]byte
	}{
		{label: "EmptyString", data: tdEmptyString, exp: map[string][]byte{"0": []byte(``)}},
		{label: "String", data: tdString, exp: map[string][]byte{"0": []byte(`some string`)}},
		{label: "Int", data: tdInt, exp: map[string][]byte{"0": []byte(`17`)}},
		{label: "Bool", data: tdBool, exp: map[string][]byte{"0": []byte(`true`)}},
		{label: "Null", data: tdNull, exp: map[string][]byte{"0": []byte(`null`)}},
		{label: "Float", data: tdFloat, exp: map[string][]byte{"0": []byte(`22.83`)}},
		{label: "StringSlice", data: tdStringSlice, exp: map[string][]byte{"0": []byte("a"), "1": []byte("b"), "2": []byte("c"), "3": []byte("d"), "4": []byte("e"), "5": []byte("t")}},
		{label: "BoolSlice", data: tdBoolSlice, exp: map[string][]byte{"0": []byte("true"), "1": []byte("false"), "2": []byte("true"), "3": []byte("false")}},
		{label: "IntSlice", data: tdIntSlice, exp: map[string][]byte{"0": []byte("-1"), "1": []byte("0"), "2": []byte("1"), "3": []byte("2"), "4": []byte("3"), "5": []byte("4")}},
		{label: "FloatSlice", data: tdFloatSlice, exp: map[string][]byte{"0": []byte("-1.1"), "1": []byte("0.0"), "2": []byte("1.1"), "3": []byte("2.2"), "4": []byte("3.3")}},
		{label: "Object", data: tdObject, exp: map[string][]byte{"a": []byte("b"), "c": []byte("d")}},
		{label: "Objects", data: tdObjects, exp: map[string][]byte{"0": []byte(`{ "e": "f", "g": "h" }`), "1": []byte(`{ "i": "j", "k": "l" }`), "2": []byte(`{ "m": "n", "o": "t" }`)}},
		{label: "Complex", data: tdComplex, exp: map[string][]byte{"0": []byte("a"), "1": []byte("2"), "2": []byte("null"), "3": []byte("false"), "4": []byte("2.2"), "5": []byte(`{ "c": "d", "empty_string": "" }`), "6": []byte(`[ "s" ]`)}},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := NewJSONReader(tc.data)
			assert.Nil(t, err)

			v := r.ToMapStringBytes()
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestGetInterface(t *testing.T) {
	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v := r.GetString("Invalid Key")
		assert.Equal(t, "", v)
	})

	t.Run("Escape Characters", func(t *testing.T) {
		b := []byte{'"', 'a', '\u0026', 'b', '\u003c', 'c', '\\', '"', '\u003e', '\\', '"', 'd', '"'}
		r, err := NewJSONReader(b)
		assert.Nil(t, err)

		assert.Equal(t, `a&b<c">"d`, r.GetString(""))
	})

	testCases := []struct {
		key string
		exp interface{}
	}{
		{key: "empty_string", exp: ``},
		{key: "string", exp: `some string`},
		{key: "int", exp: 17},
		{key: "bool", exp: true},
		{key: "null", exp: nil},
		{key: "float", exp: 22.83},
		{key: "string_slice", exp: []interface{}{"a", "b", "c", "d", "e", "t"}},
		{key: "bool_slice", exp: []interface{}{true, false, true, false}},
		{key: "int_slice", exp: []interface{}{-1, 0, 1, 2, 3, 4}},
		{key: "float_slice", exp: []interface{}{-1.1, 0.0, 1.1, 2.2, 3.3}},
		{key: "object", exp: map[string]interface{}{"a": "b", "c": "d"}},
		{key: "objects", exp: []interface{}{map[string]interface{}{`e`: `f`, `g`: `h`}, map[string]interface{}{`i`: `j`, `k`: `l`}, map[string]interface{}{`m`: `n`, `o`: `t`}}},
		{key: "objects.2.o", exp: "t"},
		{key: "complex", exp: []interface{}{"a", 2, nil, false, 2.2, map[string]interface{}{"c": "d", "empty_string": ""}, []interface{}{"s"}}},
		{key: "complex.5.c", exp: `d`},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			r, err := NewJSONReader(readerTestData)
			assert.Nil(t, err)

			v := r.GetInterface(tc.key)
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestToInterface(t *testing.T) {
	testCases := []struct {
		label string
		data  []byte
		exp   interface{}
	}{
		{label: "EmptyString", data: tdEmptyString, exp: ``},
		{label: "String", data: tdString, exp: `some string`},
		{label: "Int", data: tdInt, exp: 17},
		{label: "Bool", data: tdBool, exp: true},
		{label: "Null", data: tdNull, exp: nil},
		{label: "Float", data: tdFloat, exp: 22.83},
		{label: "StringSlice", data: tdStringSlice, exp: []interface{}{"a", "b", "c", "d", "e", "t"}},
		{label: "BoolSlice", data: tdBoolSlice, exp: []interface{}{true, false, true, false}},
		{label: "IntSlice", data: tdIntSlice, exp: []interface{}{-1, 0, 1, 2, 3, 4}},
		{label: "FloatSlice", data: tdFloatSlice, exp: []interface{}{-1.1, 0.0, 1.1, 2.2, 3.3}},
		{label: "Object", data: tdObject, exp: map[string]interface{}{"a": "b", "c": "d"}},
		{label: "Objects", data: tdObjects, exp: []interface{}{map[string]interface{}{`e`: `f`, `g`: `h`}, map[string]interface{}{`i`: `j`, `k`: `l`}, map[string]interface{}{`m`: `n`, `o`: `t`}}},
		{label: "Complex", data: tdComplex, exp: []interface{}{"a", 2, nil, false, 2.2, map[string]interface{}{"c": "d", "empty_string": ""}, []interface{}{"s"}}},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := NewJSONReader(tc.data)
			assert.Nil(t, err)

			v := r.ToInterface()
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestGetInterfaceSlice(t *testing.T) {
	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v := r.GetInterfaceSlice("Invalid Key")
		assert.Equal(t, []interface{}(nil), v)
	})

	testCases := []struct {
		key string
		exp []interface{}
	}{
		{key: "empty_string", exp: []interface{}{``}},
		{key: "string", exp: []interface{}{`some string`}},
		{key: "int", exp: []interface{}{17}},
		{key: "bool", exp: []interface{}{true}},
		{key: "null", exp: []interface{}{nil}},
		{key: "float", exp: []interface{}{22.83}},
		{key: "string_slice", exp: []interface{}{"a", "b", "c", "d", "e", "t"}},
		{key: "bool_slice", exp: []interface{}{true, false, true, false}},
		{key: "int_slice", exp: []interface{}{-1, 0, 1, 2, 3, 4}},
		{key: "float_slice", exp: []interface{}{-1.1, 0.0, 1.1, 2.2, 3.3}},
		{key: "object", exp: []interface{}{"b", "d"}},
		{key: "objects", exp: []interface{}{map[string]interface{}{`e`: `f`, `g`: `h`}, map[string]interface{}{`i`: `j`, `k`: `l`}, map[string]interface{}{`m`: `n`, `o`: `t`}}},
		{key: "objects.2.o", exp: []interface{}{"t"}},
		{key: "complex", exp: []interface{}{"a", 2, nil, false, 2.2, map[string]interface{}{"c": "d", "empty_string": ""}, []interface{}{"s"}}},
		{key: "complex.5.c", exp: []interface{}{`d`}},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			r, err := NewJSONReader(readerTestData)
			assert.Nil(t, err)

			v := r.GetInterfaceSlice(tc.key)
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestToInterfaceSlice(t *testing.T) {
	testCases := []struct {
		label string
		data  []byte
		exp   interface{}
	}{
		{label: "EmptyString", data: tdEmptyString, exp: []interface{}{``}},
		{label: "String", data: tdString, exp: []interface{}{`some string`}},
		{label: "Int", data: tdInt, exp: []interface{}{17}},
		{label: "Bool", data: tdBool, exp: []interface{}{true}},
		{label: "Null", data: tdNull, exp: []interface{}{nil}},
		{label: "Float", data: tdFloat, exp: []interface{}{22.83}},
		{label: "StringSlice", data: tdStringSlice, exp: []interface{}{"a", "b", "c", "d", "e", "t"}},
		{label: "BoolSlice", data: tdBoolSlice, exp: []interface{}{true, false, true, false}},
		{label: "IntSlice", data: tdIntSlice, exp: []interface{}{-1, 0, 1, 2, 3, 4}},
		{label: "FloatSlice", data: tdFloatSlice, exp: []interface{}{-1.1, 0.0, 1.1, 2.2, 3.3}},
		{label: "Object", data: tdObject, exp: []interface{}{"b", "d"}},
		{label: "Objects", data: tdObjects, exp: []interface{}{map[string]interface{}{`e`: `f`, `g`: `h`}, map[string]interface{}{`i`: `j`, `k`: `l`}, map[string]interface{}{`m`: `n`, `o`: `t`}}},
		{label: "Complex", data: tdComplex, exp: []interface{}{"a", 2, nil, false, 2.2, map[string]interface{}{"c": "d", "empty_string": ""}, []interface{}{"s"}}},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := NewJSONReader(tc.data)
			assert.Nil(t, err)

			v := r.ToInterfaceSlice()
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestGetMapStringInterface(t *testing.T) {
	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v := r.GetMapStringInterface("Invalid Key")
		assert.Equal(t, map[string]interface{}(nil), v)
	})

	testCases := []struct {
		key string
		exp map[string]interface{}
	}{
		{key: "empty_string", exp: map[string]interface{}{"0": ``}},
		{key: "string", exp: map[string]interface{}{"0": `some string`}},
		{key: "int", exp: map[string]interface{}{"0": 17}},
		{key: "bool", exp: map[string]interface{}{"0": true}},
		{key: "null", exp: map[string]interface{}{"0": nil}},
		{key: "float", exp: map[string]interface{}{"0": 22.83}},
		{key: "string_slice", exp: map[string]interface{}{"0": "a", "1": "b", "2": "c", "3": "d", "4": "e", "5": "t"}},
		{key: "bool_slice", exp: map[string]interface{}{"0": true, "1": false, "2": true, "3": false}},
		{key: "int_slice", exp: map[string]interface{}{"0": -1, "1": 0, "2": 1, "3": 2, "4": 3, "5": 4}},
		{key: "float_slice", exp: map[string]interface{}{"0": -1.1, "1": 0.0, "2": 1.1, "3": 2.2, "4": 3.3}},
		{key: "object", exp: map[string]interface{}{"a": "b", "c": "d"}},
		{key: "objects", exp: map[string]interface{}{"0": map[string]interface{}{`e`: `f`, `g`: `h`}, "1": map[string]interface{}{`i`: `j`, `k`: `l`}, "2": map[string]interface{}{`m`: `n`, `o`: `t`}}},
		{key: "complex", exp: map[string]interface{}{"0": "a", "1": 2, "2": nil, "3": false, "4": 2.2, "5": map[string]interface{}{"c": "d", "empty_string": ""}, "6": []interface{}{"s"}}},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			r, err := NewJSONReader(readerTestData)
			assert.Nil(t, err)

			v := r.GetMapStringInterface(tc.key)
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestToMapStringInterface(t *testing.T) {
	testCases := []struct {
		label string
		data  []byte
		exp   map[string]interface{}
	}{
		{label: "EmptyString", data: tdEmptyString, exp: map[string]interface{}{"0": ``}},
		{label: "String", data: tdString, exp: map[string]interface{}{"0": `some string`}},
		{label: "Int", data: tdInt, exp: map[string]interface{}{"0": 17}},
		{label: "Bool", data: tdBool, exp: map[string]interface{}{"0": true}},
		{label: "Null", data: tdNull, exp: map[string]interface{}{"0": nil}},
		{label: "Float", data: tdFloat, exp: map[string]interface{}{"0": 22.83}},
		{label: "StringSlice", data: tdStringSlice, exp: map[string]interface{}{"0": "a", "1": "b", "2": "c", "3": "d", "4": "e", "5": "t"}},
		{label: "BoolSlice", data: tdBoolSlice, exp: map[string]interface{}{"0": true, "1": false, "2": true, "3": false}},
		{label: "IntSlice", data: tdIntSlice, exp: map[string]interface{}{"0": -1, "1": 0, "2": 1, "3": 2, "4": 3, "5": 4}},
		{label: "FloatSlice", data: tdFloatSlice, exp: map[string]interface{}{"0": -1.1, "1": 0.0, "2": 1.1, "3": 2.2, "4": 3.3}},
		{label: "Object", data: tdObject, exp: map[string]interface{}{"a": "b", "c": "d"}},
		{label: "Objects", data: tdObjects, exp: map[string]interface{}{"0": map[string]interface{}{`e`: `f`, `g`: `h`}, "1": map[string]interface{}{`i`: `j`, `k`: `l`}, "2": map[string]interface{}{`m`: `n`, `o`: `t`}}},
		{label: "Complex", data: tdComplex, exp: map[string]interface{}{"0": "a", "1": 2, "2": nil, "3": false, "4": 2.2, "5": map[string]interface{}{"c": "d", "empty_string": ""}, "6": []interface{}{"s"}}},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			r, err := NewJSONReader(tc.data)
			assert.Nil(t, err)

			v := r.ToMapStringInterface()
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestGetIface(t *testing.T) {
	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v := r.getIface("Invalid Key")
		assert.Equal(t, nil, v)
	})

	testCases := []struct {
		key string
		exp interface{}
	}{
		{key: "empty_string", exp: ``},
		{key: "string", exp: `some string`},
		{key: "int", exp: 17},
		{key: "bool", exp: true},
		{key: "null", exp: nil},
		{key: "float", exp: 22.83},
		{key: "string_slice", exp: []interface{}{"a", "b", "c", "d", "e", "t"}},
		{key: "bool_slice", exp: []interface{}{true, false, true, false}},
		{key: "int_slice", exp: []interface{}{-1, 0, 1, 2, 3, 4}},
		{key: "float_slice", exp: []interface{}{-1.1, 0.0, 1.1, 2.2, 3.3}},
		{key: "object", exp: map[string]interface{}{"a": "b", "c": "d"}},
		{key: "objects", exp: []interface{}{map[string]interface{}{`e`: `f`, `g`: `h`}, map[string]interface{}{`i`: `j`, `k`: `l`}, map[string]interface{}{`m`: `n`, `o`: `t`}}},
		{key: "objects.2.o", exp: "t"},
		{key: "complex", exp: []interface{}{"a", 2, nil, false, 2.2, map[string]interface{}{"c": "d", "empty_string": ""}, []interface{}{"s"}}},
		{key: "complex.5.c", exp: `d`},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			r, err := NewJSONReader(readerTestData)
			assert.Nil(t, err)

			v := r.getIface(tc.key)
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestGetObject(t *testing.T) {
	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v, _ := r.getObject("Invalid Key")
		assert.Equal(t, map[string]interface{}(nil), v)
	})

	t.Run("Not an Object", func(t *testing.T) {
		r, err := NewJSONReader(tdString)
		assert.Nil(t, err)

		v, _ := r.getObject("")
		assert.Equal(t, map[string]interface{}(nil), v)
	})

	t.Run("Complex Object", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v, _ := r.getObject("")
		assert.Equal(t, map[string]interface{}{
			"bool":         true,
			"empty_string": "",
			"float":        22.83,
			"int":          17,
			"null":         interface{}(nil),
			"string":       "some string",
			"bool_slice":   []interface{}{true, false, true, false},
			"float_slice":  []interface{}{-1.1, 0.0, 1.1, 2.2, 3.3},
			"int_slice":    []interface{}{-1, 0, 1, 2, 3, 4},
			"string_slice": []interface{}{"a", "b", "c", "d", "e", "t"},
			"object":       map[string]interface{}{"a": "b", "c": "d"},
			"objects":      []interface{}{map[string]interface{}{"e": "f", "g": "h"}, map[string]interface{}{"i": "j", "k": "l"}, map[string]interface{}{"m": "n", "o": "t"}},
			"complex":      []interface{}{"a", 2, interface{}(nil), false, 2.2, map[string]interface{}{"c": "d", "empty_string": ""}, []interface{}{"s"}}},
			v)
	})
}

func TestGetSlice(t *testing.T) {
	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v := r.getSlice("Invalid Key")
		assert.Equal(t, []interface{}(nil), v)
	})

	t.Run("Not an Object", func(t *testing.T) {
		r, err := NewJSONReader(tdString)
		assert.Nil(t, err)

		v := r.getSlice("")
		assert.Equal(t, []interface{}(nil), v)
	})

	t.Run("Complex Slice", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v := r.getSlice("complex")
		assert.Equal(t, []interface{}{
			"a",
			2,
			interface{}(nil),
			false,
			2.2,
			map[string]interface{}{"c": "d", "empty_string": ""},
			[]interface{}{"s"}},
			v)
	})
}

func TestGetDataByKey(t *testing.T) {
	t.Run("Empty Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v, dt, k := r.getDataByKey("")
		assert.Equal(t, readerTestData, v)
		assert.Equal(t, "object", dt)
		assert.Equal(t, readerTestDataKeys, k)
	})

	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v, dt, k := r.getDataByKey("Invalid Key")
		assert.Equal(t, []byte(nil), v)
		assert.Equal(t, "", dt)
		assert.Equal(t, []string(nil), k)
	})

	t.Run("Missing Key Nested", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v, dt, k := r.getDataByKey("object.0.k")
		assert.Equal(t, []byte(nil), v)
		assert.Equal(t, "", dt)
		assert.Equal(t, []string(nil), k)
	})

	t.Run("Valid Key - Object", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v, dt, k := r.getDataByKey("object")
		assert.Equal(t, tdObject, v)
		assert.Equal(t, "object", dt)
		assert.Equal(t, []string{"a", "c"}, k)
	})

	t.Run("Valid Key - Slice", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v, dt, k := r.getDataByKey("bool_slice")
		assert.Equal(t, tdBoolSlice, v)
		assert.Equal(t, "array", dt)
		assert.Equal(t, []string{"0", "1", "2", "3"}, k)
	})

	t.Run("Valid Key - Nested String", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v, dt, k := r.getDataByKey("object.c")
		assert.Equal(t, []byte("d"), v)
		assert.Equal(t, "string", dt)
		assert.Equal(t, []string(nil), k)
	})

	t.Run("Valid Key - Nested Object Slice String", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		v, dt, k := r.getDataByKey("objects.2.o")
		assert.Equal(t, []byte("t"), v)
		assert.Equal(t, "string", dt)
		assert.Equal(t, []string(nil), k)
	})
}

func TestGetChildByKey(t *testing.T) {
	t.Run("Empty Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		p := r.getChildByKey("")
		assert.Equal(t, readerTestData, p.bytes)
		assert.Equal(t, "object", p.dtype)
		assert.Equal(t, readerTestDataKeys, p.keys)
	})

	t.Run("Missing Key", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		p := r.getChildByKey("Invalid Key")
		assert.Nil(t, p)
	})

	t.Run("Missing Key Nested", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		p := r.getChildByKey("object.0.p.keys")
		assert.Nil(t, p)
	})

	t.Run("Valid Key - Object", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		p := r.getChildByKey("object")
		assert.Equal(t, tdObject, p.bytes)
		assert.Equal(t, "object", p.dtype)
		assert.Equal(t, []string{"a", "c"}, p.keys)
	})

	t.Run("Valid Key - Slice", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		p := r.getChildByKey("bool_slice")
		assert.Equal(t, tdBoolSlice, p.bytes)
		assert.Equal(t, "array", p.dtype)
		assert.Equal(t, []string{"0", "1", "2", "3"}, p.keys)
	})

	t.Run("Valid Key - Nested String", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		p := r.getChildByKey("object.c")
		assert.Equal(t, []byte("d"), p.bytes)
		assert.Equal(t, "string", p.dtype)
		assert.Equal(t, []string(nil), p.keys)
	})

	t.Run("Valid Key - Nested Object Slice String", func(t *testing.T) {
		r, err := NewJSONReader(readerTestData)
		assert.Nil(t, err)

		p := r.getChildByKey("objects.2.o")
		assert.Equal(t, []byte("t"), p.bytes)
		assert.Equal(t, "string", p.dtype)
		assert.Equal(t, []string(nil), p.keys)
	})
}

func TestToIface(t *testing.T) {
	t.Run("Object with JSON Error", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				assert.Equal(t, `invalid character 'v' at position '8' in segment '{"key": value}' (expected object value)`, r.(error).Error())
			}
		}()

		toIface([]byte(`{"key": value}`), JSONObject, false)
		assert.Fail(t, "Expected Panic")
	})

	t.Run("Array with JSON Error", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				assert.Equal(t, `invalid character 'v' at position '1' in segment '[value]'`, r.(error).Error())
			}
		}()

		toIface([]byte(`[value]`), JSONArray, false)
		assert.Fail(t, "Expected Panic")
	})

	testCases := []struct {
		label string
		dtype string
		data  []byte
		exp   interface{}
	}{
		{label: "EmptyString", dtype: JSONString, data: tdEmptyString, exp: ``},
		{label: "String", dtype: JSONString, data: tdString, exp: `some string`},
		{label: "Int", dtype: JSONInt, data: tdInt, exp: 17},
		{label: "Bool", dtype: JSONBool, data: tdBool, exp: true},
		{label: "Null", dtype: JSONNull, data: tdNull, exp: nil},
		{label: "Float", dtype: JSONFloat, data: tdFloat, exp: 22.83},
		{label: "StringSlice", dtype: JSONArray, data: tdStringSlice, exp: []interface{}{"a", "b", "c", "d", "e", "t"}},
		{label: "BoolSlice", dtype: JSONArray, data: tdBoolSlice, exp: []interface{}{true, false, true, false}},
		{label: "IntSlice", dtype: JSONArray, data: tdIntSlice, exp: []interface{}{-1, 0, 1, 2, 3, 4}},
		{label: "FloatSlice", dtype: JSONArray, data: tdFloatSlice, exp: []interface{}{-1.1, 0.0, 1.1, 2.2, 3.3}},
		{label: "Object", dtype: JSONObject, data: tdObject, exp: map[string]interface{}{"a": "b", "c": "d"}},
		{label: "Objects", dtype: JSONArray, data: tdObjects, exp: []interface{}{map[string]interface{}{`e`: `f`, `g`: `h`}, map[string]interface{}{`i`: `j`, `k`: `l`}, map[string]interface{}{`m`: `n`, `o`: `t`}}},
		{label: "Complex", dtype: JSONArray, data: tdComplex, exp: []interface{}{"a", 2, nil, false, 2.2, map[string]interface{}{"c": "d", "empty_string": ""}, []interface{}{"s"}}},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {

			v := toIface(tc.data, tc.dtype, false)
			assert.Equal(t, tc.exp, v)
		})
	}
}

func TestStringInArray(t *testing.T) {
	data := []string{
		"Hey, this is cool",
		"Bye",
		"things",
		"stuff",
		"",
	}

	testCases := []struct {
		search string
		exp    bool
	}{
		{search: "Gibberish", exp: false},
		{search: "Hey this is cool", exp: false},
		{search: "Hey, this is cool", exp: true},
		{search: "Bye", exp: true},
		{search: "bye", exp: false},
		{search: "things", exp: true},
		{search: "stuff", exp: true},
		{search: "", exp: true},
	}

	for _, tc := range testCases {
		t.Run(tc.search, func(t *testing.T) {
			assert.Equal(t, tc.exp, stringInArray(tc.search, data))
		})
	}
}

func TestInt64LargeValue(t *testing.T) {
	var expected int64 = 6754210771357157538
	actual := toInt([]byte("6.754210771357157538e18"), JSONFloat, false)
	assert.Equal(t, 6754210771357157376, actual)

	actual = toInt([]byte("6.754210771357157538e17"), JSONInt, false)
	assert.Equal(t, 0, actual)

	actual = toInt([]byte("6754210771357157538"), JSONInt, false)
	assert.Equal(t, int(expected), actual)
}

func TestToStringEmoji(t *testing.T) {
	input := []byte(`"Emoji!! \ud83d\udc4f \uD83D\uDC4C \ud83d\uDC7B"`)
	// input := []byte(`"👏"`)

	var output string
	Unmarshal(input, &output)

	assert.Equal(t, `Emoji!! 👏 👌 👻`, output)
}

func TestToStringWithEscapedSlashes(t *testing.T) {
	input := `C:\t <strong>Pendulum Effect :</strong> You cannot Special Summon monsters, except "Qli" monsters. This effect cannot be negated. Once per turn: You can pay 800 LP; add 1 "Qli" card from your Deck to your hand, except "Qliphort Scout".<br><strong>Monster Text :</strong> <em><br>Booting in Replica Mode…<br>An error has occurred when executing C:\sophia\zefra.exe<br>Unknown publisher.<br>Allow C:\tierra\qliphort.exe ? <Y/N>…[Y]<br>Booting in Autonomy Mode…</em>`

	b, _ := json.Marshal(input)

	var output string
	Unmarshal(b, &output)
	assert.Equal(t, input, output)
}

func TestToStringWithEscapedFrontSlashes(t *testing.T) {
	input := []byte(`"https:\/\/www.mydomain.com\u002Fthings\/"`)

	var output string
	Unmarshal(input, &output)
	assert.Equal(t, `https://www.mydomain.com/things/`, output)
}
