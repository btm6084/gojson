package gojson

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// int, bool, []byte, string, float64, interface{}
// []int, []bool, [][]byte, []string, []float64, []interface{}
// map[string]int, map[string]bool, map[string][]byte, map[string]string, map[string]float64, map[string]interface{}
// map[string]struct, []struct, struct

type valueReceiver struct {
	A string
}

func (s valueReceiver) UnmarshalJSON(b []byte) error {
	return errors.New("valueReceiver Unmarshaler Called")
}

type pointerReceiver struct {
	A string
}

func (p *pointerReceiver) UnmarshalJSON(b []byte) error {
	return errors.New("pointerReceiver Unmarshaler Called")
}

type ImplementsPostUnmarshalerValid struct {
	Thing []string `json:"thing"`
}

func (tt *ImplementsPostUnmarshalerValid) PostUnmarshalJSON(b []byte, err error) error {
	tt.Thing = []string{"From PostUnmarshalJSON"}
	return nil
}

type ImplementsPostUnmarshalerErrorPanic struct {
	Thing []string `json:"thing"`
}

func (tt *ImplementsPostUnmarshalerErrorPanic) PostUnmarshalJSON(b []byte, err error) error {
	panic(errors.New("Error from PostUnmarshalJSON via Panic"))
}

type ImplementsPostUnmarshalerErrorReturn struct {
	Thing []string `json:"thing"`
}

func (tt *ImplementsPostUnmarshalerErrorReturn) PostUnmarshalJSON(b []byte, err error) error {
	return errors.New("Error from PostUnmarshalJSON via Return")
}

type PostUnmarshalerGetsUnmarshalError struct {
	Thing []string `json:"thing"`
}

func (tt *PostUnmarshalerGetsUnmarshalError) PostUnmarshalJSON(b []byte, err error) error {
	return err
}

