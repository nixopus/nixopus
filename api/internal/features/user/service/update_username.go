package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
)

// Existing function that needs to be updated
func (s *UserService) UpdateUsername(id string, req *types.UpdateUserNameRequest) error {
	if req == nil {
		s.logger.Log(logger.Error, "invalid request type", "request is nil")
		return types.ErrInvalidRequestType
	}

	s.logger.Log(logger.Info, "Updating req", req.Name)
	existingUser, err := s.storage.GetUserById(id)
	if err != nil {
		s.logger.Log(logger.Error, "error fetching user", err.Error())
		return err
	}

	if existingUser.ID == uuid.Nil {
		s.logger.Log(logger.Error, types.ErrUserDoesNotExist.Error(), "")
		return types.ErrUserDoesNotExist
	}

	if err := s.storage.UpdateUserName(existingUser.ID.String(), req.Name, time.Now()); err != nil {
		s.logger.Log(logger.Error, types.ErrFailedToUpdateUser.Error(), "")
		return types.ErrFailedToUpdateUser
	}

	return nil
}
