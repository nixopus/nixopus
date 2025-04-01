package utils

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

// createToken generates a JWT token for the given email.
//
// This function creates a new JWT token using the HS256 signing method. The token
// includes the user's email and an expiration time set to 24 hours from the time
// of creation. It returns the signed token string or an error if the signing process
// fails.
func CreateToken(email string, duration time.Duration, twoFactorEnabled bool, twoFactorVerified bool) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"email":        email,
			"exp":          time.Now().Add(duration).Unix(),
			"2fa_enabled":  twoFactorEnabled,
			"2fa_verified": twoFactorVerified,
		})

	tokenString, err := token.SignedString(types.JWTSecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
