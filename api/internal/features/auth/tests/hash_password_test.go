package tests

import (
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPasswordValidInput(t *testing.T) {
	password := "mysecretpassword"
	hashedPassword, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
}

func TestHashPasswordEmptyInput(t *testing.T) {
	password := ""
	_, err := service.HashPassword(password)
	if err == nil {
		t.Errorf("HashPassword should return an error for empty input")
	}
}
