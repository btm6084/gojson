package gojson

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"strings"
)

// PanicRecovery returns a general use Panic Recovery function to capture panics
// and returns them as errors. A pointer to the error to populate will be passed
// in via the err parameter. err must be addressable.
//
// usage: defer PanicRecovery(&err)()
func PanicRecovery(err *error) {
	if r := recover(); r != nil && err != nil {
		origin := panicOrigin(debug.Stack())

		// Create a new error and assign it to our pointer.
		switch r := r.(type) {
		case error:
			*err = fmt.Errorf(`%w (%s)`, r, origin)
		case string:
			*err = fmt.Errorf(`%s (%s)`, r, origin)
		default:
			*err = fmt.Errorf("panic. context: %+v (%s)", r, origin)
		}
	}
}

// Find the origin of a panic so we can annotate our error.
func panicOrigin(raw []byte) string {
	stack := bytes.Split(raw, []byte{'\n'})

	// Lines:
	// 0 is goroutineID
	// 1-2 are debug.Stack itself.
	// 3-4 are this file's PanicRecovery function.
	// 5-6 are runtime.Panic
	// 7-8 are the origination, where 8 has the info we need.

	if len(stack) < 9 {
		return "unknown panic"
	}

	return strings.TrimSpace(string(stack[8]))
}

// Truncate returns a truncated byte slice if the length of the original slice is greater
// than a given max.
func truncate(b []byte, max int) []byte {
	if len(b) <= max {
		return b
	}

	return b[:max]
}
