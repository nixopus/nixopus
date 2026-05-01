package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuth_errorsAreDistinctSentinels(t *testing.T) {
	require.False(t, errors.Is(ErrInvalidUser, ErrEmptyPassword))
	require.ErrorContains(t, ErrInvalidEmail, "email")
}
