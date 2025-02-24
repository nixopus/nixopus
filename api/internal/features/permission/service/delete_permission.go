package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
)

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
