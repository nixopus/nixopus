package realtime

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *SocketServer) verifyToken(tokenString string) (*types.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return types.JWTSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return nil, fmt.Errorf("token expired")
			}
		}

		email, ok := claims["email"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid token claims")
		}

		userStorage := user_storage.UserStorage{
			DB:  s.db,
			Ctx: s.ctx,
		}

		user, err := userStorage.FindUserByEmail(email)
		if err != nil {
			return nil, err
		}

		return user, nil
	}

	return nil, fmt.Errorf("invalid token")
}
