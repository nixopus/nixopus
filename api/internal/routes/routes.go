package routes

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-fuego/fuego"
	"github.com/joho/godotenv"
	"github.com/nixopus/nixopus/api/internal/cache"
	"github.com/nixopus/nixopus/api/internal/config"
	audit "github.com/nixopus/nixopus/api/internal/features/audit/controller"
	auth "github.com/nixopus/nixopus/api/internal/features/auth/controller"
	auth_service "github.com/nixopus/nixopus/api/internal/features/auth/service"
	user_storage "github.com/nixopus/nixopus/api/internal/features/auth/storage"
	container "github.com/nixopus/nixopus/api/internal/features/container/controller"
	deploy "github.com/nixopus/nixopus/api/internal/features/deploy/controller"
	domain "github.com/nixopus/nixopus/api/internal/features/domain/controller"
	extension "github.com/nixopus/nixopus/api/internal/features/extension/controller"
	feature_flags_controller "github.com/nixopus/nixopus/api/internal/features/feature-flags/controller"
	feature_flags_service "github.com/nixopus/nixopus/api/internal/features/feature-flags/service"
	feature_flags_storage "github.com/nixopus/nixopus/api/internal/features/feature-flags/storage"
	file_manager "github.com/nixopus/nixopus/api/internal/features/file-manager/controller"
	githubConnector "github.com/nixopus/nixopus/api/internal/features/github-connector/controller"
	healthcheck "github.com/nixopus/nixopus/api/internal/features/healthcheck/controller"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	machine_controller "github.com/nixopus/nixopus/api/internal/features/machine/controller"
	machine_storage "github.com/nixopus/nixopus/api/internal/features/machine/storage"
	mcpController "github.com/nixopus/nixopus/api/internal/features/mcp/controller"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	"github.com/nixopus/nixopus/api/internal/features/notification/channel"
	notificationController "github.com/nixopus/nixopus/api/internal/features/notification/controller"
	server_controller "github.com/nixopus/nixopus/api/internal/features/server/controller"
	telemetry "github.com/nixopus/nixopus/api/internal/features/telemetry/controller"
	trail "github.com/nixopus/nixopus/api/internal/features/trail/controller"
	"github.com/nixopus/nixopus/api/internal/openapi"

	update "github.com/nixopus/nixopus/api/internal/features/update/controller"
	update_service "github.com/nixopus/nixopus/api/internal/features/update/service"
	user "github.com/nixopus/nixopus/api/internal/features/user/controller"
	"github.com/nixopus/nixopus/api/internal/middleware"
	"github.com/nixopus/nixopus/api/internal/realtime"
	"github.com/nixopus/nixopus/api/internal/scheduler"
	"github.com/nixopus/nixopus/api/internal/storage"
	api "github.com/nixopus/nixopus/api/internal/version"
)

// Router holds the application dependencies for route handlers
type Router struct {
	app          *storage.App
	cache        *cache.Cache
	logger       logger.Logger
	socketServer *realtime.SocketServer
	schedulers   *scheduler.Schedulers
}

// MiddlewareConfig defines which middleware to apply to a route group
type MiddlewareConfig struct {
	RBAC         bool
	FeatureFlag  string // empty string means no feature flag middleware
	Audit        bool
	ResourceName string // resource name for RBAC, audit, and feature flag
}

// NewRouter creates a new Router instance with initialized dependencies
func NewRouter(app *storage.App) *Router {
	cache, err := cache.NewCache(config.AppConfig.Redis.URL)
	if err != nil {
		log.Fatal("Error creating redis client", err)
	}

	// Initialize RBAC cache for middleware
	middleware.InitRBACCache(cache)

	return &Router{
		app:    app,
		cache:  cache,
		logger: logger.NewLogger(),
	}
}

