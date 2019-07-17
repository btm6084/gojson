package gojson

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONIsJSON(t *testing.T) {
	var empty string
	testCases := []struct {
		label    string
		input    string
		expected bool
	}{
		{"Empty", empty, false},
		{"Null", ` null    `, true},
		{"True", ` true `, true},
		{"False", ` false`, true},
		{"Number", `12357.42 `, true},
		{"Exponent", ` 5.5e-09 `, true},
		{"Negative", ` -5.5e-09 `, true},
		{"String", `"string \""`, true},
		{"Array", `["member 1", "member 2", [["Hi"]]]`, true},
		{"Array Invalid", `["a""b"]`, false},
		{"Object", `{"key": "value"}`, true},
		{"Object Invalid", `{"key": "value"`, false},
		{"Object Invalid 2", `{"key": "value" "key2": "value2"}`, false},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsJSON([]byte(tc.input)))
		})
	}
}

func TestJSONIsNull(t *testing.T) {
	var empty string
	testCases := []struct {
		label    string
		input    string
		expected bool
	}{
		{"Empty", empty, false},
		{"Null", ` null `, true},
		{"Case", `NuLl`, true},
		{"True", `true`, false},
		{"String", `"null"`, false},
		{"Number", `1`, false},
		{"False", `false`, false},
		{"Whitespace", ` n u l l `, false},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsJSONNull([]byte(tc.input)))
		})
	}
}

func TestJSONIsTrue(t *testing.T) {
	var empty string
	testCases := []struct {
		label    string
		input    string
		expected bool
	}{
		{"Empty", empty, false},
		{"True", `true`, true},
		{"Trues", `trues`, false},
		{"Case", `TrUe`, true},
		{"False", `false`, false},
		{"String", `"true"`, false},
		{"Number", `1`, false},
		{"Null", `null`, false},
		{"F", `null`, false},
		{"Whitespace", ` f a l s e `, false},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsJSONTrue([]byte(tc.input)))
		})
	}
}

func TestJSONIsFalse(t *testing.T) {
	var empty string
	testCases := []struct {
		label    string
		input    string
		expected bool
	}{
		{"Empty", empty, false},
		{"False", `false`, true},
		{"Case", `FaLsE`, true},
		{"True", `true`, false},
		{"String", `"true"`, false},
		{"Number", `1`, false},
		{"Null", `null`, false},
		{"T", `null`, false},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsJSONFalse([]byte(tc.input)))
		})
	}
}

func TestIsJSONNumber(t *testing.T) {
	t.Run("Make sure strings aren't changed", func(t *testing.T) {
		var b = []byte("TEST")
		assert.False(t, IsJSONNumber(b))
		assert.Equal(t, string("TEST"), string(b))
	})

	t.Run("Empty String", func(t *testing.T) {
		assert.False(t, IsJSONNumber([]byte{}))
	})
}

