package storage

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/testutils"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/require"
)

func TestSSHKeyStorage_promoteUsesTxConnection(t *testing.T) {
	setup := testutils.NewTestSetup()

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	key := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "tx-promote",
		AuthMethod:     "key",
		IsActive:       true,
		IsDefault:      false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err = setup.DB.NewInsert().Model(key).Exec(setup.Ctx)
	require.NoError(t, err)

	tx, err := setup.DB.BeginTx(setup.Ctx, nil)
	require.NoError(t, err)

	st := &SSHKeyStorage{DB: setup.DB, Ctx: setup.Ctx, tx: &tx}
	require.NoError(t, st.PromoteToDefault(key.ID))
	require.NoError(t, tx.Commit())

	stNoTx := &SSHKeyStorage{DB: setup.DB, Ctx: setup.Ctx}
	refreshed, err := stNoTx.GetSSHKeyByID(key.ID)
	require.NoError(t, err)
	require.True(t, refreshed.IsDefault)
}
