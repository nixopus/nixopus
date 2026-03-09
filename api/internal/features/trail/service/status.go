package service

import (
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/trail/types"
)

// GetStatus retrieves the current status of a trail provision.
//
// Parameters:
//   - userID: the UUID of the requesting user
//   - sessionID: the UUID of the provision session
//
// Returns:
//   - *types.StatusResponse: the status response with progress information
//   - error: domain error if status retrieval fails
func (s *TrailService) GetStatus(userID, sessionID string) (*types.StatusResponse, error) {
	details, err := s.storage.GetUserProvisionDetailsByID(sessionID)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), userID)
		return nil, fmt.Errorf("failed to retrieve status: %w", err)
	}

	if details == nil {
		return nil, types.ErrProvisionNotFound
	}

	if details.UserID.String() != userID {
		return nil, types.ErrProvisionNotFound
	}

	userStatus, err := s.storage.GetUserProvisionStatus(userID)
	if err != nil {
		s.logger.Log(logger.Warning, fmt.Sprintf("Failed to get user provision status: %v", err), userID)
		userStatus = types.UserProvisionStatusNotStarted
	}

	progress, message := s.calculateProgress(details.Step, userStatus, details.Error)

	trailURL := ""
	if details.Subdomain != nil && details.Domain != nil {
		trailURL = fmt.Sprintf("https://%s", *details.Domain)
	}

	stepStr := ""
	if details.Step != nil {
		stepStr = string(*details.Step)
	}

	return &types.StatusResponse{
		SessionID: details.ID.String(),
		Status:    string(userStatus),
		Step:      stepStr,
		Progress:  progress,
		Message:   message,
		Subdomain: getStringValue(details.Subdomain),
		TrailURL:  trailURL,
	}, nil
}

// calculateProgress returns progress percentage and message based on step and status.
func (s *TrailService) calculateProgress(step *types.ProvisionStep, status types.UserProvisionStatus, errorMsg *string) (int, string) {
	if status == types.UserProvisionStatusActive {
		return 100, "Provisioning completed successfully"
	}

	if status == types.UserProvisionStatusFailed {
		msg := "Provisioning failed"
		if errorMsg != nil {
			msg = fmt.Sprintf("Provisioning failed: %s", *errorMsg)
		}
		return 0, msg
	}

	if status == types.UserProvisionStatusNotStarted {
		return 0, "Waiting to start..."
	}

	if step == nil {
		return 5, "Initializing..."
	}

	switch *step {
	case types.ProvisionStepInitializing:
		return 5, "Initializing..."
	case types.ProvisionStepCreatingContainer:
		return 15, "Creating container..."
	case types.ProvisionStepSetupNetworking:
		return 25, "Setting up networking..."
	case types.ProvisionStepInstallingDeps:
		return 45, "Installing dependencies (this may take a few minutes)..."
	case types.ProvisionStepConfiguringSSH:
		return 65, "Configuring SSH..."
	case types.ProvisionStepSetupSSHForwarding:
		return 75, "Setting up SSH forwarding..."
	case types.ProvisionStepVerifyingSSH:
		return 85, "Verifying connection..."
	case types.ProvisionStepCompleted:
		return 100, "Completed"
	default:
		return 50, "Provisioning in progress..."
	}
}

// getStringValue safely extracts string value from pointer.
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
