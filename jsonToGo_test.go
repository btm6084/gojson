package gojson

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewUnmarshalInt(t *testing.T) {
	value := []byte(`-142`)
	uValue := []byte(`142`)

	t.Run("int", func(t *testing.T) {
		var a, b int
		var c float64

		err := UnmarshalJSON(value, &a)
		require.Nil(t, err)

		err = Unmarshal(value, &b)
		require.Nil(t, err)

		err = json.Unmarshal(value, &c)
		require.Nil(t, err)

		require.Equal(t, a, b)
		require.Equal(t, a, int(c))
		require.Equal(t, b, int(c))
	})
	t.Run("int32", func(t *testing.T) {
		var a, b int32
		var c float64

		err := UnmarshalJSON(value, &a)
		require.Nil(t, err)

		err = Unmarshal(value, &b)
		require.Nil(t, err)

		err = json.Unmarshal(value, &c)
		require.Nil(t, err)

		require.Equal(t, a, b)
		require.Equal(t, a, int32(c))
		require.Equal(t, b, int32(c))
	})
	t.Run("int64", func(t *testing.T) {
		var a, b int64
		var c float64

		err := UnmarshalJSON(value, &a)
		require.Nil(t, err)

		err = Unmarshal(value, &b)
		require.Nil(t, err)

		err = json.Unmarshal(value, &c)
		require.Nil(t, err)

		require.Equal(t, a, b)
		require.Equal(t, a, int64(c))
		require.Equal(t, b, int64(c))
	})
	t.Run("int16", func(t *testing.T) {
		var a, b int16
		var c float64

		err := UnmarshalJSON(value, &a)
		require.Nil(t, err)

		err = Unmarshal(value, &b)
		require.Nil(t, err)

		err = json.Unmarshal(value, &c)
		require.Nil(t, err)

		require.Equal(t, a, b)
		require.Equal(t, a, int16(c))
		require.Equal(t, b, int16(c))
	})
	t.Run("int8", func(t *testing.T) {
		var a, b int8
		var c float64

		err := UnmarshalJSON(value, &a)
		require.Nil(t, err)

		err = Unmarshal(value, &b)
		require.Nil(t, err)

		err = json.Unmarshal(value, &c)
		require.Nil(t, err)

		require.Equal(t, a, b)
		require.Equal(t, a, int8(c))
		require.Equal(t, b, int8(c))
	})

	t.Run("uint", func(t *testing.T) {
		var a, b uint
		var c float64

		err := UnmarshalJSON(uValue, &a)
		require.Nil(t, err)

		err = Unmarshal(uValue, &b)
		require.Nil(t, err)

		err = json.Unmarshal(uValue, &c)
		require.Nil(t, err)

		require.Equal(t, a, b)
		require.Equal(t, a, uint(c))
		require.Equal(t, b, uint(c))
	})
	t.Run("uint64", func(t *testing.T) {
		var a, b uint64
		var c float64

		err := UnmarshalJSON(uValue, &a)
		require.Nil(t, err)

		err = Unmarshal(uValue, &b)
		require.Nil(t, err)

		err = json.Unmarshal(uValue, &c)
		require.Nil(t, err)

		require.Equal(t, a, b)
		require.Equal(t, a, uint64(c))
		require.Equal(t, b, uint64(c))
	})
	t.Run("uint32", func(t *testing.T) {
		var a, b uint32
		var c float64

		err := UnmarshalJSON(uValue, &a)
		require.Nil(t, err)

		err = Unmarshal(uValue, &b)
		require.Nil(t, err)

		err = json.Unmarshal(uValue, &c)
		require.Nil(t, err)

		require.Equal(t, a, b)
		require.Equal(t, a, uint32(c))
		require.Equal(t, b, uint32(c))
	})
	t.Run("uint16", func(t *testing.T) {
		var a, b uint16
		var c float64

		err := UnmarshalJSON(uValue, &a)
		require.Nil(t, err)

		err = Unmarshal(uValue, &b)
		require.Nil(t, err)

		err = json.Unmarshal(uValue, &c)
		require.Nil(t, err)

		require.Equal(t, a, b)
		require.Equal(t, a, uint16(c))
		require.Equal(t, b, uint16(c))
	})
	t.Run("uint8", func(t *testing.T) {
		var a, b uint8
		var c float64

		err := UnmarshalJSON(uValue, &a)
		require.Nil(t, err)

		err = Unmarshal(uValue, &b)
		require.Nil(t, err)

		err = json.Unmarshal(uValue, &c)
		require.Nil(t, err)

		require.Equal(t, a, b)
		require.Equal(t, a, uint8(c))
		require.Equal(t, b, uint8(c))
	})
}