func TestJSONIsPositiveNumber(t *testing.T) {
	testCases := []struct {
		label    string
		input    []byte
		expected bool
	}{
		{"Empty", []byte(""), false},
		{"NoDigits e", []byte("e"), false},
		{"NoDigits E", []byte("E"), false},
		{"Zero e", []byte("e0"), false},
		{"Zero E", []byte("E0"), false},
		{"Zeros", []byte("00"), false},
		{"Double Period", []byte("0.0.0"), false},
		{"Double Exponent", []byte("0.2e+10E17"), false},
		{"No Digits After Period", []byte("0.e+10E17"), false},
		{"Double Zeros", []byte("00"), false},
		{"Decimal Must be Followed by Number", []byte("."), false},
		{"Decimal Must be Followed by Number", []byte("1."), false},
		{"Multiple e Symbols", []byte("1e.2e4"), false},
		{"Multiple E Symbols", []byte("1E.2E4"), false},
		{"Multiple Periods", []byte("1.0e2.4"), false},
		{"Not A Number", []byte("notanumber"), false},

		// DecimalNumber
		{"Zero", []byte("0"), true},
		{"Zeros", []byte("00"), false},
		{"One", []byte("1"), true},
		{"Ones", []byte("11"), true},
		{"Two", []byte("2"), true},
		{"Twos", []byte("22"), true},
		{"Three", []byte("3"), true},
		{"Threes", []byte("33"), true},
		{"Four", []byte("4"), true},
		{"Fours", []byte("44"), true},
		{"Five", []byte("5"), true},
		{"Fives", []byte("55"), true},
		{"Six", []byte("6"), true},
		{"Sixs", []byte("66"), true},
		{"Seven", []byte("7"), true},
		{"Sevens", []byte("77"), true},
		{"Eight", []byte("8"), true},
		{"Eights", []byte("88"), true},
		{"Nine", []byte("9"), true},
		{"Nines", []byte("99"), true},
		{"Plus", []byte("+11"), false},
		{"Minus", []byte("-11"), false},
		{"Numeric", []byte("230847"), true},
		{"Numeric Leading 0", []byte("083274"), false},

		// DecimalNumber ExponentPart
		{"Zero", []byte("0e17"), true},
		{"Zeros", []byte("00e17"), false},
		{"One", []byte("1E14"), true},
		{"Ones", []byte("11E14"), true},
		{"Two", []byte("2e+223"), true},
		{"Twos", []byte("22e+223"), true},
		{"Three", []byte("3E-2938"), true},
		{"Threes", []byte("33E-2938"), true},
		{"Four", []byte("4e0923"), true},
		{"Fours", []byte("44e0923"), true},
		{"Five", []byte("5E-029382"), true},
		{"Fives", []byte("55E-029382"), true},
		{"Six", []byte("6e1723"), true},
		{"Sixs", []byte("66e1723"), true},
		{"Seven", []byte("7E01234"), true},
		{"Sevens", []byte("77E01234"), true},
		{"Eight", []byte("8e-213213"), true},
		{"Eights", []byte("88e-213213"), true},
		{"Nine", []byte("9E+1927364"), true},
		{"Nines", []byte("99E+1927364"), true},
		{"Numeric", []byte("230847e0"), true},
		{"Plus", []byte("+11E11"), false},
		{"Minus", []byte("-11e42"), false},
		{"Numeric Leading 0", []byte("083274e12"), false},
		{"Missing E", []byte("99+1927364"), false},
		{"Period in Exponent", []byte("77E0.1234"), false},

		// DecimalNumber . Digits
		{"Float Zeros", []byte("0.00"), true},
		{"Double Float Zeros", []byte("00.00"), false},
		{"Float Zero", []byte("0.0"), true},
		{"Float One", []byte("1.1"), true},
		{"Float Two", []byte("2.2"), true},
		{"Float Three", []byte("3.3"), true},
		{"Float Four", []byte("4.4"), true},
		{"Float Five", []byte("5.5"), true},
		{"Float Six", []byte("6.6"), true},
		{"Float Seven", []byte("7.7"), true},
		{"Float Eight", []byte("8.8"), true},
		{"Float Nine", []byte("9.9"), true},
		{"Float Plus", []byte("+1.1"), false},
		{"Float Minus", []byte("-1.1"), false},
		{"Float Numeric", []byte("230.847"), true},
		{"Float Numeric Leading 0", []byte("0.274"), true},
		{"Float Numeric Leading Period", []byte(".274"), false},
		{"Float Numeric Leading 0 Period", []byte("0.274"), true},
		{"Float Numeric Leading Digits", []byte("083.274"), false},

		// DecimalNumber . Digits ExponentPart
		{"Exponent Float Zeros", []byte("0.00e1"), true},
		{"Exponent Float Zero", []byte("0.0E2"), true},
		{"Exponent Float One", []byte("1.1e+3"), true},
		{"Exponent Float Two", []byte("2.2E-4"), true},
		{"Exponent Float Three", []byte("3.3e07"), true},
		{"Exponent Float Four", []byte("4.4E12"), true},
		{"Exponent Float Five", []byte("5.5e-09"), true},
		{"Exponent Float Six", []byte("6.6E+017"), true},
		{"Exponent Float Seven", []byte("7.7e-09"), true},
		{"Exponent Float Eight", []byte("8.8E+017"), true},
		{"Exponent Float Nine", []byte("9.9e1"), true},
		{"Exponent Float Plus", []byte("+1.1E2"), false},
		{"Exponent Float Minus", []byte("-1.1e3"), false},
		{"Exponent Float Numeric", []byte("230.847E4"), true},
		{"Exponent Float Numeric Leading 0", []byte("0.274e5"), true},
		{"Exponent Float Numeric Leading Period", []byte(".274E6"), false},
		{"Exponent Float Numeric Leading 0 Period", []byte("0.274e7"), true},
		{"Exponent Float Numeric Leading Digits", []byte("083.274E8"), false},

		{"AlphaNum", []byte("9723hj78"), false},
		{"AlphaNumPlus", []byte("+9723hj78"), false},
		{"AlphaNumMinus", []byte("-9723hj78"), false},
		{"Alpha", []byte("hello"), false},
		{"Newline", []byte("5\n10"), false},
		{"Null", []byte("\x00\x00\x00"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, isPositiveNumber(tc.input))
		})
	}
}

