package service

import (
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetRole retrieves a role by its ID from the storage.
// It returns the role and nil if found, or an error if the operation fails.
func (s *RoleService) GetRole(id string) (*shared_types.Role, error) {
	return s.storage.GetRole(id)
}

// GetRoles retrieves all roles from the storage.
// It returns a slice of roles and nil if the operation is successful, or an error if the operation fails.
func (s *RoleService) GetRoles() ([]shared_types.Role, error) {
	return s.storage.GetRoles()
}
