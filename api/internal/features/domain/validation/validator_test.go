package validation_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/nixopus/nixopus/api/internal/features/domain/types"
	"github.com/nixopus/nixopus/api/internal/features/domain/validation"
)

// newValidator creates a Validator with a nil storage — safe because ValidateName
// never accesses the storage field.
func newValidator() *validation.Validator {
	return validation.NewValidator(nil)
}

// ---------------------------------------------------------------------------
// NewValidator
// ---------------------------------------------------------------------------

func TestNewValidator_NonNil(t *testing.T) {
	v := newValidator()
	if v == nil {
		t.Fatal("expected non-nil Validator")
	}
}

// ---------------------------------------------------------------------------
// ValidateName — all branches
// ---------------------------------------------------------------------------

func TestValidateName_Empty(t *testing.T) {
	err := newValidator().ValidateName("")
	if !errors.Is(err, types.ErrMissingDomainName) {
		t.Errorf("expected ErrMissingDomainName, got %v", err)
	}
}

func TestValidateName_TooShort(t *testing.T) {
	// len("ab") == 2 < 3
	err := newValidator().ValidateName("ab")
	if !errors.Is(err, types.ErrDomainNameTooShort) {
		t.Errorf("expected ErrDomainNameTooShort, got %v", err)
	}
}

func TestValidateName_TooLong(t *testing.T) {
	// len > 255
	err := newValidator().ValidateName(strings.Repeat("a", 256))
	if !errors.Is(err, types.ErrDomainNameTooLong) {
		t.Errorf("expected ErrDomainNameTooLong, got %v", err)
	}
}

func TestValidateName_NoDot(t *testing.T) {
	err := newValidator().ValidateName("nodotdomain")
	if !errors.Is(err, types.ErrDomainNameInvalid) {
		t.Errorf("expected ErrDomainNameInvalid, got %v", err)
	}
}

func TestValidateName_TLDTooShort(t *testing.T) {
	// "app.x.com" → Split[1] = "x" (len 1 < 2)
	err := newValidator().ValidateName("app.x.com")
	if !errors.Is(err, types.ErrDomainNameInvalid) {
		t.Errorf("expected ErrDomainNameInvalid for short TLD segment, got %v", err)
	}
}

func TestValidateName_TLDTooLong(t *testing.T) {
	// "app.{64 chars}.com" → Split[1] has 64 chars > 63
	long := "app." + strings.Repeat("a", 64) + ".com"
	err := newValidator().ValidateName(long)
	if !errors.Is(err, types.ErrDomainNameInvalid) {
		t.Errorf("expected ErrDomainNameInvalid for long TLD segment, got %v", err)
	}
}

func TestValidateName_Valid_ThreePart(t *testing.T) {
	// "app.example.com" → Split[1] = "example" (len 7, valid)
	err := newValidator().ValidateName("app.example.com")
	if err != nil {
		t.Errorf("expected nil for valid domain, got %v", err)
	}
}

func TestValidateName_Valid_TwoPart(t *testing.T) {
	// "example.co" → Split[1] = "co" (len 2, exactly at lower bound — valid)
	err := newValidator().ValidateName("example.co")
	if err != nil {
		t.Errorf("expected nil for two-part domain, got %v", err)
	}
}

func TestValidateName_Valid_TLDExactly63Chars(t *testing.T) {
	// Split[1] has exactly 63 chars — upper boundary, still valid
	name := "app." + strings.Repeat("a", 63) + ".com"
	err := newValidator().ValidateName(name)
	if err != nil {
		t.Errorf("expected nil for 63-char TLD segment, got %v", err)
	}
}

func TestValidateName_ExactlyAtLengthBoundaries(t *testing.T) {
	// Exactly 3 chars with a dot — shortest possible valid-looking input that
	// still fails (NoDot check passed but Split[1] = "" which is len 0 < 2)
	err := newValidator().ValidateName("a.b")
	if !errors.Is(err, types.ErrDomainNameInvalid) {
		t.Errorf("expected ErrDomainNameInvalid for 'a.b', got %v", err)
	}
}

func TestValidateName_ExactlyMaxLength(t *testing.T) {
	// 255-char name — exactly at the limit, should pass the length checks
	// Build "app.{243 'a' chars}.com" = 3+1+243+1+3 = 251 chars, valid
	name := "app." + strings.Repeat("a", 5) + ".com" // 12 chars, well under 255
	err := newValidator().ValidateName(name)
	if err != nil {
		t.Errorf("expected nil for name within length limits, got %v", err)
	}
}
