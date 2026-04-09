package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"github.com/google/uuid"
	ff_service "github.com/nixopus/nixopus/api/internal/features/feature-flags/service"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/queue"
	api_types "github.com/nixopus/nixopus/api/internal/types"
	cryptossh "golang.org/x/crypto/ssh"
)

const defaultMaxBYOSMachines = 2

type MachineBillingChecker interface {
	CanProvision(orgID uuid.UUID) error
}

type NoOpBillingChecker struct{}

func (n *NoOpBillingChecker) CanProvision(orgID uuid.UUID) error {
	return nil
}

type RegistrationService struct {
	storage            *storage.RegistrationStorage
	featureFlagService *ff_service.FeatureFlagService
	billingChecker     MachineBillingChecker
	logger             logger.Logger
	ctx                context.Context
}

func NewRegistrationService(
	s *storage.RegistrationStorage,
	ffs *ff_service.FeatureFlagService,
	bc MachineBillingChecker,
	l logger.Logger,
	ctx context.Context,
) *RegistrationService {
	if bc == nil {
		bc = &NoOpBillingChecker{}
	}
	return &RegistrationService{
		storage:            s,
		featureFlagService: ffs,
		billingChecker:     bc,
		logger:             l,
		ctx:                ctx,
	}
}

func (s *RegistrationService) CreateMachine(orgID uuid.UUID, userID uuid.UUID, req types.CreateMachineRequest) (*types.CreateMachineResponse, error) {
	count, err := s.storage.CountUserOwnedMachines(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to count machines: %w", err)
	}
	if count >= defaultMaxBYOSMachines {
		return nil, types.ErrMachineLimitReached
	}

	port := req.Port
	if port == 0 {
		port = 22
	}
	user := req.User
	if user == "" {
		user = "root"
	}

	exists, err := s.storage.HostPortExists(orgID, req.Host, port)
	if err != nil {
		return nil, fmt.Errorf("failed to check host uniqueness: %w", err)
	}
	if exists {
		return nil, types.ErrDuplicateHost
	}

	privateKeyPEM, publicKeyStr, fingerprint, err := generateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	keyType := "rsa"
	keySize := 4096
	authMethod := "key"
	sshKey := &api_types.SSHKey{
		ID:                  uuid.New(),
		OrganizationID:      orgID,
		Name:                req.Name,
		Host:                &req.Host,
		User:                &user,
		Port:                &port,
		PublicKey:           &publicKeyStr,
		PrivateKeyEncrypted: &privateKeyPEM,
		KeyType:             &keyType,
		KeySize:             &keySize,
		Fingerprint:         &fingerprint,
		AuthMethod:          authMethod,
		IsActive:            false,
		IsDefault:           false,
	}

	if err := s.storage.InsertSSHKey(sshKey); err != nil {
		return nil, fmt.Errorf("failed to insert ssh key: %w", err)
	}

	if err := s.storage.InsertProvisionDetails(userID, orgID, sshKey.ID, "user_owned", "COMPLETED"); err != nil {
		return nil, fmt.Errorf("failed to insert provision details: %w", err)
	}

	return &types.CreateMachineResponse{
		ID:        sshKey.ID.String(),
		Name:      req.Name,
		Host:      req.Host,
		Port:      port,
		User:      user,
		PublicKey: publicKeyStr,
	}, nil
}

func (s *RegistrationService) VerifyMachine(orgID uuid.UUID, machineID uuid.UUID) error {
	_, err := s.storage.GetSSHKeyByID(machineID, orgID)
	if err != nil {
		return fmt.Errorf("machine not found: %w", err)
	}

	return queue.EnqueueMachineVerifyTask(s.ctx, queue.MachineVerifyPayload{
		MachineID: machineID.String(),
		OrgID:     orgID.String(),
	})
}

func (s *RegistrationService) DeleteMachine(orgID uuid.UUID, machineID uuid.UUID) error {
	_, err := s.storage.GetSSHKeyByID(machineID, orgID)
	if err != nil {
		return fmt.Errorf("machine not found: %w", err)
	}

	hasApps, err := s.storage.HasActiveAppServers(machineID)
	if err != nil {
		return fmt.Errorf("failed to check app servers: %w", err)
	}
	if hasApps {
		return types.ErrMachineHasApps
	}

	return s.storage.SoftDeleteSSHKey(machineID)
}

func (s *RegistrationService) GetSSHStatus(orgID uuid.UUID, machineID uuid.UUID) (*types.SSHStatusResponse, error) {
	isActive, lastUsedAt, err := s.storage.GetSSHKeyStatus(machineID, orgID)
	if err != nil {
		return nil, fmt.Errorf("machine not found: %w", err)
	}

	resp := &types.SSHStatusResponse{
		IsActive: isActive,
	}
	if lastUsedAt != nil {
		resp.LastUsedAt = lastUsedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return resp, nil
}

func generateKeyPair() (privateKeyPEM, publicKeyStr, fingerprint string, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate RSA key: %w", err)
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicSSHKey, err := cryptossh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create SSH public key: %w", err)
	}

	publicKeyBytes := cryptossh.MarshalAuthorizedKey(publicSSHKey)

	hash := sha256.Sum256(publicSSHKey.Marshal())
	fp := "SHA256:" + base64.StdEncoding.EncodeToString(hash[:])

	return string(privatePEM), string(publicKeyBytes), fp, nil
}
