package service

import (
	"database/sql"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"

	"github.com/google/uuid"
)

// GetOrganization retrieves an organization by its ID.
//
// It queries the storage layer to find the organization with the given ID.
// If the organization exists, it returns the organization and a nil error.
// If the organization does not exist or an error occurs, it returns an empty organization
// and the sql.ErrNoRows error.
func (o *OrganizationService) GetOrganization(id string) (shared_types.Organization, error) {
	o.logger.Log(logger.Info, "getting organization", id)
	existingOrganization, err := o.storage.GetOrganization(id)
	if err == nil && existingOrganization.ID != uuid.Nil {
		o.logger.Log(logger.Info, "organization found", id)
		return *existingOrganization, nil
	}

	return shared_types.Organization{}, sql.ErrNoRows
}
