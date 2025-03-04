package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
)

// DeletePermission deletes a permission by its ID.
//
// It first checks if the permission exists using the provided ID.
// If the permission does not exist, it returns ErrPermissionDoesNotExist.
// If the permission exists, it calls the storage layer's DeletePermission method to delete it.
// If the deletion fails, it returns ErrFailedToDeletePermission.
// Upon successful deletion, it returns nil.
func (s *PermissionService) DeletePermission(permissionID string) error {
	s.logger.Log(logger.Info, "Deleting permission", "")
	existingPermission, err := s.storage.GetPermission(permissionID)
	if err == nil && existingPermission.ID == uuid.Nil {
		s.logger.Log(logger.Error, types.ErrPermissionDoesNotExist.Error(), "")
		return types.ErrPermissionDoesNotExist
	}

	if err := s.storage.DeletePermission(permissionID); err != nil {
		s.logger.Log(logger.Error, types.ErrFailedToDeletePermission.Error(), err.Error())
		return types.ErrFailedToDeletePermission
	}

	return nil
}
