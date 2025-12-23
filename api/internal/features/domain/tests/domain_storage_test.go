package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/storage"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestDomainStorage(t *testing.T) {
	setup := testutils.NewTestSetup()
	domainStorage := &storage.DomainStorage{
		DB:  setup.DB,
		Ctx: setup.Ctx,
	}

	testUser, testOrg, err := setup.CreateTestUserAndOrg()
	assert.NoError(t, err)

	t.Run("CreateDomain", func(t *testing.T) {
		domain := &shared_types.Domain{
			ID:             uuid.New(),
			Name:           "test.com",
			UserID:         testUser.ID,
			OrganizationID: testOrg.ID,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err := domainStorage.CreateDomain(domain)
		assert.NoError(t, err)

		foundDomain, err := domainStorage.GetDomain(domain.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, domain.Name, foundDomain.Name)
		assert.Equal(t, domain.UserID, foundDomain.UserID)
	})

	t.Run("GetDomain", func(t *testing.T) {
		domain := &shared_types.Domain{
			ID:             uuid.New(),
			Name:           "getdomain.com",
			UserID:         testUser.ID,
			OrganizationID: testOrg.ID,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err := domainStorage.CreateDomain(domain)
		assert.NoError(t, err)

		foundDomain, err := domainStorage.GetDomain(domain.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, domain.Name, foundDomain.Name)
		assert.Equal(t, domain.UserID, foundDomain.UserID)
	})

	t.Run("UpdateDomain", func(t *testing.T) {
		domain := &shared_types.Domain{
			ID:             uuid.New(),
			Name:           "update.com",
			UserID:         testUser.ID,
			OrganizationID: testOrg.ID,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err := domainStorage.CreateDomain(domain)
		assert.NoError(t, err)

		err = domainStorage.UpdateDomain(domain.ID.String(), "updated.com")
		assert.NoError(t, err)

		updatedDomain, err := domainStorage.GetDomain(domain.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, "updated.com", updatedDomain.Name)
	})

	t.Run("DeleteDomain", func(t *testing.T) {
		domain := &shared_types.Domain{
			ID:             uuid.New(),
			Name:           "delete.com",
			UserID:         testUser.ID,
			OrganizationID: testOrg.ID,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err := domainStorage.CreateDomain(domain)
		assert.NoError(t, err)

		err = domainStorage.DeleteDomain(domain)
		assert.NoError(t, err)

		_, err = domainStorage.GetDomain(domain.ID.String())
		assert.Error(t, err)
	})

	t.Run("GetDomains", func(t *testing.T) {
		domains := []*shared_types.Domain{
			{
				ID:             uuid.New(),
				Name:           "domain1.com",
				UserID:         testUser.ID,
				OrganizationID: testOrg.ID,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			{
				ID:             uuid.New(),
				Name:           "domain2.com",
				UserID:         testUser.ID,
				OrganizationID: testOrg.ID,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
		}

		for _, domain := range domains {
			err := domainStorage.CreateDomain(domain)
			assert.NoError(t, err)
		}

		foundDomains, err := domainStorage.GetDomains(testOrg.ID.String(), testUser.ID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(foundDomains), 2)
	})

	t.Run("GetDomainByName", func(t *testing.T) {
		domain := &shared_types.Domain{
			ID:             uuid.New(),
			Name:           "byname.com",
			UserID:         testUser.ID,
			OrganizationID: testOrg.ID,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err := domainStorage.CreateDomain(domain)
		assert.NoError(t, err)

		foundDomain, err := domainStorage.GetDomainByName(domain.Name, testOrg.ID)
		assert.NoError(t, err)
		assert.Equal(t, domain.Name, foundDomain.Name)
		assert.Equal(t, domain.UserID, foundDomain.UserID)
	})

	t.Run("IsDomainExists", func(t *testing.T) {
		domain := &shared_types.Domain{
			ID:             uuid.New(),
			Name:           "exists.com",
			UserID:         testUser.ID,
			OrganizationID: testOrg.ID,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err := domainStorage.CreateDomain(domain)
		assert.NoError(t, err)

		exists, err := domainStorage.IsDomainExists(domain.ID.String())
		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = domainStorage.IsDomainExists(uuid.New().String())
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("GetDomainOwnerByID", func(t *testing.T) {
		domain := &shared_types.Domain{
			ID:             uuid.New(),
			Name:           "owner.com",
			UserID:         testUser.ID,
			OrganizationID: testOrg.ID,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err := domainStorage.CreateDomain(domain)
		assert.NoError(t, err)

		ownerID, err := domainStorage.GetDomainOwnerByID(domain.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, testUser.ID.String(), ownerID)
	})

	t.Run("Transaction", func(t *testing.T) {
		domain := &shared_types.Domain{
			ID:             uuid.New(),
			Name:           "transaction.com",
			UserID:         testUser.ID,
			OrganizationID: testOrg.ID,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		tx, err := domainStorage.BeginTx()
		assert.NoError(t, err)

		domainStorageWithTx := domainStorage.WithTx(tx)

		err = domainStorageWithTx.CreateDomain(domain)
		assert.NoError(t, err)

		err = tx.Rollback()
		assert.NoError(t, err)

		_, err = domainStorage.GetDomain(domain.ID.String())
		assert.Error(t, err)
	})
}
