package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
)

func (s *UserService) UpdateAvatar(ctx context.Context, userID string, req *types.UpdateAvatarRequest) error {
	if req == nil {
		s.logger.Log(logger.Error, "invalid request type", "request is nil")
		return types.ErrInvalidRequestType
	}

	s.logger.Log(logger.Info, "Updating avatar for user", userID)
	if err := s.storage.UpdateUserAvatar(ctx, userID, req.AvatarData); err != nil {
		s.logger.Log(logger.Error, "failed to update avatar", err.Error())
		return err
	}

	return nil
}
