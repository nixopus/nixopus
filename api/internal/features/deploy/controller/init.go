package controller

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/service"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/validation"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	github_service "github.com/raghavyuva/nixopus-api/internal/features/github-connector/service"
	github_storage "github.com/raghavyuva/nixopus-api/internal/features/github-connector/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type DeployController struct {
	store         *shared_storage.Store
	validator     *validation.Validator
	service       *service.DeployService
	storage       *storage.DeployStorage
	ctx           context.Context
	logger        logger.Logger
	notifier      shared_types.Notifier
	taskService   *tasks.TaskService
	githubService *github_service.GithubConnectorService
}

func NewDeployController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	notifier shared_types.Notifier,
) (*DeployController, error) {
	deployStorage := storage.DeployStorage{DB: store.DB, Ctx: ctx}
	github_service := github_service.NewGithubConnectorService(store, ctx, l, &github_storage.GithubConnectorStorage{DB: store.DB, Ctx: ctx})
	taskService := tasks.NewTaskService(&deployStorage, l, github_service, store, notifier)
	taskService.SetupCreateDeploymentQueue()

	// TODO: Re-enable reconciler and health monitor once systemd-based Caddy
	// support is fully validated on trail VMs.
	// orgFetcher := newOrgFetcher(store)
	// reconcilerDaemon := caddy.NewReconcilerDaemon(&deployStorage, l, 5*time.Minute, orgFetcher)
	// reconcilerDaemon.SetupQueues()
	// healthMonitor := caddy.NewHealthMonitor(l, reconcilerDaemon.Reconciler(), 30*time.Second, orgFetcher)
	// healthMonitor.SetupQueue()
	// reconcilerDaemon.Start(ctx)
	// healthMonitor.Start(ctx)

	taskService.StartConsumers(ctx)

	return &DeployController{
		store:         store,
		validator:     validation.NewValidator(),
		service:       service.NewDeployService(store, ctx, l, &deployStorage),
		storage:       &deployStorage,
		ctx:           ctx,
		logger:        l,
		notifier:      notifier,
		taskService:   taskService,
		githubService: github_service,
	}, nil
}

// newOrgFetcher returns a function that queries the DB for all organization IDs
// that have active SSH keys (i.e. servers attached). These are the orgs whose
// Caddy instances need monitoring and reconciliation.
func newOrgFetcher(store *shared_storage.Store) func(ctx context.Context) ([]uuid.UUID, error) {
	return func(ctx context.Context) ([]uuid.UUID, error) {
		var orgIDs []uuid.UUID
		err := store.DB.NewSelect().
			TableExpr("ssh_keys").
			ColumnExpr("DISTINCT organization_id").
			Where("is_active = ?", true).
			Where("deleted_at IS NULL").
			Scan(ctx, &orgIDs)
		if err != nil {
			return nil, err
		}
		return orgIDs, nil
	}
}

// Service returns the deploy service instance.
func (c *DeployController) Service() *service.DeployService {
	return c.service
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
