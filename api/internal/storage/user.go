package storage

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
	"golang.org/x/net/context"
)

// FindUserByEmail finds a user by email in the database.
//
// The function returns an error if the user does not exist or if the query
// fails.
func FindUserByEmail(db *bun.DB, email string, ctx context.Context) (*types.User, error) {
	user := &types.User{}
	err := db.NewSelect().Model(user).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func FindUserByID(db *bun.DB, id string, ctx context.Context) (*types.User, error) {
	user := &types.User{}
	err := db.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func CreateUser(db *bun.DB, user *types.User, ctx context.Context) error {
	_, err := db.NewInsert().Model(user).Exec(ctx)
	return err
}

func UpdateUser(db *bun.DB, user *types.User, ctx context.Context) error {
	_, err := db.NewUpdate().Model(user).Where("id = ?", user.ID).Exec(ctx)
	return err
}

func GetResetToken(db *bun.DB, token string, ctx context.Context) (*types.User, error) {
	user := &types.User{}
	err := db.NewSelect().Model(user).Where("reset_token = ?", token).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func CreateRefreshToken(db *bun.DB, userID uuid.UUID, ctx context.Context) (*types.RefreshToken, error) {
    refreshToken := &types.RefreshToken{
        ID:        uuid.New(),
        UserID:    userID,
        Token:     uuid.New().String(),
        ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
        CreatedAt: time.Now(),
    }

    _, err := db.NewInsert().Model(refreshToken).Exec(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to store refresh token: %w", err)
    }

    return refreshToken, nil
}

func GetRefreshToken(db *bun.DB, token string, ctx context.Context) (*types.RefreshToken, error) {
    var refreshToken types.RefreshToken
    err := db.NewSelect().Model(&refreshToken).Where("token = ?", token).Scan(ctx)
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

func RevokeRefreshToken(db *bun.DB, token string, ctx context.Context) error {
	_, err := db.NewUpdate().Model(&types.RefreshToken{RevokedAt: &time.Time{}}).Where("token = ?", token).Exec(ctx)
	return err
}