func TestJSONIsDecimalNumber(t *testing.T) {
	testCases := []struct {
		label    string
		input    []byte
		expected bool
	}{
		{"Empty", []byte(""), false},
		{"Zero", []byte("0"), true},
		{"Zeros", []byte("00"), false},
		{"One", []byte("1"), true},
		{"Ones", []byte("11"), true},
		{"Two", []byte("2"), true},
		{"Twos", []byte("22"), true},
		{"Three", []byte("3"), true},
		{"Threes", []byte("33"), true},
		{"Four", []byte("4"), true},
		{"Fours", []byte("44"), true},
		{"Five", []byte("5"), true},
		{"Fives", []byte("55"), true},
		{"Six", []byte("6"), true},
		{"Sixs", []byte("66"), true},
		{"Seven", []byte("7"), true},
		{"Sevens", []byte("77"), true},
		{"Eight", []byte("8"), true},
		{"Eights", []byte("88"), true},
		{"Nine", []byte("9"), true},
		{"Nines", []byte("99"), true},
		{"Plus", []byte("+11"), false},
		{"Minus", []byte("-11"), false},
		{"Numeric", []byte("230847"), true},
		{"Numeric Leading 0", []byte("083274"), false},
		{"AlphaNum", []byte("9723hj78"), false},
		{"AlphaNumPlus", []byte("+9723hj78"), false},
		{"AlphaNumMinus", []byte("-9723hj78"), false},
		{"Alpha", []byte("hello"), false},
		{"Newline", []byte("5\n10"), false},
		{"Null", []byte("\x00\x00\x00"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, isDecimalNumber(tc.input))
		})
	}
}

func TestJSONIsExponent(t *testing.T) {
	testCases := []struct {
		label    string
		input    []byte
		expected bool
	}{
		{"Empty", []byte(""), false},
		{"Zero", []byte("0"), true},
		{"Zeros", []byte("00"), true},
		{"One", []byte("1"), true},
		{"Ones", []byte("11"), true},
		{"Two", []byte("2"), true},
		{"Twos", []byte("22"), true},
		{"Three", []byte("3"), true},
		{"Threes", []byte("33"), true},
		{"Four", []byte("4"), true},
		{"Fours", []byte("44"), true},
		{"Five", []byte("5"), true},
		{"Fives", []byte("55"), true},
		{"Six", []byte("6"), true},
		{"Sixs", []byte("66"), true},
		{"Seven", []byte("7"), true},
		{"Sevens", []byte("77"), true},
		{"Eight", []byte("8"), true},
		{"Eights", []byte("88"), true},
		{"Nine", []byte("9"), true},
		{"Nines", []byte("99"), true},
		{"Plus", []byte("+11"), true},
		{"Minus", []byte("-11"), true},
		{"Numeric", []byte("230847"), true},
		{"Numeric Leading 0", []byte("083274"), true},
		{"AlphaNum", []byte("9723hj78"), false},
		{"AlphaNumPlus", []byte("+9723hj78"), false},
		{"AlphaNumMinus", []byte("-9723hj78"), false},
		{"Alpha", []byte("hello"), false},
		{"Newline", []byte("5\n10"), false},
		{"Null", []byte("\x00\x00\x00"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, isExponent(tc.input))
		})
	}
}

