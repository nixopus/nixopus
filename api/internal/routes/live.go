package routes

import (
	"context"
	"log"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"

	deploy_storage "github.com/raghavyuva/nixopus-api/internal/features/deploy/storage"
	deploy_tasks "github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	github_connector_service "github.com/raghavyuva/nixopus-api/internal/features/github-connector/service"
	github_connector_storage "github.com/raghavyuva/nixopus-api/internal/features/github-connector/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/live"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	api "github.com/raghavyuva/nixopus-api/internal/version"
)

// LiveDeployController handles live deploy routes
type LiveDeployController struct {
	gateway *live.Gateway
	store   *storage.Store
	logger  logger.Logger
}

// NewLiveDeployController creates a new live deploy controller
func NewLiveDeployController(store *storage.Store) *LiveDeployController {
	// Initialize shared logger
	sharedLogger := logger.NewLogger()

	// Initialize TaskService for live dev deployments
	ctx := context.Background()
	deployStorage := deploy_storage.DeployStorage{DB: store.DB, Ctx: ctx}
	githubConnectorStorage := github_connector_storage.GithubConnectorStorage{DB: store.DB, Ctx: ctx}
	githubConnectorService := github_connector_service.NewGithubConnectorService(store, ctx, sharedLogger, &githubConnectorStorage)
	taskService := deploy_tasks.NewTaskService(&deployStorage, sharedLogger, githubConnectorService, store)
	taskService.SetupCreateDeploymentQueue() // This also sets up LiveDevQueue via onceQueues
	taskService.StartConsumers(ctx)

	// Initialize staging manager (reuses existing GetClonePath function)
	stagingManager := live.NewStagingManager(githubConnectorService, &deployStorage)

	gateway := live.NewGateway(stagingManager, taskService, store)

	// Wire callback: when live dev build completes and container is healthy, mark deployed
	// so file injection is used instead of full rebuilds.
	taskService.SetOnLiveDevDeployed(func(appID uuid.UUID, workdir string) {
		gateway.BuildFirstManager().MarkDeployed(appID, workdir)
	})

	log.Println("Live deploy components initialized")

	return &LiveDeployController{
		gateway: gateway,
		store:   store,
		logger:  sharedLogger,
	}
}

// RegisterLiveDeployRoutes registers live deploy routes
func (router *Router) RegisterLiveDeployRoutes(server *fuego.Server, apiV1 api.Version) {
	controller := NewLiveDeployController(router.app.Store)

	// WebSocket endpoint for live deploy (register at root level to match /ws/live/{application_id})
	wsHandler := func(c fuego.ContextNoBody) (interface{}, error) {
		controller.gateway.HandleWebSocket(c.Response(), c.Request())
		return nil, nil
	}
	fuego.Get(server, "/ws/live/{application_id}", wsHandler)

	// Pause live dev service
	liveGroup := fuego.Group(server, apiV1.Path+"/live")
	fuego.Post(liveGroup, "/pause", controller.HandlePause)
}

// PauseRequest holds the optional request body for pause
type PauseRequest struct {
	ApplicationID string `json:"application_id"`
}

// HandlePause pauses the live dev service for the given application.
// Accepts application_id via query param or JSON body.
func (c *LiveDeployController) HandlePause(f fuego.ContextWithBody[PauseRequest]) (*types.Response, error) {
	r := f.Request()
	applicationIDStr := r.URL.Query().Get("application_id")
	if applicationIDStr == "" {
		if body, err := f.Body(); err == nil && body.ApplicationID != "" {
			applicationIDStr = body.ApplicationID
		}
	}
	if applicationIDStr == "" {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	applicationID, err := uuid.Parse(applicationIDStr)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		if auth := r.Header.Get("Authorization"); len(auth) > 7 && auth[:7] == "Bearer " {
			token = auth[7:]
		}
	}
	if token == "" {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	ctx := r.Context()
	user, orgID, err := c.gateway.VerifySession(ctx, token, r)
	if err != nil {
		c.logger.Log(logger.Error, "pause: session verification failed", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusUnauthorized,
		}
	}

	organizationID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusUnauthorized,
		}
	}

	deployStorage := deploy_storage.DeployStorage{DB: c.store.DB, Ctx: ctx}
	application, err := deployStorage.GetApplicationById(applicationID.String(), organizationID)
	if err != nil || application.UserID != user.ID || application.OrganizationID != organizationID {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusNotFound,
		}
	}

	orgCtx := context.WithValue(ctx, types.OrganizationIDKey, orgID)
	if err := deploy_tasks.PauseLiveDevService(orgCtx, applicationID); err != nil {
		c.logger.Log(logger.Error, "pause: failed to pause service", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.Response{
		Status:  "success",
		Message: "Live dev service paused",
		Data:    nil,
	}, nil
}
