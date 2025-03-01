package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// CreateOrganization creates a new organization in the database.
//
// It takes a CreateOrganizationRequest which contains the name and description of the organization.
//
// If an organization with the same name already exists, it returns ErrOrganizationAlreadyExists.
// If there is an error while creating the organization, it returns ErrFailedToCreateOrganization.
func (o *OrganizationService) CreateOrganization(organization *types.CreateOrganizationRequest) (shared_types.Organization, error) {
	o.logger.Log(logger.Info, "creating organization", organization.Name)
	existingOrganization, err := o.storage.GetOrganizationByName(organization.Name)
	if err == nil && existingOrganization.ID != uuid.Nil {
		o.logger.Log(logger.Error, types.ErrOrganizationAlreadyExists.Error(), "")
		return shared_types.Organization{}, types.ErrOrganizationAlreadyExists
	}

	organizationToCreate := types.NewOrganization(organization.Name, organization.Description)

	if err := o.storage.CreateOrganization(organizationToCreate); err != nil {
		o.logger.Log(logger.Error, types.ErrFailedToCreateOrganization.Error(), err.Error())
		return shared_types.Organization{}, types.ErrFailedToCreateOrganization
	}

	return organizationToCreate, nil
}
