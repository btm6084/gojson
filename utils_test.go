package gojson

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPanicRecoveryString(t *testing.T) {
	var err error
	defer func() {
		assert.False(t, err == nil)
		assert.True(t, strings.HasPrefix(err.Error(), "Test"))
	}()
	defer PanicRecovery(&err)

	assert.True(t, err == nil)

	panic("Test")
}

func TestPanicRecoveryError(t *testing.T) {
	var err error
	defer func() {
		assert.False(t, err == nil)
		assert.True(t, strings.HasPrefix(err.Error(), "From Error"))
	}()
	defer PanicRecovery(&err)

	assert.True(t, err == nil)

	panic(errors.New("From Error"))
}

func TestPanicRecoveryInterface(t *testing.T) {
	var err error
	defer func() {
		assert.False(t, err == nil)
		assert.True(t, strings.HasPrefix(err.Error(), "panic. context: {A:From Struct B:true C:17}"))
	}()
	defer PanicRecovery(&err)

	assert.True(t, err == nil)

	things := struct {
		A string
		B bool
		C int
	}{"From Struct", true, 17}

	panic(things)
}
