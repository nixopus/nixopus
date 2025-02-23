package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
)

// DeleteOrganization deletes an organization by its ID.
//
// It first checks if the organization exists and if not, returns ErrOrganizationDoesNotExist.
// If the organization exists, it calls the storage layer's DeleteOrganization method to delete the organization.
// If the storage layer returns an error, it returns ErrFailedToDeleteOrganization.
// If the storage layer succeeds in deleting the organization, it returns nil.
func (o *OrganizationService) DeleteOrganization(organizationID uuid.UUID) error {
	existingOrganization, err := o.storage.GetOrganization(organizationID.String())
	if err == nil && existingOrganization.ID == uuid.Nil {
		return types.ErrOrganizationDoesNotExist
	}

	if err := o.storage.DeleteOrganization(organizationID.String()); err != nil {
		return types.ErrFailedToDeleteOrganization
	}

	return nil
}
