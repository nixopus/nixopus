package routes

import (
	"context"
	"log"

	"github.com/go-fuego/fuego"

	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	deploy_storage "github.com/raghavyuva/nixopus-api/internal/features/deploy/storage"
	deploy_tasks "github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	github_connector_service "github.com/raghavyuva/nixopus-api/internal/features/github-connector/service"
	github_connector_storage "github.com/raghavyuva/nixopus-api/internal/features/github-connector/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/live"
	"github.com/raghavyuva/nixopus-api/internal/storage"
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
	dockerRepo, err := docker.GetDockerManager().GetDefaultService()
	if err != nil {
		log.Fatalf("Failed to get Docker service: %v", err)
	}
	if dockerRepo == nil {
		log.Fatalf("Docker service is nil")
	}
	githubConnectorStorage := github_connector_storage.GithubConnectorStorage{DB: store.DB, Ctx: ctx}
	githubConnectorService := github_connector_service.NewGithubConnectorService(store, ctx, sharedLogger, &githubConnectorStorage)
	taskService := deploy_tasks.NewTaskService(&deployStorage, sharedLogger, dockerRepo, githubConnectorService, store)
	taskService.SetupCreateDeploymentQueue() // This also sets up LiveDevQueue via onceQueues
	taskService.StartConsumers(ctx)

	// Initialize staging manager (reuses existing GetClonePath function)
	stagingManager := live.NewStagingManager(githubConnectorService, &deployStorage)

	gateway := live.NewGateway(stagingManager, taskService, store)

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
}
