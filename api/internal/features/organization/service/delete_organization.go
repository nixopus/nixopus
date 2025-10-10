package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
)

// DeleteOrganization deletes an organization by its ID.
//
// It first checks if the organization exists and if not, returns ErrOrganizationDoesNotExist.
// If the organization exists, it calls the storage layer's DeleteOrganization method to delete the organization.
// If the storage layer returns an error, it returns ErrFailedToDeleteOrganization.
// If the storage layer succeeds in deleting the organization, it returns nil.
func (o *OrganizationService) DeleteOrganization(organizationID uuid.UUID) error {
	o.logger.Log(logger.Info, "deleting organization", organizationID.String())
	existingOrganization, err := o.storage.GetOrganization(organizationID.String())
	if err == nil && existingOrganization.ID == uuid.Nil {
		o.logger.Log(logger.Error, types.ErrOrganizationDoesNotExist.Error(), "")
		return types.ErrOrganizationDoesNotExist
	}

	tx, err := o.storage.BeginTx()
	if err != nil {
		o.logger.Log(logger.Error, "failed to begin transaction", err.Error())
		return types.ErrFailedToDeleteOrganization
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	storage := o.storage.WithTx(tx)

	if err := storage.DeleteOrganization(organizationID.String()); err != nil {
		o.logger.Log(logger.Error, types.ErrFailedToDeleteOrganization.Error(), err.Error())
		return types.ErrFailedToDeleteOrganization
	}

	if err := tx.Commit(); err != nil {
		o.logger.Log(logger.Error, "failed to commit transaction", err.Error())
		return types.ErrFailedToDeleteOrganization
	}

	//TODO: deleting organization should send a notification to all the users in the organization

	return nil
}
