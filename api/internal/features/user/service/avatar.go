package service

import (
	"context"
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/user/types"
)

func (s *UserService) UpdateAvatar(ctx context.Context, userID string, req *types.UpdateAvatarRequest) error {
	ctxStr := fmt.Sprintf("user_id=%s", userID)
	if req == nil {
		s.logger.Log(logger.Error, "user service: UpdateAvatar: nil request", ctxStr)
		return types.ErrInvalidRequestType
	}

	payload := fmt.Sprintf("%s payload_len=%d", ctxStr, len(req.AvatarData))
	s.logger.Log(logger.Info, "user service: UpdateAvatar", payload)
	if err := s.storage.UpdateUserAvatar(ctx, userID, req.AvatarData); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("user service: UpdateAvatar: %v", err), ctxStr)
		return err
	}

	return nil
}
