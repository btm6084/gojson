package gojson

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIterator(t *testing.T) {
	t.Run("Complex Objects and Arrays", func(t *testing.T) {
		data := []byte(`[{"a":"b"},["c","d"],[{"e":"f"},{"g":"h"}],[["i","j"],{"k":"l"}]]`)
		i, err := NewIterator(data)
		assert.Nil(t, err)

		b, dt, err := i.Next()
		assert.Equal(t, `{"a":"b"}`, string(b))
		assert.Equal(t, JSONObject, dt)
		assert.Nil(t, err)

		b, dt, err = i.Next()
		assert.Equal(t, `["c","d"]`, string(b))
		assert.Equal(t, JSONArray, dt)
		assert.Nil(t, err)

		b, dt, err = i.Next()
		assert.Equal(t, `[{"e":"f"},{"g":"h"}]`, string(b))
		assert.Equal(t, JSONArray, dt)
		assert.Nil(t, err)

		b, dt, err = i.Next()
		assert.Equal(t, `[["i","j"],{"k":"l"}]`, string(b))
		assert.Equal(t, JSONArray, dt)
		assert.Nil(t, err)

		b, dt, err = i.Next()
		assert.Nil(t, b)
		assert.Equal(t, "", dt)
		assert.Equal(t, ErrEndOfInput, err)

		i.Reset()

		b, dt, err = i.Next()
		assert.Equal(t, `{"a":"b"}`, string(b))
		assert.Equal(t, JSONObject, dt)
		assert.Nil(t, err)

		b, dt, err = i.Next()
		assert.Equal(t, `["c","d"]`, string(b))
		assert.Equal(t, JSONArray, dt)
		assert.Nil(t, err)

		b, dt, err = i.Next()
		assert.Equal(t, `[{"e":"f"},{"g":"h"}]`, string(b))
		assert.Equal(t, JSONArray, dt)
		assert.Nil(t, err)

		b, dt, err = i.Next()
		assert.Equal(t, `[["i","j"],{"k":"l"}]`, string(b))
		assert.Equal(t, JSONArray, dt)
		assert.Nil(t, err)

		b, dt, err = i.Next()
		assert.Nil(t, b)
		assert.Equal(t, "", dt)
		assert.Equal(t, ErrEndOfInput, err)
	})

	t.Run("Simple Stuff", func(t *testing.T) {
		data := []byte(`["String", true, false, null, 17, 42.42]`)
		i, err := NewIterator(data)
		assert.Nil(t, err)

		b, dt, err := i.Next()
		assert.Equal(t, `"String"`, string(b))
		assert.Equal(t, JSONString, dt)
		assert.Nil(t, err)

		b, dt, err = i.Next()
		assert.Equal(t, `true`, string(b))
		assert.Equal(t, JSONBool, dt)
		assert.Nil(t, err)

		b, dt, err = i.Next()
		assert.Equal(t, `false`, string(b))
		assert.Equal(t, JSONBool, dt)
		assert.Nil(t, err)

		b, dt, err = i.Next()
		assert.Equal(t, `null`, string(b))
		assert.Equal(t, JSONNull, dt)
		assert.Nil(t, err)

		b, dt, err = i.Next()
		assert.Equal(t, `17`, string(b))
		assert.Equal(t, JSONInt, dt)
		assert.Nil(t, err)

		b, dt, err = i.Next()
		assert.Equal(t, `42.42`, string(b))
		assert.Equal(t, JSONFloat, dt)
		assert.Nil(t, err)

		b, dt, err = i.Next()
		assert.Nil(t, b)
		assert.Equal(t, "", dt)
		assert.Equal(t, ErrEndOfInput, err)
	})

	t.Run("Not Array or Object", func(t *testing.T) {
		data := []byte(`"String"`)
		_, err := NewIterator(data)
		assert.Equal(t, ErrRequiresObject, err)
	})

	t.Run("Malformed JSON", func(t *testing.T) {
		data := []byte(`String"`)
		_, err := NewIterator(data)
		assert.Equal(t, ErrMalformedJSON, err)

		data = []byte(`["a" "b"]`)
		_, err = NewIterator(data)
		assert.Equal(t, ErrMalformedJSON, err)

		data = []byte(`["a", "b"`)
		_, err = NewIterator(data)
		assert.Equal(t, ErrMalformedJSON, err)

		data = []byte(`"a", "b"]`)
		_, err = NewIterator(data)
		assert.Equal(t, ErrMalformedJSON, err)

		data = []byte(`123456a`)
		_, err = NewIterator(data)
		assert.Equal(t, ErrMalformedJSON, err)
	})

	t.Run("Last", func(t *testing.T) {
		data := []byte(`["String", true, false, null, 17, 42.42]`)
		i, err := NewIterator(data)
		assert.Nil(t, err)

		b, dt, err := i.Next()
		assert.Equal(t, `"String"`, string(b))
		assert.Equal(t, JSONString, dt)
		assert.Nil(t, err)

		b, dt, err = i.Last()
		assert.Equal(t, `"String"`, string(b))
		assert.Equal(t, JSONString, dt)
		assert.Nil(t, err)

		b, dt, err = i.Next()
		assert.Equal(t, `true`, string(b))
		assert.Equal(t, JSONBool, dt)
		assert.Nil(t, err)

		i.Next()
		i.Next()
		i.Next()

		b, dt, err = i.Last()
		assert.Equal(t, `17`, string(b))
		assert.Equal(t, JSONInt, dt)
		assert.Nil(t, err)
	})

	t.Run("Reset", func(t *testing.T) {
		data := []byte(`["String", true, false, null, 17, 42.42]`)
		i, err := NewIterator(data)
		assert.Nil(t, err)

		b, dt, err := i.Next()
		assert.Equal(t, `"String"`, string(b))
		assert.Equal(t, JSONString, dt)
		assert.Nil(t, err)

		b, dt, err = i.Last()
		assert.Equal(t, `"String"`, string(b))
		assert.Equal(t, JSONString, dt)
		assert.Nil(t, err)

		b, dt, err = i.Next()
		assert.Equal(t, `true`, string(b))
		assert.Equal(t, JSONBool, dt)
		assert.Nil(t, err)

		i.Reset()

		b, dt, err = i.Next()
		assert.Equal(t, `"String"`, string(b))
		assert.Equal(t, JSONString, dt)
		assert.Nil(t, err)

		b, dt, err = i.Last()
		assert.Equal(t, `"String"`, string(b))
		assert.Equal(t, JSONString, dt)
		assert.Nil(t, err)

		b, dt, err = i.Next()
		assert.Equal(t, `true`, string(b))
		assert.Equal(t, JSONBool, dt)
		assert.Nil(t, err)
	})

	for k, test := range []struct {
		expected     string
		expectedType string
	}{
		{`{"a":"b"}`, JSONObject},
		{`["c","d"]`, JSONArray},
		{`[{"e":"f"},{"g":"h"}]`, JSONArray},
		{`[["i","j"],{"k":"l"}]`, JSONArray},
		{`"Hello!"`, JSONString},
	} {
		t.Run("Test Index", func(t *testing.T) {
			data := []byte(`[{"a":"b"},["c","d"],[{"e":"f"},{"g":"h"}],[["i","j"],{"k":"l"}], "Hello!"]`)
			i, err := NewIterator(data)
			assert.Nil(t, err)

			b, typ, err := i.Index(k)
			assert.Nil(t, err)
			assert.Equal(t, test.expected, string(b))
			assert.Equal(t, test.expectedType, typ)
		})
	}

	t.Run("Test Index Sets Position Correctly", func(t *testing.T) {
		data := []byte(`[{"a":"b"},["c","d"],[{"e":"f"},{"g":"h"}],[["i","j"],{"k":"l"}], "Hello!"]`)
		i, err := NewIterator(data)
		assert.Nil(t, err)

		i.Next()
		i.Next()
		i.Next()

		b, typ, err := i.Index(1)
		assert.Nil(t, err)
		assert.Equal(t, `["c","d"]`, string(b))
		assert.Equal(t, JSONArray, typ)

		b, typ, err = i.Last()
		assert.Nil(t, err)
		assert.Equal(t, `["c","d"]`, string(b))
		assert.Equal(t, JSONArray, typ)

		b, typ, err = i.Next()
		assert.Nil(t, err)
		assert.Equal(t, `[{"e":"f"},{"g":"h"}]`, string(b))
		assert.Equal(t, JSONArray, typ)
	})

	t.Run("Test Index Sets Last Position", func(t *testing.T) {
		data := []byte(`[{"a":"b"},["c","d"],[{"e":"f"},{"g":"h"}],[["i","j"],{"k":"l"}], "Hello!"]`)
		i, err := NewIterator(data)
		assert.Nil(t, err)

		b, typ, err := i.Index(1)
		assert.Nil(t, err)
		assert.Equal(t, `["c","d"]`, string(b))
		assert.Equal(t, JSONArray, typ)

		b, typ, err = i.Last()
		assert.Nil(t, err)
		assert.Equal(t, `["c","d"]`, string(b))
		assert.Equal(t, JSONArray, typ)

		b, typ, err = i.Next()
		assert.Nil(t, err)
		assert.Equal(t, `[{"e":"f"},{"g":"h"}]`, string(b))
		assert.Equal(t, JSONArray, typ)
	})

	t.Run("Test Index Out of Bounds", func(t *testing.T) {
		data := []byte(`[{"a":"b"},["c","d"],[{"e":"f"},{"g":"h"}],[["i","j"],{"k":"l"}], "Hello!"]`)
		i, err := NewIterator(data)
		assert.Nil(t, err)

		b, typ, err := i.Index(10)
		assert.Equal(t, ErrNoSuchIndex, err)
		assert.Equal(t, ``, string(b))
		assert.Equal(t, ``, typ)

		b, typ, err = i.Index(-1)
		assert.Equal(t, ErrNoSuchIndex, err)
		assert.Equal(t, ``, string(b))
		assert.Equal(t, ``, typ)
	})
}
