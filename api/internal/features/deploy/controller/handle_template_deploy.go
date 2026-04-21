package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *DeployController) HandleTemplateDeploy(f fuego.ContextWithBody[types.CreateTemplateDeploymentRequest]) (*types.ApplicationResponse, error) {
	c.logger.Log(logger.Info, "starting template deployment", "")

	data, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, "failed to read request body", err.Error())
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	if err := c.validator.ValidateRequest(&data); err != nil {
		c.logger.Log(logger.Error, "template deployment validation failed", err.Error())
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		return nil, fuego.UnauthorizedError{Detail: "organization not found"}
	}

	ext, err := c.extensionLoader.GetExtensionByID(f.Request().Context(), data.TemplateID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: fmt.Sprintf("template %q not found", data.TemplateID),
			Status: http.StatusNotFound,
		}
	}

	envVars := make(map[string]string)
	var proxyDomain string
	for _, v := range ext.Variables {
		varName := strings.ToUpper(v.VariableName)
		if val, ok := data.Variables[v.VariableName]; ok {
			envVars[varName] = fmt.Sprintf("%v", val)
		} else if v.DefaultValue != nil {
			envVars[varName] = strings.Trim(string(v.DefaultValue), `"`)
		}
		lower := strings.ToLower(v.VariableName)
		if lower == "proxy_domain" || lower == "domain" {
			if val, ok := data.Variables[v.VariableName]; ok {
				proxyDomain = fmt.Sprintf("%v", val)
			}
		}
	}

	env := data.Environment
	if env == "" {
		env = "production"
	}

	deployReq := &types.CreateDeploymentRequest{
		Name:                 data.Name,
		Environment:          env,
		BuildPack:            shared_types.DockerCompose,
		Repository:           "",
		Branch:               "main",
		Port:                 0,
		EnvironmentVariables: envVars,
		Source:               shared_types.SourceTemplate,
		ServerIDs:            data.ServerIDs,
		PrimaryServerID:      data.PrimaryServerID,
		RoutingStrategy:      data.RoutingStrategy,
	}

	if proxyDomain != "" {
		deployReq.Domains = []string{proxyDomain}
	}

	application, err := c.taskService.CreateTemplateDeploymentTask(deployReq, user.ID, organizationID, data.TemplateID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to create template deployment", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ApplicationResponse{
		Status:  "success",
		Message: "Template deployment created successfully",
		Data:    application,
	}, nil
}
