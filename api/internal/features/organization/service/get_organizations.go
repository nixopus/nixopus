package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetOrganizations fetches all organizations.
//
// It queries the storage layer to fetch all organizations.
// If the storage layer returns an error, it returns ErrFailedToGetOrganizations.
// If the storage layer succeeds in fetching the organizations, it returns the organizations.
func (o *OrganizationService) GetOrganizations() ([]shared_types.Organization, error) {
	organizations, err := o.storage.GetOrganizations()

	if err != nil {
		return nil, types.ErrFailedToGetOrganizations
	}

	return organizations, nil
}