func TestJSONIsDigits(t *testing.T) {
	testCases := []struct {
		label    string
		input    []byte
		expected bool
	}{
		{"Empty", []byte(""), false},
		{"Zero", []byte("0"), true},
		{"Zeros", []byte("00"), true},
		{"One", []byte("11"), true},
		{"Two", []byte("22"), true},
		{"Three", []byte("33"), true},
		{"Four", []byte("44"), true},
		{"Five", []byte("55"), true},
		{"Six", []byte("66"), true},
		{"Seven", []byte("77"), true},
		{"Eight", []byte("88"), true},
		{"Nine", []byte("99"), true},
		{"Numeric", []byte("230847"), true},
		{"Numeric Leading 0", []byte("083274"), true},
		{"AlphaNum", []byte("9723hj78"), false},
		{"Alpha", []byte("hello"), false},
		{"Newline", []byte("5\n10"), false},
		{"Null", []byte("\x00\x00\x00"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, isDigits(tc.input))
		})
	}
}

func TestJSONIsDigit(t *testing.T) {
	var empty byte

	testCases := []struct {
		label    string
		input    byte
		expected bool
	}{
		{"Empty", empty, false},
		{"Zero", '0', true},
		{"One", '1', true},
		{"Two", '2', true},
		{"Three", '3', true},
		{"Four", '4', true},
		{"Five", '5', true},
		{"Six", '6', true},
		{"Seven", '7', true},
		{"Eight", '8', true},
		{"Nine", '9', true},
		{"Alpha", 'a', false},
		{"Newline", '\n', false},
		{"Null", '\x00', false},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, isDigit(tc.input))
		})
	}
}

func TestJSONIsOneToNine(t *testing.T) {
	var empty byte
	testCases := []struct {
		label    string
		input    byte
		expected bool
	}{
		{"Empty", empty, false},
		{"Zero", '0', false},
		{"One", '1', true},
		{"Two", '2', true},
		{"Three", '3', true},
		{"Four", '4', true},
		{"Five", '5', true},
		{"Six", '6', true},
		{"Seven", '7', true},
		{"Eight", '8', true},
		{"Nine", '9', true},
		{"Alpha", 'a', false},
		{"Newline", '\n', false},
		{"Null", '\x00', false},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, isOneToNine(tc.input))
		})
	}
}

func TestJSONIsString(t *testing.T) {
	var empty []byte
	testCases := []struct {
		label    string
		input    []byte
		expected bool
	}{
		{"Empty", empty, false},
		{"String 1", []byte(`This is string characters`), false},
		{"String 2", []byte(`This is string characters\n`), false},
		{"String 3", []byte(`\"This is string characters\"`), false},
		{"String 4", []byte(`"Missing Quote`), false},
		{"String 5", []byte(`"Text Outside" Quote`), false},
		{"String 6", []byte(`"Escaped \" Quote"`), true},
		{"String 7", []byte(`"With Newline \n"`), true},
		{"String 8", []byte(`"This string is [\n\b99%\b\n] better than the last \"string\""`), true},
		{"String 9", []byte(`""`), true},
		{"String 10", []byte(`"Escaped \a thing"`), false},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsJSONString(tc.input))
		})
	}
}

func TestJSONIsStringCharacters(t *testing.T) {
	var empty []byte
	testCases := []struct {
		label    string
		input    []byte
		expected bool
	}{
		{"Empty", empty, false},
		{"String 1", []byte(`This is string characters`), true},
		{"String 2", []byte(`This is string characters\n`), true},
		{"String 3", []byte(`\"This is string characters\"`), true},
		{"String 4", []byte(`"This is not string characters"`), false},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, isStringCharacters(tc.input))
		})
	}
}

