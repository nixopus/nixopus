package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	machine_storage "github.com/nixopus/nixopus/api/internal/features/machine/storage"
	machine_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/queue"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// TrailService handles business logic for trial machine provisioning.
type TrailService struct {
	storage machine_storage.TrailRepository
	store   *shared_storage.Store
	ctx     context.Context
	logger  logger.Logger
	config  *shared_types.TrailConfig
}

// NewTrailService creates a new TrailService instance.
func NewTrailService(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	repository machine_storage.TrailRepository,
) *TrailService {
	return &TrailService{
		storage: repository,
		store:   store,
		ctx:     ctx,
		logger:  l,
		config:  &config.AppConfig.Trail,
	}
}

func (s *TrailService) IsImageAllowed(image string) bool {
	if len(s.config.AllowedImages) == 0 {
		return true
	}
	for _, allowed := range s.config.AllowedImages {
		if allowed == image {
			return true
		}
	}
	return false
}

func (s *TrailService) GenerateSubdomain() (string, error) {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		subdomain, err := generateRandomSubdomain()
		if err != nil {
			return "", fmt.Errorf("failed to generate subdomain: %w", err)
		}

		taken, err := s.storage.IsSubdomainTaken(subdomain)
		if err != nil {
			return "", fmt.Errorf("failed to check subdomain availability: %w", err)
		}

		if !taken {
			return subdomain, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique subdomain after %d attempts", maxAttempts)
}

func (s *TrailService) GenerateContainerName(displayName string) string {
	name := strings.ToLower(displayName)

	re := regexp.MustCompile(`[^a-z0-9-]`)
	name = re.ReplaceAllString(name, "-")

	re = regexp.MustCompile(`-+`)
	name = re.ReplaceAllString(name, "-")

	name = strings.Trim(name, "-")

	if len(name) > 20 {
		name = name[:20]
	}

	if name == "" {
		name = "trail"
	}

	randomSuffix, _ := generateRandomString(6)
	return fmt.Sprintf("%s-%s", name, randomSuffix)
}

func (s *TrailService) EnqueueProvisionTask(ctx context.Context, payload machine_types.ProvisionPayload) error {
	return queue.EnqueueProvisionTask(ctx, payload)
}

// ProvisionTrail handles the business logic for provisioning a new trial machine.
func (s *TrailService) ProvisionTrail(userID, orgID string, req machine_types.ProvisionRequest) (*machine_types.ProvisionResponse, error) {
	image := req.Image
	if image == "" {
		image = s.config.DefaultImage
	}

	if !s.IsImageAllowed(image) {
		s.logger.Log(logger.Warning, fmt.Sprintf("User %s requested disallowed image: %s", userID, image), userID)
		return nil, machine_types.ErrImageNotAllowed
	}

	activeProvision, err := s.storage.GetActiveProvisionByUserAndOrg(userID, orgID)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), userID)
		return nil, fmt.Errorf("failed to check active provisions: %w", err)
	}

	if activeProvision != nil {
		return nil, machine_types.ErrActiveProvisionExists
	}

	count, err := s.storage.CountActiveProvisions()
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, fmt.Errorf("failed to check system capacity: %w", err)
	}

	if count >= s.config.MaxConcurrentTrails {
		s.logger.Log(logger.Warning, fmt.Sprintf("Max concurrent trails reached (%d/%d)", count, s.config.MaxConcurrentTrails), "")
		return nil, machine_types.ErrSystemAtCapacity
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

	initialStep := shared_types.ProvisionStepInitializing
	provisionDetails := &shared_types.UserProvisionDetails{
		UserID:           userUUID,
		OrganizationID:   orgUUID,
		LXDContainerName: &containerName,
		Subdomain:        &subdomain,
		Domain:           &fullDomain,
		Step:             &initialStep,
	}

	if err := s.storage.CreateActiveUserProvision(provisionDetails); err != nil {
		if strings.Contains(err.Error(), "active_provision_per_user_org") || strings.Contains(err.Error(), "duplicate") {
			return nil, machine_types.ErrActiveProvisionExists
		}
		s.logger.Log(logger.Error, err.Error(), userID)
		return nil, fmt.Errorf("failed to create provision record: %w", err)
	}

	if err := s.storage.UpdateUserProvisionStatus(userID, machine_types.UserProvisionStatusProvisioning); err != nil {
		s.logger.Log(logger.Warning, fmt.Sprintf("Failed to set user provision_status=provisioning: %v", err), userID)
	}

	serverID, err := s.storage.SelectBestServer(1, 1024, 25)
	if err != nil {
		s.logger.Log(logger.Warning, fmt.Sprintf("Server scheduling failed, falling back to legacy queue: %v", err), userID)
	}

	payload := machine_types.ProvisionPayload{
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

		if updateErr := s.storage.UpdateUserProvisionStatus(userID, machine_types.UserProvisionStatusFailed); updateErr != nil {
			s.logger.Log(logger.Warning, fmt.Sprintf("Failed to update user provision_status: %v", updateErr), userID)
		}

		return nil, machine_types.ErrFailedToEnqueueTask
	}

	return &machine_types.ProvisionResponse{
		SessionID: provisionDetails.ID.String(),
		Status:    string(machine_types.UserProvisionStatusProvisioning),
		Message:   "Trail provisioning started",
	}, nil
}

