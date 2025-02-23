package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/role/types"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)


func (s *RoleService) UpdateRole(id string, role types.UpdateRoleRequest) error {
	existingRole, err := s.storage.GetRole(role.ID)
	if err == nil && existingRole.ID == uuid.Nil {
		return types.ErrRoleDoesNotExist
	}

	updatingRole := shared_types.Role{
		ID:          existingRole.ID,
		Name:        role.Name,
		Description: role.Description,
		UpdatedAt:   time.Now(),
		DeletedAt:   existingRole.DeletedAt,
	}

	if err := s.storage.UpdateRole(&updatingRole); err != nil {
		return types.ErrFailedToUpdateRole
	}

	return nil
}