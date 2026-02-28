package controller

import (
	"fmt"
	"net/http"
	"path"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *DeployController) PreviewComposeServices(f fuego.ContextWithBody[types.PreviewComposeRequest]) (*types.PreviewComposeResponse, error) {
	data, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Status: http.StatusBadRequest}
	}

	if data.Repository == "" {
		return nil, fuego.HTTPError{Err: types.ErrMissingRepository, Status: http.StatusBadRequest}
	}
	if data.Branch == "" {
		return nil, fuego.HTTPError{Err: types.ErrMissingBranch, Status: http.StatusBadRequest}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusUnauthorized}
	}

	filePath := composeFilePath(data.BasePath, data.DockerfilePath)
	c.logger.Log(logger.Info, "preview-compose: resolved file path", fmt.Sprintf("repo=%s branch=%s basePath=%q dockerfilePath=%q -> filePath=%q", data.Repository, data.Branch, data.BasePath, data.DockerfilePath, filePath))

	content, err := c.githubService.GetRepositoryFileContent(
		user.ID.String(), data.Repository, data.Branch, filePath,
	)
	if err != nil {
		c.logger.Log(logger.Warning, "preview-compose: failed to fetch from GitHub", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusUnprocessableEntity,
		}
	}
	c.logger.Log(logger.Info, "preview-compose: fetched file", fmt.Sprintf("content length=%d bytes, first 200 chars: %s", len(content), truncate(string(content), 200)))

	parsed, err := tasks.ParseComposeYAML(content)
	if err != nil {
		c.logger.Log(logger.Warning, "preview-compose: YAML parse error", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusUnprocessableEntity,
		}
	}
	c.logger.Log(logger.Info, "preview-compose: parsed services", fmt.Sprintf("count=%d", len(parsed)))
	for _, p := range parsed {
		c.logger.Log(logger.Info, "preview-compose: service detail", fmt.Sprintf("name=%s ports=%v", p.ServiceName, p.Ports))
	}

	var services []types.PreviewComposeService
	for _, p := range parsed {
		port := 0
		if len(p.Ports) > 0 {
			port = p.Ports[0]
		}
		services = append(services, types.PreviewComposeService{
			ServiceName: p.ServiceName,
			Port:        port,
		})
	}

	c.logger.Log(logger.Info, "preview-compose: returning services", fmt.Sprintf("count=%d", len(services)))
	return &types.PreviewComposeResponse{Services: services}, nil
}

func composeFilePath(basePath, dockerfilePath string) string {
	if basePath == "" || basePath == "/" {
		basePath = "."
	}
	fileName := "docker-compose.yml"
	if dockerfilePath != "" && dockerfilePath != "Dockerfile" {
		fileName = dockerfilePath
	}
	return path.Join(basePath, fileName)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