func TestUnmarshal(t *testing.T) {
	t.Run("Unmarshal empty array", func(t *testing.T) {
		var m []string
		assert.Nil(t, Unmarshal([]byte(`[]`), &m))
		assert.Len(t, m, 0)
	})

	t.Run("Unmarshal empty object", func(t *testing.T) {
		var m map[string]string
		assert.Nil(t, Unmarshal([]byte(`{}`), &m))
		assert.Len(t, m, 0)
	})

	t.Run("InvalidJSON default case", func(t *testing.T) {
		var m string
		assert.Equal(t, "malformed json provided", Unmarshal([]byte(`This is not json`), &m).Error())
	})

	t.Run("InvalidJSON interface case", func(t *testing.T) {
		var m interface{}
		assert.Equal(t, "malformed json provided", Unmarshal([]byte(`This is not json`), &m).Error())
	})

	t.Run("Empty Input", func(t *testing.T) {
		var m string
		assert.Equal(t, `empty json value provided`, Unmarshal([]byte{}, &m).Error())
	})

	type PointerTest struct {
		Level1 *string  `json:"Level 1"`
		Level2 **string `json:"Level 2"`
	}

	t.Run("Map Map Pointer", func(t *testing.T) {
		var m map[string]*map[string]PointerTest
		data := []byte(`{"test": { "Level 1": { "Level 2": "Level 2 String"}}}`)

		Unmarshal(data, &m)
		mp := *m["test"]
		assert.Equal(t, "Level 2 String", **mp["Level 1"].Level2)
	})

	t.Run("Map with Struct with Pointer", func(t *testing.T) {
		var m map[string]PointerTest
		data := []byte(`{"test": { "Level 1": "Level 1 String"}}`)

		Unmarshal(data, &m)
		assert.Equal(t, "Level 1 String", *m["test"].Level1)
	})

	t.Run("Map Pointer with Struct with Pointer", func(t *testing.T) {
		var m *map[string]PointerTest
		data := []byte(`{"test": { "Level 1": "Level 1 String"}}`)

		Unmarshal(data, &m)
		assert.Equal(t, "Level 1 String", *(*m)["test"].Level1)
	})

	t.Run("Pointer Slice", func(t *testing.T) {
		var m []*PointerTest
		data := []byte(`[{ "Level 1": "Array 1 String"}, { "Level 1": "Array 2 String"}, { "Level 1": "Array 3 String"}]`)

		Unmarshal(data, &m)
		assert.Equal(t, "Array 1 String", *m[0].Level1)
		assert.Equal(t, "Array 2 String", *m[1].Level1)
		assert.Equal(t, "Array 3 String", *m[2].Level1)
	})

	t.Run("Pointer Struct", func(t *testing.T) {
		var m *PointerTest
		data := []byte(`{ "Level 1": "Struct 1 String", "Level 2": "Struct 2 String"}`)

		Unmarshal(data, &m)
		assert.Equal(t, "Struct 1 String", *(*m).Level1)
		assert.Equal(t, "Struct 2 String", **(*m).Level2)
	})

	t.Run("Pointer Interface", func(t *testing.T) {
		var m *interface{}
		data := []byte(`[{ "Level 1": "Array 1 String"}, { "Level 1": "Array 2 String"}, { "Level 1": "Array 3 String"}]`)

		Unmarshal(data, &m)
		iface := (*m).([]interface{})

		level1 := iface[0].(map[string]interface{})
		assert.Equal(t, "Array 1 String", level1["Level 1"])

		level2 := iface[1].(map[string]interface{})
		assert.Equal(t, "Array 2 String", level2["Level 1"])

		level3 := iface[2].(map[string]interface{})
		assert.Equal(t, "Array 3 String", level3["Level 1"])
	})

	t.Run("Pointer String", func(t *testing.T) {
		var m *string
		data := []byte(`"This is a good string"`)

		Unmarshal(data, &m)
		assert.Equal(t, "This is a good string", *m)
	})

	t.Run("Large Struct", func(t *testing.T) {
		var m TestComponentResponse

		err := Unmarshal([]byte(largeJSONTestBlob), &m)
		assert.Len(t, m.Items, 19)
		assert.Nil(t, err)
	})

	t.Run("Large Map String Interface", func(t *testing.T) {
		var m map[string]interface{}

		err := Unmarshal([]byte(largeJSONTestBlob), &m)
		assert.Len(t, m["items"], 19)
		assert.Nil(t, err)
	})

	t.Run("Struct Invalid JSON Strict", func(t *testing.T) {
		type A struct {
			A string
			B int
		}

		input := `"a": "test value", "b": 762}`

		var m A
		err := UnmarshalStrict([]byte(input), &m)
		assert.Equal(t, "attempt to unmarshal JSON value with type 'string' into struct", err.Error())
	})

	t.Run("Struct Invalid JSON", func(t *testing.T) {
		type A struct {
			A string
			B int
		}

		input := `"a": "test value", "b": 762}`

		var m A
		err := Unmarshal([]byte(input), &m)
		assert.Nil(t, err)
	})

	t.Run("Struct", func(t *testing.T) {
		type subTest struct {
			A string
			C string
		}
		type test struct {
			Bool        bool   `json:"bool"`
			String      string `json:"-,omitempty"`
			EmptyString string
			Float       float64     `json:"-"`
			Int         int         `json:",omitempty"`
			StringSlice interface{} `json:"string_slice,omitempty"`
			FloatSlice  []float64   `json:"float_slice"`
			IgnoreMe    interface{}
			dontTouch   interface{}
			SubObject   subTest `json:"object"`
			Complex     map[string]interface{}
		}

		var m test
		err := Unmarshal([]byte(readerTestData), &m)

		assert.Nil(t, err)
		assert.Equal(t, true, m.Bool)
		assert.Equal(t, "", m.String)
		assert.Equal(t, "", m.EmptyString)
		assert.Equal(t, 0.0, m.Float)
		assert.Equal(t, 17, m.Int)
		assert.Equal(t, []interface{}{"a", "b", "c", "d", "e", "t"}, m.StringSlice)
		assert.Equal(t, []float64{-1.1, 0, 1.1, 2.2, 3.3}, m.FloatSlice)
		assert.Equal(t, interface{}(nil), m.IgnoreMe)
		assert.Equal(t, interface{}(nil), m.dontTouch)
		assert.Equal(t, subTest{"b", "d"}, m.SubObject)
		assert.Equal(t, map[string]interface{}{"0": "a", "1": 2, "2": interface{}(nil), "3": false, "4": 2.2, "5": map[string]interface{}{"c": "d", "empty_string": ""}, "6": []interface{}{"s"}}, m.Complex)
	})

	t.Run("Struct ValueReceiver Unmarshaler", func(t *testing.T) {
		var m valueReceiver
		data := []byte(`{"A": "B"}`)
		err := Unmarshal([]byte(data), &m)
		assert.Equal(t, "valueReceiver Unmarshaler Called", err.Error())
	})

	t.Run("Struct PointerReceiver Unmarshaler", func(t *testing.T) {
		var m pointerReceiver
		data := []byte(`{"A": "B"}`)
		err := Unmarshal([]byte(data), &m)
		assert.Equal(t, "pointerReceiver Unmarshaler Called", err.Error())
	})

	t.Run("Struct Key Error Unmarshaler", func(t *testing.T) {
		var m struct {
			A string
		}
		data := []byte(`{A: "B"}`)
		err := Unmarshal([]byte(data), &m)
		assert.Equal(t, "expected object key at position 1 in segment '{A: \"B\"}'", err.Error())
	})

	t.Run("Unmarshal Struct, Slice Error", func(t *testing.T) {
		var m struct {
			A []string `json:"a"`
		}
		data := `{ "a": [ a b c d ] }`
		err := Unmarshal([]byte(data), &m)
		assert.Equal(t, "invalid character 'a' at position '2' in segment '[ a b c d ]'", err.Error())
	})

	t.Run("Unmarshal Struct, Map Error", func(t *testing.T) {
		var m struct {
			A map[string]string `json:"a"`
		}
		data := `{ "a": { b: "c" } }`
		err := Unmarshal([]byte(data), &m)
		assert.Equal(t, "expected object key at position 1 in segment '{ b: \"c\" }'", err.Error())
	})

	t.Run("Unmarshal Struct, Struct Error", func(t *testing.T) {
		type mb struct {
			B string `json:"b"`
		}
		var m struct {
			A mb `json:"a"`
		}
		data := `{ "a": { b: "c" } }`
		err := Unmarshal([]byte(data), &m)
		assert.Equal(t, "expected object key at position 1 in segment '{ b: \"c\" }'", err.Error())
	})

	t.Run("Struct Slice", func(t *testing.T) {
		type B struct {
			Name string
			ID   string
		}

		type A struct {
			A int
			B string
			C B
		}

		input := `[{"a": 1, "b": "two", "c":{"name": "First", "id": "Test"}}, {"a": 2, "b": "three", "c":{"name": "Second", "id": "Unittest"}}]`

		var m []A
		err := Unmarshal([]byte(input), &m)

		assert.Nil(t, err)
		assert.Equal(t, 2, len(m))

		assert.Equal(t, 1, m[0].A)
		assert.Equal(t, "two", m[0].B)
		assert.Equal(t, "First", m[0].C.Name)
		assert.Equal(t, "Test", m[0].C.ID)

		assert.Equal(t, 2, m[1].A)
		assert.Equal(t, "three", m[1].B)
		assert.Equal(t, "Second", m[1].C.Name)
		assert.Equal(t, "Unittest", m[1].C.ID)
	})

	t.Run("Struct Map", func(t *testing.T) {
		type A struct {
			A string
			B int
		}

		input := `{"keyname1":{"a": "test data", "b": 101}, "keyname2":{"a": "test value", "b": 762}}`

		var m map[string]A
		err := Unmarshal([]byte(input), &m)

		assert.Nil(t, err)
		assert.Equal(t, 2, len(m))

		assert.Equal(t, "test data", m["keyname1"].A)
		assert.Equal(t, 101, m["keyname1"].B)

		assert.Equal(t, "test value", m["keyname2"].A)
		assert.Equal(t, 762, m["keyname2"].B)
	})

	t.Run("Map String Interface", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected interface{}
		}{
			{"String", `"This is a good string"`, map[string]interface{}{"0": "This is a good string"}},
			{"Integer", "193", map[string]interface{}{"0": 193}},
			{"Float", "-122.54", map[string]interface{}{"0": -122.54}},
			{"True", "true", map[string]interface{}{"0": true}},
			{"False", "false", map[string]interface{}{"0": false}},
			{"Array", `["a", 1, false]`, map[string]interface{}{"0": "a", "1": 1, "2": false}},
			{"Object", `{"a":null, "b":1, "c":false}`, map[string]interface{}{"a": nil, "b": 1, "c": false}},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m map[string]interface{}
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Map String Float", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected map[string]float64
		}{
			{"BoolSlice", "[true, false, null]", map[string]float64{"0": 1, "1": 0, "2": 0}},
			{"StringSlice", `["-1", "0", "1", "2.2", "3e+4", "7e-4"]`, map[string]float64{"0": -1, "1": 0, "2": 1, "3": 2.2, "4": 30000, "5": 0.0007}},
			{"IntegerSlice", "[-1e-1, -1, 0, 1, 1e1, 2e-2, 42]", map[string]float64{"0": -0.1, "1": -1, "2": 0, "3": 1, "4": 10, "5": 0.02, "6": 42}},
			{"FloatSlice", "[-1.1, 0.0, 0.1, 2.2, -1.0e4, -1e-4]", map[string]float64{"0": -1.1, "1": 0, "2": 0.1, "3": 2.2, "4": -10000, "5": -0.0001}},
			{"String", `"This is a good string 98.7"`, map[string]float64{"0": 0}},
			{"Integer", "193", map[string]float64{"0": 193}},
			{"Float", "-122.54", map[string]float64{"0": -122.54}},
			{"True", "true", map[string]float64{"0": 1}},
			{"False", "false", map[string]float64{"0": 0}},
			{"Array", `["a", 1, false]`, map[string]float64{"0": 0, "1": 1, "2": 0}},
			{"Object", `{"a":null, "b":1, "c":34.7322, "d": "199.01"}`, map[string]float64{"a": 0, "b": 1, "c": 34.7322, "d": 199.01}},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m map[string]float64
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Map String Bytes", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected map[string][]byte
		}{
			{"BoolSlice", "[true, false, null]", map[string][]byte{"0": []byte("true"), "1": []byte("false"), "2": []byte(nil)}},
			{"StringSlice", `["-1", "0", "1", "2.2", "3e+4", "7e-4"]`, map[string][]byte{"0": []byte("-1"), "1": []byte("0"), "2": []byte("1"), "3": []byte("2.2"), "4": []byte("3e+4"), "5": []byte("7e-4")}},
			{"IntegerSlice", "[-1e-1, -1, 0, 1, 1e1, 2e-2, 42]", map[string][]byte{"0": []byte("-1e-1"), "1": []byte("-1"), "2": []byte("0"), "3": []byte("1"), "4": []byte("1e1"), "5": []byte("2e-2"), "6": []byte("42")}},
			{"FloatSlice", "[-1.1, 0.0, 0.1, 2.2, -1.0e4, -1e-4]", map[string][]byte{"0": []byte("-1.1"), "1": []byte("0.0"), "2": []byte("0.1"), "3": []byte("2.2"), "4": []byte("-1.0e4"), "5": []byte("-1e-4")}},
			{"String", `"This is a good string"`, map[string][]byte{"0": []byte("This is a good string")}},
			{"Integer", "193", map[string][]byte{"0": []byte("193")}},
			{"Float", "-122.54", map[string][]byte{"0": []byte("-122.54")}},
			{"True", "true", map[string][]byte{"0": []byte("true")}},
			{"False", "false", map[string][]byte{"0": []byte("false")}},
			{"Array", `["a", 1, false]`, map[string][]byte{"0": []byte("a"), "1": []byte("1"), "2": []byte("false")}},
			{"Object", `{"a":null, "b":1, "c":false}`, map[string][]byte{"a": []byte(nil), "b": []byte("1"), "c": []byte("false")}},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m map[string][]byte
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Map String Byte", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected map[string]byte
		}{
			{"BoolSlice", "[true, false, null]", map[string]byte{"16": 0x6c, "9": 0x6c, "11": 0x65, "15": 0x75, "3": 0x75, "6": 0x20, "8": 0x61, "13": 0x20, "18": 0x5d, "1": 0x74, "5": 0x2c, "10": 0x73, "7": 0x66, "12": 0x2c, "14": 0x6e, "17": 0x6c, "0": 0x5b, "2": 0x72, "4": 0x65}},
			{"StringSlice", `["-1", "0", "1", "2.2", "3e+4", "7e-4"]`, map[string]byte{"8": 0x30, "11": 0x20, "21": 0x22, "24": 0x22, "31": 0x20, "32": 0x22, "0": 0x5b, "14": 0x22, "15": 0x2c, "25": 0x33, "37": 0x22, "10": 0x2c, "27": 0x2b, "33": 0x37, "35": 0x2d, "3": 0x31, "18": 0x32, "22": 0x2c, "28": 0x34, "9": 0x22, "23": 0x20, "38": 0x5d, "2": 0x2d, "6": 0x20, "7": 0x22, "16": 0x20, "34": 0x65, "36": 0x34, "1": 0x22, "5": 0x2c, "12": 0x22, "17": 0x22, "19": 0x2e, "20": 0x32, "26": 0x65, "29": 0x22, "4": 0x22, "30": 0x2c, "13": 0x31}},
			{"IntegerSlice", "[-1e-1, -1, 0, 1, 1e1, 2e-2, 42]", map[string]byte{"15": 0x31, "18": 0x31, "27": 0x2c, "3": 0x65, "2": 0x31, "7": 0x20, "19": 0x65, "23": 0x32, "28": 0x20, "0": 0x5b, "9": 0x31, "13": 0x2c, "6": 0x2c, "17": 0x20, "30": 0x32, "10": 0x2c, "4": 0x2d, "5": 0x31, "8": 0x2d, "12": 0x30, "14": 0x20, "16": 0x2c, "26": 0x32, "1": 0x2d, "31": 0x5d, "29": 0x34, "21": 0x2c, "20": 0x31, "22": 0x20, "24": 0x65, "25": 0x2d, "11": 0x20}},
			{"FloatSlice", "[-1.1, 0.0, 0.1, 2.2, -1.0e4, -1e-4]", map[string]byte{"8": 0x2e, "19": 0x32, "22": 0x2d, "25": 0x30, "31": 0x31, "7": 0x30, "11": 0x20, "13": 0x2e, "21": 0x20, "27": 0x34, "28": 0x2c, "35": 0x5d, "1": 0x2d, "10": 0x2c, "16": 0x20, "3": 0x2e, "9": 0x30, "33": 0x2d, "32": 0x65, "34": 0x34, "2": 0x31, "12": 0x30, "23": 0x31, "20": 0x2c, "24": 0x2e, "30": 0x2d, "4": 0x31, "17": 0x32, "18": 0x2e, "14": 0x31, "15": 0x2c, "26": 0x65, "29": 0x20, "0": 0x5b, "5": 0x2c, "6": 0x20}},
			{"String", `"This is a good string"`, map[string]byte{"14": 0x20, "18": 0x69, "5": 0x69, "12": 0x6f, "17": 0x72, "0": 0x54, "3": 0x73, "7": 0x20, "8": 0x61, "11": 0x6f, "10": 0x67, "13": 0x64, "15": 0x73, "1": 0x68, "2": 0x69, "4": 0x20, "6": 0x73, "9": 0x20, "16": 0x74, "19": 0x6e, "20": 0x67}},
			{"Integer", "193", map[string]byte{"2": 0x33, "0": 0x31, "1": 0x39}},
			{"Float", "-122.54", map[string]byte{"1": 0x31, "2": 0x32, "3": 0x32, "4": 0x2e, "5": 0x35, "6": 0x34, "0": 0x2d}},
			{"True", "true", map[string]byte{"0": 0x74, "1": 0x72, "2": 0x75, "3": 0x65}},
			{"False", "false", map[string]byte{"3": 0x73, "4": 0x65, "0": 0x66, "1": 0x61, "2": 0x6c}},
			{"Array", `["a", 1, false]`, map[string]byte{"12": 0x73, "2": 0x61, "3": 0x22, "6": 0x31, "9": 0x66, "10": 0x61, "13": 0x65, "1": 0x22, "4": 0x2c, "11": 0x6c, "0": 0x5b, "5": 0x20, "7": 0x2c, "8": 0x20, "14": 0x5d}},
			{"Object", `{"a":null, "b":1, "c":false}`, map[string]byte{"3": 0x22, "26": 0x65, "11": 0x22, "16": 0x2c, "22": 0x66, "6": 0x75, "12": 0x62, "23": 0x61, "27": 0x7d, "1": 0x22, "8": 0x6c, "5": 0x6e, "7": 0x6c, "13": 0x22, "21": 0x3a, "2": 0x61, "4": 0x3a, "15": 0x31, "18": 0x22, "19": 0x63, "0": 0x7b, "10": 0x20, "20": 0x22, "24": 0x6c, "14": 0x3a, "17": 0x20, "9": 0x2c, "25": 0x73}},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m map[string]byte
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Map String String", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected map[string]string
		}{
			{"BoolSlice", "[true, false, null]", map[string]string{"0": "true", "1": "false", "2": ""}},
			{"StringSlice", `["-1", "0", "1", "2.2", "3e+4", "7e-4"]`, map[string]string{"0": "-1", "1": "0", "2": "1", "3": "2.2", "4": "3e+4", "5": "7e-4"}},
			{"IntegerSlice", "[-1e-1, -1, 0, 1, 1e1, 2e-2, 42]", map[string]string{"0": "-1e-1", "1": "-1", "2": "0", "3": "1", "4": "1e1", "5": "2e-2", "6": "42"}},
			{"FloatSlice", "[-1.1, 0.0, 0.1, 2.2, -1.0e4, -1e-4]", map[string]string{"0": "-1.1", "1": "0.0", "2": "0.1", "3": "2.2", "4": "-1.0e4", "5": "-1e-4"}},
			{"String", `"This is a good string"`, map[string]string{"0": "This is a good string"}},
			{"Integer", "193", map[string]string{"0": "193"}},
			{"Float", "-122.54", map[string]string{"0": "-122.54"}},
			{"True", "true", map[string]string{"0": "true"}},
			{"False", "false", map[string]string{"0": "false"}},
			{"Array", `["a", 1, false]`, map[string]string{"0": "a", "1": "1", "2": "false"}},
			{"Object", `{"a":null, "b":1, "c":false}`, map[string]string{"a": "", "b": "1", "c": "false"}},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m map[string]string
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Map String Bool", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected map[string]bool
		}{
			{"BoolSlice", "[true, false, null]", map[string]bool{"0": true, "1": false, "2": false}},
			{"StringSlice", `["true", "false", "null", "", "A", "T"]`, map[string]bool{"0": true, "1": false, "2": false, "3": false, "4": false, "5": true}},
			{"IntegerSlice", "[-1e-1, -1, 0, 1, 1e1, 2e-2, 42]", map[string]bool{"0": true, "1": true, "2": false, "3": true, "4": true, "5": true, "6": true}},
			{"FloatSlice", "[-1.1, 0.0, 0.1, 2.2, -1.0e4, -1e-4]", map[string]bool{"0": true, "1": false, "2": true, "3": true, "4": true, "5": true}},
			{"String", `"This is a good string"`, map[string]bool{"0": false}},
			{"String False", `"f"`, map[string]bool{"0": false}},
			{"Integer", "193", map[string]bool{"0": true}},
			{"Float", "-122.54", map[string]bool{"0": true}},
			{"True", "true", map[string]bool{"0": true}},
			{"False", "false", map[string]bool{"0": false}},
			{"Array", `["a", 1, false]`, map[string]bool{"2": false, "0": false, "1": true}},
			{"Object", `{"a":null, "b":1, "c":false}`, map[string]bool{"a": false, "b": true, "c": false}},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m map[string]bool
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Map String Int", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected map[string]int
		}{
			{"BoolSlice", "[true, false, null]", map[string]int{"0": 1, "1": 0, "2": 0}},
			{"StringSlice", `["-1", "0", "1", "2.2", "3e+4", "7e-4"]`, map[string]int{"0": -1, "1": 0, "2": 1, "3": 2, "4": 3e4, "5": 0}},
			{"IntegerSlice", "[-1e-1, -1, 0, 1, 1e1, 2e-2, 42]", map[string]int{"0": 0, "1": -1, "2": 0, "3": 1, "4": 10, "5": 0, "6": 42}},
			{"FloatSlice", "[-1.1, 0.0, 0.1, 2.2, -1.0e4, -1e-4]", map[string]int{"0": -1, "1": 0, "2": 0, "3": 2, "4": -10000, "5": 0}},
			{"String", `"This is a good string"`, map[string]int{"0": 0}},
			{"Integer", "193", map[string]int{"0": 193}},
			{"Float", "-122.54", map[string]int{"0": -122}},
			{"True", "true", map[string]int{"0": 1}},
			{"False", "false", map[string]int{"0": 0}},
			{"Array", `["a", 1, false]`, map[string]int{"0": 0, "1": 1, "2": 0}},
			{"Object", `{"a":null, "b":1, "c":22}`, map[string]int{"a": 0, "b": 1, "c": 22}},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m map[string]int
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Unmarshal Map, Node Error", func(t *testing.T) {
		var m map[string]interface{}
		data := `[ a b c d ]`
		err := Unmarshal([]byte(data), &m)
		assert.Equal(t, "invalid character 'a' at position '2' in segment '[ a b c d ]'", err.Error())
	})

	t.Run("Unmarshal Map, Slice Error", func(t *testing.T) {
		var m map[string][]interface{}
		data := `[[ a b c d ]]`
		err := Unmarshal([]byte(data), &m)
		assert.Equal(t, "invalid character 'a' at position '2' in segment '[ a b c d ]'", err.Error())
	})

	t.Run("Unmarshal Map, Struct Error Strict", func(t *testing.T) {
		var m map[string]struct {
			a string
		}
		data := `[[ a b c d ]]`
		err := UnmarshalStrict([]byte(data), &m)
		assert.Equal(t, "strict standards: attempt to unmarshal JSON value with type 'array' into map", err.Error())
	})

	t.Run("Unmarshal Map, Struct Error", func(t *testing.T) {
		var m map[string]struct {
			a string
		}
		data := `[[ a b c d ]]`
		err := Unmarshal([]byte(data), &m)
		assert.Nil(t, err)
	})

	t.Run("Unmarshal Map, Map Error", func(t *testing.T) {
		var m map[string]map[string]string
		data := `[{"a":"some string","b":true},{"c":false "d":null},{"e":["a"],"f":{"suba":1}},{"g":123,"h":-123,"i":123.0,"j":1e73}]`
		err := Unmarshal([]byte(data), &m)
		assert.Nil(t, m)
		assert.Equal(t, `expected value terminator ('}', ']' or ',') at position '10' in segment '{"c":false "d":null}'`, err.Error())
	})

	t.Run("Interface Slice", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected []interface{}
		}{
			{"BoolSlice", "[true, false, null]", []interface{}{true, false, interface{}(nil)}},
			{"StringSlice", `["-1", "0", "1", "2.2", "3e+4", "7e-4"]`, []interface{}{"-1", "0", "1", "2.2", "3e+4", "7e-4"}},
			{"IntegerSlice", "[-1e-1, -1, 0, 1, 1e1, 2e-2, 42]", []interface{}{-0.1, -1, 0, 1, 10.0, 0.02, 42}},
			{"FloatSlice", "[-1.1, 0.0, 0.1, 2.2, -1.0e4, -1e-4]", []interface{}{-1.1, 0.0, 0.1, 2.2, -10000.0, -0.0001}},
			{"String", `"This is a good string"`, []interface{}{"This is a good string"}},
			{"Integer", "193", []interface{}{193}},
			{"Float", "-122.54", []interface{}{-122.54}},
			{"True", "true", []interface{}{true}},
			{"False", "false", []interface{}{false}},
			{"Array", `["a", 1, false]`, []interface{}{"a", 1, false}},
			{"Object", `{"a":null, "b":1, "c":22}`, []interface{}{interface{}(nil), 1, 22}},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m []interface{}
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Map Slice", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected []map[string]string
		}{
			{
				"TestData",
				`[{"a":"some string","b":true},{"c":false,"d":null},{"e":["a"],"f":{"suba":1}},{"g":123,"h":-123,"i":123.0,"j":1e73}]`,
				[]map[string]string{
					map[string]string{"a": "some string", "b": "true"},
					map[string]string{"c": "false", "d": ""},
					map[string]string{"e": `["a"]`, "f": `{"suba":1}`},
					map[string]string{"g": "123", "h": "-123", "i": "123.0", "j": "1e73"},
				},
			},
		}

		for _, tc := range testCases {
			t.Run("Float64: "+tc.Label, func(t *testing.T) {
				var m []map[string]string
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Float Slice", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected []float64
		}{
			{"BoolSlice", "[true, false, null]", []float64{1.0, 0.0, 0.0}},
			{"StringSlice", `["-1", "0", "1", "2.2", "3e+4", "7e-4"]`, []float64{-1.0, 0.0, 1.0, 2.2, 30000.0, 0.0007}},
			{"IntegerSlice", "[-1e-1, -1, 0, 1, 1e1, 2e-2, 42]", []float64{-0.1, -1.0, 0.0, 1.0, 10.0, 0.02, 42.0}},
			{"FloatSlice", "[-1.1, 0.0, 0.1, 2.2, -1.0e4, -1e-4]", []float64{-1.1, 0.0, 0.1, 2.2, -10000.0, -0.0001}},
			{"String", `"This is a good string"`, []float64{0.0}},
			{"Integer", "193", []float64{193.0}},
			{"Float", "-122.54", []float64{-122.54}},
			{"True", "true", []float64{1.0}},
			{"False", "false", []float64{0.0}},
			{"Array", `["a", 1, false]`, []float64{0.0, 1.0, 0.0}},
			{"Object", `{"a":null, "b":1, "c":22}`, []float64{0.0, 1.0, 22.0}},
		}

		for _, tc := range testCases {
			t.Run("Float64: "+tc.Label, func(t *testing.T) {
				var m []float64
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Byte Slices", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected [][]byte
		}{
			{"BoolSlice", "[true, false, null]", [][]byte{[]byte("true"), []byte("false"), []byte(nil)}},
			{"StringSlice", `["-1", "0", "1", "2.2", "3e+4", "7e-4"]`, [][]byte{[]byte("-1"), []byte("0"), []byte("1"), []byte("2.2"), []byte("3e+4"), []byte("7e-4")}},
			{"IntegerSlice", "[-1e-1, -1, 0, 1, 1e1, 2e-2, 42]", [][]byte{[]byte("-1e-1"), []byte("-1"), []byte("0"), []byte("1"), []byte("1e1"), []byte("2e-2"), []byte("42")}},
			{"FloatSlice", "[-1.1, 0.0, 0.1, 2.2, -1.0e4, -1e-4]", [][]byte{[]byte("-1.1"), []byte("0.0"), []byte("0.1"), []byte("2.2"), []byte("-1.0e4"), []byte("-1e-4")}},
			{"String", `"This is a good byte"`, [][]byte{[]byte("This is a good byte")}},
			{"Integer", "193", [][]byte{[]byte("193")}},
			{"Float", "-122.54", [][]byte{[]byte("-122.54")}},
			{"True", "true", [][]byte{[]byte("true")}},
			{"False", "false", [][]byte{[]byte("false")}},
			{"Array", `["a", 1, false]`, [][]byte{[]byte("a"), []byte("1"), []byte("false")}},
			{"Object", `{"a":null, "b":1, "c":false, "d": "test string"}`, [][]byte{[]byte(nil), []byte("1"), []byte("false"), []byte("test string")}},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m [][]byte
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Byte Slice", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected []byte
		}{
			{"String", `"This is a good string"`, []byte(`This is a good string`)},
			{"Integer", "193", []byte("193")},
			{"Float", "-122.54", []byte("-122.54")},
			{"True", "true", []byte("true")},
			{"False", "false", []byte("false")},
			{"Array", `["a", 1, false]`, []byte(`["a", 1, false]`)},
			{"Object", `{"a":null, "b":1, "c":false}`, []byte(`{"a":null, "b":1, "c":false}`)},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m []byte
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("String Slice", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected []string
		}{
			{"BoolSlice", "[true, false, null]", []string{"true", "false", ""}},
			{"StringSlice", `["-1", "0", "1", "2.2", "3e+4", "7e-4"]`, []string{"-1", "0", "1", "2.2", "3e+4", "7e-4"}},
			{"IntegerSlice", "[-1e-1, -1, 0, 1, 1e1, 2e-2, 42]", []string{"-1e-1", "-1", "0", "1", "1e1", "2e-2", "42"}},
			{"FloatSlice", "[-1.1, 0.0, 0.1, 2.2, -1.0e4, -1e-4]", []string{"-1.1", "0.0", "0.1", "2.2", "-1.0e4", "-1e-4"}},
			{"String", `"This is a good string"`, []string{"This is a good string"}},
			{"Integer", "193", []string{"193"}},
			{"Float", "-122.54", []string{"-122.54"}},
			{"True", "true", []string{"true"}},
			{"False", "false", []string{"false"}},
			{"Array", `["a", 1, false]`, []string{"a", "1", "false"}},
			{"Object", `{"a":null, "b":1, "c":false}`, []string{"", "1", "false"}},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m []string
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Bool Slice", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected []bool
		}{
			{"BoolSlice", "[true, false, null]", []bool{true, false, false}},
			{"StringSlice", `["true", "false", "null", "", "A", "T"]`, []bool{true, false, false, false, false, true}},
			{"IntegerSlice", "[-1e-1, -1, 0, 1, 1e1, 2e-2, 42]", []bool{true, true, false, true, true, true, true}},
			{"FloatSlice", "[-1.1, 0.0, 0.1, 2.2, -1.0e4, -1e-4]", []bool{true, false, true, true, true, true}},
			{"String", `"This is a good string"`, []bool{false}},
			{"String False", `"f"`, []bool{false}},
			{"Integer", "193", []bool{true}},
			{"Float", "-122.54", []bool{true}},
			{"True", "true", []bool{true}},
			{"False", "false", []bool{false}},
			{"Array", `["a", 1, false]`, []bool{false, true, false}},
			{"Object", `{"a":null, "b":1, "c":false}`, []bool{false, true, false}},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m []bool
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Int Slice", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected []int
		}{
			{"BoolSlice", "[true, false, null]", []int{1, 0, 0}},
			{"StringSlice", `["-1", "0", "1", "2.2", "3e+4", "7e-44"]`, []int{-1, 0, 1, 2, 3e4, 0}},
			{"IntegerSlice", "[-1e-1, -1, 0, 1, 1e1, 2e-2, 42]", []int{0, -1, 0, 1, 10, 0, 42}},
			{"FloatSlice", "[-1.1, 0.0, 0.1, 2.2, -1.0e4, -1e-4]", []int{-1, 0, 0, 2, -10000, 0}},
			{"String", `"This is a good string"`, []int{0}},
			{"Integer", "193", []int{193}},
			{"Float", "-122.54", []int{-122}},
			{"True", "true", []int{1}},
			{"False", "false", []int{0}},
			{"Array", `["a", 1, false]`, []int{0, 1, 0}},
			{"Object", `{"a":null, "b":1, "c":22}`, []int{0, 1, 22}},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m []int
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Unmarshal Slice, Node Error", func(t *testing.T) {
		var m []interface{}
		data := `[ a b c d ]`
		err := Unmarshal([]byte(data), &m)
		assert.Equal(t, "invalid character 'a' at position '2' in segment '[ a b c d ]'", err.Error())
	})

	t.Run("Unmarshal Slice, Slice Error", func(t *testing.T) {
		var m [][]interface{}
		data := `[[ a b c d ]]`
		err := Unmarshal([]byte(data), &m)
		assert.Equal(t, "invalid character 'a' at position '2' in segment '[ a b c d ]'", err.Error())
	})

	t.Run("Unmarshal Slice, Struct Error Strict", func(t *testing.T) {
		var m []struct {
			a string
		}
		data := `[[ a b c d ]]`
		err := UnmarshalStrict([]byte(data), &m)
		assert.Equal(t, "attempt to unmarshal JSON value with type 'array' into struct", err.Error())
	})

	t.Run("Unmarshal Slice, Struct Error", func(t *testing.T) {
		var m []struct {
			a string
		}
		data := `[[ a b c d ]]`
		err := Unmarshal([]byte(data), &m)
		assert.Nil(t, err)
	})

	t.Run("Unmarshal Slice, Map Error", func(t *testing.T) {
		var m []map[string]string
		data := `[{"a":"some string","b":true},{"c":false "d":null},{"e":["a"],"f":{"suba":1}},{"g":123,"h":-123,"i":123.0,"j":1e73}]`
		err := Unmarshal([]byte(data), &m)
		assert.Nil(t, m)
		assert.Equal(t, `expected value terminator ('}', ']' or ',') at position '10' in segment '{"c":false "d":null}'`, err.Error())
	})

	t.Run("Interface Slice", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected interface{}
		}{
			{"String", `"This is a good string"`, "This is a good string"},
			{"Integer", "193", 193},
			{"Float", "-122.54", -122.54},
			{"True", "true", true},
			{"False", "false", false},
			{"Array", `["a", 1, false]`, []interface{}{"a", 1, false}},
			{"Object", `{"a":null, "b":1, "c":false}`, map[string]interface{}{"a": nil, "b": 1, "c": false}},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m interface{}
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Floats", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected float64
		}{
			{"Integer", "193", 193.0},
			{"Exponent Pos", "1e3", 1.0e3},
			{"Exponent Neg", "4e-3", 0.004},
			{"Exponent Plus", "1e+3", 1.0e+3},
			{"Float", "173.22", 173.22},
			{"Integer 0", "0", 0.0},
			{"Float 0", "0.0", 0.0},
			{"Negative Integer", "-193", -193.0},
			{"Negative Exponent Pos", "-1e3", -1.0e3},
			{"Negative Exponent Neg", "-4e-3", -0.004},
			{"Negative Exponent Plus", "-1e+3", -1.0e+3},
			{"Negative Float", "-173.22", -173.22},
			{"Negative Integer 0", "-0", 0.0},
			{"Negative Float 0", "-0.0", 0.0},
		}

		for _, tc := range testCases {
			t.Run("Float64: "+tc.Label, func(t *testing.T) {
				var m float64
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}

		for _, tc := range testCases {
			t.Run("Float32: "+tc.Label, func(t *testing.T) {
				var m float32
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, float32(tc.Expected), m)
			})
		}
	})

	t.Run("Strings", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected string
		}{
			{"Good String", `"This is a good string"`, "This is a good string"},
			{"Empty String", `""`, ``},
			{"String with Quotes", `"This is a \"quoted\" string"`, `This is a "quoted" string`},
			{"String with newline escape sequence", `"Line 1\nLine2"`, "Line 1\nLine2"},
			{"String with Quotes, Line Feed, Carriage Return", "\f\"true\"\n\r", `true`},
			{"String with Escaped Literals", `"\n\t\r" `, "\n\t\r"},
			{"String that is nothing but a quote", `"\"" `, `"`},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m string
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Booleans", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected bool
		}{
			{"Bool True", "true", true},
			{"Bool False", "false", false},
			{"String True", `"true"`, true},
			{"String False", `"false"`, false},
			{"Integer True", "1", true},
			{"Integer False", "0", false},
			{"Float True", "1.0", true},
			{"Float False", "0.0", false},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m bool
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Ints", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected int
		}{
			{"Integer", "193", 193},
			{"Exponent Pos", "1e3", 1e3},
			{"Exponent Neg", "4e-3", 0},
			{"Exponent Plus", "1e+3", 1e+3},
			{"Float", "173.22", 173},
			{"Integer 0", "0", 0},
			{"Float 0", "0.0", 0},
			{"Negative Integer", "-193", -193},
			{"Negative Exponent Pos", "-1e3", -1e3},
			{"Negative Exponent Neg", "-4e-3", 0},
			{"Negative Exponent Plus", "-1e+3", -1e+3},
			{"Negative Float", "-173.22", -173},
			{"Negative Integer 0", "-0", 0},
			{"Negative Float 0", "-0.0", 0},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m int
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Int64s", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected int64
		}{
			{"Integer", "193", 193},
			{"Exponent Pos", "1e3", 1e3},
			{"Exponent Neg", "4e-3", 0},
			{"Exponent Plus", "1e+3", 1e+3},
			{"Float", "173.22", 173},
			{"Integer 0", "0", 0},
			{"Float 0", "0.0", 0},
			{"Negative Exponent Neg", "-4e-3", 0},
			{"Negative Integer 0", "-0", 0},
			{"Negative Float 0", "-0.0", 0},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m int64
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("UInts", func(t *testing.T) {
		testCases := []struct {
			Label    string
			Input    string
			Expected uint
		}{
			{"Integer", "193", 193},
			{"Exponent Pos", "1e3", 1e3},
			{"Exponent Neg", "4e-3", 0},
			{"Exponent Plus", "1e+3", 1e+3},
			{"Float", "173.22", 173},
			{"Integer 0", "0", 0},
			{"Float 0", "0.0", 0},
			{"Negative Exponent Neg", "-4e-3", 0},
			{"Negative Integer 0", "-0", 0},
			{"Negative Float 0", "-0.0", 0},
		}

		for _, tc := range testCases {
			t.Run(tc.Label, func(t *testing.T) {
				var m uint
				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Channel Container", func(t *testing.T) {
		m := make(chan int)
		err := Unmarshal([]byte(readerTestData), &m)
		assert.Equal(t, "Unmarshal: Invalid Container Type 'chan'", err.Error())
	})

	t.Run("Slice Channel Container", func(t *testing.T) {
		type ch chan int
		m := make([]ch, 1)
		err := Unmarshal([]byte(`{"a":"b"}`), &m)
		assert.Equal(t, "Unmarshal: Invalid Container Type 'chan'", err.Error())
	})

	t.Run("Map Channel Container", func(t *testing.T) {
		type ch chan int
		m := make(map[string]ch, 1)
		err := Unmarshal([]byte(`{"a":"b"}`), &m)
		assert.Equal(t, "Unmarshal: Invalid Container Type 'chan'", err.Error())
	})

	t.Run("Map Channel Container", func(t *testing.T) {
		type ch chan int
		var m struct {
			A ch `json:"a"`
		}
		err := Unmarshal([]byte(`{"a":"b"}`), &m)
		assert.Equal(t, "Unmarshal: Invalid Container Type 'chan'", err.Error())
	})

	t.Run("Non-Pointer Container", func(t *testing.T) {
		var m map[string]interface{}
		err := Unmarshal([]byte(`{"a":"b"}`), m)
		assert.Equal(t, err, fmt.Errorf("supplied container (v) must be a pointer"))

		var s []interface{}
		err = Unmarshal([]byte(`{"a":"b"}`), s)
		assert.Equal(t, err, fmt.Errorf("supplied container (v) must be a pointer"))

		var i interface{}
		err = Unmarshal([]byte(`{"a":"b"}`), i)
		assert.Equal(t, err, fmt.Errorf("supplied container (v) must be a pointer"))

		var b bool
		err = Unmarshal([]byte(`{"a":"b"}`), b)
		assert.Equal(t, err, fmt.Errorf("supplied container (v) must be a pointer"))
	})

	t.Run("Complex Object with Large Number of Nodes", func(t *testing.T) {
		var m []int
		data := []byte{'['}

		for i := 0; i < 99999; i++ {
			data = append(data, []byte{'5', ',', ' '}...)
		}

		data = append(data, []byte{'5', ']'}...)

		err := Unmarshal(data, &m)
		assert.Nil(t, err)
		assert.Len(t, m, 100000)
	})

	t.Run("Valid Extraction, Invalid JSON", func(t *testing.T) {
		testCases := []struct {
			Input    string
			Expected string
		}{
			{`[[]`, `expected value terminator ('}', ']' or ',') at position '3' in segment '[[]'`},
			{`[]]`, `invalid character ']' at position '1' in segment '[]]'`},
			{`[]}`, `invalid character ']' at position '1' in segment '[]}'`},
			{`[{}`, `expected value terminator ('}', ']' or ',') at position '3' in segment '[{}'`},
			{`{[]`, `expected object key at position 1 in segment '{[]'`},
			{`{{}`, `expected object key at position 1 in segment '{{}'`},
			{`{}]`, `expected object key at position 1 in segment '{}]'`},
			{`{}}`, `expected object key at position 1 in segment '{}}'`},
		}

		for i, tc := range testCases {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				var m map[string]interface{}

				err := Unmarshal([]byte(tc.Input), &m)
				assert.Equal(t, tc.Expected, err.Error())
			})
		}

	})

	t.Run("With Empty Object", func(t *testing.T) {
		var m map[string]interface{}

		err := Unmarshal([]byte(`{"a": {}, "b": 42}`), &m)
		assert.Nil(t, err)
		assert.Equal(t, m["a"], map[string]interface{}{})
		assert.Equal(t, m["b"], 42)
	})

	t.Run("With Empty Array Member", func(t *testing.T) {
		var m map[string]interface{}

		err := Unmarshal([]byte(`{"a": false, "b": [], "c": 17.9}`), &m)
		assert.Nil(t, err)
		assert.Equal(t, false, m["a"])
		assert.Equal(t, []interface{}{}, m["b"])
		assert.Equal(t, 17.9, m["c"])
	})

	t.Run("Empty Array, Extra Array Close", func(t *testing.T) {
		var m map[string]interface{}

		err := Unmarshal([]byte(`{"a": false, "b": [], "c": 17.9} ]`), &m)
		assert.Equal(t, `expected object key at position 32 in segment '{"a": false, "b": [], "c": 17.9} ]'`, err.Error())
	})

	t.Run("Extra Close", func(t *testing.T) {
		var m map[string]interface{}

		err := Unmarshal([]byte(`{"a": "b"}}`), &m)
		assert.Equal(t, `expected object key at position 10 in segment '{"a": "b"}}'`, err.Error())
	})

	t.Run("Valid JSON that Terminates Early", func(t *testing.T) {
		var m map[string]interface{}

		err := Unmarshal([]byte(`["a", {"b":4}, false]  ]`), &m)
		assert.Equal(t, `invalid character ']' at position '23' in segment '["a", {"b":4}, false]  ]'`, err.Error())
	})

	t.Run("Null Value into Interface", func(t *testing.T) {
		var m interface{}

		err := Unmarshal([]byte(`null`), &m)
		assert.Nil(t, err)
		assert.Equal(t, interface{}(nil), m)
	})

	t.Run("Array with Null Value into Interface", func(t *testing.T) {
		var m interface{}

		err := Unmarshal([]byte(`[true, null]`), &m)
		assert.Nil(t, err)
		assert.Equal(t, []interface{}{true, interface{}(nil)}, m)
	})

	t.Run("Array with Null Value into Interface Slice", func(t *testing.T) {
		var m []interface{}

		err := Unmarshal([]byte(`[true, null]`), &m)
		assert.Nil(t, err)
		assert.Equal(t, []interface{}{true, interface{}(nil)}, m)
	})

	t.Run("Array with Null Value into Interface Map", func(t *testing.T) {
		var m map[string]interface{}

		err := Unmarshal([]byte(`[true, null]`), &m)
		assert.Nil(t, err)
		assert.Equal(t, map[string]interface{}{"0": true, "1": interface{}(nil)}, m)
	})

	t.Run("Object with Null Value into Interface Struct", func(t *testing.T) {
		var m struct {
			IsNull interface{} `json:"is_null"`
		}

		err := Unmarshal([]byte(`{"is_null": null}`), &m)
		assert.Nil(t, err)
		assert.Equal(t, interface{}(nil), m.IsNull)
	})

	t.Run("Dealing With Unicode", func(t *testing.T) {
		testCases := []struct {
			Input    string
			Expected string
		}{
			{`"We\u2019ve been had"`, `We’ve been had`},
			{`"\u2018Hello there.\u2019"`, `‘Hello there.’`},
			{`"\u003cGeneral Kenobi\u003e"`, `<General Kenobi>`},
			{`"Shoots \u0026 Giggles"`, `Shoots & Giggles`},
			{`"Quoted String"`, `Quoted String`},
		}

		for i, tc := range testCases {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				var m string

				err := Unmarshal([]byte(tc.Input), &m)
				assert.Nil(t, err)
				assert.Equal(t, tc.Expected, m)
			})
		}
	})

	t.Run("Required Field Exists", func(t *testing.T) {
		type testType struct {
			IsRequired interface{} `json:"is_required,required"`
		}

		var m testType

		err := Unmarshal([]byte(`{"is_required": "Hello!"}`), &m)
		assert.Nil(t, err)
		assert.Equal(t, "Hello!", m.IsRequired)
	})

	t.Run("Required Field Does Not Exist", func(t *testing.T) {
		type testType struct {
			IsRequired interface{} `json:"is_required,required"`
		}

		var m testType

		err := Unmarshal([]byte(`{"not_required": "Hello!"}`), &m)
		assert.Equal(t, err, errors.New(`required key 'is_required' for struct 'testType' was not found`))
	})

	t.Run("NonEmpty Field Does Not Exist", func(t *testing.T) {
		type testType struct {
			NotEmpty interface{} `json:"not_empty,nonempty"`
		}

		var m testType

		err := Unmarshal([]byte(`{"is_not_right_key": "Hello!"}`), &m)
		assert.Equal(t, err, errors.New(`required key 'not_empty' for struct 'testType' was not found`))
	})

	t.Run("NonEmpty Field Exists But Is Empty String", func(t *testing.T) {
		type testType struct {
			IsEmpty string `json:"is_empty,nonempty"`
		}

		var m testType

		err := Unmarshal([]byte(`{"is_empty": ""}`), &m)
		assert.Equal(t, err, errors.New(`nonempty key 'is_empty' for struct 'testType' has string zero value`))
	})

	t.Run("NonEmpty Field Exists But Is Empty Int", func(t *testing.T) {
		type testType struct {
			IsEmpty int `json:"is_empty,nonempty"`
		}

		var m testType

		err := Unmarshal([]byte(`{"is_empty": 0}`), &m)
		assert.Equal(t, err, errors.New(`nonempty key 'is_empty' for struct 'testType' has int zero value`))
	})

	t.Run("NonEmpty Field Exists But Is Empty Float", func(t *testing.T) {
		type testType struct {
			IsEmpty float64 `json:"is_empty,nonempty"`
		}

		var m testType

		err := Unmarshal([]byte(`{"is_empty": 0}`), &m)
		assert.Equal(t, errors.New(`nonempty key 'is_empty' for struct 'testType' has int zero value`), err)

		err = Unmarshal([]byte(`{"is_empty": 0.0}`), &m)
		assert.Equal(t, errors.New(`nonempty key 'is_empty' for struct 'testType' has float zero value`), err)
	})

	t.Run("NonEmpty Field Exists But Is Empty Array", func(t *testing.T) {
		type testType struct {
			IsEmpty []string `json:"is_empty,nonempty"`
		}

		var m testType

		err := Unmarshal([]byte(`{"is_empty": []}`), &m)
		assert.Equal(t, errors.New(`nonempty key 'is_empty' for struct 'testType' has array zero value`), err)

		err = Unmarshal([]byte(`{"is_empty": [    ]}`), &m)
		assert.Equal(t, errors.New(`nonempty key 'is_empty' for struct 'testType' has array zero value`), err)
	})

	t.Run("NonEmpty Field Exists But Is Empty Array", func(t *testing.T) {
		type testType struct {
			IsEmpty struct{ A string } `json:"is_empty,nonempty"`
		}

		var m testType

		err := Unmarshal([]byte(`{"is_empty": {}}`), &m)
		assert.Equal(t, errors.New(`nonempty key 'is_empty' for struct 'testType' has object zero value`), err)

		err = Unmarshal([]byte(`{"is_empty": {    }}`), &m)
		assert.Equal(t, errors.New(`nonempty key 'is_empty' for struct 'testType' has object zero value`), err)
	})

	// See definition for ImplementsPostUnmarshalerValid type at the top of the file.
	t.Run("Test PostUnmarshalJSON", func(t *testing.T) {

		var tt ImplementsPostUnmarshalerValid

		err := Unmarshal([]byte(`{"thing": ["stuff", "junk"]}`), &tt)
		assert.Nil(t, err)
		assert.Equal(t, []string{"From PostUnmarshalJSON"}, tt.Thing)

	})

	// See definition for ImplementsPostUnmarshalerError type at the top of the file.
	t.Run("Test PostUnmarshalJSON Error from Panic", func(t *testing.T) {
		var tt ImplementsPostUnmarshalerErrorPanic

		err := Unmarshal([]byte(`{"thing": ["stuff", "junk"]}`), &tt)
		assert.Equal(t, "Error from PostUnmarshalJSON via Panic", err.Error())
		assert.Equal(t, []string{"stuff", "junk"}, tt.Thing)
	})

	t.Run("Test PostUnmarshalJSON Error from Return", func(t *testing.T) {
		var tt ImplementsPostUnmarshalerErrorReturn

		err := Unmarshal([]byte(`{"thing": ["stuff", "junk"]}`), &tt)
		assert.Equal(t, "Error from PostUnmarshalJSON via Return", err.Error())
		assert.Equal(t, []string{"stuff", "junk"}, tt.Thing)
	})

	t.Run("Test PostUnmarshalJSON Gets Error from Unmarshal", func(t *testing.T) {
		var tt PostUnmarshalerGetsUnmarshalError

		err := Unmarshal([]byte(`{"key": value}`), &tt)
		assert.Equal(t, `invalid character 'v' at position '8' in segment '{"key": value}' (expected object value)`, err.Error())
		assert.Equal(t, []string(nil), tt.Thing)
	})

	t.Run("Interface Resolution", func(t *testing.T) {
		type testType struct {
			Thing struct {
				A string `json:"stuff"`
				B bool   `json:"things"`
			} `json:"thing"`
		}

		testUnmarshalInterface := func(d []byte, v interface{}) ([]byte, error) {
			assert.Nil(t, Unmarshal(d, &v))
			return json.Marshal(v)
		}

		var tt testType
		d, err := testUnmarshalInterface([]byte(`{"thing": {"stuff": "hello", "junk": 123}}`), &tt)
		assert.Nil(t, err)
		assert.Equal(t, `{"thing":{"stuff":"hello","things":false}}`, string(d))
	})

	t.Run("Interface Resolution UnSetable", func(t *testing.T) {
		type testType struct {
			Thing struct {
				A string `json:"stuff"`
				B bool   `json:"things"`
			} `json:"thing"`
		}

		testUnmarshalInterface := func(d []byte, v interface{}) ([]byte, error) {
			assert.Nil(t, Unmarshal(d, &v))
			return json.Marshal(v)
		}

		var tt testType
		d, err := testUnmarshalInterface([]byte(`{"thing": {"stuff": "hello", "junk": 123}}`), tt)
		assert.Nil(t, err)
		assert.Equal(t, `{"thing":{"junk":123,"stuff":"hello"}}`, string(d))
	})
}

func TestUnmarshalStrict(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		var m *string

		err := UnmarshalStrict([]byte(`42`), &m)
		assert.Equal(t, errors.New("strict standards error, expected string, got int"), err)
	})

	t.Run("Int", func(t *testing.T) {
		var m int

		err := UnmarshalStrict([]byte(`42.0`), &m)
		assert.Equal(t, errors.New("strict standards error, expected int, got float"), err)
	})

	t.Run("Float", func(t *testing.T) {
		var m float64

		err := UnmarshalStrict([]byte(`42`), &m)
		assert.Equal(t, errors.New("strict standards error, expected float, got int"), err)
	})

	t.Run("Bool", func(t *testing.T) {
		var m bool

		err := UnmarshalStrict([]byte(`"42"`), &m)
		assert.Equal(t, errors.New("strict standards error, expected bool, got string"), err)
	})

	t.Run("Nested in Struct", func(t *testing.T) {
		var m struct{ A string }

		err := UnmarshalStrict([]byte(`{"a": 42}`), &m)
		assert.Equal(t, errors.New("strict standards error, expected string, got int"), err)
	})

	t.Run("Nested in Map", func(t *testing.T) {
		var m map[string]string

		err := UnmarshalStrict([]byte(`{"a": 42}`), &m)
		assert.Equal(t, errors.New("strict standards error, expected string, got int"), err)
	})

	t.Run("Array Into Struct", func(t *testing.T) {
		var m struct{ A string }

		err := UnmarshalStrict([]byte(`[42]`), &m)
		assert.Equal(t, errors.New("attempt to unmarshal JSON value with type 'array' into struct"), err)
	})

	t.Run("Array Into Map", func(t *testing.T) {
		var m map[string]int

		err := UnmarshalStrict([]byte(`[42]`), &m)
		assert.Equal(t, errors.New("strict standards: attempt to unmarshal JSON value with type 'array' into map"), err)
	})
}

