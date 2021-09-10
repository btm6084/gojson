package gojson

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewUnmarshalInterface(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		var a, b, c interface{}
		value := []byte(`12345`)
		expected := 12345

		err := UnmarshalJSON(value, &a)
		require.Nil(t, err)
		require.Equal(t, expected, a)

		err = Unmarshal(value, &b)
		require.Nil(t, err)
		require.Equal(t, expected, b)

		err = json.Unmarshal(value, &c)
		require.Nil(t, err)
		require.Equal(t, float64(expected), c)

		require.Equal(t, a, b)
	})

	t.Run("float", func(t *testing.T) {
		var a, b, c interface{}
		value := []byte(`12.345`)
		expected := 12.345

		err := UnmarshalJSON(value, &a)
		require.Nil(t, err)
		require.Equal(t, expected, a)

		err = Unmarshal(value, &b)
		require.Nil(t, err)
		require.Equal(t, expected, b)

		err = json.Unmarshal(value, &c)
		require.Nil(t, err)
		require.Equal(t, float64(expected), c)

		require.Equal(t, a, b)
	})

	t.Run("string", func(t *testing.T) {
		var a, b, c interface{}
		value := []byte(`"Emoji!! \ud83d\udc4f \uD83D\uDC4C \ud83d\uDC7B"`)
		expected := `Emoji!! 👏 👌 👻`

		err := UnmarshalJSON(value, &a)
		require.Nil(t, err)
		require.Equal(t, expected, a)

		err = Unmarshal(value, &b)
		require.Nil(t, err)
		require.Equal(t, expected, b)

		err = json.Unmarshal(value, &c)
		require.Nil(t, err)
		require.Equal(t, expected, c)

		require.Equal(t, a, b)
	})

	t.Run("true", func(t *testing.T) {
		var a, b, c interface{}
		value := []byte(`true`)
		expected := true

		err := UnmarshalJSON(value, &a)
		require.Nil(t, err)
		require.Equal(t, expected, a)

		err = Unmarshal(value, &b)
		require.Nil(t, err)
		require.Equal(t, expected, b)

		err = json.Unmarshal(value, &c)
		require.Nil(t, err)
		require.Equal(t, expected, c)

		require.Equal(t, a, b)
	})

	t.Run("false", func(t *testing.T) {
		var a, b, c interface{}
		value := []byte(`false`)
		expected := false

		err := UnmarshalJSON(value, &a)
		require.Nil(t, err)
		require.Equal(t, expected, a)

		err = Unmarshal(value, &b)
		require.Nil(t, err)
		require.Equal(t, expected, b)

		err = json.Unmarshal(value, &c)
		require.Nil(t, err)
		require.Equal(t, expected, c)

		require.Equal(t, a, b)
	})

	t.Run("null", func(t *testing.T) {
		var a, b, c interface{}
		value := []byte(`null`)

		err := UnmarshalJSON(value, &a)
		require.Nil(t, err)
		require.Nil(t, a)

		err = Unmarshal(value, &b)
		require.Nil(t, err)
		require.Nil(t, b)

		err = json.Unmarshal(value, &c)
		require.Nil(t, err)
		require.Nil(t, c)

		require.Equal(t, a, b)
	})
}

func BenchmarkNewUnmarshalInterface(b *testing.B) {
	b.Run("Interface String", func(b *testing.B) {
		value := []byte(`"This is a string"`)

		b.Run("Default", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := json.Unmarshal(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})

		b.Run("Old", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := Unmarshal(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})

		b.Run("New", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := UnmarshalJSON(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})
	})

	b.Run("Interface Int", func(b *testing.B) {
		value := []byte(`12345`)

		b.Run("Default", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := json.Unmarshal(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})

		b.Run("Old", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := Unmarshal(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})

		b.Run("New", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := UnmarshalJSON(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})
	})

	b.Run("Interface Float", func(b *testing.B) {
		value := []byte(`12.345`)

		b.Run("Default", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := json.Unmarshal(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})

		b.Run("Old", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := Unmarshal(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})

		b.Run("New", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := UnmarshalJSON(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})
	})

	b.Run("Interface False", func(b *testing.B) {
		value := []byte(`false`)

		b.Run("Default", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := json.Unmarshal(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})

		b.Run("Old", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := Unmarshal(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})

		b.Run("New", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := UnmarshalJSON(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})
	})

	b.Run("Interface True", func(b *testing.B) {
		value := []byte(`true`)

		b.Run("Default", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := json.Unmarshal(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})

		b.Run("Old", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := Unmarshal(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})

		b.Run("New", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := UnmarshalJSON(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})
	})

	b.Run("Interface Null", func(b *testing.B) {
		value := []byte(`null`)

		b.Run("Default", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := json.Unmarshal(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})

		b.Run("Old", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := Unmarshal(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})

		b.Run("New", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := UnmarshalJSON(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})
	})

	b.Run("Interface Zero", func(b *testing.B) {
		value := []byte(`0`)

		b.Run("Default", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := json.Unmarshal(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})

		b.Run("Old", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := Unmarshal(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})

		b.Run("New", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var m interface{}

				err := UnmarshalJSON(value, &m)
				if err != nil {
					log.Fatal(err)
				}
			}
		})
	})
}

