package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh/storage"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type SSHKeyService struct {
	store   *shared_storage.Store
	ctx     context.Context
	logger  logger.Logger
	storage storage.SSHKeyRepository
}

func NewSSHKeyService(store *shared_storage.Store, ctx context.Context, l logger.Logger) *SSHKeyService {
	sshStorage := &storage.SSHKeyStorage{DB: store.DB, Ctx: ctx}
	return &SSHKeyService{
		store:   store,
		ctx:     ctx,
		logger:  l,
		storage: sshStorage,
	}
}

// GetSSHConfigForOrganization retrieves the active SSH key for an organization
// and converts it to SSHConfig format ready for use with SSHManager.
// Returns an error if no active SSH key is found for the organization.
func (s *SSHKeyService) GetSSHConfigForOrganization(orgID uuid.UUID) (*types.SSHConfig, error) {
	sshKey, err := s.storage.GetActiveSSHKeyByOrganizationID(orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active SSH key found for organization %s", orgID.String())
		}
		s.logger.Log(logger.Error, fmt.Sprintf("failed to get SSH key for organization %s: %v", orgID.String(), err), "")
		return nil, fmt.Errorf("failed to get SSH key for organization: %w", err)
	}

	// Convert SSHKey to SSHConfig
	// Note: Decryption of PrivateKeyEncrypted and PasswordEncrypted will be implemented later
	privateKey := getStringValue(sshKey.PrivateKeyEncrypted)
	password := getStringValue(sshKey.PasswordEncrypted)
	host := getStringValue(sshKey.Host)
	user := getStringValue(sshKey.User)
	port := getUintFromInt(sshKey.Port, 22)

	config := &types.SSHConfig{
		Host:                host,
		User:                user,
		Port:                port,
		PrivateKey:          privateKey,
		Password:            password,
		PrivateKeyProtected: "", // Not used in current implementation
	}

	// Log SSH credentials info (without sensitive data)
	hasPrivateKey := len(privateKey) > 0
	hasPassword := len(password) > 0
	authMethod := "none"
	if hasPrivateKey {
		authMethod = "private_key"
	} else if hasPassword {
		authMethod = "password"
	}
	s.logger.Log(logger.Info, fmt.Sprintf("SSH config loaded for organization %s: host=%s, user=%s, port=%d, auth_method=%s, has_private_key=%v, has_password=%v", orgID.String(), host, user, port, authMethod, hasPrivateKey, hasPassword), "")

	return config, nil
}

// getStringValue safely extracts string value from pointer, returning empty string if nil
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// getUintFromInt safely converts *int to uint, returning default value if nil
func getUintFromInt(ptr *int, defaultValue uint) uint {
	if ptr == nil {
		return defaultValue
	}
	if *ptr < 0 {
		return defaultValue
	}
	return uint(*ptr)
}
