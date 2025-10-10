package realtime

import (
	"fmt"
	"net/http"

	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	session "github.com/supertokens/supertokens-golang/recipe/session"
)

// verifyToken verifies the SuperTokens session token and returns the user if the token is valid.
//
// Parameters:
//
//	tokenString - the SuperTokens session token string to verify.
//
// Returns:
//   - the user if the token is valid.
//   - an error if the token is invalid.
func (s *SocketServer) verifyToken(tokenString string) (*types.User, error) {
	// Create a mock request with the token to verify the session
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	var user *types.User
	var err error

	session.VerifySession(nil, func(w http.ResponseWriter, r *http.Request) {
		sessionContainer := session.GetSessionFromRequestContext(r.Context())
		userID := sessionContainer.GetUserID()
		userStorage := user_storage.UserStorage{
			DB:  s.db,
			Ctx: s.ctx,
		}
		user, err = userStorage.FindUserBySupertokensID(userID)
		if err != nil {
			fmt.Printf("Error finding user: %v\n", err)
		}
	}).ServeHTTP(nil, req)

	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	return user, nil
}
