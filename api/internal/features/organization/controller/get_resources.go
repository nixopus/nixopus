package controller

import (
	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *OrganizationsController) GetResources(f fuego.ContextNoBody) (*shared_types.Response, error) {
	c.logger.Log(logger.Info, "getting all available resources", "")

	resources := []types.ResourceType{
		types.ResourceTypeUser,
		types.ResourceTypeOrganization,
		types.ResourceTypeRole,
		types.ResourceTypePermission,
		types.ResourceTypeDomain,
		types.ResourceTypeGithubConnector,
		types.ResourceTypeNotification,
		types.ResourceTypeFileManager,
		types.ResourceTypeDeploy,
		types.ResourceTypeAudit,
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Resources fetched successfully",
		Data:    resources,
	}, nil
}
