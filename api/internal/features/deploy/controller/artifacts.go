package controller

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *DeployController) ListArtifacts(f fuego.ContextNoBody) (*types.ArtifactListResponse, error) {
	appIDStr := f.QueryParam("application_id")
	if appIDStr == "" {
		return nil, fuego.BadRequestError{Detail: "application_id is required"}
	}
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		return nil, fuego.BadRequestError{Detail: "invalid application_id"}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		return nil, fuego.UnauthorizedError{Detail: "organization not found"}
	}

	artifacts, err := c.service.ListArtifacts(appID, orgID)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to list artifacts: "+err.Error(), "")
		if errors.Is(err, types.ErrPermissionDenied) {
			return nil, fuego.ForbiddenError{Detail: "permission denied"}
		}
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: "failed to list artifacts",
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ArtifactListResponse{
		Status:  "success",
		Message: "Artifacts retrieved successfully",
		Data:    artifacts,
	}, nil
}

func (c *DeployController) GetArtifactDownloadURL(f fuego.ContextNoBody) (*types.ArtifactDownloadResponse, error) {
	deploymentID := f.PathParam("deployment_id")
	if deploymentID == "" {
		return nil, fuego.BadRequestError{Detail: "deployment_id is required"}
	}

	if _, err := uuid.Parse(deploymentID); err != nil {
		return nil, fuego.BadRequestError{Detail: "deployment_id must be a valid UUID"}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		return nil, fuego.UnauthorizedError{Detail: "organization not found"}
	}

	url, err := c.service.GetArtifactDownloadURL(f.Context(), deploymentID, orgID)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to get artifact download URL: "+err.Error(), "")
		if errors.Is(err, types.ErrPermissionDenied) {
			return nil, fuego.ForbiddenError{Detail: "permission denied"}
		}
		if strings.Contains(err.Error(), "deployment not found") {
			return nil, fuego.NotFoundError{Detail: "deployment not found"}
		}
		if strings.Contains(err.Error(), "no artifact available") {
			return nil, fuego.BadRequestError{Detail: "no artifact available"}
		}
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: "failed to get download URL",
			Status: http.StatusInternalServerError,
		}
	}

	resp := &types.ArtifactDownloadResponse{
		Status:  "success",
		Message: "Download URL generated successfully",
	}
	resp.Data.URL = url
	resp.Data.ExpiresIn = 900
	return resp, nil
}

func (c *DeployController) DeleteArtifact(f fuego.ContextNoBody) (*types.ArtifactDeleteResponse, error) {
	deploymentID := f.PathParam("deployment_id")
	if deploymentID == "" {
		return nil, fuego.BadRequestError{Detail: "deployment_id is required"}
	}

	if _, err := uuid.Parse(deploymentID); err != nil {
		return nil, fuego.BadRequestError{Detail: "deployment_id must be a valid UUID"}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		return nil, fuego.UnauthorizedError{Detail: "organization not found"}
	}

	if err := c.service.DeleteArtifact(f.Context(), deploymentID, orgID); err != nil {
		c.logger.Log(logger.Error, "Failed to delete artifact: "+err.Error(), "")
		if errors.Is(err, types.ErrPermissionDenied) {
			return nil, fuego.ForbiddenError{Detail: "permission denied"}
		}
		if strings.Contains(err.Error(), "deployment not found") {
			return nil, fuego.NotFoundError{Detail: "deployment not found"}
		}
		if strings.Contains(err.Error(), "no artifact") {
			return nil, fuego.BadRequestError{Detail: "no artifact available"}
		}
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: "failed to delete artifact",
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ArtifactDeleteResponse{
		Status:  "success",
		Message: "Artifact deleted successfully",
	}, nil
}
