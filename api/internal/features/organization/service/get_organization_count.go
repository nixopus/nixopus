package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
)

// GetOrganizationCount returns the count of organizations for a given user.
//
// It queries the storage layer to get the count of organizations for the given user ID.
// If the storage layer returns an error, it returns 0 and the error.
// If the storage layer succeeds, it returns the count of organizations.
func (o *OrganizationService) GetOrganizationCount(userID string) (int, error) {
	o.logger.Log(logger.Info, "getting organization count for user", userID)

	organizations, err := o.storage.GetOrganizations()
	if err != nil {
		o.logger.Log(logger.Error, types.ErrFailedToGetOrganizations.Error(), err.Error())
		return 0, types.ErrFailedToGetOrganizations
	}

	return len(organizations), nil
}
