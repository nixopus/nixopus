package service

import (
	"context"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	ssh_key_storage "github.com/raghavyuva/nixopus-api/internal/features/ssh-key/storage"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"golang.org/x/crypto/ssh"
)

type SSHKeyService struct {
	store   *shared_storage.Store
	ctx     context.Context
	logger  logger.Logger
	storage ssh_key_storage.SSHKeyRepository
}

func NewSSHKeyService(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	repository ssh_key_storage.SSHKeyRepository,
) *SSHKeyService {
	return &SSHKeyService{
		store:   store,
		ctx:     ctx,
		logger:  l,
		storage: repository,
	}
}

// SetupSelfHostedSSHKey sets up SSH keys for a self-hosted user after registration
func (s *SSHKeyService) SetupSelfHostedSSHKey(userID uuid.UUID, organizationID uuid.UUID) error {
	s.logger.Log(logger.Info, fmt.Sprintf("Setting up SSH key for self-hosted user %s, org %s", userID, organizationID), "")

	// Check if SSH keys already exist for this organization
	existingKeys, err := s.storage.GetSSHKeysByOrganizationID(organizationID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to check existing SSH keys", err.Error())
		return fmt.Errorf("failed to check existing SSH keys: %w", err)
	}

	// If keys already exist, skip setup
	if len(existingKeys) > 0 {
		s.logger.Log(logger.Info, "SSH keys already exist for organization, skipping setup", organizationID.String())
		return nil
	}

	// Read SSH private key from environment variable
	privateKeyPath := config.AppConfig.SSH.PrivateKey
	if privateKeyPath == "" {
		s.logger.Log(logger.Warning, "SSH_PRIVATE_KEY not set in environment, skipping SSH key setup", "")
		return nil
	}

	// Read private key file
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to read SSH private key from %s", privateKeyPath), err.Error())
		return fmt.Errorf("failed to read SSH private key: %w", err)
	}

	privateKeyContent := string(privateKeyBytes)

	// Extract public key from private key
	publicKeyContent, err := s.extractPublicKeyFromPrivate(privateKeyContent)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to extract public key from private key", err.Error())
		return fmt.Errorf("failed to extract public key: %w", err)
	}

	// Generate fingerprint from public key
	fingerprint, err := s.generateFingerprint(publicKeyContent)
	if err != nil {
		s.logger.Log(logger.Warning, "Failed to generate fingerprint", err.Error())
		fingerprint = nil
	}

	// Encrypt private key (base64 encode for now - proper encryption should be added)
	encryptedPrivateKey := base64.StdEncoding.EncodeToString([]byte(privateKeyContent))

	// Determine key type and size
	keyType, keySize := s.parseKeyInfo(privateKeyContent)

	// Get SSH connection details from config
	var sshHost, sshUser *string
	var sshPort *int

	if config.AppConfig.SSH.Host != "" {
		sshHost = &config.AppConfig.SSH.Host
	}
	if config.AppConfig.SSH.User != "" {
		sshUser = &config.AppConfig.SSH.User
	}
	if config.AppConfig.SSH.Port > 0 {
		port := int(config.AppConfig.SSH.Port)
		sshPort = &port
	} else {
		// Default SSH port
		defaultPort := 22
		sshPort = &defaultPort
	}

	// Create SSH key record
	sshKey := &types.SSHKey{
		ID:                  uuid.New(),
		OrganizationID:      organizationID,
		Name:                "Default SSH Key",
		Description:         stringPtr("SSH key generated during installation"),
		Host:                sshHost,
		User:                sshUser,
		Port:                sshPort,
		PublicKey:           publicKeyContent,
		PrivateKeyEncrypted: &encryptedPrivateKey,
		KeyType:             keyType,
		KeySize:             keySize,
		Fingerprint:         fingerprint,
		AuthMethod:          "key",
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Save to database
	if err := s.storage.CreateSSHKey(sshKey); err != nil {
		s.logger.Log(logger.Error, "Failed to create SSH key in database", err.Error())
		return fmt.Errorf("failed to create SSH key: %w", err)
	}

	s.logger.Log(logger.Info, "Successfully set up SSH key for self-hosted user", fmt.Sprintf("user_id=%s, org_id=%s", userID, organizationID))
	return nil
}

// extractPublicKeyFromPrivate extracts the public key from a private key
func (s *SSHKeyService) extractPublicKeyFromPrivate(privateKeyContent string) (string, error) {
	// Try to parse as RSA private key
	block, _ := pem.Decode([]byte(privateKeyContent))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	var privateKey interface{}
	var err error

	switch block.Type {
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse RSA private key: %w", err)
		}
	case "PRIVATE KEY":
		privateKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse PKCS8 private key: %w", err)
		}
	default:
		return "", fmt.Errorf("unsupported private key type: %s", block.Type)
	}

	// Convert to SSH public key
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key is not RSA")
	}

	publicKey, err := ssh.NewPublicKey(&rsaKey.PublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to create SSH public key: %w", err)
	}

	// Format as authorized_keys format
	publicKeyContent := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(publicKey)))
	return publicKeyContent, nil
}

// generateFingerprint generates MD5 fingerprint from public key
func (s *SSHKeyService) generateFingerprint(publicKeyContent string) (*string, error) {
	// Parse public key to get the key bytes
	parts := strings.Fields(publicKeyContent)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid public key format")
	}

	keyBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	// Generate MD5 fingerprint
	hash := md5.Sum(keyBytes)
	fingerprint := fmt.Sprintf("%x", hash)
	// Format as colon-separated hex (standard SSH fingerprint format)
	formatted := ""
	for i := 0; i < len(fingerprint); i += 2 {
		if i > 0 {
			formatted += ":"
		}
		formatted += fingerprint[i : i+2]
	}

	return &formatted, nil
}

// parseKeyInfo extracts key type and size from private key
func (s *SSHKeyService) parseKeyInfo(privateKeyContent string) (string, int) {
	block, _ := pem.Decode([]byte(privateKeyContent))
	if block == nil {
		return "rsa", 4096 // Default
	}

	var privateKey interface{}
	var err error

	switch block.Type {
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err == nil {
			if rsaKey, ok := privateKey.(*rsa.PrivateKey); ok {
				return "rsa", rsaKey.N.BitLen()
			}
		}
	case "PRIVATE KEY":
		privateKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err == nil {
			if rsaKey, ok := privateKey.(*rsa.PrivateKey); ok {
				return "rsa", rsaKey.N.BitLen()
			}
		}
	}

	return "rsa", 4096 // Default fallback
}

func stringPtr(s string) *string {
	return &s
}
