package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// UpdateOrganization updates an existing organization in the database.
//
// It takes an UpdateOrganizationRequest which contains the updated name and description of the organization.
// If the organization does not exist, it returns ErrOrganizationDoesNotExist.
// If there is an error while updating the organization, it returns ErrFailedToUpdateOrganization.
// Upon successful update, it returns nil.
func (o *OrganizationService) UpdateOrganization(organization *types.UpdateOrganizationRequest) error {
	o.logger.Log(logger.Info, "updating organization", organization.ID)
	existingOrganization, err := o.storage.GetOrganization(organization.ID)
	if err == nil && existingOrganization.ID == uuid.Nil {
		o.logger.Log(logger.Error, types.ErrOrganizationDoesNotExist.Error(), "")
		return types.ErrOrganizationDoesNotExist
	}

	organizationToUpdate := shared_types.Organization{
		Name:        organization.Name,
		Description: organization.Description,
		UpdatedAt:   time.Now(),
		CreatedAt:   existingOrganization.CreatedAt,
		DeletedAt:   existingOrganization.DeletedAt,
		ID:          existingOrganization.ID,
	}

	if err := o.storage.UpdateOrganization(&organizationToUpdate); err != nil {
		o.logger.Log(logger.Error, types.ErrFailedToUpdateOrganization.Error(), "")
		return types.ErrFailedToUpdateOrganization
	}

	return nil
}
