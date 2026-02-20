package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// shellUnsafe matches characters that can break out of shell argument context.
var shellUnsafe = regexp.MustCompile(`[;&|` + "`" + `$(){}\\<>!'"*?\n\r]`)

// safeGitRef matches valid git ref characters (branches, tags, commit hashes).
var safeGitRef = regexp.MustCompile(`^[a-zA-Z0-9._/\-]+$`)

// safePath matches a reasonable filesystem path (no shell metacharacters).
var safePath = regexp.MustCompile(`^[a-zA-Z0-9._/\-~@ ]+$`)

// ValidateShellArg rejects strings containing shell metacharacters.
// Use this for values that will be interpolated into shell commands.
func ValidateShellArg(value, fieldName string) error {
	if shellUnsafe.MatchString(value) {
		return fmt.Errorf("%s contains invalid characters", fieldName)
	}
	return nil
}

// ValidateGitRef validates that a string is a safe git reference (branch, tag, commit hash).
func ValidateGitRef(value, fieldName string) error {
	if value == "" {
		return fmt.Errorf("%s is empty", fieldName)
	}
	if !safeGitRef.MatchString(value) {
		return fmt.Errorf("%s contains invalid characters: only alphanumeric, dots, slashes, and hyphens are allowed", fieldName)
	}
	return nil
}

// ValidatePath validates that a string is a safe filesystem path.
func ValidatePath(value, fieldName string) error {
	if value == "" {
		return fmt.Errorf("%s is empty", fieldName)
	}
	if !safePath.MatchString(value) {
		return fmt.Errorf("%s contains invalid characters", fieldName)
	}
	if strings.Contains(value, "..") {
		return fmt.Errorf("%s must not contain path traversal sequences", fieldName)
	}
	return nil
}

// ShellQuote wraps a value in single quotes, escaping any embedded single quotes.
// This is the safest way to pass a value into a shell command.
func ShellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}
