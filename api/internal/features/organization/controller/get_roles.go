package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetRoles godoc
// @Summary Get all roles with permissions
// @Description Retrieves all roles with their permissions in the database
// @Tags organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} types.Role "Success response with roles and permissions"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /organizations/roles [get]
func (c *OrganizationsController) GetRoles(f fuego.ContextNoBody) (*shared_types.Response, error) {
	c.logger.Log(logger.Info, "getting all roles with permissions", "")

	roles, err := c.service.GetRoles()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Roles with permissions fetched successfully",
		Data:    roles,
	}, nil
}
