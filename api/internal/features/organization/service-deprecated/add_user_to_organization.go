package service_deprecated

import (
	// "time"

	// "github.com/google/uuid"
	// "github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	// shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

// AddUserToOrganization adds a user to an organization.
//
// It first checks if the organization and role exist using the IDs from the request.
// If the organization does not exist, it returns ErrOrganizationDoesNotExist.
// If the role does not exist, it returns ErrRoleDoesNotExist.
// It then checks if the user exists using the user ID from the request.
// If the user does not exist, it returns ErrUserDoesNotExist.
// It also checks if the user is already part of the organization using both IDs.
// If the user is already in the organization, it returns ErrUserAlreadyInOrganization.
// If all checks pass, it calls the storage layer's AddUserToOrganization method to add the user to the organization.
// If the addition fails, it returns ErrFailedToAddUserToOrganization.
// Upon successful addition, it returns nil.
func (o *OrganizationService) AddUserToOrganization(request types.AddUserToOrganizationRequest, tx ...bun.Tx) error {
	// o.logger.Log(logger.Info, "adding user to organization", request.UserID)
	// roleId, err := uuid.Parse(request.RoleId)
	// if err != nil {
	// 	o.logger.Log(logger.Error, types.ErrInvalidRoleID.Error(), err.Error())
	// 	return types.ErrInvalidRoleID
	// }

	// var dbTx bun.Tx
	// var shouldCommit bool

	// if len(tx) == 0 {
	// 	dbTx, err = o.storage.BeginTx()
	// 	if err != nil {
	// 		o.logger.Log(logger.Error, "failed to begin transaction", err.Error())
	// 		return types.ErrInternalServer
	// 	}
	// 	shouldCommit = true
	// } else {
	// 	dbTx = tx[0]
	// }

	// storageWithTx := o.storage.WithTx(dbTx)
	// userStorageWithTx := o.user_storage.WithTx(dbTx)
	// roleStorageWithTx := o.role_storage.WithTx(dbTx)

	// existingOrganization, err := storageWithTx.GetOrganization(request.OrganizationID)
	// if err != nil {
	// 	o.logger.Log(logger.Error, types.ErrOrganizationDoesNotExist.Error(), err.Error())
	// 	if shouldCommit {
	// 		dbTx.Rollback()
	// 	}
	// 	return err
	// }

	// if existingOrganization.ID == uuid.Nil {
	// 	o.logger.Log(logger.Error, types.ErrOrganizationDoesNotExist.Error(), "")
	// 	if shouldCommit {
	// 		dbTx.Rollback()
	// 	}
	// 	return types.ErrOrganizationDoesNotExist
	// }

	// existingUser, err := userStorageWithTx.FindUserByID(request.UserID)
	// if err != nil {
	// 	o.logger.Log(logger.Error, types.ErrUserDoesNotExist.Error(), err.Error())
	// 	if shouldCommit {
	// 		dbTx.Rollback()
	// 	}
	// 	return err
	// }

	// if existingUser.ID == uuid.Nil {
	// 	o.logger.Log(logger.Error, types.ErrUserDoesNotExist.Error(), "")
	// 	if shouldCommit {
	// 		dbTx.Rollback()
	// 	}
	// 	return types.ErrUserDoesNotExist
	// }

	// existingRole, err := roleStorageWithTx.GetRole(roleId.String())
	// if err != nil {
	// 	o.logger.Log(logger.Error, types.ErrRoleDoesNotExist.Error(), err.Error())
	// 	if shouldCommit {
	// 		dbTx.Rollback()
	// 	}
	// 	return err
	// }
	// if existingRole.ID == uuid.Nil {
	// 	o.logger.Log(logger.Error, types.ErrRoleDoesNotExist.Error(), "")
	// 	if shouldCommit {
	// 		dbTx.Rollback()
	// 	}
	// 	return types.ErrRoleDoesNotExist
	// }

	// existingUserInOrganization, err := storageWithTx.FindUserInOrganization(request.UserID, request.OrganizationID)
	// if err != nil {
	// 	o.logger.Log(logger.Error, types.ErrFailedToAddUserToOrganization.Error(), err.Error())
	// 	if shouldCommit {
	// 		dbTx.Rollback()
	// 	}
	// 	return err
	// }
	// if existingUserInOrganization.ID != uuid.Nil {
	// 	o.logger.Log(logger.Error, types.ErrUserAlreadyInOrganization.Error(), "")
	// 	if shouldCommit {
	// 		dbTx.Rollback()
	// 	}
	// 	return types.ErrUserAlreadyInOrganization
	// }

	// organizationUser := shared_types.OrganizationUsers{
	// 	UserID:         existingUser.ID,
	// 	OrganizationID: existingOrganization.ID,
	// 	RoleID:         roleId,
	// 	CreatedAt:      time.Now(),
	// 	UpdatedAt:      time.Now(),
	// 	DeletedAt:      nil,
	// 	ID:             uuid.New(),
	// }

	// if err := storageWithTx.AddUserToOrganization(organizationUser); err != nil {
	// 	o.logger.Log(logger.Error, types.ErrFailedToAddUserToOrganization.Error(), err.Error())
	// 	if shouldCommit {
	// 		dbTx.Rollback()
	// 	}
	// 	return types.ErrFailedToAddUserToOrganization
	// }

	// // Invalidate cache for organization membership
	// if err := o.cache.InvalidateOrgMembership(o.Ctx, request.UserID, request.OrganizationID); err != nil {
	// 	o.logger.Log(logger.Error, "failed to invalidate organization membership cache", err.Error())
	// }

	// if shouldCommit {
	// 	if err := dbTx.Commit(); err != nil {
	// 		o.logger.Log(logger.Error, "failed to commit transaction", err.Error())
	// 		return types.ErrInternalServer
	// 	}
	// }

	return nil
}
