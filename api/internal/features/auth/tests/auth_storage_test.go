package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestUserStorage(t *testing.T) {
	setup := testutils.NewTestSetup()
	userStorage := setup.UserStorage

	t.Run("CreateUser", func(t *testing.T) {
		user := &shared_types.User{
			ID:        uuid.New(),
			Email:     "test@example.com",
			Password:  "hashedpassword",
			Username:  "testuser",
			Type:      "viewer",
			CreatedAt: time.Now(),
		}

		err := userStorage.CreateUser(user)
		assert.NoError(t, err)

		// Verify user was created
		foundUser, err := userStorage.FindUserByID(user.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, user.Email, foundUser.Email)
		assert.Equal(t, user.Username, foundUser.Username)
	})

	t.Run("FindUserByEmail", func(t *testing.T) {
		user := &shared_types.User{
			ID:        uuid.New(),
			Email:     "findbyemail@example.com",
			Password:  "hashedpassword",
			Username:  "findbyemail",
			Type:      "viewer",
			CreatedAt: time.Now(),
		}

		err := userStorage.CreateUser(user)
		assert.NoError(t, err)

		foundUser, err := userStorage.FindUserByEmail(user.Email)
		assert.NoError(t, err)
		assert.Equal(t, user.Email, foundUser.Email)
		assert.Equal(t, user.Username, foundUser.Username)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		user := &shared_types.User{
			ID:        uuid.New(),
			Email:     "update@example.com",
			Password:  "hashedpassword",
			Username:  "updateuser",
			Type:      "viewer",
			CreatedAt: time.Now(),
		}

		err := userStorage.CreateUser(user)
		assert.NoError(t, err)

		user.Username = "updatedusername"
		err = userStorage.UpdateUser(user)
		assert.NoError(t, err)

		updatedUser, err := userStorage.FindUserByID(user.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, "updatedusername", updatedUser.Username)
	})

	t.Run("CreateRefreshToken", func(t *testing.T) {
		user := &shared_types.User{
			ID:        uuid.New(),
			Email:     "refreshtoken@example.com",
			Password:  "hashedpassword",
			Username:  "refreshtokenuser",
			Type:      "viewer",
			CreatedAt: time.Now(),
		}

		err := userStorage.CreateUser(user)
		assert.NoError(t, err)

		refreshToken, err := userStorage.CreateRefreshToken(user.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, refreshToken.Token)
		assert.True(t, refreshToken.ExpiresAt.After(time.Now()))
	})

	t.Run("GetRefreshToken", func(t *testing.T) {
		user := &shared_types.User{
			ID:        uuid.New(),
			Email:     "getrefreshtoken@example.com",
			Password:  "hashedpassword",
			Username:  "getrefreshtokenuser",
			Type:      "viewer",
			CreatedAt: time.Now(),
		}

		err := userStorage.CreateUser(user)
		assert.NoError(t, err)

		refreshToken, err := userStorage.CreateRefreshToken(user.ID)
		assert.NoError(t, err)

		foundToken, err := userStorage.GetRefreshToken(refreshToken.Token)
		assert.NoError(t, err)
		assert.Equal(t, refreshToken.Token, foundToken.Token)
		assert.Equal(t, user.ID, foundToken.UserID)
	})

	t.Run("RevokeRefreshToken", func(t *testing.T) {
		user := &shared_types.User{
			ID:        uuid.New(),
			Email:     "revoke@example.com",
			Password:  "hashedpassword",
			Username:  "revokeuser",
			Type:      "viewer",
			CreatedAt: time.Now(),
		}

		err := userStorage.CreateUser(user)
		assert.NoError(t, err)

		refreshToken, err := userStorage.CreateRefreshToken(user.ID)
		assert.NoError(t, err)

		err = userStorage.RevokeRefreshToken(refreshToken.Token)
		assert.NoError(t, err)

		_, err = userStorage.GetRefreshToken(refreshToken.Token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "refresh token revoked")
	})

	t.Run("StoreVerificationToken", func(t *testing.T) {
		user := &shared_types.User{
			ID:        uuid.New(),
			Email:     "verification@example.com",
			Password:  "hashedpassword",
			Username:  "verificationuser",
			Type:      "viewer",
			CreatedAt: time.Now(),
		}

		err := userStorage.CreateUser(user)
		assert.NoError(t, err)

		token := uuid.New().String()
		expiresAt := time.Now().Add(24 * time.Hour)

		err = userStorage.StoreVerificationToken(user.ID.String(), token, expiresAt)
		assert.NoError(t, err)

		userID, tokenExpiresAt, err := userStorage.GetVerificationToken(token)
		assert.NoError(t, err)
		assert.Equal(t, user.ID.String(), userID)
		assert.Equal(t, expiresAt.Unix(), tokenExpiresAt.Unix())
	})

	t.Run("UpdateUserEmailVerification", func(t *testing.T) {
		user := &shared_types.User{
			ID:        uuid.New(),
			Email:     "verifyemail@example.com",
			Password:  "hashedpassword",
			Username:  "verifyemailuser",
			Type:      "viewer",
			CreatedAt: time.Now(),
		}

		err := userStorage.CreateUser(user)
		assert.NoError(t, err)

		err = userStorage.UpdateUserEmailVerification(user.ID.String(), true)
		assert.NoError(t, err)

		updatedUser, err := userStorage.FindUserByID(user.ID.String())
		assert.NoError(t, err)
		assert.True(t, updatedUser.IsVerified)
	})

	t.Run("FindUserByType", func(t *testing.T) {
		user := &shared_types.User{
			ID:        uuid.New(),
			Email:     "findbytype@example.com",
			Password:  "hashedpassword",
			Username:  "findbytypeuser",
			Type:      "admin",
			CreatedAt: time.Now(),
		}

		err := userStorage.CreateUser(user)
		assert.NoError(t, err)

		foundUser, err := userStorage.FindUserByType("admin")
		assert.NoError(t, err)
		assert.Equal(t, user.Email, foundUser.Email)
		assert.Equal(t, user.Type, foundUser.Type)
	})

	t.Run("UserBelongsToOrganization", func(t *testing.T) {
		user := &shared_types.User{
			ID:        uuid.New(),
			Email:     "orgmember@example.com",
			Password:  "hashedpassword",
			Username:  "orgmember",
			Type:      "viewer",
			CreatedAt: time.Now(),
		}

		err := userStorage.CreateUser(user)
		assert.NoError(t, err)

		orgID := uuid.New().String()
		belongs, err := userStorage.UserBelongsToOrganization(user.ID.String(), orgID)
		assert.NoError(t, err)
		assert.False(t, belongs)
	})
}
