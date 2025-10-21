package controller

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/service"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/validation"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	github_service "github.com/raghavyuva/nixopus-api/internal/features/github-connector/service"
	github_storage "github.com/raghavyuva/nixopus-api/internal/features/github-connector/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type DeployController struct {
	store        *shared_storage.Store
	validator    *validation.Validator
	service      *service.DeployService
	ctx          context.Context
	logger       logger.Logger
	notification *notification.NotificationManager
	taskService  *tasks.TaskService
}

func NewDeployController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	notificationManager *notification.NotificationManager,
) *DeployController {
	storage := storage.DeployStorage{DB: store.DB, Ctx: ctx}
	docker_repo := docker.NewDockerService()
	github_service := github_service.NewGithubConnectorService(store, ctx, l, &github_storage.GithubConnectorStorage{DB: store.DB, Ctx: ctx})
	taskService := tasks.NewTaskService(&storage, l, docker_repo, github_service, store)
	taskService.SetupCreateDeploymentQueue()
	taskService.StartConsumers(ctx)

	return &DeployController{
		store:        store,
		validator:    validation.NewValidator(),
		service:      service.NewDeployService(store, ctx, l, &storage, docker_repo, github_service),
		ctx:          ctx,
		logger:       l,
		notification: notificationManager,
		taskService:  taskService,
	}
}

// parseAndValidate parses and validates the request body.
//
// This method attempts to parse the request body into the provided 'req' interface
// using the controller's validator. If parsing fails, an error response is sent
// and the method returns false. It also validates the parsed request object and
// returns false if validation fails. If both operations are successful, it returns true.
//
// Parameters:
//
//	w - the HTTP response writer to send error responses.
//	r - the HTTP request containing the body to parse.
//	req - the interface to populate with the parsed request body.
//
// Returns:
//
//	bool - true if parsing and validation succeed, false otherwise.
func (c *DeployController) parseAndValidate(w http.ResponseWriter, r *http.Request, req interface{}) bool {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		c.logger.Log(logger.Error, shared_types.ErrFailedToDecodeRequest.Error(), err.Error())
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return false
	}
	defer r.Body.Close()

	if err := c.validator.ParseRequestBody(r, io.NopCloser(bytes.NewReader(body)), req); err != nil {
		c.logger.Log(logger.Error, shared_types.ErrFailedToDecodeRequest.Error(), err.Error())
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return false
	}

	if err := c.validator.ValidateRequest(req); err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}
