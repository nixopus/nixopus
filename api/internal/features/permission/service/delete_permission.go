package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/permission/types"
)

func (s *PermissionService) DeletePermission(permissionID string) error {
	existingPermission, err := s.storage.GetPermission(permissionID)
	if err == nil && existingPermission.ID == uuid.Nil {
		return types.ErrPermissionDoesNotExist
	}

	if err := s.storage.DeletePermission(permissionID); err != nil {
		return types.ErrFailedToDeletePermission
	}

	return nil
}
