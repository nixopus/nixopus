package service

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/trail/types"
)

// ProvisionTrail handles the business logic for provisioning a new trail instance.
//
// Parameters:
//   - userID: the UUID of the requesting user
//   - orgID: the UUID of the organization
//   - req: the provision request containing optional image selection
//
// Returns:
//   - *types.ProvisionResponse: the provision response with session ID
//   - error: domain error if provisioning fails
func (s *TrailService) ProvisionTrail(userID, orgID string, req types.ProvisionRequest) (*types.ProvisionResponse, error) {
	image := req.Image
	if image == "" {
		image = s.config.DefaultImage
	}

	if !s.IsImageAllowed(image) {
		s.logger.Log(logger.Warning, fmt.Sprintf("User %s requested disallowed image: %s", userID, image), userID)
		return nil, types.ErrImageNotAllowed
	}

	activeProvision, err := s.storage.GetActiveProvisionByUserAndOrg(userID, orgID)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), userID)
		return nil, fmt.Errorf("failed to check active provisions: %w", err)
	}

	if activeProvision != nil {
		return nil, types.ErrActiveProvisionExists
	}

	count, err := s.storage.CountActiveProvisions()
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, fmt.Errorf("failed to check system capacity: %w", err)
	}

	if count >= s.config.MaxConcurrentTrails {
		s.logger.Log(logger.Warning, fmt.Sprintf("Max concurrent trails reached (%d/%d)", count, s.config.MaxConcurrentTrails), "")
		return nil, types.ErrSystemAtCapacity
	}

	subdomain, err := s.GenerateSubdomain()
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), userID)
		return nil, fmt.Errorf("failed to generate subdomain: %w", err)
	}

	user, err := s.storage.GetUserByID(userID)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), userID)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	displayName := ""
	if user != nil {
		if user.Name != "" {
			displayName = user.Name
		} else {
			displayName = user.Email
		}
	}

	containerName := s.GenerateContainerName(displayName)

	fullDomain := fmt.Sprintf("%s.%s", subdomain, s.config.TrailDomain)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	initialStep := types.ProvisionStepInitializing
	provisionDetails := &types.UserProvisionDetails{
		UserID:           userUUID,
		OrganizationID:   orgUUID,
		LXDContainerName: &containerName,
		Subdomain:        &subdomain,
		Domain:           &fullDomain,
		Step:             &initialStep,
	}

	if err := s.storage.CreateActiveUserProvision(provisionDetails); err != nil {
		if strings.Contains(err.Error(), "active_provision_per_user_org") || strings.Contains(err.Error(), "duplicate") {
			return nil, types.ErrActiveProvisionExists
		}
		s.logger.Log(logger.Error, err.Error(), userID)
		return nil, fmt.Errorf("failed to create provision record: %w", err)
	}

	if err := s.storage.UpdateUserProvisionStatus(userID, types.UserProvisionStatusProvisioning); err != nil {
		s.logger.Log(logger.Warning, fmt.Sprintf("Failed to set user provision_status=provisioning: %v", err), userID)
	}

	serverID, err := s.storage.SelectBestServer(1, 1024, 25)
	if err != nil {
		s.logger.Log(logger.Warning, fmt.Sprintf("Server scheduling failed, falling back to legacy queue: %v", err), userID)
	}

	payload := types.ProvisionPayload{
		SessionID:          provisionDetails.ID.String(),
		Subdomain:          subdomain,
		ContainerName:      containerName,
		Image:              image,
		UserID:             userID,
		OrgID:              orgID,
		ProvisionDetailsID: provisionDetails.ID.String(),
		ServerID:           serverID,
	}

	if err := s.EnqueueProvisionTask(s.ctx, payload); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to enqueue provision task: %v", err), userID)

		if updateErr := s.storage.UpdateUserProvisionDetailsWithError(provisionDetails.ID.String(), fmt.Sprintf("Failed to enqueue task: %v", err)); updateErr != nil {
			s.logger.Log(logger.Warning, fmt.Sprintf("Failed to update provision details error: %v", updateErr), userID)
		}

		if updateErr := s.storage.UpdateUserProvisionStatus(userID, types.UserProvisionStatusFailed); updateErr != nil {
			s.logger.Log(logger.Warning, fmt.Sprintf("Failed to update user provision_status: %v", updateErr), userID)
		}

		return nil, types.ErrFailedToEnqueueTask
	}

	return &types.ProvisionResponse{
		SessionID: provisionDetails.ID.String(),
		Status:    string(types.UserProvisionStatusProvisioning),
		Message:   "Trail provisioning started",
	}, nil
}
