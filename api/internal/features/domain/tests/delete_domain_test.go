package tests

import (
	"testing"

	"github.com/google/uuid"
	domainService "github.com/raghavyuva/nixopus-api/internal/features/domain/service"
	domainStorage "github.com/raghavyuva/nixopus-api/internal/features/domain/storage"
	domainTypes "github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestDeleteDomain(t *testing.T) {
	t.Run("should delete domain successfully", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		storage := &domainStorage.DomainStorage{DB: setup.DB, Ctx: setup.Ctx}
		service := domainService.NewDomainsService(setup.Store, setup.Ctx, setup.Logger, storage)

		user, org, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		req := domainTypes.CreateDomainRequest{
			Name:           "test.domain.com",
			OrganizationID: org.ID,
		}

		resp, err := service.CreateDomain(req, user.ID.String())
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.ID)

		err = service.DeleteDomain(resp.ID)
		assert.NoError(t, err)

		domains, err := service.GetDomains(org.ID.String(), user.ID)
		assert.NoError(t, err)
		assert.Len(t, domains, 0)
	})

	t.Run("should return error when deleting domain with invalid ID", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		storage := &domainStorage.DomainStorage{DB: setup.DB, Ctx: setup.Ctx}
		service := domainService.NewDomainsService(setup.Store, setup.Ctx, setup.Logger, storage)

		err := service.DeleteDomain("invalid-uuid")
		assert.Error(t, err)
		assert.ErrorIs(t, err, domainTypes.ErrInvalidDomainID)
	})

	t.Run("should return error when deleting non-existent domain", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		storage := &domainStorage.DomainStorage{DB: setup.DB, Ctx: setup.Ctx}
		service := domainService.NewDomainsService(setup.Store, setup.Ctx, setup.Logger, storage)

		nonExistentID := uuid.New().String()
		err := service.DeleteDomain(nonExistentID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domainTypes.ErrDomainNotFound)
	})
}
