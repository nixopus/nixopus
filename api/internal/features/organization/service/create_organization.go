package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
)

// CreateOrganization creates a new organization in the database.
//
// It takes a CreateOrganizationRequest which contains the name and description of the organization.
//
// If an organization with the same name already exists, it returns ErrOrganizationAlreadyExists.
// If there is an error while creating the organization, it returns ErrFailedToCreateOrganization.
func (o *OrganizationService) CreateOrganization(organization *types.CreateOrganizationRequest) error {
	existingOrganization, err := o.storage.GetOrganizationByName(organization.Name)
	if err == nil && existingOrganization.ID != uuid.Nil {
		return types.ErrOrganizationAlreadyExists
	}

	organizationToCreate := types.NewOrganization(organization.Name, organization.Description)

	if err := o.storage.CreateOrganization(organizationToCreate); err != nil {
		return types.ErrFailedToCreateOrganization
	}

	return nil
}
