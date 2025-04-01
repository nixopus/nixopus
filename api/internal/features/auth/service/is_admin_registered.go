package service

import "github.com/raghavyuva/nixopus-api/internal/types"

// IsAdminRegistered checks if an admin user is registered in the database.
//
// The function queries the database to find a user with the type "admin".
// If a user is found, it returns true. Otherwise, it returns false.
//
// Returns:
// - bool: true if an admin user is registered, false otherwise
func (s *AuthService) IsAdminRegistered() (bool, error) {
	user, err := s.storage.FindUserByType(types.UserTypeAdmin)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, nil
	}
	return true, nil
}
