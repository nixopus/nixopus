package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword generates a hashed version of the given password using bcrypt.
//
// This function takes a plaintext password as input and returns a hashed version
// of the password. It uses the bcrypt algorithm with the default cost for hashing.
// If the hashing process fails, it returns an error describing the failure.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedPassword), nil
}
