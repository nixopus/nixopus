package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"errors"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}

func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case *types.CreateDeploymentRequest:
		return validateDeploymentRequest(r)
	case *types.CreateProjectRequest:
		return validateCreateProjectRequest(r)
	case *types.DeployProjectRequest:
		return validateDeployProjectRequest(*r)
	case *types.UpdateDeploymentRequest:
		return validateUpdateDeploymentRequest(r)
	case *types.DeleteDeploymentRequest:
		return validateDeleteDeploymentRequest(*r)
	case *types.ReDeployApplicationRequest:
		return validateRedeployApplicationRequest(*r)
	case *types.RollbackDeploymentRequest:
		return validateRollbackDeploymentRequest(*r)
	case *types.RestartDeploymentRequest:
		return validateRestartDeploymentRequest(*r)
	case *types.DuplicateProjectRequest:
		return validateDuplicateProjectRequest(*r)
	case *types.GetProjectFamilyRequest:
		return validateGetProjectFamilyRequest(*r)
	case *types.AddApplicationToFamilyRequest:
		return validateAddApplicationToFamilyRequest(r)
	default:
		return types.ErrInvalidRequestType
	}
}

func validateDeploymentRequest(req *types.CreateDeploymentRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.Environment == "" {
		return errors.New("environment is required")
	}
	if !shared_types.IsValidEnvironment(string(req.Environment)) {
		return types.ErrInvalidEnvironment
	}
	if req.BuildPack == "" {
		return errors.New("build_pack is required")
	}
	if req.Repository == "" {
		return errors.New("repository is required")
	}
	if req.Branch == "" {
		return errors.New("branch is required")
	}
	if req.Port == 0 {
		return errors.New("port is required")
	}
	if err := validateDomains(req.Domains); err != nil {
		return err
	}
	if req.BasePath == "" {
		req.BasePath = "/"
	} else if req.BasePath[0] != '/' {
		req.BasePath = "/" + req.BasePath
	}
	return nil
}

func validateUpdateDeploymentRequest(req *types.UpdateDeploymentRequest) error {
	if req.Name != "" {
		if len(req.Name) < 3 {
			return errors.New("name must be at least 3 characters")
		}
	}
	if req.Environment != "" && !shared_types.IsValidEnvironment(string(req.Environment)) {
		return types.ErrInvalidEnvironment
	}
	if req.BuildPack != "" && !shared_types.IsValidBuildPack(string(req.BuildPack)) {
		return types.ErrInvalidBuildPack
	}
	if req.Port != 0 {
		if req.Port < 1 || req.Port > 65535 {
			return errors.New("port must be between 1 and 65535")
		}
	}
	if req.BasePath != "" {
		if req.BasePath[0] != '/' {
			req.BasePath = "/" + req.BasePath
		}
	}
	if req.Domains != nil && len(req.Domains) > 5 {
		return errors.New("maximum 5 domains allowed per application")
	}
	return nil
}

func validateDeleteDeploymentRequest(req types.DeleteDeploymentRequest) error {
	// 	// here we need to validate whether user has access to delete the deployment
	if req.ID == uuid.Nil {
		return types.ErrMissingID
	}
	return nil
}

func validateRedeployApplicationRequest(req types.ReDeployApplicationRequest) error {
	if req.ID == uuid.Nil {
		return types.ErrMissingID
	}
	return nil
}

func validateRollbackDeploymentRequest(req types.RollbackDeploymentRequest) error {
	if req.ID == uuid.Nil {
		return types.ErrMissingID
	}
	return nil
}

func validateRestartDeploymentRequest(req types.RestartDeploymentRequest) error {
	if req.ID == uuid.Nil {
		return types.ErrMissingID
	}
	return nil
}

// validateCreateProjectRequest validates a request to create a project without deploying.
// Only name and repository are required. Domain is optional. Other fields have defaults.
func validateCreateProjectRequest(req *types.CreateProjectRequest) error {
	if req.Name == "" {
		return types.ErrMissingName
	}
	if err := validateDomains(req.Domains); err != nil {
		return err
	}
	if req.Repository == "" {
		return types.ErrMissingRepository
	}
	// Set defaults for optional fields
	if req.Environment == "" {
		req.Environment = "production"
	}
	if !shared_types.IsValidEnvironment(string(req.Environment)) {
		return types.ErrInvalidEnvironment
	}
	if req.BuildPack == "" {
		req.BuildPack = "dockerfile"
	}
	if req.Branch == "" {
		req.Branch = "main"
	}
	if req.Port == 0 {
		req.Port = 3000
	}
	if req.BasePath == "" {
		req.BasePath = "/"
	} else if req.BasePath[0] != '/' {
		req.BasePath = "/" + req.BasePath
	}
	if req.DockerfilePath == "" {
		req.DockerfilePath = "Dockerfile"
	}
	return nil
}

// validateDeployProjectRequest validates a request to deploy an existing project.
func validateDeployProjectRequest(req types.DeployProjectRequest) error {
	if req.ID == uuid.Nil {
		return types.ErrMissingID
	}
	return nil
}

// validateDuplicateProjectRequest validates a request to duplicate a project.
func validateDuplicateProjectRequest(req types.DuplicateProjectRequest) error {
	if req.SourceProjectID == uuid.Nil {
		return types.ErrMissingSourceProjectID
	}
	if err := validateDomains(req.Domains); err != nil {
		return err
	}
	if req.Environment == "" {
		return types.ErrInvalidEnvironment
	}
	if !shared_types.IsValidEnvironment(string(req.Environment)) {
		return types.ErrInvalidEnvironment
	}
	return nil
}

// validateGetProjectFamilyRequest validates a request to get project family.
func validateGetProjectFamilyRequest(req types.GetProjectFamilyRequest) error {
	if req.FamilyID == uuid.Nil {
		return types.ErrMissingID
	}
	return nil
}

// validateAddApplicationToFamilyRequest validates a request to add an application to a family.
func validateAddApplicationToFamilyRequest(req *types.AddApplicationToFamilyRequest) error {
	if req.Name == "" {
		return types.ErrMissingName
	}
	if req.Repository == "" {
		return types.ErrMissingRepository
	}
	if err := validateDomains(req.Domains); err != nil {
		return err
	}
	// Set defaults for optional fields
	if req.Environment == "" {
		req.Environment = "development"
	}
	if !shared_types.IsValidEnvironment(string(req.Environment)) {
		return types.ErrInvalidEnvironment
	}
	if req.BuildPack == "" {
		req.BuildPack = "dockerfile"
	}
	if req.Branch == "" {
		req.Branch = "main"
	}
	if req.Port == 0 {
		req.Port = 8080
	}
	if req.Path == "" {
		req.Path = "/"
	}
	if req.DockerfilePath == "" {
		req.DockerfilePath = "Dockerfile"
	}
	return nil
}

// isDomainValid performs RFC 1035-compliant domain validation (pure string check, no DB).
func isDomainValid(domain string) bool {
	if domain == "" || len(domain) > 253 {
		return false
	}
	for _, c := range domain {
		if c == '/' || c == '\\' || c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			return false
		}
	}
	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return false
	}
	for _, label := range labels {
		if label == "" || len(label) > 63 {
			return false
		}
		for i, c := range label {
			isAlnum := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
			isHyphen := c == '-'
			if !isAlnum && !isHyphen {
				return false
			}
			if isHyphen && (i == 0 || i == len(label)-1) {
				return false
			}
		}
	}
	return true
}

func validateDomains(domains []string) error {
	for _, d := range domains {
		if !isDomainValid(d) {
			return fmt.Errorf("invalid domain: %q", d)
		}
	}
	return nil
}