func TestJSONIsStringCharacter(t *testing.T) {
	testCases := []struct {
		label    string
		input    []byte
		expected bool
	}{
		// String characters:
		{"Slash", []byte{'\\'}, false},
		{"Quote", []byte{'"'}, false},
		{"Hex 1", []byte{'\x0a'}, false},
		{"Hex 2", []byte{'\u001F'}, false},
		{"Hex 3", []byte{'\u000F'}, false},
		{"Hex 4", []byte{'\u0020'}, true},
		{"Hex 5", []byte{'\x00'}, false},
		{"Char 1", []byte{'a'}, true},
		{"Char 2", []byte{'z'}, true},
		{"Char 3", []byte{'!'}, true},
		{"Char 4", []byte{'1'}, true},
		{"Char 5", []byte{'9'}, true},
		{"Char 5", []byte{'9'}, true},
		{"Quote", []byte{'\\', '"'}, true},

		// Escape Sequences:
		{"FrontSlash", []byte{'\\', '/'}, true},
		{"BackSlash", []byte{'\\', '\\'}, true},
		{"B", []byte{'\\', 'b'}, true},
		{"F", []byte{'\\', 'f'}, true},
		{"N", []byte{'\\', 'n'}, true},
		{"R", []byte{'\\', 'r'}, true},
		{"T", []byte{'\\', 't'}, true},
		{"Hex", []byte{'\\', 'u', '0', '0', '0', '0'}, true},
		{"Hex 2", []byte{'u', '0', '0', '0', '0'}, false},
		{"Hex 3", []byte{'\\', '0', '0', '0', '0'}, false},
		{"Hex 4", []byte{'\\', 'u', '1', 'd', '4', 'F'}, true},
		{"Hex 5", []byte{'\\', 'u', '1', 'G', '4', 'F'}, false},
		{"Hex 6", []byte{'\\', 'u', '1', '1', '4', 'F', 'F'}, false},
		{"A", []byte{'\\', 'a'}, false},
		{"1", []byte{'\\', '1'}, false},
		{"Long", []byte{'\\', 'b', 'f'}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, isStringCharacter(tc.input))
		})
	}
}

func TestJSONIsEscapeSequence(t *testing.T) {
	testCases := []struct {
		label    string
		input    []byte
		expected bool
	}{
		{"Empty", []byte{}, false},
		{"Quote", []byte{'\\', '"'}, true},
		{"FrontSlash", []byte{'\\', '/'}, true},
		{"BackSlash", []byte{'\\', '\\'}, true},
		{"B", []byte{'\\', 'b'}, true},
		{"F", []byte{'\\', 'f'}, true},
		{"N", []byte{'\\', 'n'}, true},
		{"R", []byte{'\\', 'r'}, true},
		{"T", []byte{'\\', 't'}, true},
		{"Hex", []byte{'\\', 'u', '0', '0', '0', '0'}, true},
		{"Hex 2", []byte{'u', '0', '0', '0', '0'}, false},
		{"Hex 3", []byte{'\\', '0', '0', '0', '0'}, false},
		{"Hex 4", []byte{'\\', 'u', '1', 'd', '4', 'F'}, true},
		{"Hex 5", []byte{'\\', 'u', '1', 'G', '4', 'F'}, false},
		{"Hex 6", []byte{'\\', 'u', '1', '1', '4', 'F', 'F'}, false},
		{"A", []byte{'\\', 'a'}, false},
		{"1", []byte{'\\', '1'}, false},
		{"Long", []byte{'\\', 'b', 'f'}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, isEscapeSequence([]byte(tc.input)))
		})
	}
}

func TestJSONIsHexDigit(t *testing.T) {
	testCases := []struct {
		label    string
		input    byte
		expected bool
	}{
		{"Zero", '0', true},
		{"One", '1', true},
		{"Two", '2', true},
		{"Three", '3', true},
		{"Four", '4', true},
		{"Five", '5', true},
		{"Six", '6', true},
		{"Seven", '7', true},
		{"Eight", '8', true},
		{"Nine", '9', true},
		{"Invalid", '-', false},
		{"a", 'a', true},
		{"b", 'b', true},
		{"c", 'c', true},
		{"d", 'd', true},
		{"e", 'e', true},
		{"f", 'f', true},
		{"z", 'z', false},
		{"A", 'A', true},
		{"B", 'B', true},
		{"C", 'C', true},
		{"D", 'D', true},
		{"E", 'E', true},
		{"F", 'F', true},
		{"G", 'G', false},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, isHexDigit(tc.input))
		})
	}
}