func BenchmarkNewUnmarshalInt(b *testing.B) {
	value := []byte(`-124e7`)

	b.Run("Default", func(b *testing.B) {
		var m *int

		for i := 0; i < b.N; i++ {
			json.Unmarshal(value, &m)
		}
	})

	b.Run("Old", func(b *testing.B) {
		var m *int

		for i := 0; i < b.N; i++ {
			Unmarshal(value, &m)
		}
	})

	b.Run("New", func(b *testing.B) {
		var m *int

		for i := 0; i < b.N; i++ {
			UnmarshalJSON(value, &m)
		}
	})
}

func TestNewUnmarshalFloat(t *testing.T) {
	value := []byte(`-2311.2423123`)
	t.Run("float64", func(t *testing.T) {
		var a, b float64
		var c float64

		err := UnmarshalJSON(value, &a)
		require.Nil(t, err)

		err = Unmarshal(value, &b)
		require.Nil(t, err)

		err = json.Unmarshal(value, &c)
		require.Nil(t, err)

		require.Equal(t, a, b)
		require.Equal(t, a, float64(c))
		require.Equal(t, b, float64(c))
	})
	t.Run("float32", func(t *testing.T) {
		var a, b float32
		var c float64

		err := UnmarshalJSON(value, &a)
		require.Nil(t, err)

		err = Unmarshal(value, &b)
		require.Nil(t, err)

		err = json.Unmarshal(value, &c)
		require.Nil(t, err)

		require.Equal(t, a, b)
		require.Equal(t, a, float32(c))
		require.Equal(t, b, float32(c))
	})
}

func BenchmarkNewUnmarshalFloat(b *testing.B) {
	value := []byte(`-124e7`)

	b.Run("Default", func(b *testing.B) {
		var m *int

		for i := 0; i < b.N; i++ {
			json.Unmarshal(value, &m)
		}
	})

	b.Run("Old", func(b *testing.B) {
		var m *int

		for i := 0; i < b.N; i++ {
			Unmarshal(value, &m)
		}
	})

	b.Run("New", func(b *testing.B) {
		var m *int

		for i := 0; i < b.N; i++ {
			UnmarshalJSON(value, &m)
		}
	})
}

func TestNewUnmarshalString(t *testing.T) {
	var a, b, c string

	err := UnmarshalJSON([]byte(massiveQuotedString), &a)
	require.Nil(t, err)

	err = Unmarshal([]byte(massiveQuotedString), &b)
	require.Nil(t, err)

	err = json.Unmarshal([]byte(massiveQuotedString), &c)
	require.Nil(t, err)

	require.Equal(t, a, b)
	require.Equal(t, a, c)
	require.Equal(t, b, c)
}

func BenchmarkNewUnmarshalString(b *testing.B) {
	value := []byte(massiveQuotedString)

	b.Run("Default", func(b *testing.B) {
		var m *string

		for i := 0; i < b.N; i++ {
			json.Unmarshal(value, &m)
		}
	})

	b.Run("Old", func(b *testing.B) {
		var m *string

		for i := 0; i < b.N; i++ {
			Unmarshal(value, &m)
		}
	})

	b.Run("New", func(b *testing.B) {
		var m *string

		for i := 0; i < b.N; i++ {
			UnmarshalJSON(value, &m)
		}
	})
}

func TestNewUnmarshalBool(t *testing.T) {
	var a, b, c bool
	value := []byte(`"true"`)
	defaultValue := []byte(`true`)

	err := UnmarshalJSON(value, &a)
	require.Nil(t, err)

	err = Unmarshal(value, &b)
	require.Nil(t, err)

	err = json.Unmarshal(defaultValue, &c)
	require.Nil(t, err)

	require.True(t, a)
	require.True(t, b)
	require.True(t, c)

	require.Equal(t, a, b)
	require.Equal(t, a, c)
	require.Equal(t, b, c)
}

func BenchmarkNewUnmarshalBool(b *testing.B) {
	value := []byte(`"TrUe"`)

	b.Run("Default", func(b *testing.B) {
		var m *bool

		for i := 0; i < b.N; i++ {
			json.Unmarshal(value, &m)
		}
	})

	b.Run("Old", func(b *testing.B) {
		var m *bool

		for i := 0; i < b.N; i++ {
			Unmarshal(value, &m)
		}
	})

	b.Run("New", func(b *testing.B) {
		var m *bool

		for i := 0; i < b.N; i++ {
			UnmarshalJSON(value, &m)
		}
	})
}
