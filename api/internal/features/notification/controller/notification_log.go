package controller

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *NotificationController) logNotificationDebug(handler, reason, data string) {
	c.logger.Log(logger.Debug, fmt.Sprintf("notification: %s: %s", handler, reason), data)
}

func notificationRequestLogData(r *http.Request, user *types.User) string {
	orgID := utils.GetOrganizationID(r)
	if user != nil && orgID != uuid.Nil {
		return fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID)
	}
	if user != nil {
		return fmt.Sprintf("user_id=%s", user.ID)
	}
	if orgID != uuid.Nil {
		return fmt.Sprintf("org_id=%s", orgID)
	}
	return ""
}
