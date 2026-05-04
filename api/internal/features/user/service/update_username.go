package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/user/types"
)

// UpdateUsername updates the authenticated user's display name (stored as name on the user row).
func (s *UserService) UpdateUsername(id string, req *types.UpdateUserNameRequest) error {
	data := fmt.Sprintf("user_id=%s", id)
	if req == nil {
		s.logger.Log(logger.Error, "user service: UpdateUsername: nil request", data)
		return types.ErrInvalidRequestType
	}

	existingUser, err := s.storage.GetUserById(id)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("user service: UpdateUsername: get user: %v", err), data)
		return err
	}

	if existingUser.ID == uuid.Nil {
		s.logger.Log(logger.Error, fmt.Sprintf("user service: UpdateUsername: %v", types.ErrUserDoesNotExist), data)
		return types.ErrUserDoesNotExist
	}

	if err := s.storage.UpdateUserName(existingUser.ID.String(), req.Name, time.Now()); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("user service: UpdateUsername: %v", types.ErrFailedToUpdateUser), data)
		return types.ErrFailedToUpdateUser
	}

	return nil
}
