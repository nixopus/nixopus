package validation

import (
	"encoding/json"
	"io"

	"errors"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
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
	// Domain is now optional - validation removed
	if req.Repository == "" {
		return types.ErrMissingRepository
	}
	// Set defaults for optional fields
	if req.Environment == "" {
		req.Environment = "production"
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
	if req.Environment == "" {
		return types.ErrInvalidEnvironment
	}
	// Validate environment value
	switch req.Environment {
	case "development", "staging", "production":
		// Valid environment
	default:
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
