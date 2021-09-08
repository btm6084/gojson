package gojson

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkNewUnmarshalString(b *testing.B) {
	var m *string

	for i := 0; i < b.N; i++ {
		UnmarshalJSON([]byte(`"\u2018Hello there.\u2019, \u003cGeneral Kenobi\u003e Emoji!! \ud83d\udc4f \uD83D\uDC4C \ud83d\uDC7B"`), &m)
	}
}
func BenchmarkNewUnmarshalStringOld(b *testing.B) {
	var m *string

	for i := 0; i < b.N; i++ {
		Unmarshal([]byte(`"\u2018Hello there.\u2019, \u003cGeneral Kenobi\u003e Emoji!! \ud83d\udc4f \uD83D\uDC4C \ud83d\uDC7B"`), &m)
	}
}

func BenchmarkNewUnmarshalStringDefault(b *testing.B) {
	var m *string

	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(`"\u2018Hello there.\u2019, \u003cGeneral Kenobi\u003e Emoji!! \ud83d\udc4f \uD83D\uDC4C \ud83d\uDC7B"`), &m)
	}
}

func TestNewTests1(t *testing.T) {
	var a, b string
	Unmarshal([]byte(`"\u2018Hello there.\u2019, \u003cGeneral Kenobi\u003e Emoji!! \ud83d\udc4f \uD83D\uDC4C \ud83d\uDC7B"`), &a)
	UnmarshalJSON([]byte(`"\u2018Hello there.\u2019, \u003cGeneral Kenobi\u003e Emoji!! \ud83d\udc4f \uD83D\uDC4C \ud83d\uDC7B"`), &b)
	require.Equal(t, a, b)
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
