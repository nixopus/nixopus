package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// HandleDuplicateProject duplicates an existing project with a different environment.
// All configurations are copied and a new project is created in draft status.
func (c *DeployController) HandleDuplicateProject(f fuego.ContextWithBody[types.DuplicateProjectRequest]) (*types.ApplicationResponse, error) {
	c.logger.Log(logger.Info, "duplicating project", "")

	data, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, "failed to read request body", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	c.logger.Log(logger.Info, "request body parsed successfully", "source_id: "+data.SourceProjectID.String())

	if err := c.validator.ValidateRequest(&data); err != nil {
		c.logger.Log(logger.Error, "request validation failed", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "user authentication failed", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "organization not found", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	c.logger.Log(logger.Info, "attempting to duplicate project", "source_id: "+data.SourceProjectID.String()+", user_id: "+user.ID.String())

	application, err := c.service.DuplicateProject(&data, user.ID, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to duplicate project", err.Error())

		status := http.StatusInternalServerError
		if err == types.ErrApplicationNotFound {
			status = http.StatusNotFound
		} else if err == types.ErrSameEnvironmentAsDuplicate || err == types.ErrEnvironmentAlreadyExistsInFamily {
			status = http.StatusConflict
		}

		return nil, fuego.HTTPError{
			Err:    err,
			Status: status,
		}
	}

	c.logger.Log(logger.Info, "project duplicated successfully", "new_id: "+application.ID.String())

	return &types.ApplicationResponse{
		Status:  "success",
		Message: "Project duplicated successfully",
		Data:    application,
	}, nil
}

// HandleGetProjectFamily retrieves all projects that belong to a family.
func (c *DeployController) HandleGetProjectFamily(f fuego.ContextNoBody) (*types.ProjectFamilyResponse, error) {
	familyIDStr := f.QueryParam("family_id")
	if familyIDStr == "" {
		c.logger.Log(logger.Error, "family_id is required", "")
		return nil, fuego.HTTPError{
			Err:    types.ErrMissingID,
			Status: http.StatusBadRequest,
		}
	}

	familyID, err := uuid.Parse(familyIDStr)
	if err != nil {
		c.logger.Log(logger.Error, "invalid family_id", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "user authentication failed", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "organization not found", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	c.logger.Log(logger.Info, "getting project family", "family_id: "+familyID.String())

	projects, err := c.service.GetProjectFamily(familyID, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to get project family", err.Error())

		status := http.StatusInternalServerError
		if err == types.ErrProjectFamilyNotFound {
			status = http.StatusNotFound
		}

		return nil, fuego.HTTPError{
			Err:    err,
			Status: status,
		}
	}

	return &types.ProjectFamilyResponse{
		Status:  "success",
		Message: "Project family retrieved successfully",
		Data: types.ProjectFamilyResponseData{
			Projects: projects,
		},
	}, nil
}

// HandleGetEnvironmentsInFamily retrieves all environments that exist in a project family.
func (c *DeployController) HandleGetEnvironmentsInFamily(f fuego.ContextNoBody) (*types.EnvironmentsInFamilyResponse, error) {
	familyIDStr := f.QueryParam("family_id")
	if familyIDStr == "" {
		c.logger.Log(logger.Error, "family_id is required", "")
		return nil, fuego.HTTPError{
			Err:    types.ErrMissingID,
			Status: http.StatusBadRequest,
		}
	}

	familyID, err := uuid.Parse(familyIDStr)
	if err != nil {
		c.logger.Log(logger.Error, "invalid family_id", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "user authentication failed", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "organization not found", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	c.logger.Log(logger.Info, "getting environments in family", "family_id: "+familyID.String())

	environments, err := c.service.GetEnvironmentsInFamily(familyID, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to get environments in family", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.EnvironmentsInFamilyResponse{
		Status:  "success",
		Message: "Environments retrieved successfully",
		Data: types.EnvironmentsInFamilyResponseData{
			Environments: environments,
		},
	}, nil
}
