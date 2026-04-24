package utils

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsClosedConnectionError(t *testing.T) {
	assert.False(t, isClosedConnectionError(nil))
	assert.True(t, isClosedConnectionError(io.EOF))
	for _, msg := range []string{
		"use of closed network connection",
		"connection closed",
		"connection lost",
		"broken pipe",
		"EOF",
		"connection reset by peer",
	} {
		assert.True(t, isClosedConnectionError(errors.New(msg)), msg)
	}
	assert.False(t, isClosedConnectionError(errors.New("some other failure")))
}
