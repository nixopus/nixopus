package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/role/types"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *RoleService) CreateRole(role *types.CreateRoleRequest) error {
	existingRole, err := s.storage.GetRoleByName(role.Name)
	if err == nil && existingRole.ID != uuid.Nil {
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
		return types.ErrFailedToCreateRole
	}

	return nil
}
