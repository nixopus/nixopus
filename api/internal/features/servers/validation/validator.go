package validation

import (
	"net"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/types"
)

// Validator handles server validation logic
type Validator struct {
	storage storage.ServerStorageInterface
}

// NewValidator creates a new validator instance
func NewValidator(storage storage.ServerStorageInterface) *Validator {
	return &Validator{
		storage: storage,
	}
}

// ValidateID validates the server ID is a valid UUID
func (v *Validator) ValidateID(id string) error {
	if id == "" {
		return types.ErrMissingServerID
	}
	if _, err := uuid.Parse(id); err != nil {
		return types.ErrInvalidServerID
	}
	return nil
}

// ValidateName validates server name meets requirements
func (v *Validator) ValidateName(name string) error {
	if name == "" {
		return types.ErrMissingServerName
	}

	if len(name) < 2 {
		return types.ErrServerNameTooShort
	}

	if len(name) > 255 {
		return types.ErrServerNameTooLong
	}

	// Allow alphanumeric characters, hyphens, underscores, and spaces
	validName := regexp.MustCompile(`^[a-zA-Z0-9\-_\s]+$`)
	if !validName.MatchString(name) {
		return types.ErrServerNameInvalid
	}

	return nil
}

// ValidateHost validates server host (IP address or hostname)
func (v *Validator) ValidateHost(host string) error {
	if host == "" {
		return types.ErrMissingHost
	}

	// Check if it's a valid IP address
	if net.ParseIP(host) != nil {
		return nil
	}

	// Check if it's a valid hostname
	validHostname := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !validHostname.MatchString(host) {
		return types.ErrInvalidHost
	}

	return nil
}

// ValidatePort validates server port number
func (v *Validator) ValidatePort(port int) error {
	if port <= 0 {
		return types.ErrMissingPort
	}

	if port < 1 || port > 65535 {
		return types.ErrInvalidPort
	}

	return nil
}

// ValidateUsername validates SSH username
func (v *Validator) ValidateUsername(username string) error {
	if username == "" {
		return types.ErrMissingUsername
	}

	// Basic username validation (alphanumeric, hyphens, underscores)
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	if !validUsername.MatchString(username) {
		return types.ErrMissingUsername
	}

	return nil
}

// ValidateSSHAuth validates SSH authentication methods for creation (requires auth)
func (v *Validator) ValidateSSHAuth(password *string, privateKeyPath *string) error {
	hasPassword := password != nil && *password != ""
	hasPrivateKey := privateKeyPath != nil && *privateKeyPath != ""

	// Must have exactly one authentication method
	if !hasPassword && !hasPrivateKey {
		return types.ErrMissingSSHAuth
	}

	if hasPassword && hasPrivateKey {
		return types.ErrBothSSHAuthProvided
	}

	// Validate private key path if provided
	if hasPrivateKey {
		if !filepath.IsAbs(*privateKeyPath) {
			return types.ErrInvalidSSHPrivateKeyPath
		}
		// Check file extension
		ext := filepath.Ext(*privateKeyPath)
		if ext != "" && !strings.Contains(".pem.key.ppk", ext) {
			return types.ErrInvalidSSHPrivateKeyPath
		}
	}

	return nil
}

// ValidateSSHAuthForUpdate validates SSH authentication methods for updates (optional)
func (v *Validator) ValidateSSHAuthForUpdate(password *string, privateKeyPath *string) error {
	hasPassword := password != nil && *password != ""
	hasPrivateKey := privateKeyPath != nil && *privateKeyPath != ""

	// In update mode, auth fields are optional - only validate if both are provided
	if hasPassword && hasPrivateKey {
		return types.ErrBothSSHAuthProvided
	}

	// Validate private key path if provided
	if hasPrivateKey {
		if !filepath.IsAbs(*privateKeyPath) {
			return types.ErrInvalidSSHPrivateKeyPath
		}
		// Check file extension
		ext := filepath.Ext(*privateKeyPath)
		if ext != "" && !strings.Contains(".pem.key.ppk", ext) {
			return types.ErrInvalidSSHPrivateKeyPath
		}
	}

	return nil
}

// ValidateRequest validates different server request types
func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case *types.CreateServerRequest:
		return v.ValidateCreateServerRequest(*r)
	case *types.UpdateServerRequest:
		return v.ValidateUpdateServerRequest(*r)
	case *types.UpdateServerStatusRequest:
		return v.ValidateUpdateServerStatusRequest(*r)
	case *types.DeleteServerRequest:
		return v.ValidateDeleteServerRequest(*r)
	default:
		return types.ErrInvalidRequestType
	}
}

// ValidateCreateServerRequest validates server creation requests
func (v *Validator) ValidateCreateServerRequest(req types.CreateServerRequest) error {
	if err := v.ValidateName(req.Name); err != nil {
		return err
	}

	if err := v.ValidateHost(req.Host); err != nil {
		return err
	}

	if err := v.ValidatePort(req.Port); err != nil {
		return err
	}

	if err := v.ValidateUsername(req.Username); err != nil {
		return err
	}

	if err := v.ValidateSSHAuth(req.SSHPassword, req.SSHPrivateKeyPath); err != nil {
		return err
	}

	return nil
}

// ValidateUpdateServerRequest validates server update requests
func (v *Validator) ValidateUpdateServerRequest(req types.UpdateServerRequest) error {
	if err := v.ValidateID(req.ID); err != nil {
		return err
	}

	if err := v.ValidateName(req.Name); err != nil {
		return err
	}

	if err := v.ValidateHost(req.Host); err != nil {
		return err
	}

	if err := v.ValidatePort(req.Port); err != nil {
		return err
	}

	if err := v.ValidateUsername(req.Username); err != nil {
		return err
	}

	if err := v.ValidateSSHAuthForUpdate(req.SSHPassword, req.SSHPrivateKeyPath); err != nil {
		return err
	}

	return nil
}

// ValidateDeleteServerRequest validates server deletion requests
func (v *Validator) ValidateDeleteServerRequest(req types.DeleteServerRequest) error {
	if err := v.ValidateID(req.ID); err != nil {
		return err
	}

	return nil
}

func (v *Validator) ValidateUpdateServerStatusRequest(req types.UpdateServerStatusRequest) error {
	if err := v.ValidateID(req.ID); err != nil {
		return err
	}

	if err := v.ValidateStatus(req.Status); err != nil {
		return err
	}

	return nil
}

func (v *Validator) ValidateStatus(status string) error {
	if status == "" {
		return types.ErrMissingStatus
	}

	validStatuses := []string{"active", "inactive", "maintenance"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}

	return types.ErrInvalidStatus
}
