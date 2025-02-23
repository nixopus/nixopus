package service

import (
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *RoleService) GetRole(id string) (*shared_types.Role, error) {
	return s.storage.GetRole(id)
}

func (s *RoleService) GetRoles() ([]shared_types.Role, error) {
	return s.storage.GetRoles()
}