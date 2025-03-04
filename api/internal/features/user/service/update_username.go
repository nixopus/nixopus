package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
)

// UpdateUsername updates a user's name in the application.
//
// It first checks if the user exists using the provided ID.
// If the user does not exist, it returns ErrUserDoesNotExist.
// If the user exists, it updates the user with the provided details and saves it to the database.
// If the update fails, it returns ErrFailedToUpdateUser.
// Upon successful update, it returns nil.
func (s *UserService) UpdateUsername(id string, req *types.UpdateUserNameRequest) error {
	s.logger.Log(logger.Info, "Updating req", req.Name)
	existingUser, err := s.storage.GetUserById(id)
	if err == nil && existingUser.ID == uuid.Nil {
		s.logger.Log(logger.Error, types.ErrUserDoesNotExist.Error(), "")
		return types.ErrUserDoesNotExist
	}
	if err := s.storage.UpdateUserName(existingUser.ID.String(), req.Name, time.Now()); err != nil {
		s.logger.Log(logger.Error, types.ErrFailedToUpdateUser.Error(), "")
		return types.ErrFailedToUpdateUser
	}

	return nil
}