// GetStatus retrieves the current status of a trial provision.
func (s *TrailService) GetStatus(userID, sessionID string) (*machine_types.StatusResponse, error) {
	details, err := s.storage.GetUserProvisionDetailsByID(sessionID)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), userID)
		return nil, fmt.Errorf("failed to retrieve status: %w", err)
	}

	if details == nil {
		return nil, machine_types.ErrProvisionNotFound
	}

	if details.UserID.String() != userID {
		return nil, machine_types.ErrProvisionNotFound
	}

	userStatus, err := s.storage.GetUserProvisionStatus(userID)
	if err != nil {
		s.logger.Log(logger.Warning, fmt.Sprintf("Failed to get user provision status: %v", err), userID)
		userStatus = machine_types.UserProvisionStatusPending
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

	return &machine_types.StatusResponse{
		SessionID: details.ID.String(),
		Status:    string(userStatus),
		Step:      stepStr,
		Progress:  progress,
		Message:   message,
		Subdomain: getStringValue(details.Subdomain),
		TrailURL:  trailURL,
	}, nil
}

func (s *TrailService) calculateProgress(step *shared_types.ProvisionStep, status machine_types.UserProvisionStatus, errorMsg *string) (int, string) {
	if status == machine_types.UserProvisionStatusCompleted {
		return 100, "Provisioning completed successfully"
	}

	if status == machine_types.UserProvisionStatusFailed {
		msg := "Provisioning failed"
		if errorMsg != nil {
			msg = fmt.Sprintf("Provisioning failed: %s", *errorMsg)
		}
		return 0, msg
	}

	if status == machine_types.UserProvisionStatusPending {
		return 0, "Waiting to start..."
	}

	if step == nil {
		return 5, "Initializing..."
	}

	switch *step {
	case shared_types.ProvisionStepInitializing:
		return 5, "Initializing..."
	case shared_types.ProvisionStepCreatingContainer:
		return 15, "Creating container..."
	case shared_types.ProvisionStepSetupNetworking:
		return 25, "Setting up networking..."
	case shared_types.ProvisionStepInstallingDeps:
		return 45, "Installing dependencies (this may take a few minutes)..."
	case shared_types.ProvisionStepConfiguringSSH:
		return 65, "Configuring SSH..."
	case shared_types.ProvisionStepSetupSSHForwarding:
		return 75, "Setting up SSH forwarding..."
	case shared_types.ProvisionStepVerifyingSSH:
		return 85, "Verifying connection..."
	case shared_types.ProvisionStepCompleted:
		return 100, "Completed"
	default:
		return 50, "Provisioning in progress..."
	}
}

// UpgradeResources enqueues a resource update task for a completed trial provision.
func (s *TrailService) UpgradeResources(userID, orgID string, vcpu, memoryMB int) error {
	provision, err := s.storage.GetCompletedProvisionByUserID(userID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to look up completed provision: %v", err), userID)
		return fmt.Errorf("failed to look up provision: %w", err)
	}

	if provision == nil {
		s.logger.Log(logger.Warning, "No completed provision found for resource upgrade", userID)
		return machine_types.ErrProvisionNotFound
	}

	if provision.LXDContainerName == nil || *provision.LXDContainerName == "" {
		s.logger.Log(logger.Error, "Completed provision has no container name", userID)
		return fmt.Errorf("provision missing container name")
	}

	payload := queue.ResourceUpdatePayload{
		VMName:    *provision.LXDContainerName,
		VcpuCount: vcpu,
		MemoryMB:  memoryMB,
		UserID:    userID,
		OrgID:     orgID,
	}

	if provision.ServerID != nil {
		payload.ServerID = provision.ServerID.String()
	}

	if err := queue.EnqueueResourceUpdateTask(s.ctx, payload); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to enqueue resource update: %v", err), userID)
		return machine_types.ErrFailedToEnqueueTask
	}

	s.logger.Log(logger.Info, fmt.Sprintf("Resource upgrade enqueued: vm=%s vcpu=%d mem=%d", payload.VMName, vcpu, memoryMB), userID)
	return nil
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func generateRandomSubdomain() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}