func TestUnmarshalEscapedBackslash(t *testing.T) {
	data := `{"results":[{"keywords":"\\","canonical":"nbc-world_of_dance:srank_world_finale_front_row-hulu2"}]}`

	var m map[string]interface{}
	err := Unmarshal([]byte(data), &m)

	assert.Nil(t, err)
	assert.Equal(t, `nbc-world_of_dance:srank_world_finale_front_row-hulu2`, m["results"].([]interface{})[0].(map[string]interface{})["canonical"])
}

func TestUnmarshalEmbededStructs(t *testing.T) {
	var data = []byte(`{"field":"first_level", "embed_me": "Hi from level 1", "notembeded": "This shouldn't be embeded", "secondembed": {"field":"second_level", "embed_me": "Hi from level 2", "notembeded": "This shouldn't exist", "secondembed": {"client": "level_three", "embed_me": "Hi from level 3"}}}`)

	// For our expected data, notice that level 3 is dropped. This is because of the name conflict that arrises from the recursive embedding of Metadata.
	var expected = `{"field":"first_level","embed_me":"Hi from level 1","NotEmbeded":"This shouldn't be embeded","SecondEmbed":{"field":"second_level","embed_me":"Hi from level 2"}}`

	type Embed struct {
		IgnoreMe string `json:"-"`
		AndMe    string `json:"-"`
		EmbedMe  string `json:"embed_me"`
	}

	type Metadata struct {
		Type string `json:"type,omitempty"`
		Embed
		Field string `json:"field,omitempty"`
	}

	type TopLevel struct {
		NotEmbeded string
		Metadata
		SecondEmbed struct{ Metadata }
	}

	// Run against encoding/json to ensure we have no difference.
	var encodingJSON TopLevel
	json.Unmarshal(data, &encodingJSON)
	s, err := json.Marshal(encodingJSON)

	assert.Nil(t, err)
	assert.JSONEq(t, expected, string(s))

	var gojson TopLevel
	Unmarshal(data, &gojson)
	s, err = json.Marshal(gojson)

	assert.Nil(t, err)
	assert.JSONEq(t, expected, string(s))
}

