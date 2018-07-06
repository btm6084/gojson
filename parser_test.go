package gojson

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	parseTestNumbers = [][]byte{
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

	parseTestConstants = [][]byte{
		[]byte("true"),
		[]byte("false"),
	}

	parseTestStrings = [][]byte{
		[]byte(`"This is a good string"`),
		[]byte("\f\"true\"\n\r"),
		[]byte("\f\"true\n\"\n\r"),
		[]byte(`""`),
		[]byte(`"\n\t\r" `),
		[]byte("\"\n\t\r\""),
		[]byte(`"\"" `),
	}

	parseTestStringsExpected = [][]byte{
		[]byte(`This is a good string`),
		[]byte(`true`),
		[]byte("true\n"),
		[]byte(``),
		[]byte(`\n\t\r`),
		[]byte("\n\t\r"),
		[]byte(`\"`),
	}

	parseTestArrays = [][]byte{
		[]byte(`[]`),
		[]byte(`[1, 2, 3, 4]`),
		[]byte(`["a", "b", "c", "d"]`),
		[]byte(`[true, false, null]`),
		[]byte(`[1.1, 2.2, 3.3, 4.4]`),
		[]byte(`[{"a": 1, "b": 2, "c": 3}]`),
		[]byte(`[["a", 1], ["b", 2], ["c", 3]]`),
		[]byte(`[1, 2.2, "c", true, false, null, ["a"], {"a": 3}]`),
	}

	parseTestObjects = [][]byte{
		[]byte(`{}`),
		[]byte(`{"a": 1, "b": 2, "c": 3}`),
		[]byte(`{"a": [["a", 1], ["b", 2], ["c", 3]], "b": {"b": 3}, "c": null}`),
	}
)

func TestParse(t *testing.T) {
	t.Run("Empty ByteString", func(t *testing.T) {
		r := JSONReader{rawData: []byte(``)}
		assert.Equal(t, ErrEmpty, r.parse())
	})

	t.Run("ParseValue Error", func(t *testing.T) {
		r := JSONReader{rawData: []byte(`["Missing close"`)}

		var err error
		defer func() {
			assert.True(t, r.Empty)
			assert.Equal(t, `expected ']', found '"' at position 15`, err.Error())
		}()
		defer PanicRecovery(&err)

		r.parse()
	})

	t.Run("Malformed ByteString", func(t *testing.T) {
		r := JSONReader{rawData: []byte(`Totally not JSON`)}
		assert.Equal(t, ErrMalformedJSON, r.parse())
	})
}

func TestParseKey(t *testing.T) {
	t.Run("Now With More Whitespace", func(t *testing.T) {
		r := JSONReader{rawData: []byte(`{ "a" : "b" }`)}
		b, i := r.parseKey(1)
		assert.Equal(t, []byte(`a`), b)
		assert.Equal(t, 7, i)
	})

	t.Run("Malformed JSON before Key Separator", func(t *testing.T) {
		r := JSONReader{rawData: []byte(`{ "a" a : "b" }`)}
		b, i := r.parseKey(1)
		assert.Equal(t, []byte(nil), b)
		assert.Equal(t, -1, i)
	})

	t.Run("No Key Separator", func(t *testing.T) {
		r := JSONReader{rawData: []byte(`{ "a" "b" }`)}
		b, i := r.parseKey(1)
		assert.Equal(t, []byte(nil), b)
		assert.Equal(t, -1, i)
	})
}

func TestParseKeyValue(t *testing.T) {
	t.Run("Malformed Value", func(t *testing.T) {
		r, _ := NewJSONReader([]byte(`{"a": b }`))
		b, i := r.parseKeyValue(5)
		assert.Equal(t, parsed{}, b)
		assert.Equal(t, -1, i)
	})
}

func TestParseString(t *testing.T) {
	t.Run("Malformed Value", func(t *testing.T) {
		var err error
		var b parsed

		defer func() {
			assert.Equal(t, `expected '"', found 'T' at position 0`, err.Error())
			assert.Equal(t, parsed{}, b)
		}()
		defer PanicRecovery(&err)

		r := JSONReader{rawData: []byte(`Totally not a string`)}
		b, _ = r.parseString(0)

	})

	t.Run("No Terminating Quote", func(t *testing.T) {
		var err error
		var b parsed

		defer func() {
			assert.Equal(t, "unterminated string at starting position 9", err.Error())
			assert.Equal(t, parsed{}, b)
		}()
		defer PanicRecovery(&err)

		r := JSONReader{rawData: []byte(`{ "key": "Totally no terminal quote}`)}
		b, _ = r.parseString(9)
	})
}

func TestParseConst(t *testing.T) {
	t.Run("Invalid Value", func(t *testing.T) {
		var b parsed
		var err error
		r := JSONReader{rawData: []byte(`TotallyNotTrue`)}

		defer func() {
			assert.True(t, r.Empty)
			assert.Equal(t, parsed{}, b)
			assert.Equal(t, `expected const at position 0`, err.Error())
		}()
		defer PanicRecovery(&err)

		b, _ = r.parseConst(0)
	})
}

func TestParseNumber(t *testing.T) {
	t.Run("Invalid Value", func(t *testing.T) {
		var b parsed
		var err error
		r := JSONReader{rawData: []byte(`a43`)}

		defer func() {
			assert.True(t, r.Empty)
			assert.Equal(t, parsed{}, b)
			assert.Equal(t, `expected number at position 0, found 'a43'`, err.Error())
		}()
		defer PanicRecovery(&err)

		b, _ = r.parseNumber(0)
	})
}

func TestParserObjects(t *testing.T) {
	for k, i := range parseTestObjects {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			r := JSONReader{rawData: i}
			err := r.parse()
			assert.Nil(t, err)
			assert.Equal(t, i, r.rawData)
			assert.Equal(t, "object", r.Type)
		})
	}
}

func TestParserArrays(t *testing.T) {
	for k, i := range parseTestArrays {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			r := JSONReader{rawData: i}
			err := r.parse()
			assert.Nil(t, err)
			assert.Equal(t, i, r.rawData)
			assert.Equal(t, "array", r.Type)
		})
	}
}

func TestParserStrings(t *testing.T) {
	for k, i := range parseTestStrings {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			r := JSONReader{rawData: i}
			err := r.parse()
			assert.Nil(t, err)
			assert.Equal(t, parseTestStringsExpected[k], r.GetByteSlice("0"))
			assert.Equal(t, "string", r.Type)
		})
	}
}

func TestParserNumbers(t *testing.T) {
	for k, i := range parseTestNumbers {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			r := JSONReader{rawData: i}
			err := r.parse()
			assert.Nil(t, err)
			assert.Equal(t, i, r.GetByteSlice("0"))
			assert.True(t, stringInArray(r.Type, []string{"int", JSONFloat}))
		})
	}
}

func TestParserNull(t *testing.T) {
	r := JSONReader{rawData: []byte(`null`)}
	err := r.parse()
	assert.Nil(t, err)
	assert.Equal(t, []byte{'n', 'u', 'l', 'l'}, r.GetByteSlice("0"))
}

func TestParserConstants(t *testing.T) {
	for k, i := range parseTestConstants {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			r := JSONReader{rawData: i}
			err := r.parse()
			assert.Nil(t, err)
			assert.Equal(t, i, r.GetByteSlice("0"))
			assert.True(t, stringInArray(r.Type, []string{"bool", "null"}))
		})
	}
}
