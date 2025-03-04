package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/role/types"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// UpdateRole updates a role in the application.
//
// It first checks if the role exists using the provided ID.
// If the role does not exist, it returns ErrRoleDoesNotExist.
// If the role exists, it updates the role with the provided details and saves it to the database.
// If the update fails, it returns ErrFailedToUpdateRole.
// Upon successful update, it returns nil.
func (s *RoleService) UpdateRole(id string, role types.UpdateRoleRequest) error {
	s.logger.Log(logger.Info, "Updating role", role.Name)
	existingRole, err := s.storage.GetRole(role.ID)
	if err == nil && existingRole.ID == uuid.Nil {
		s.logger.Log(logger.Error, types.ErrRoleDoesNotExist.Error(), "")
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
		s.logger.Log(logger.Error, types.ErrFailedToUpdateRole.Error(), "")
		return types.ErrFailedToUpdateRole
	}

	return nil
}