func TestUnmarshalGoJSONTags(t *testing.T) {
	t.Run("Mixed Tags", func(t *testing.T) {
		type Example struct {
			GoJSON        string `gojson:"gojson"`
			Both          string `json:"both"`
			Neither       string `json:"-" gojson:"-"`
			OnlyUnmarshal string `json:"-" gojson:"only_unmarshal"`
		}

		var e Example
		data := []byte(`{
			"gojson": "Unmarshalled with gojson tag, marshaled with struct name",
			"both": "Will be unmarshalled AND marshalled",
			"neither": "Won't be unmarshalled or marshalled",
			"only_unmarshal": "This should only be unmarshalled"
		}`)

		err := Unmarshal(data, &e)
		assert.Nil(t, err)

		assert.Equal(t, e.GoJSON, "Unmarshalled with gojson tag, marshaled with struct name")
		assert.Equal(t, e.Both, "Will be unmarshalled AND marshalled")
		assert.Equal(t, e.Neither, "")
		assert.Equal(t, e.OnlyUnmarshal, "This should only be unmarshalled")

		m, err := json.Marshal(e)
		assert.Nil(t, err)
		assert.JSONEq(t, `{ "both": "Will be unmarshalled AND marshalled", "GoJSON": "Unmarshalled with gojson tag, marshaled with struct name" }`, string(m))
	})
}
