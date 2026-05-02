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
	svc := NewAuthService(stubAuthRepo{u: &u}, nil, l, ctx, "")
	got, err := svc.GetUserByEmail("e@x")
	require.NoError(t, err)
	require.Equal(t, u.ID, got.ID)

	wantErr := errors.New("boom")
	svc2 := NewAuthService(stubAuthRepo{err: wantErr}, nil, l, ctx, "")
	_, err = svc2.GetUserByEmail("e@x")
	require.ErrorIs(t, err, wantErr)
}

func TestNewAuthService_withRedisCache(t *testing.T) {
	mr := miniredis.RunT(t)
	ctx := context.Background()
	l := logger.NewLogger()
	svc := NewAuthService(stubAuthRepo{}, nil, l, ctx, "redis://"+mr.Addr())
	require.NotNil(t, svc.Cache)
}

func TestNewAuthService_invalidRedisProceedsWithoutCache(t *testing.T) {
	ctx := context.Background()
	l := logger.NewLogger()
	svc := NewAuthService(stubAuthRepo{}, nil, l, ctx, "::not-a-redis-url")
	require.Nil(t, svc.Cache)

	svc2 := NewAuthService(stubAuthRepo{}, nil, l, ctx, "")
	require.Nil(t, svc2.Cache)
}

func TestAuthService_GetAdminRegistered_cacheHitSkipsDB(t *testing.T) {
	mr := miniredis.RunT(t)
	ctx := context.Background()
	l := logger.NewLogger()
	svc := NewAuthService(stubAuthRepo{}, nil, l, ctx, "redis://"+mr.Addr())
	require.NoError(t, svc.Cache.SetAdminRegistered(ctx, true))

	got, err := svc.GetAdminRegistered(ctx)
	require.NoError(t, err)
	require.True(t, got)
}

func TestAuthService_GetAdminRegistered_cacheMissRequiresDB(t *testing.T) {
	mr := miniredis.RunT(t)
	ctx := context.Background()
	l := logger.NewLogger()
	svc := NewAuthService(stubAuthRepo{}, nil, l, ctx, "redis://"+mr.Addr())

	_, err := svc.GetAdminRegistered(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "database not configured")
}

func TestAuthService_GetAdminRegistered_noCacheRequiresDB(t *testing.T) {
	ctx := context.Background()
	l := logger.NewLogger()
	svc := NewAuthService(stubAuthRepo{}, nil, l, ctx, "")

	_, err := svc.GetAdminRegistered(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "database not configured")
}

func TestAuthService_BuildBootstrap_requiresDB(t *testing.T) {
	ctx := context.Background()
	l := logger.NewLogger()
	svc := NewAuthService(stubAuthRepo{}, nil, l, ctx, "")
	u := &shared_types.User{
		ID:        uuid.New(),
		Name:      "n",
		Email:     "e@x",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := svc.BuildBootstrap(ctx, u, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "database not configured")
}
