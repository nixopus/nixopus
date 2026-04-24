package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateShellArg(t *testing.T) {
	require.NoError(t, ValidateShellArg("safe-value_1", "ref"))
	err := ValidateShellArg("a;b", "ref")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid characters")
}

func TestValidateGitRef(t *testing.T) {
	require.Error(t, ValidateGitRef("", "ref"))
	require.NoError(t, ValidateGitRef("main", "ref"))
	require.NoError(t, ValidateGitRef("feature/foo-bar", "ref"))
	err := ValidateGitRef("bad ref", "ref")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid characters")
}

func TestValidatePath(t *testing.T) {
	require.Error(t, ValidatePath("", "p"))
	require.NoError(t, ValidatePath("/home/x/file.txt", "p"))
	err := ValidatePath("/../etc", "p")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "traversal")
	err = ValidatePath("/tmp/x;y", "p")
	require.Error(t, err)
}

func TestShellQuote(t *testing.T) {
	assert.Equal(t, "'plain'", ShellQuote("plain"))
	assert.Equal(t, "'a'\"'\"'b'", ShellQuote("a'b"))
}
