package gojson

import (
	"encoding/json"
	"fmt"
	"testing"
)

// "\u2018Hello there.\u2019, \u003cGeneral Kenobi\u003e Emoji!! \ud83d\udc4f \uD83D\uDC4C \ud83d\uDC7B"

func BenchmarkNewUnmarshalString(b *testing.B) {
	var m *int

	for i := 0; i < b.N; i++ {
		UnmarshalJSON([]byte(`-124e7`), &m)
	}
}
func BenchmarkNewUnmarshalStringOld(b *testing.B) {
	var m *int

	for i := 0; i < b.N; i++ {
		Unmarshal([]byte(`-124e7`), &m)
	}
}

func BenchmarkNewUnmarshalStringDefault(b *testing.B) {
	var m *int

	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(`-124e7`), &m)
	}
}

func TestNewTests1(t *testing.T) {
	// fmt.Println(string(findInt([]byte(`"124.762`))))
	// fmt.Println(string(findInt([]byte(`"124e762e2`))))
	// fmt.Println(string(findInt([]byte(`"124e762e2`))))
	// fmt.Println(string(findInt([]byte(`"-124.762`))))
	// fmt.Println(string(findInt([]byte(`"-124e-762e-2`))))
	// fmt.Println(string(findInt([]byte(`"-124e-762e2`))))

	var m int
	err := UnmarshalJSON([]byte(`"-124"`), &m)
	fmt.Println(err)
	fmt.Println(m)

	// var a, b int
	// Unmarshal([]byte(`-124e762`), &a)
	// UnmarshalJSON([]byte(`-124e762`), &b)
	// require.Equal(t, a, b)
}

// func BenchmarkNewTests3(b *testing.B) {
// 	// var m string

// 	for i := 0; i < b.N; i++ {
// 		manualUnescapeString([]byte(`"\u2018Hello there.\u2019, \u003cGeneral Kenobi\u003e Emoji!! \ud83d\udc4f \uD83D\uDC4C \ud83d\uDC7B"`))
// 	}
// }
// func BenchmarkNewTests4(b *testing.B) {
// 	// var m string

// 	for i := 0; i < b.N; i++ {
// 		toJSONString([]byte(`"\u2018Hello there.\u2019, \u003cGeneral Kenobi\u003e Emoji!! \ud83d\udc4f \uD83D\uDC4C \ud83d\uDC7B"`))
// 	}
// }
