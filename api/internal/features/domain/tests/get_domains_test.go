package tests

import (
	"testing"

	"github.com/google/uuid"
	domainService "github.com/raghavyuva/nixopus-api/internal/features/domain/service"
	domainStorage "github.com/raghavyuva/nixopus-api/internal/features/domain/storage"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestGetDomains(t *testing.T) {
	t.Run("should return empty list when no domains exist", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		storage := &domainStorage.DomainStorage{DB: setup.DB, Ctx: setup.Ctx}
		service := domainService.NewDomainsService(setup.Store, setup.Ctx, setup.Logger, storage)

		user, org, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		domains, err := service.GetDomains(org.ID.String(), user.ID)
		assert.NoError(t, err)
		assert.Empty(t, domains)
	})

	t.Run("should return list of domains for organization", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		storage := &domainStorage.DomainStorage{DB: setup.DB, Ctx: setup.Ctx}
		service := domainService.NewDomainsService(setup.Store, setup.Ctx, setup.Logger, storage)

		user, org, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		testDomains := []*shared_types.Domain{
			{
				ID:             uuid.New(),
				Name:           "test-domain-1",
				OrganizationID: org.ID,
				UserID:         user.ID,
			},
			{
				ID:             uuid.New(),
				Name:           "test-domain-2",
				OrganizationID: org.ID,
				UserID:         user.ID,
			},
		}

		for _, domain := range testDomains {
			err := storage.CreateDomain(domain)
			assert.NoError(t, err)
		}

		domains, err := service.GetDomains(org.ID.String(), user.ID)
		assert.NoError(t, err)
		assert.Len(t, domains, 2)

		for i, domain := range domains {
			assert.Equal(t, testDomains[i].ID, domain.ID)
			assert.Equal(t, testDomains[i].Name, domain.Name)
			assert.Equal(t, testDomains[i].OrganizationID, domain.OrganizationID)
			assert.Equal(t, testDomains[i].UserID, domain.UserID)
		}
	})

	t.Run("should return only domains for specified organization", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		storage := &domainStorage.DomainStorage{DB: setup.DB, Ctx: setup.Ctx}
		service := domainService.NewDomainsService(setup.Store, setup.Ctx, setup.Logger, storage)

		user, org1, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		org2 := &shared_types.Organization{
			ID:          uuid.New(),
			Name:        "test-org-2",
			Description: "Test organization 2",
		}

		err = setup.OrgStorage.CreateOrganization(*org2)
		assert.NoError(t, err)

		orgUser := &shared_types.OrganizationUsers{
			ID:             uuid.New(),
			UserID:         user.ID,
			OrganizationID: org2.ID,
		}

		err = setup.OrgStorage.AddUserToOrganization(*orgUser)
		assert.NoError(t, err)

		org1Domain := &shared_types.Domain{
			ID:             uuid.New(),
			Name:           "test-domain-org1",
			OrganizationID: org1.ID,
			UserID:         user.ID,
		}
		err = storage.CreateDomain(org1Domain)
		assert.NoError(t, err)

		org2Domain := &shared_types.Domain{
			ID:             uuid.New(),
			Name:           "test-domain-org2",
			OrganizationID: org2.ID,
			UserID:         user.ID,
		}
		err = storage.CreateDomain(org2Domain)
		assert.NoError(t, err)

		domains, err := service.GetDomains(org1.ID.String(), user.ID)
		assert.NoError(t, err)
		assert.Len(t, domains, 1)
		assert.Equal(t, org1Domain.ID, domains[0].ID)
		assert.Equal(t, org1Domain.OrganizationID, domains[0].OrganizationID)

		domains, err = service.GetDomains(org2.ID.String(), user.ID)
		assert.NoError(t, err)
		assert.Len(t, domains, 1)
		assert.Equal(t, org2Domain.ID, domains[0].ID)
		assert.Equal(t, org2Domain.OrganizationID, domains[0].OrganizationID)
	})
}
