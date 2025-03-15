package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

// TestCreateToken tests the CreateToken function by creating multiple tokens with
// different durations and ensuring that the resulting tokens are valid and have
// the correct claims. It also tests that an invalid token results in an error.
func TestCreateToken(t *testing.T) {
	table := []struct {
		name     string
		email    string
		duration time.Duration
	}{
		{"24 hour token", "nixopus_user1@nixopus.com", time.Hour * 24},
		{"1 hour token", "nixopus_user2@nixopus.com", time.Hour},
		{"15 minute token", "nixopus_user3@nixopus.com", time.Minute * 15},
		{"30 second token", "nixopus_user4@nixopus.com", time.Second * 30},
	}

	for _, test := range table {
		t.Run(test.name, func(t *testing.T) {
			token, err := service.CreateToken(test.email, test.duration)
			if err != nil {
				t.Fatalf("createToken failed: %v", err)
			}

			parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return types.JWTSecretKey, nil
			})

			if err != nil {
				t.Fatalf("failed to parse token: %v", err)
			}

			if !parsedToken.Valid {
				t.Errorf("token is invalid")
			}

			if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
				if email, ok := claims["email"].(string); !ok || email != test.email {
					t.Errorf("expected email %s, got %v", test.email, claims["email"])
				}

				now := time.Now().Unix()
				expectedExp := now + int64(test.duration.Seconds())
				if exp, ok := claims["exp"].(float64); !ok || int64(exp) < expectedExp-1 || int64(exp) > expectedExp+1 {
					t.Errorf("expected expiration around %d, got %v", expectedExp, claims["exp"])
				}
			} else {
				t.Errorf("failed to extract claims from token")
			}
		})
	}

	t.Run("invalid token", func(t *testing.T) {
		token := "invalid.jwt.token"
		_, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return types.JWTSecretKey, nil
		})
		if err == nil {
			t.Errorf("expected error when parsing invalid token, got nil")
		}
	})
}
