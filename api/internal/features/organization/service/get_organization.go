package service

import (
	"database/sql"

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
	existingOrganization, err := o.storage.GetOrganization(id)
	if err == nil && existingOrganization.ID != uuid.Nil {
		return *existingOrganization, nil
	}

	return shared_types.Organization{}, sql.ErrNoRows
}