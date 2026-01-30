package storage

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type UserStorage struct {
	DB  *bun.DB
	Ctx context.Context
	tx  *bun.Tx
}

type AuthRepository interface {
	FindUserByEmail(email string) (*types.User, error)
	BeginTx() (bun.Tx, error)
	WithTx(tx bun.Tx) AuthRepository
}

func (u *UserStorage) WithTx(tx bun.Tx) AuthRepository {
	return &UserStorage{
		DB:  u.DB,
		Ctx: u.Ctx,
		tx:  &tx,
	}
}

func (u *UserStorage) BeginTx() (bun.Tx, error) {
	return u.DB.BeginTx(u.Ctx, nil)
}

func (u *UserStorage) getDB() bun.IDB {
	if u.tx != nil {
		return *u.tx
	}
	return u.DB
}

func (u *UserStorage) FindUserByEmail(email string) (*types.User, error) {
	user := &types.User{}
	err := u.getDB().NewSelect().
		Model(user).
		Where("email = ?", email).
		Relation("Organizations").
		Scan(u.Ctx)
	if err != nil {
		return nil, err
	}

	err = u.getDB().NewSelect().
		Model(&user.OrganizationUsers).
		Where("user_id = ?", user.ID).
		Relation("Organization").
		Scan(u.Ctx)
	if err != nil {
		return nil, err
	}

	user.ComputeCompatibilityFields()

	return user, nil
}
