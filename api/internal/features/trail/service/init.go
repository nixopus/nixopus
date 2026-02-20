package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/trail/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/trail/types"
	"github.com/raghavyuva/nixopus-api/internal/queue"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// TrailService handles business logic for trail provisioning.
type TrailService struct {
	storage storage.TrailRepository
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
	repository storage.TrailRepository,
) *TrailService {
	return &TrailService{
		storage: repository,
		store:   store,
		ctx:     ctx,
		logger:  l,
		config:  &config.AppConfig.Trail,
	}
}

// IsImageAllowed checks if the given image is in the allowed images list.
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

// GenerateSubdomain generates a unique subdomain for a trail instance.
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

// GenerateContainerName generates a container name from the user's display name.
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
	return fmt.Sprintf("trail-%s-%s", name, randomSuffix)
}

// EnqueueProvisionTask enqueues a provision task to the Redis queue.
func (s *TrailService) EnqueueProvisionTask(ctx context.Context, payload types.ProvisionPayload) error {
	return queue.EnqueueProvisionTask(ctx, payload)
}

// generateRandomSubdomain generates a random subdomain string.
func generateRandomSubdomain() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "trail-" + hex.EncodeToString(bytes), nil
}

// generateRandomString generates a random alphanumeric string of the given length.
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}
