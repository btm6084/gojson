package gojson

import (
	"errors"
	"fmt"
)

// PanicRecovery returns a general use Panic Recovery function to capture panics
// and returns them as errors. A pointer to the error to populate will be passed
// in via the err parameter. err must be addressable.
//
// usage: defer PanicRecovery(&err)()
func PanicRecovery(err *error) {
	if r := recover(); r != nil && err != nil {
		// Create a new error and assign it to our pointer.
		switch r.(type) {
		case error:
			*err = r.(error)
		case string:
			*err = errors.New(r.(string))
		default:
			*err = fmt.Errorf("Panic. Context: %+v", r)
		}
	}
}

// Truncate returns a truncated byte slice if the length of the original slice is greater
// than a given max.
func truncate(b []byte, max int) []byte {
	if len(b) <= max {
		return b
	}

	return b[:max]
}