func TestNewUnmarshalString(t *testing.T) {
	t.Run("Massive", func(t *testing.T) {
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
	})

	t.Run("Small", func(t *testing.T) {
		value := []byte(`"\u2018Hello there.\u2019, \u003cGeneral Kenobi\u003e"`)
		expected := `‘Hello there.’, <General Kenobi>`
		var a, b, c *string

		err := UnmarshalJSON([]byte(value), &a)
		require.Nil(t, err)
		require.Equal(t, expected, *a, "a")

		err = Unmarshal([]byte(value), &b)
		require.Nil(t, err)
		require.Equal(t, expected, *b, "b")

		err = json.Unmarshal([]byte(value), &c)
		require.Nil(t, err)
		require.Equal(t, expected, *c, "b")

		require.Equal(t, *a, *b)
		require.Equal(t, *a, *c)
		require.Equal(t, *b, *c)
	})
}

func BenchmarkNewUnmarshalString(b *testing.B) {
	value := []byte(massiveQuotedString)
	smallValue := []byte(`"\u2018Hello there.\u2019, \u003cGeneral Kenobi\u003e"`)
	tinyValue := []byte(`"Simple String"`)

	b.Run("DefaultMassive", func(b *testing.B) {
		var m *string

		for i := 0; i < b.N; i++ {
			err := json.Unmarshal(value, &m)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	b.Run("OldMassive", func(b *testing.B) {
		var m *string

		for i := 0; i < b.N; i++ {
			err := Unmarshal(value, &m)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	b.Run("NewMassive", func(b *testing.B) {
		var m *string

		for i := 0; i < b.N; i++ {
			err := UnmarshalJSON(value, &m)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	b.Run("DefaultSmall", func(b *testing.B) {
		var m *string

		for i := 0; i < b.N; i++ {
			err := json.Unmarshal(smallValue, &m)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	b.Run("OldSmall", func(b *testing.B) {
		var m *string

		for i := 0; i < b.N; i++ {
			err := Unmarshal(smallValue, &m)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	b.Run("NewSmall", func(b *testing.B) {
		var m *string

		for i := 0; i < b.N; i++ {
			err := UnmarshalJSON(smallValue, &m)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	b.Run("DefaultTiny", func(b *testing.B) {
		var m *string

		for i := 0; i < b.N; i++ {
			err := json.Unmarshal(tinyValue, &m)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	b.Run("OldTiny", func(b *testing.B) {
		var m *string

		for i := 0; i < b.N; i++ {
			err := Unmarshal(tinyValue, &m)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	b.Run("NewTiny", func(b *testing.B) {
		var m *string

		for i := 0; i < b.N; i++ {
			err := UnmarshalJSON(tinyValue, &m)
			if err != nil {
				log.Fatal(err)
			}
		}
	})
}

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
	t.Run("intptr", func(t *testing.T) {
		var a, b *int
		var c *float64

		err := UnmarshalJSON(value, &a)
		require.Nil(t, err)

		err = Unmarshal(value, &b)
		require.Nil(t, err)

		err = json.Unmarshal(value, &c)
		require.Nil(t, err)

		require.Equal(t, *a, *b)
		require.Equal(t, *a, int(*c))
		require.Equal(t, *b, int(*c))
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
		var a, b *float64
		var c *float64

		err := UnmarshalJSON(value, &a)
		require.Nil(t, err)

		err = Unmarshal(value, &b)
		require.Nil(t, err)

		err = json.Unmarshal(value, &c)
		require.Nil(t, err)

		require.Equal(t, *a, *b)
		require.Equal(t, *a, float64(*c))
		require.Equal(t, *b, float64(*c))
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
	value := []byte(`-1.24`)

	b.Run("Default", func(b *testing.B) {
		var m *float64

		for i := 0; i < b.N; i++ {
			err := json.Unmarshal(value, &m)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	b.Run("Old", func(b *testing.B) {
		var m *float64

		for i := 0; i < b.N; i++ {
			err := Unmarshal(value, &m)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	b.Run("New", func(b *testing.B) {
		var m *float64

		for i := 0; i < b.N; i++ {
			err := UnmarshalJSON(value, &m)
			if err != nil {
				log.Fatal(err)
			}
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

func TestNewUnmarshalMap(t *testing.T) {
	t.Run("ints", func(t *testing.T) {
		var a, b []int
		// value := []byte(`    [123, 234, 345,        456, 567, 678      , 789, 890, 901, 1012]   `)
		value := []byte(`["123", "234", "345", "456", "567", "678", "789", "890", "901", "1012"]`)

		err := UnmarshalJSON(value, &a)
		require.Nil(t, err)

		// err = Unmarshal(value, &b)
		// require.Nil(t, err)

		// err = Unmarshal(value, &b)
		// require.Nil(t, err)

		// err = json.Unmarshal(value, &c)
		// require.Nil(t, err)

		require.Equal(t, a, b, "a, b")
		// require.Equal(t, a, c, "a, c")
		// require.Equal(t, b, c, "b, c")
	})
}

func BenchmarkNewUnmarshalMap(b *testing.B) {
	ints := []byte(`["123", "234", "345", "456", "567", "678", "789", "890", "901", "1012"]`)

	b.Run("Old", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var m []int
			Unmarshal(ints, &m)
		}
	})
}
