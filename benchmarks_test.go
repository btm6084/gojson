package gojson

import (
	"encoding/json"
	"strconv"
	"testing"
	"unsafe"
)

var benchData = `{"string":"some string","int":17,"bool":true,"float":22.83,"string_slice":["a","b","c","d","","\""],"bool_slice":[true,false,true,false],"int_slice":[1,2,3,4],"float_slice":[0.0,1.1,2.2,3.3],"object":{"a":"b","c":"d"},"objects":[{"e":"f","g":"h"},{"i":"j","k":"l"},{"m":"n","o":"p"}],"complex":["a", 2, null, false, 2.2, {"c":"d"}, ["s"]]}`
var largeJSONTestBlobBytes = []byte(largeJSONTestBlob)

func BenchmarkUnmarshalFloat(b *testing.B) {
	var m float64

	for i := 0; i < b.N; i++ {
		Unmarshal([]byte(`14.23e12`), &m)
	}
}

func BenchmarkUnmarshalFloatDefault(b *testing.B) {
	var m float64

	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(`14.23e12`), &m)
	}
}
func BenchmarkUnmarshalNull(b *testing.B) {
	var m *string

	for i := 0; i < b.N; i++ {
		Unmarshal([]byte(`null`), &m)
	}
}

func BenchmarkUnmarshalNullDefault(b *testing.B) {
	var m *string

	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(`null`), &m)
	}
}
func BenchmarkUnmarshalBool(b *testing.B) {
	var m bool

	for i := 0; i < b.N; i++ {
		Unmarshal([]byte(`true`), &m)
	}
}

func BenchmarkUnmarshalBoolDefault(b *testing.B) {
	var m bool

	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(`true`), &m)
	}
}

func BenchmarkUnmarshalInt(b *testing.B) {
	var m int

	for i := 0; i < b.N; i++ {
		Unmarshal([]byte(`4e012`), &m)
	}
}

func BenchmarkUnmarshalIntDefault(b *testing.B) {
	var m int

	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(`4e102`), &m)
	}
}

func BenchmarkUnmarshalString(b *testing.B) {
	var m string

	for i := 0; i < b.N; i++ {
		Unmarshal([]byte(`"\u2018Hello there.\u2019, \u003cGeneral Kenobi\u003e"`), &m)
	}
}

func BenchmarkUnmarshalStringDefault(b *testing.B) {
	var m string

	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(`"\u2018Hello there.\u2019, \u003cGeneral Kenobi\u003e"`), &m)
	}
}

func BenchmarkUnmarshalSlice(b *testing.B) {
	var m []string

	for i := 0; i < b.N; i++ {
		Unmarshal([]byte(tdStringSlice), &m)
	}
}

func BenchmarkUnmarshalSliceDefault(b *testing.B) {
	var m []string

	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(tdStringSlice), &m)
	}
}

func BenchmarkUnmarshalMap(b *testing.B) {
	var m map[string]interface{}

	for i := 0; i < b.N; i++ {
		Unmarshal([]byte(tdObject), &m)
	}
}

func BenchmarkUnmarshalMapDefault(b *testing.B) {
	var m map[string]interface{}

	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(tdObject), &m)
	}
}

func BenchmarkUnmarshalInterface(b *testing.B) {
	var m interface{}

	for i := 0; i < b.N; i++ {
		Unmarshal([]byte(largeJSONTestBlob), &m)
	}
}

func BenchmarkUnmarshalInterfaceDefault(b *testing.B) {
	var m interface{}

	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(largeJSONTestBlob), &m)
	}
}

func BenchmarkUnmarshalStruct(b *testing.B) {
	var m TestComponentResponse

	for i := 0; i < b.N; i++ {
		Unmarshal([]byte(largeJSONTestBlob), &m)
	}
}

func BenchmarkUnmarshalStructDefault(b *testing.B) {
	var m TestComponentResponse

	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(largeJSONTestBlob), &m)
	}
}

func BenchmarkExtract(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Extract(largeJSONTestBlobBytes, "items.18.data.assets.0.begins")
	}
}

