package storage_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/ssh/storage"
	"github.com/nixopus/nixopus/api/internal/testutils"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertSSHKey(setup *testutils.TestSetup, key *types.SSHKey) error {
	_, err := setup.DB.NewInsert().Model(key).Exec(setup.Ctx)
	return err
}

func TestGetDefaultSSHKeyByOrganizationID_ActiveDefault(t *testing.T) {
	setup := testutils.NewTestSetup()
	sshStorage := &storage.SSHKeyStorage{DB: setup.DB, Ctx: setup.Ctx}

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)
	require.NotNil(t, org)

	key := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "default-active-key",
		AuthMethod:     "key",
		IsActive:       true,
		IsDefault:      true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, insertSSHKey(setup, key))

	result, err := sshStorage.GetDefaultSSHKeyByOrganizationID(org.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, key.ID, result.ID)
	assert.True(t, result.IsDefault)
	assert.True(t, result.IsActive)
}

func TestGetDefaultSSHKeyByOrganizationID_InactiveDefault(t *testing.T) {
	setup := testutils.NewTestSetup()
	sshStorage := &storage.SSHKeyStorage{DB: setup.DB, Ctx: setup.Ctx}

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)
	require.NotNil(t, org)

	key := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "default-inactive-key",
		AuthMethod:     "key",
		IsActive:       false,
		IsDefault:      true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, insertSSHKey(setup, key))

	result, err := sshStorage.GetDefaultSSHKeyByOrganizationID(org.ID)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, sql.ErrNoRows), "expected sql.ErrNoRows, got: %v", err)
}

func TestGetDefaultSSHKeyByOrganizationID_NoDefault(t *testing.T) {
	setup := testutils.NewTestSetup()
	sshStorage := &storage.SSHKeyStorage{DB: setup.DB, Ctx: setup.Ctx}

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)
	require.NotNil(t, org)

	key := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "non-default-key",
		AuthMethod:     "key",
		IsActive:       true,
		IsDefault:      false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, insertSSHKey(setup, key))

	result, err := sshStorage.GetDefaultSSHKeyByOrganizationID(org.ID)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, sql.ErrNoRows), "expected sql.ErrNoRows, got: %v", err)
}

func sp(s string) *string { return &s }

func TestGetActiveSSHKeyByOrganizationID(t *testing.T) {
	setup := testutils.NewTestSetup()
	st := &storage.SSHKeyStorage{DB: setup.DB, Ctx: setup.Ctx}

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	old := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "older-active",
		AuthMethod:     "key",
		IsActive:       true,
		IsDefault:      false,
		CreatedAt:      time.Now().Add(-time.Hour),
		UpdatedAt:      time.Now().Add(-time.Hour),
	}
	newer := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "newer-active",
		AuthMethod:     "key",
		IsActive:       true,
		IsDefault:      false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	inactive := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "inactive",
		AuthMethod:     "key",
		IsActive:       false,
		IsDefault:      false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, insertSSHKey(setup, old))
	require.NoError(t, insertSSHKey(setup, newer))
	require.NoError(t, insertSSHKey(setup, inactive))

	got, err := st.GetActiveSSHKeyByOrganizationID(org.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, newer.ID, got.ID)

	missingOrg := uuid.New()
	gotNil, err := st.GetActiveSSHKeyByOrganizationID(missingOrg)
	assert.Nil(t, gotNil)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestGetSSHKeyByID(t *testing.T) {
	setup := testutils.NewTestSetup()
	st := &storage.SSHKeyStorage{DB: setup.DB, Ctx: setup.Ctx}

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	host := "10.0.0.1"
	key := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "by-id",
		AuthMethod:     "key",
		IsActive:       true,
		IsDefault:      false,
		Host:           sp(host),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, insertSSHKey(setup, key))

	got, err := st.GetSSHKeyByID(key.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, host, *got.Host)

	wrong := uuid.New()
	_, err = st.GetSSHKeyByID(wrong)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestGetSSHKeyByID_excludesDeleted(t *testing.T) {
	setup := testutils.NewTestSetup()
	st := &storage.SSHKeyStorage{DB: setup.DB, Ctx: setup.Ctx}

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	now := time.Now()
	key := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "deleted-row",
		AuthMethod:     "key",
		IsActive:       true,
		IsDefault:      false,
		CreatedAt:      now,
		UpdatedAt:      now,
		DeletedAt:      &now,
	}
	require.NoError(t, insertSSHKey(setup, key))

	_, err = st.GetSSHKeyByID(key.ID)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestPromoteToDefault(t *testing.T) {
	setup := testutils.NewTestSetup()
	st := &storage.SSHKeyStorage{DB: setup.DB, Ctx: setup.Ctx}

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	a := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "k-a",
		AuthMethod:     "key",
		IsActive:       true,
		IsDefault:      false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	b := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "k-b",
		AuthMethod:     "key",
		IsActive:       true,
		IsDefault:      false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, insertSSHKey(setup, a))
	require.NoError(t, insertSSHKey(setup, b))

	require.NoError(t, st.PromoteToDefault(a.ID))

	refreshed, err := st.GetSSHKeyByID(a.ID)
	require.NoError(t, err)
	assert.True(t, refreshed.IsDefault)
}

func TestListSSHKeysByOrganizationID(t *testing.T) {
	setup := testutils.NewTestSetup()
	st := &storage.SSHKeyStorage{DB: setup.DB, Ctx: setup.Ctx}

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	k1 := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "list-one",
		AuthMethod:     "key",
		IsActive:       true,
		IsDefault:      false,
		CreatedAt:      time.Now().Add(-time.Minute),
		UpdatedAt:      time.Now().Add(-time.Minute),
	}
	k2 := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "list-two",
		AuthMethod:     "key",
		IsActive:       false,
		IsDefault:      false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, insertSSHKey(setup, k1))
	require.NoError(t, insertSSHKey(setup, k2))

	list, err := st.ListSSHKeysByOrganizationID(org.ID)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, k2.ID, list[0].ID)
	assert.Equal(t, k1.ID, list[1].ID)

	empty, err := st.ListSSHKeysByOrganizationID(uuid.New())
	require.NoError(t, err)
	assert.Len(t, empty, 0)
}