func TestJSONIsObject(t *testing.T) {
	var blank []byte
	testCases := []struct {
		label    string
		input    []byte
		expected bool
	}{
		{"Blank", blank, false},
		{"Empty", []byte("{}"), true},
		{"String", []byte(`{"key": "string"}`), true},
		{"Number", []byte(`{"key": 129E72}`), true},
		{"True", []byte(`{"key": true}`), true},
		{"False", []byte(`{"key": false}`), true},
		{"Null", []byte(`{"key":null}`), true},
		{"Combo", []byte(` {"key1": "string", "key2": 129E72, "key3": true, "key4": false} `), true},
		{"Trailing Comma", []byte(`{ "key1": "string", "key2": 129E72, "key3": true, "key4": false, } `), false},
		{"Complex", []byte(` { "a":"string", "b":129E72, "c":true, "d":false, "e":["subarray", ["Hello", "Goodbye"]], "f":{"key": [null, "true"]} } `), true},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsJSONObject(tc.input))
		})
	}
}

func TestJSONIsMembers(t *testing.T) {
	var blank []byte
	testCases := []struct {
		label    string
		input    []byte
		expected bool
	}{
		{"Blank", blank, false},
		{"Empty", []byte(`{}`), false},
		{"Malformed", []byte(`"key": val`), false},
		{"String", []byte(`"key": "string"`), true},
		{"Number", []byte(`"key": 129E72`), true},
		{"True", []byte(`"key": true`), true},
		{"False", []byte(`"key": false`), true},
		{"Null", []byte(`"key": null`), true},
		{"Combo", []byte(` "key1": "string", "key2": 129E72, "key3": true, "key4": false `), true},
		{"Trailing Comma", []byte(` "key1": "string", "key2": 129E72, "key3": true, "key4": false,  `), false},
		{"Complex", []byte(` "a":"string", "b":129E72, "c":true, "d":false, "e":["subarray", ["Hello", "Goodbye"]], "f":{"key": [null, "true"]} `), true},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, isMembers(tc.input))
		})
	}
}

func TestJSONIsArray(t *testing.T) {
	var blank []byte
	testCases := []struct {
		label    string
		input    []byte
		expected bool
	}{
		{"Blank", blank, false},
		{"Empty", []byte(`[]`), true},
		{"No Open", []byte(`"string"]`), false},
		{"No Close", []byte(`["string"`), false},
		{"String", []byte(`["string"]`), true},
		{"Embeded Open", []byte(`["str[ing"]`), true},
		{"Embeded Close", []byte(`["str]ing"]`), true},
		{"Number", []byte(`[129E72]`), true},
		{"True", []byte(`[true]`), true},
		{"False", []byte(`[false]`), true},
		{"Null", []byte(`[null]`), true},
		{"Combo", []byte(`  [  "string", 129E72, true, false ] `), true},
		{"Embeded", []byte(`["string", 129E72, true, false, [ "a", 2, ["b"], true]]`), true},
		{"Trailing Comma", []byte(`[ "string", 129E72, true, false,  ]`), false},
		{"Extra Open", []byte(`[ ["string", 129E72, true, false ]`), false},
		{"Extra Close", []byte(`[ "string", 129E72, true, false  ]]`), false},
		{"Test", []byte(`[ false,  ]]`), false},
		//{"Complex", []byte(` ["string", 129E72, true, false, ["subarray", ["Hello", "Goodbye"]], {"key": [null, "true"]}] `), true},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsJSONArray(tc.input))
		})
	}
}

func TestJSONIsArrayElements(t *testing.T) {
	var empty []byte
	testCases := []struct {
		label    string
		input    []byte
		expected bool
	}{
		{"Empty", empty, false},
		{"String", []byte(`"string"`), true},
		{"Number", []byte(`129E72`), true},
		{"True", []byte(`true`), true},
		{"False", []byte(`false`), true},
		{"Null", []byte(`null`), true},
		{"Combo", []byte(` "string", 129E72, true, false `), true},
		{"Trailing Comma", []byte(` "string", 129E72, true, false,  `), false},
		//{"Complex", []byte(` "string", 129E72, true, false, ["subarray", ["Hello", "Goodbye"]], {"key": [null, "true"]} `), true},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			assert.Equal(t, tc.expected, isArrayElements(tc.input))
		})
	}
}

