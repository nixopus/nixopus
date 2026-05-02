package auth

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/auth/storage"
	"github.com/nixopus/nixopus/api/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestAuthUserStorage_FindUserByEmail(t *testing.T) {
	setup := testutils.NewTestSetup()
	us := &storage.UserStorage{DB: setup.DB, Ctx: setup.Ctx}

	user, _, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	got, err := us.FindUserByEmail(user.Email)
	require.NoError(t, err)
	require.Equal(t, user.ID, got.ID)
	require.Equal(t, user.Email, got.Email)

	_, err = us.FindUserByEmail("no-one-" + uuid.New().String() + "@example.test")
	require.Error(t, err)
}

func TestAuthUserStorage_FindUserByEmail_usesTransactionalRepo(t *testing.T) {
	setup := testutils.NewTestSetup()
	base := &storage.UserStorage{DB: setup.DB, Ctx: setup.Ctx}

	user, _, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	tx, err := base.BeginTx()
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback() })

	scoped := base.WithTx(tx)
	got, err := scoped.FindUserByEmail(user.Email)
	require.NoError(t, err)
	require.Equal(t, user.ID, got.ID)

	require.NoError(t, tx.Rollback())

	_, err = scoped.FindUserByEmail(user.Email)
	require.Error(t, err)
}

func TestAuthUserStorage_BeginTx(t *testing.T) {
	setup := testutils.NewTestSetup()
	us := &storage.UserStorage{DB: setup.DB, Ctx: setup.Ctx}

	tx, err := us.BeginTx()
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())
}
