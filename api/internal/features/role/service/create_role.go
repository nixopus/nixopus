package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/role/types"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// CreateRole creates a new role in the application.
//
// It first checks if a role with the same name already exists.
// If it does, it returns ErrRoleAlreadyExists.
// If not, it creates a new role with the provided details and saves it to the database.
// If the creation fails, it returns ErrFailedToCreateRole.
// On successful creation, it returns nil.
func (s *RoleService) CreateRole(role *types.CreateRoleRequest) error {
	s.logger.Log(logger.Info, "Creating role", role.Name)
	existingRole, err := s.storage.GetRoleByName(role.Name)
	if err == nil && existingRole.ID != uuid.Nil {
		s.logger.Log(logger.Error, types.ErrRoleAlreadyExists.Error(), "")
		return types.ErrRoleAlreadyExists
	}

	insertingRole := shared_types.Role{
		ID:          uuid.New(),
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DeletedAt:   nil,
	}

	if err := s.storage.CreateRole(insertingRole); err != nil {
		s.logger.Log(logger.Error, types.ErrFailedToCreateRole.Error(), "")
		return types.ErrFailedToCreateRole
	}

	return nil
}