// applyMiddleware applies middleware chain to a route group based on configuration
func (router *Router) applyMiddleware(group *fuego.Server, cfg MiddlewareConfig) {
	if cfg.RBAC {
		fuego.Use(group, func(next http.Handler) http.Handler {
			return middleware.RBACMiddleware(next, router.app, cfg.ResourceName)
		})
	}
	if cfg.FeatureFlag != "" {
		fuego.Use(group, func(next http.Handler) http.Handler {
			return middleware.FeatureFlagMiddleware(next, router.app, cfg.FeatureFlag, router.cache)
		})
	}
	if cfg.Audit {
		fuego.Use(group, func(next http.Handler) http.Handler {
			return middleware.AuditMiddleware(next, router.app, router.logger, cfg.ResourceName)
		})
	}
}

// createServer initializes the Fuego server with global middleware and security settings
func (router *Router) createServer(port string) *fuego.Server {
	return fuego.NewServer(
		fuego.WithEngineOptions(
			fuego.WithOpenAPIConfig(fuego.OpenAPIConfig{
				PrettyFormatJSON: true,
				SwaggerURL:       "/swagger",
				SpecURL:          "/swagger/openapi.json",
				JSONFilePath:     "doc/openapi.json",
			}),
			fuego.WithOpenAPIGeneratorSchemaCustomizer(openapi.SchemaCustomizer),
		),
		fuego.WithGlobalMiddlewares(
			middleware.RecoveryMiddleware,
			middleware.CorsMiddleware,
			middleware.LoggingMiddleware,
			api.VersionMiddleware,
			api.MigrationMiddleware,
		),
		fuego.WithSecurity(openapi3.SecuritySchemes{
			"bearerAuth": &openapi3.SecuritySchemeRef{
				Value: openapi3.NewSecurityScheme().
					WithType("http").
					WithScheme("bearer").
					WithBearerFormat("JWT").
					WithDescription("Enter your JWT token in the format: Bearer <token>"),
			},
		}),
		fuego.WithAddr(":"+port),
		fuego.WithRouteOptions(openapi.RouteContractOption()),
	)
}

// setupAuthentication configures the authentication middleware
func (router *Router) setupAuthentication(server *fuego.Server) {
	fuego.Use(server, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.AppConfig.App.Environment == "development" && strings.HasPrefix(r.URL.Path, "/swagger") {
				next.ServeHTTP(w, r)
				return
			}
			if strings.HasPrefix(r.URL.Path, "/api/v1/live") || strings.HasPrefix(r.URL.Path, "/ws/live") {
				next.ServeHTTP(w, r)
				return
			}
			middleware.AuthMiddleware(next, router.app, router.cache).ServeHTTP(w, r)
		})
	})
}

// SetSchedulers sets the schedulers on the router
func (router *Router) SetSchedulers(schedulers *scheduler.Schedulers) {
	router.schedulers = schedulers
}

// initChannels creates and returns the notification channel adapters
// backed by the application database.
func (router *Router) initChannels() map[string]channel.Channel {
	db := router.app.Store.DB
	ctx := router.app.Ctx
	channels := map[string]channel.Channel{
		"email":   channel.NewEmailChannel(db, ctx),
		"slack":   channel.NewSlackChannel(db, ctx),
		"discord": channel.NewDiscordChannel(db, ctx),
	}

	resendCfg := config.AppConfig.Resend
	if resendCfg.APIKey != "" {
		channels["system_email"] = channel.NewSystemEmailChannel(resendCfg.APIKey, resendCfg.FromEmail)
	}

	agentCfg := config.AppConfig.AgentChannel
	authURL := strings.TrimRight(config.AppConfig.BetterAuth.URL, "/")

	if agentCfg.URL != "" && authURL != "" && agentCfg.ClientID != "" && agentCfg.ClientSecret != "" {
		webhookURL := strings.TrimRight(agentCfg.URL, "/") + "/api/webhooks/events"
		tokenURL := authURL + "/api/auth/oauth2/token"
		channels["agent"] = channel.NewAgentChannel(webhookURL, tokenURL, agentCfg.ClientID, agentCfg.ClientSecret)
	}

	return channels
}

