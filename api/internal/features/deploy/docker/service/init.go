package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker/storage"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type DockerConfigService struct {
	store   *shared_storage.Store
	ctx     context.Context
	logger  logger.Logger
	storage storage.DockerConfigRepository
}

func NewDockerConfigService(store *shared_storage.Store, ctx context.Context, l logger.Logger) *DockerConfigService {
	dockerStorage := &storage.DockerConfigStorage{DB: store.DB, Ctx: ctx}
	return &DockerConfigService{
		store:   store,
		ctx:     ctx,
		logger:  l,
		storage: dockerStorage,
	}
}

// GetDockerConfigForOrganization retrieves the active Docker config for an organization
// and converts it to runtime DockerRuntimeConfig format ready for use with DockerManager.
// Returns an error if no active config is found for the organization.
func (s *DockerConfigService) GetDockerConfigForOrganization(orgID uuid.UUID) (*types.DockerRuntimeConfig, error) {
	dockerConfig, err := s.storage.GetActiveDockerConfigByOrganizationID(orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active Docker config found for organization %s", orgID.String())
		}
		s.logger.Log(logger.Error, fmt.Sprintf("failed to get Docker config for organization %s: %v", orgID.String(), err), "")
		return nil, fmt.Errorf("failed to get Docker config for organization: %w", err)
	}

	// Decrypt certificates (implement encryption/decryption logic)
	caCert := getStringValue(dockerConfig.CACertEncrypted)
	clientCert := getStringValue(dockerConfig.ClientCertEncrypted)
	clientKey := getStringValue(dockerConfig.ClientKeyEncrypted)

	port := 2376
	if dockerConfig.Port != nil {
		port = *dockerConfig.Port
	}

	config := &types.DockerRuntimeConfig{
		Host:       dockerConfig.Host,
		Port:       port,
		TLSEnabled: dockerConfig.TLSEnabled,
		CACert:     caCert,
		ClientCert: clientCert,
		ClientKey:  clientKey,
	}

	s.logger.Log(logger.Info, fmt.Sprintf("Docker config loaded for organization %s: host=%s, port=%d, tls_enabled=%v", 
		orgID.String(), config.Host, config.Port, config.TLSEnabled), "")

	return config, nil
}

// getStringValue safely extracts string value from pointer, returning empty string if nil
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