func BenchmarkParse(b *testing.B) {

	for i := 0; i < b.N; i++ {
		NewJSONReader([]byte(largeJSONTestBlob))
	}
}

func BenchmarkGetInterface(b *testing.B) {
	r, _ := NewJSONReader([]byte(benchData))

	for i := 0; i < b.N; i++ {
		r.GetInterface("complex")
	}
}

func BenchmarkGetString(b *testing.B) {
	r, _ := NewJSONReader([]byte(benchData))

	for i := 0; i < b.N; i++ {
		r.GetString("string")
	}
}

func BenchmarkGetInt(b *testing.B) {
	r, _ := NewJSONReader([]byte(benchData))

	for i := 0; i < b.N; i++ {
		r.GetInt("int")
	}
}

func BenchmarkGetBool(b *testing.B) {
	r, _ := NewJSONReader([]byte(benchData))

	for i := 0; i < b.N; i++ {
		r.GetBool("bool")
	}
}

func BenchmarkGetFloat(b *testing.B) {
	r, _ := NewJSONReader([]byte(benchData))

	for i := 0; i < b.N; i++ {
		r.GetFloat("float")
	}
}

func BenchmarkGetInterfaceSlice(b *testing.B) {
	r, _ := NewJSONReader([]byte(benchData))

	for i := 0; i < b.N; i++ {
		r.GetInterfaceSlice("complex")
	}
}

func BenchmarkGetStringSlice(b *testing.B) {
	r, _ := NewJSONReader([]byte(benchData))

	for i := 0; i < b.N; i++ {
		r.GetStringSlice("string_slice")
	}
}

func BenchmarkGetBoolSlice(b *testing.B) {
	r, _ := NewJSONReader([]byte(benchData))

	for i := 0; i < b.N; i++ {
		r.GetBoolSlice("bool_slice")
	}
}

func BenchmarkGetIntSlice(b *testing.B) {
	r, _ := NewJSONReader([]byte(benchData))

	for i := 0; i < b.N; i++ {
		r.GetIntSlice("int_slice")
	}
}

func BenchmarkGetFloatSlice(b *testing.B) {
	r, _ := NewJSONReader([]byte(benchData))

	for i := 0; i < b.N; i++ {
		r.GetFloatSlice("float_slice")
	}
}

func BenchmarkGetMapStringInterface(b *testing.B) {
	r, _ := NewJSONReader([]byte(benchData))

	for i := 0; i < b.N; i++ {
		r.GetMapStringInterface("object")
	}
}

func BenchmarkGet(b *testing.B) {
	r, _ := NewJSONReader([]byte(benchData))

	for i := 0; i < b.N; i++ {
		r.Get("object")
	}
}

func BenchmarkGetCollection(b *testing.B) {
	r, _ := NewJSONReader([]byte(benchData))

	for i := 0; i < b.N; i++ {
		r.GetCollection("objects")
	}
}

func BenchmarkManualUnquote(b *testing.B) {
	s := []byte(`"Shoots \u0026 Giggles \u003c\tGeneral\nKenobi\r\"\u003e\""`)

	for i := 0; i < b.N; i++ {
		manualUnescapeString(s)
	}
}

func BenchmarkUnquoteDefault(b *testing.B) {
	s := []byte(`"Shoots \u0026 Giggles \u003c\tGeneral\nKenobi\r\"\u003e\""`)

	for i := 0; i < b.N; i++ {
		strconv.Unquote(*(*string)(unsafe.Pointer(&s)))
	}
}

func BenchmarkMassivelyQuotesString(b *testing.B) {
	s := []byte(massiveQuotedString)
	var out string

	for i := 0; i < b.N; i++ {
		Unmarshal(s, &out)
	}
}

func BenchmarkMassivelyQuotesStringDefault(b *testing.B) {
	s := []byte(massiveQuotedString)
	var out string

	for i := 0; i < b.N; i++ {
		json.Unmarshal(s, &out)
	}
}
