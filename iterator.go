package gojson

import (
	"errors"
	"fmt"
)

var (
	// ErrEndOfInput is returned when there are no further items to extract via Next()
	ErrEndOfInput = errors.New("Reached end of JSON Input")

	// ErrRequiresObject is returned when the input is neither an array or object.
	ErrRequiresObject = errors.New("NewIterator requires a valid JSONArray or JSONObject")
)

// Iterator receives a raw JSONArray or JSONObject, and provides an interface for extracting
// each member item one-by-one.
type Iterator struct {
	data      []byte
	close     byte
	lastStart int
	pos       int
	start     int
	end       bool
}

// NewIterator returns a primed Iterator
func NewIterator(raw []byte) (*Iterator, error) {
	if !IsJSON(raw) {
		return nil, ErrMalformedJSON
	}

	raw = trim(raw)
	var close byte = ']'

	if raw[0] != '[' && raw[0] != '{' {
		return nil, ErrRequiresObject
	}

	if raw[0] == '{' {
		close = '}'
	}

	if raw[len(raw)-1] != close && raw[len(raw)-1] != close {
		return nil, ErrRequiresObject
	}

	return &Iterator{
		data:      raw,
		close:     close,
		lastStart: 1,
		pos:       1,
		start:     1,
	}, nil
}

// Next returns the next member element in the container.
func (i *Iterator) Next() ([]byte, string, error) {
	if i.end {
		return nil, "", ErrEndOfInput
	}

	b, t, pos, err := extractValue(i.data, i.pos)
	if err != nil {
		return nil, "", err
	}

	pos = findTerminator(i.data, pos)
	if pos < 0 {
		return nil, "", fmt.Errorf("expected value terminator ('}', ']' or ',') at position '%d' in segment '%s'", pos, truncate(b, 50))
	}

	// We have run out of elements if the last terminator is not a comma
	if i.data[pos-1] != ',' {
		i.end = true
	}

	i.lastStart = i.pos
	i.pos = pos

	return b, t, err
}

// Last returns the most recently accessed member element in the container,
// or the first element if never accessed.
func (i *Iterator) Last() ([]byte, string, error) {
	b, t, _, err := extractValue(i.data, i.lastStart)

	return b, t, err
}

// Reset moves the internal pointer to the beginning of the JSON Input.
func (i *Iterator) Reset() {
	i.pos = i.start
	i.lastStart = i.start
	i.end = false
}