// SetupRoutes initializes and configures all application routes
func (router *Router) SetupRoutes() {
	if err := godotenv.Load(); err != nil {
		log.Println("Info: .env file not found, using environment variables and secret manager")
	}

	// Initialize notification dispatcher with channel adapters
	channels := router.initChannels()
	dispatcher := notification.NewDispatcher(router.app.Store.DB, router.app.Ctx, router.logger, channels)
	dispatcher.SetupQueue()

	if router.schedulers != nil && router.schedulers.HealthCheck != nil {
		router.schedulers.HealthCheck.SetNotifier(dispatcher)
	}
	if router.schedulers != nil && router.schedulers.TrialExpiry != nil {
		router.schedulers.TrialExpiry.SetNotifier(dispatcher)
	}

	PORT := config.AppConfig.Server.Port
	server := router.createServer(PORT)
	apiV1 := api.NewVersion(api.CurrentVersion)

	deployController, err := deploy.NewDeployController(router.app.Store, router.app.Ctx, router.logger, dispatcher)
	if err != nil {
		log.Fatalf("Failed to create deploy controller: %v", err)
	}

	router.registerPublicRoutes(server, apiV1, dispatcher, deployController)
	router.setupAuthentication(server)
	router.registerProtectedRoutes(server, apiV1, dispatcher, deployController)

	log.Printf("Server starting on port %s", PORT)
	log.Printf("Swagger UI available at: http://localhost:%s/swagger/", PORT)
	_ = os.Remove("doc/openapi.json")
	go func() {
		if err := openapi.PostProcessSpecWithRetry("doc/openapi.json", 30*time.Second); err != nil {
			log.Printf("Warning: failed to post-process OpenAPI spec: %v", err)
		}
	}()
	server.Run()
}

// registerPublicRoutes registers routes that don't require authentication
func (router *Router) registerPublicRoutes(server *fuego.Server, apiV1 api.Version, dispatcher *notification.Dispatcher, deployController *deploy.DeployController) {
	healthGroup := fuego.Group(server, apiV1.Path+"/health")
	router.RegisterHealthRoutes(healthGroup)

	webhookGroup := fuego.Group(server, apiV1.Path+"/webhook")
	fuego.Post(
		webhookGroup,
		"",
		deployController.HandleGithubWebhook,
		fuego.OptionSummary("Handle GitHub webhook"),
	)

	router.RegisterWebSocketRoutes(server, deployController, router.schedulers.HealthCheck)

	trailInternalController := trail.NewTrailController(router.app.Store, router.app.Ctx, router.logger, router.cache)
	trailInternalGroup := fuego.Group(server, apiV1.Path+"/trail")
	router.RegisterTrailInternalRoutes(trailInternalGroup, trailInternalController)

	authController := router.createAuthController(dispatcher)
	authGroup := fuego.Group(server, apiV1.Path+"/auth")
	router.RegisterAuthRoutes(authGroup, authController)

	mcpPublicCtrl := mcpController.NewMCPController(router.app.Store, router.app.Ctx, router.logger)
	mcpPublicGroup := fuego.Group(server, apiV1.Path+"/mcp")
	router.RegisterMCPPublicRoutes(mcpPublicGroup, mcpPublicCtrl)

	telemetryCtrl := telemetry.NewTelemetryController(router.app.Store.DB, router.app.Ctx, router.logger)
	telemetryGroup := fuego.Group(server, apiV1.Path+"/cli/telemetry")
	fuego.Use(telemetryGroup, middleware.NewRateLimiterWithConfig(0.01, 3))
	router.RegisterTelemetryRoutes(telemetryGroup, telemetryCtrl)
}