func TestGetJSONType(t *testing.T) {
	testCases := []struct {
		label    string
		input    string
		expected string
	}{
		{label: `Invalid`, input: `http://not.actually.json/`, expected: JSONInvalid},
		{label: `Empty`, input: ``, expected: JSONInvalid},
		{label: `String Valid`, input: `"http://is.actually.json/"`, expected: JSONString},
		{label: `String Invalid`, input: `"No closing quote`, expected: JSONString},
		{label: `Int`, input: `21`, expected: JSONInt},
		{label: `Negative Int`, input: `-21`, expected: JSONInt},
		{label: `Float`, input: `21.0`, expected: JSONFloat},
		{label: `Negative Float`, input: `-21.0`, expected: JSONFloat},
		{label: `Invalid Int`, input: `-a21`, expected: JSONInvalid},
		{label: `Invalid Float`, input: `-a21.0`, expected: JSONInvalid},
		{label: `Invalid Bool`, input: `trust`, expected: JSONInvalid},
		{label: `Invalid with N`, input: `NotNull`, expected: JSONInvalid},
		{label: `Bool t`, input: `true`, expected: JSONBool},
		{label: `Bool T`, input: `True`, expected: JSONBool},
		{label: `Bool f`, input: `false`, expected: JSONBool},
		{label: `Bool F`, input: `False`, expected: JSONBool},
		{label: `Null n`, input: `null`, expected: JSONNull},
		{label: `Null N`, input: `Null`, expected: JSONNull},
		{label: `Object Valid`, input: `{"a": "b"}`, expected: JSONObject},
		{label: `Object InValid`, input: `{"a""b"}`, expected: JSONObject},
		{label: `Array Valid`, input: `["a", "b"]`, expected: JSONArray},
		{label: `Array InValid`, input: `["a""b"]`, expected: JSONArray},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			dt := GetJSONType([]byte(tc.input), 0)
			assert.Equal(t, tc.expected, dt)
		})
	}
}

func TestGetJSONTypeStrict(t *testing.T) {
	testCases := []struct {
		label    string
		input    string
		expected string
	}{
		{label: `Invalid`, input: `http://not.actually.json/`, expected: JSONInvalid},
		{label: `Empty`, input: ``, expected: JSONInvalid},
		{label: `String Valid`, input: `"http://is.actually.json/"`, expected: JSONString},
		{label: `String Invalid`, input: `"No closing quote`, expected: JSONInvalid},
		{label: `Int`, input: `21`, expected: JSONInt},
		{label: `Negative Int`, input: `-21`, expected: JSONInt},
		{label: `Float`, input: `21.0`, expected: JSONFloat},
		{label: `Negative Float`, input: `-21.0`, expected: JSONFloat},
		{label: `Invalid Int`, input: `-a21`, expected: JSONInvalid},
		{label: `Invalid Float`, input: `-a21.0`, expected: JSONInvalid},
		{label: `Invalid Bool`, input: `trust`, expected: JSONInvalid},
		{label: `Invalid with N`, input: `NotNull`, expected: JSONInvalid},
		{label: `Bool t`, input: `true`, expected: JSONBool},
		{label: `Bool T`, input: `True`, expected: JSONBool},
		{label: `Bool f`, input: `false`, expected: JSONBool},
		{label: `Bool F`, input: `False`, expected: JSONBool},
		{label: `Null n`, input: `null`, expected: JSONNull},
		{label: `Null N`, input: `Null`, expected: JSONNull},
		{label: `Object Valid`, input: `{"a": "b"}`, expected: JSONObject},
		{label: `Object Invalid`, input: `{"a""b"}`, expected: JSONInvalid},
		{label: `Array Valid`, input: `["a", "b"]`, expected: JSONArray},
		{label: `Array Invalid`, input: `["a""b"]`, expected: JSONInvalid},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			dt := GetJSONTypeStrict([]byte(tc.input), 0)
			assert.Equal(t, tc.expected, dt)
		})
	}
}

func TestFindTerminator(t *testing.T) {
	t.Run("Start LessThan 0", func(t *testing.T) {
		assert.Equal(t, -1, findTerminator([]byte(`[ "a" , "b"]`), -1))
	})
	t.Run("With Whitespace", func(t *testing.T) {
		assert.Equal(t, 7, findTerminator([]byte(`[ "a" , "b" ]`), 5))
	})
}

func TestIsEmptyArray(t *testing.T) {
	t.Run("No Opener", func(t *testing.T) {
		assert.False(t, IsEmptyArray([]byte(`"a", "b", "c"]`)))
	})
}
