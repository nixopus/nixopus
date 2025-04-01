package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	domainService "github.com/raghavyuva/nixopus-api/internal/features/domain/service"
	domainStorage "github.com/raghavyuva/nixopus-api/internal/features/domain/storage"
	domainTypes "github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestCreateDomain(t *testing.T) {
	t.Run("should create domain successfully", func(t *testing.T) {
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

		// Verify domain was created
		domains, err := service.GetDomains(org.ID.String(), user.ID)
		assert.NoError(t, err)
		assert.Len(t, domains, 1)
		assert.Equal(t, req.Name, domains[0].Name)
		assert.Equal(t, org.ID, domains[0].OrganizationID)
		assert.Equal(t, user.ID, domains[0].UserID)
	})

	t.Run("should not create domain with duplicate name in same organization", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		storage := &domainStorage.DomainStorage{DB: setup.DB, Ctx: setup.Ctx}
		service := domainService.NewDomainsService(setup.Store, setup.Ctx, setup.Logger, storage)

		user, org, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		req := domainTypes.CreateDomainRequest{
			Name:           "test.domain.com",
			OrganizationID: org.ID,
		}

		// Create first domain
		resp, err := service.CreateDomain(req, user.ID.String())
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.ID)

		// Try to create domain with same name
		_, err = service.CreateDomain(req, user.ID.String())
		assert.Error(t, err)
		assert.ErrorIs(t, err, domainTypes.ErrDomainAlreadyExists)
	})

	t.Run("should not create domain with invalid organization ID", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		storage := &domainStorage.DomainStorage{DB: setup.DB, Ctx: setup.Ctx}
		service := domainService.NewDomainsService(setup.Store, setup.Ctx, setup.Logger, storage)

		user, _, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		req := domainTypes.CreateDomainRequest{
			Name:           "test.domain.com",
			OrganizationID: uuid.New(), // Random non-existent org ID
		}

		_, err = service.CreateDomain(req, user.ID.String())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "organization not found")
	})

	t.Run("should not create domain with invalid name format", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		storage := &domainStorage.DomainStorage{DB: setup.DB, Ctx: setup.Ctx}
		service := domainService.NewDomainsService(setup.Store, setup.Ctx, setup.Logger, storage)

		user, org, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		invalidNames := []string{
			"",                                // Empty name
			"test",                            // No TLD
			"t.c",                             // TLD too short
			"test." + strings.Repeat("x", 64), // TLD too long
			"a",                               // Name too short
			strings.Repeat("x", 256) + ".com", // Name too long
		}

		for _, name := range invalidNames {
			req := domainTypes.CreateDomainRequest{
				Name:           name,
				OrganizationID: org.ID,
			}

			_, err = service.CreateDomain(req, user.ID.String())
			assert.Error(t, err, "Expected error for invalid name: %s", name)
		}
	})

	t.Run("should not create domain that does not belong to server", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		storage := &domainStorage.DomainStorage{DB: setup.DB, Ctx: setup.Ctx}
		service := domainService.NewDomainsService(setup.Store, setup.Ctx, setup.Logger, storage)

		user, org, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		serverHost := os.Getenv("SSH_HOST")
		if serverHost == "" {
			serverHost, err = os.Hostname()
			assert.NoError(t, err)
		}

		req := domainTypes.CreateDomainRequest{
			Name:           "example.com",
			OrganizationID: org.ID,
		}

		_, err = service.CreateDomain(req, user.ID.String())
		assert.Error(t, err)
		assert.ErrorIs(t, err, domainTypes.ErrDomainDoesNotBelongToServer)
	})

	t.Run("should create domain that belongs to server", func(t *testing.T) {
		setup := testutils.NewTestSetup()
		storage := &domainStorage.DomainStorage{DB: setup.DB, Ctx: setup.Ctx}
		service := domainService.NewDomainsService(setup.Store, setup.Ctx, setup.Logger, storage)

		user, org, err := setup.CreateTestUserAndOrg()
		assert.NoError(t, err)

		serverHost := os.Getenv("SSH_HOST")
		if serverHost == "" {
			serverHost, err = os.Hostname()
			assert.NoError(t, err)
		}

		req := domainTypes.CreateDomainRequest{
			Name:           "test." + serverHost,
			OrganizationID: org.ID,
		}

		_, err = service.CreateDomain(req, user.ID.String())
		assert.NoError(t, err)
	})
}