// registerProtectedRoutes registers routes that require authentication
func (router *Router) registerProtectedRoutes(server *fuego.Server, apiV1 api.Version, dispatcher *notification.Dispatcher, deployController *deploy.DeployController) {
	authController := router.createAuthController(dispatcher)
	authProtectedGroup := fuego.Group(server, apiV1.Path+"/auth")
	router.applyMiddleware(authProtectedGroup, MiddlewareConfig{RBAC: false, Audit: false, ResourceName: "auth"})
	router.RegisterAuthProtectedRoutes(authProtectedGroup, authController)

	userController := user.NewUserController(router.app.Store, router.app.Ctx, router.logger, router.cache)
	userGroup := fuego.Group(server, apiV1.Path+"/user")
	router.applyMiddleware(userGroup, MiddlewareConfig{RBAC: false, Audit: false, ResourceName: "user"})
	router.RegisterUserRoutes(userGroup, userController)

	domainController := domain.NewDomainsController(router.app.Store, router.app.Ctx, router.logger, dispatcher)
	domainGroup := fuego.Group(server, apiV1.Path+"/domain")
	router.applyMiddleware(domainGroup, MiddlewareConfig{RBAC: true, FeatureFlag: "domain", Audit: true, ResourceName: "domain"})
	router.RegisterDomainRoutes(domainGroup, domainController)

	githubConnectorController := githubConnector.NewGithubConnectorController(router.app.Store, router.app.Ctx, router.logger, dispatcher)
	githubConnectorGroup := fuego.Group(server, apiV1.Path+"/github-connector")
	router.applyMiddleware(githubConnectorGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "github_connector",
		Audit:        true,
		ResourceName: "github-connector",
	})
	router.RegisterGithubConnectorRoutes(githubConnectorGroup, githubConnectorController)

	notifController := notificationController.NewNotificationController(router.app.Store, router.app.Ctx, router.logger, dispatcher)
	notificationGroup := fuego.Group(server, apiV1.Path+"/notification")
	router.applyMiddleware(notificationGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "notifications",
		Audit:        true,
		ResourceName: "notification",
	})
	router.RegisterNotificationRoutes(notificationGroup, notifController)

	fileManagerController := file_manager.NewFileManagerController(router.app.Store, router.app.Ctx, router.logger, dispatcher)
	fileManagerGroup := fuego.Group(server, apiV1.Path+"/file-manager")
	router.applyMiddleware(fileManagerGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "file_manager",
		Audit:        true,
		ResourceName: "file-manager",
	})
	router.RegisterFileManagerRoutes(fileManagerGroup, fileManagerController)

	deployGroup := fuego.Group(server, apiV1.Path+"/deploy")
	router.applyMiddleware(deployGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "deploy",
		Audit:        true,
		ResourceName: "deploy",
	})
	router.RegisterDeployRoutes(deployGroup, deployController)

	auditController := audit.NewAuditController(router.app.Store.DB, router.app.Ctx, router.logger)
	auditGroup := fuego.Group(server, apiV1.Path+"/audit")
	router.applyMiddleware(auditGroup, MiddlewareConfig{RBAC: true, FeatureFlag: "audit", Audit: true, ResourceName: "audit"})
	router.RegisterAuditRoutes(auditGroup, auditController)

	updateService := update_service.NewUpdateService(router.app, &router.logger, router.app.Ctx)
	updateController := update.NewUpdateController(updateService, &router.logger)
	updateGroup := fuego.Group(server, apiV1.Path+"/update")
	router.RegisterUpdateRoutes(updateGroup, updateController)

	featureFlagController := router.createFeatureFlagController()
	featureFlagReadGroup := fuego.Group(server, apiV1.Path+"/feature-flags")
	featureFlagWriteGroup := fuego.Group(server, apiV1.Path+"/feature-flags")
	featureFlagMiddleware := MiddlewareConfig{RBAC: true, Audit: true, ResourceName: "feature_flags"}
	router.applyMiddleware(featureFlagReadGroup, featureFlagMiddleware)
	router.applyMiddleware(featureFlagWriteGroup, featureFlagMiddleware)
	router.RegisterFeatureFlagRoutes(featureFlagReadGroup, featureFlagWriteGroup, featureFlagController)

	containerController, err := container.NewContainerController(router.app.Store, router.app.Ctx, router.logger, dispatcher)
	if err != nil {
		log.Fatalf("Failed to create container controller: %v", err)
	}
	containerGroup := fuego.Group(server, apiV1.Path+"/container")
	router.applyMiddleware(containerGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "container",
		Audit:        true,
		ResourceName: "container",
	})
	router.RegisterContainerRoutes(containerGroup, containerController)

	healthCheckController := healthcheck.NewHealthCheckController(router.app.Store, router.app.Ctx, router.logger)
	healthCheckGroup := fuego.Group(server, apiV1.Path+"/healthcheck")
	router.applyMiddleware(healthCheckGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "deploy",
		Audit:        true,
		ResourceName: "healthcheck",
	})
	router.RegisterHealthCheckRoutes(healthCheckGroup, healthCheckController)

	extensionController := extension.NewExtensionsController(router.app.Store, router.app.Ctx, router.logger, config.AppConfig.Redis.URL)
	extensionGroup := fuego.Group(server, apiV1.Path+"/extensions")
	router.applyMiddleware(extensionGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "extension",
		Audit:        true,
		ResourceName: "extension",
	})
	router.RegisterExtensionRoutes(extensionGroup, extensionController)

	serverController := server_controller.NewServerController(router.app.Store, router.app.Ctx, router.logger, dispatcher)
	serverGroup := fuego.Group(server, apiV1.Path+"/servers")
	router.applyMiddleware(serverGroup, MiddlewareConfig{
		RBAC:         true,
		Audit:        true,
		ResourceName: "server",
	})
	router.RegisterServerRoutes(serverGroup, serverController)

	machineTimescaleStore, _ := machine_storage.NewTimescaleStore(router.app.Ctx, config.AppConfig.Timescale.URL)
	machineController := machine_controller.NewMachineController(router.app.Store, router.app.Ctx, router.logger, machineTimescaleStore)
	machineGroup := fuego.Group(server, apiV1.Path+"/machine")
	router.applyMiddleware(machineGroup, MiddlewareConfig{
		RBAC:         true,
		Audit:        true,
		ResourceName: "machine",
	})
	router.RegisterMachineRoutes(machineGroup, machineController)

	machineBillingGroup := fuego.Group(server, apiV1.Path+"/machine")
	router.applyMiddleware(machineBillingGroup, MiddlewareConfig{
		RBAC:         false,
		Audit:        true,
		ResourceName: "machine",
	})
	router.RegisterMachineBillingRoutes(machineBillingGroup, machineController)

	machineRegGroup := fuego.Group(server, apiV1.Path)
	router.applyMiddleware(machineRegGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "machine_byos",
		Audit:        true,
		ResourceName: "machine",
	})
	router.RegisterMachineRegistrationRoutes(machineRegGroup, machineController)

	trailController := trail.NewTrailController(router.app.Store, router.app.Ctx, router.logger, router.cache)
	trailGroup := fuego.Group(server, apiV1.Path+"/trail")
	router.applyMiddleware(trailGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "trail",
		Audit:        true,
		ResourceName: "trail",
	})
	router.RegisterTrailRoutes(trailGroup, trailController)

	mcpCtrl := mcpController.NewMCPController(router.app.Store, router.app.Ctx, router.logger)
	mcpGroup := fuego.Group(server, apiV1.Path+"/mcp")
	router.applyMiddleware(mcpGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "mcp",
		Audit:        true,
		ResourceName: "mcp",
	})
	router.RegisterMCPRoutes(mcpGroup, mcpCtrl)
}

func (router *Router) createAuthController(dispatcher *notification.Dispatcher) *auth.AuthController {
	userStorage := &user_storage.UserStorage{DB: router.app.Store.DB, Ctx: router.app.Ctx}
	authService := auth_service.NewAuthService(userStorage, router.logger, router.app.Ctx, config.AppConfig.Redis.URL)
	return auth.NewAuthController(router.app.Ctx, router.logger, dispatcher, *authService, router.app.Store)
}

func (router *Router) createFeatureFlagController() *feature_flags_controller.FeatureFlagController {
	featureFlagStorage := &feature_flags_storage.FeatureFlagStorage{DB: router.app.Store.DB, Ctx: router.app.Ctx}
	featureFlagService := feature_flags_service.NewFeatureFlagService(featureFlagStorage, router.logger, router.app.Ctx)
	return feature_flags_controller.NewFeatureFlagController(featureFlagService, router.logger, router.app.Ctx, router.cache)
}
