package gojson

import (
	"encoding/json"
	"testing"
)

var benchData = `{"string":"some string","int":17,"bool":true,"float":22.83,"string_slice":["a","b","c","d","","\""],"bool_slice":[true,false,true,false],"int_slice":[1,2,3,4],"float_slice":[0.0,1.1,2.2,3.3],"object":{"a":"b","c":"d"},"objects":[{"e":"f","g":"h"},{"i":"j","k":"l"},{"m":"n","o":"p"}],"complex":["a", 2, null, false, 2.2, {"c":"d"}, ["s"]]}`
var largeJSONTestBlobBytes = []byte(largeJSONTestBlob)

func BenchmarkUnmarshalString(b *testing.B) {
	var m string

	for i := 0; i < b.N; i++ {
		Unmarshal([]byte(`"Hello"`), &m)
	}
}

func BenchmarkUnmarshalStringDefault(b *testing.B) {
	var m string

	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(`"Hello"`), &m)
	}
}

func BenchmarkUnmarshalSlice(b *testing.B) {
	var m []string

	for i := 0; i < b.N; i++ {
		Unmarshal([]byte(readerTestData), &m)
	}
}

func BenchmarkUnmarshalSliceDefault(b *testing.B) {
	var m []string

	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(readerTestData), &m)
	}
}

func BenchmarkUnmarshalMap(b *testing.B) {
	var m map[string]interface{}

	for i := 0; i < b.N; i++ {
		Unmarshal([]byte(largeJSONTestBlob), &m)
	}
}

func BenchmarkUnmarshalMapDefault(b *testing.B) {
	var m map[string]interface{}

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
