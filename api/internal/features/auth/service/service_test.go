package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/auth/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

type stubAuthRepo struct {
	u   *shared_types.User
	err error
}

func (s stubAuthRepo) FindUserByEmail(string) (*shared_types.User, error) {
	return s.u, s.err
}

func (stubAuthRepo) BeginTx() (bun.Tx, error) {
	return bun.Tx{}, errors.New("stub: no tx")
}

func (s stubAuthRepo) WithTx(bun.Tx) storage.AuthRepository { return s }

func TestAuthService_GetUserByEmail(t *testing.T) {
	ctx := context.Background()
	l := logger.NewLogger()

	u := shared_types.User{
		ID:        uuid.New(),
		Name:      "n",
		Email:     "e@x",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	svc := NewAuthService(stubAuthRepo{u: &u}, l, ctx, "")
	got, err := svc.GetUserByEmail("e@x")
	require.NoError(t, err)
	require.Equal(t, u.ID, got.ID)

	wantErr := errors.New("boom")
	svc2 := NewAuthService(stubAuthRepo{err: wantErr}, l, ctx, "")
	_, err = svc2.GetUserByEmail("e@x")
	require.ErrorIs(t, err, wantErr)
}

func TestNewAuthService_withRedisCache(t *testing.T) {
	mr := miniredis.RunT(t)
	ctx := context.Background()
	l := logger.NewLogger()
	svc := NewAuthService(stubAuthRepo{}, l, ctx, "redis://"+mr.Addr())
	require.NotNil(t, svc.Cache)
}

func TestNewAuthService_invalidRedisProceedsWithoutCache(t *testing.T) {
	ctx := context.Background()
	l := logger.NewLogger()
	svc := NewAuthService(stubAuthRepo{}, l, ctx, "::not-a-redis-url")
	require.Nil(t, svc.Cache)

	svc2 := NewAuthService(stubAuthRepo{}, l, ctx, "")
	require.Nil(t, svc2.Cache)
}
