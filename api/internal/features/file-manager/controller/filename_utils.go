package controller

import (
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"
)

// SanitizeFilenameForHeader sanitizes a filename for safe use in HTTP Content-Disposition header
// and returns a properly formatted header value with both filename and filename* parameters
func SanitizeFilenameForHeader(originalName string) string {
	const maxFilenameLength = 255

	// Step 1: Remove control characters (including CR/LF) and other dangerous chars
	sanitized := strings.Map(func(r rune) rune {
		// Remove control characters, quotes, semicolons, and other header injection chars
		if r < 32 || r == 127 || r == '"' || r == ';' || r == '\\' || r == '\r' || r == '\n' {
			return -1
		}
		return r
	}, originalName)

	// Step 2: Trim whitespace and limit length
	sanitized = strings.TrimSpace(sanitized)

	// Convert to rune slice to handle UTF-8 properly
	runes := []rune(sanitized)
	if len(runes) > maxFilenameLength {
		runes = runes[:maxFilenameLength]
		// Try to cut at word boundary if possible
		lastSpace := -1
		for i := len(runes) - 1; i >= 0; i-- {
			if runes[i] == ' ' {
				lastSpace = i
				break
			}
		}
		if lastSpace > maxFilenameLength/2 {
			runes = runes[:lastSpace]
		}
	}
	sanitized = string(runes)

	// Ensure we didn't break a UTF-8 sequence
	for !utf8.ValidString(sanitized) && len(sanitized) > 0 {
		sanitized = sanitized[:len(sanitized)-1]
	}

	// Step 3: Ensure we have a non-empty filename
	if sanitized == "" {
		sanitized = "download"
	}

	// Step 4: Create ASCII-safe version for filename parameter
	asciiSafe := strings.Map(func(r rune) rune {
		if r >= 32 && r < 127 && r != '"' && r != ';' && r != '\\' {
			return r
		}
		return '_'
	}, sanitized)

	// Step 5: Create RFC 5987 encoded version for filename*
	var encoded string
	if utf8.ValidString(sanitized) {
		// Use proper percent-encoding for RFC 5987 (not QueryEscape which uses ' ' for spaces)
		encoded = strings.ReplaceAll(url.PathEscape(sanitized), " ", "%20")
	} else {
		// If not valid UTF-8, fall back to ASCII version
		encoded = strings.ReplaceAll(url.PathEscape(asciiSafe), " ", "%20")
	}

	// Step 6: Format the Content-Disposition header
	// Use filename for ASCII-safe names, filename* for UTF-8
	if asciiSafe == sanitized {
		// ASCII-safe, use simple filename
		return fmt.Sprintf("attachment; filename=\"%s\"", asciiSafe)
	} else {
		// UTF-8, use both filename (ASCII fallback) and filename*
		return fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", asciiSafe, encoded)
	}
}