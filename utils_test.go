package gojson

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPanicRecoveryString(t *testing.T) {
	var err error
	defer func() {
		assert.False(t, err == nil)
		assert.Equal(t, "Test", err.Error())
	}()
	defer PanicRecovery(&err)

	assert.True(t, err == nil)

	panic("Test")
}

func TestPanicRecoveryError(t *testing.T) {
	var err error
	defer func() {
		assert.False(t, err == nil)
		assert.Equal(t, "From Error", err.Error())
	}()
	defer PanicRecovery(&err)

	assert.True(t, err == nil)

	panic(errors.New("From Error"))
}

func TestPanicRecoveryInterface(t *testing.T) {
	var err error
	defer func() {
		assert.False(t, err == nil)
		assert.Equal(t, "Panic. Context: {A:From Struct B:true C:17}", err.Error())
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
