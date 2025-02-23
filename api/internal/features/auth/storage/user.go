package storage

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
	"golang.org/x/net/context"
)

type UserStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

// FindUserByEmail finds a user by email in the database.
//
// The function returns an error if the user does not exist or if the query
// fails.
func (u *UserStorage) FindUserByEmail(email string) (*types.User, error) {
	user := &types.User{}
	err := u.DB.NewSelect().Model(user).Where("email = ?", email).Scan(u.Ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserStorage) FindUserByID(id string) (*types.User, error) {
	user := &types.User{}
	err := u.DB.NewSelect().Model(user).Where("id = ?", id).Scan(u.Ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserStorage) CreateUser(user *types.User) error {
	_, err := u.DB.NewInsert().Model(user).Exec(u.Ctx)
	return err
}

func (u *UserStorage) UpdateUser(user *types.User) error {
	_, err := u.DB.NewUpdate().Model(user).Where("id = ?", user.ID).Exec(u.Ctx)
	return err
}

func (u *UserStorage) GetResetToken(token string) (*types.User, error) {
	user := &types.User{}
	err := u.DB.NewSelect().Model(user).Where("reset_token = ?", token).Scan(u.Ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserStorage) CreateRefreshToken(userID uuid.UUID) (*types.RefreshToken, error) {
	refreshToken := &types.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     uuid.New().String(),
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
		CreatedAt: time.Now(),
	}

	_, err := u.DB.NewInsert().Model(refreshToken).Exec(u.Ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return refreshToken, nil
}

func (u *UserStorage) GetRefreshToken(token string) (*types.RefreshToken, error) {
	var refreshToken types.RefreshToken
	err := u.DB.NewSelect().Model(&refreshToken).Where("token = ?", token).Scan(u.Ctx)
	if err != nil {
		return nil, fmt.Errorf("refresh token not found: %w", err)
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("refresh token expired")
	}

	if refreshToken.RevokedAt != nil {
		return nil, fmt.Errorf("refresh token revoked")
	}

	return &refreshToken, nil
}

func (u *UserStorage) RevokeRefreshToken(token string) error {
	_, err := u.DB.NewUpdate().Model(&types.RefreshToken{RevokedAt: &time.Time{}}).Where("token = ?", token).Exec(u.Ctx)
	return err
}
