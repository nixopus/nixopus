package routes

import (
	"net/http"
	"os"
	"strings"
	"time"

	apilog "github.com/nixopus/nixopus/api/internal/log"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/joho/godotenv"
	"github.com/nixopus/nixopus/api/internal/cache"
	"github.com/nixopus/nixopus/api/internal/config"
	audit "github.com/nixopus/nixopus/api/internal/features/audit/controller"
	authsession "github.com/nixopus/nixopus/api/internal/features/auth"
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
	githubConnector "github.com/nixopus/nixopus/api/internal/features/github-connector/controller"
	healthcheck "github.com/nixopus/nixopus/api/internal/features/healthcheck/controller"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	machine_controller "github.com/nixopus/nixopus/api/internal/features/machine/controller"
	machine_storage "github.com/nixopus/nixopus/api/internal/features/machine/storage"
	mcpController "github.com/nixopus/nixopus/api/internal/features/mcp/controller"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	"github.com/nixopus/nixopus/api/internal/features/notification/channel"
	notificationController "github.com/nixopus/nixopus/api/internal/features/notification/controller"
	telemetry "github.com/nixopus/nixopus/api/internal/features/telemetry/controller"
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
		apilog.Fatal("Error creating redis client", err)
	}

	// Initialize RBAC cache for middleware
	middleware.InitRBACCache(cache)

	router := &Router{
		app:    app,
		cache:  cache,
		logger: logger.NewLogger(),
	}
	authsession.VerifySessionLogger = &router.logger
	return router
}

// applyMiddleware applies middleware chain to a route group based on configuration
func (router *Router) applyMiddleware(group *fuego.Server, cfg MiddlewareConfig) {
	if cfg.RBAC {
		fuego.Use(group, func(next http.Handler) http.Handler {
			return middleware.RBACMiddleware(next, router.app, cfg.ResourceName, router.logger)
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
		fuego.WithoutAutoGroupTags(),
		fuego.WithLoggingMiddleware(fuego.LoggingConfig{
			DisableRequest:  true,
			DisableResponse: true,
		}),
		fuego.WithGlobalMiddlewares(
			middleware.RecoveryMiddleware(router.logger),
			middleware.RequestIDMiddleware,
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
			middleware.AuthMiddleware(next, router.app, router.cache, router.logger).ServeHTTP(w, r)
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
		apilog.Println("Info: .env file not found, using environment variables and secret manager")
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

	deployController, err := deploy.NewDeployController(router.app.Store, router.app.Ctx, router.logger, dispatcher, router.app.Store.ExtensionLoader)
	if err != nil {
		apilog.Fatalf("Failed to create deploy controller: %v", err)
	}

	router.registerPublicRoutes(server, apiV1, dispatcher, deployController)
	router.setupAuthentication(server)
	router.registerProtectedRoutes(server, apiV1, dispatcher, deployController)

	apilog.Printf("Server starting on port %s", PORT)
	apilog.Printf("Swagger UI available at: http://localhost:%s/swagger/", PORT)
	_ = os.Remove("doc/openapi.json")
	go func() {
		if err := openapi.PostProcessSpecWithRetry("doc/openapi.json", 30*time.Second); err != nil {
			apilog.Printf("Warning: failed to post-process OpenAPI spec: %v", err)
		}
	}()

	openapi.CleanSpecTags(server.OpenAPI)
	server.Run()
}

// registerPublicRoutes registers routes that don't require authentication
func (router *Router) registerPublicRoutes(server *fuego.Server, apiV1 api.Version, dispatcher *notification.Dispatcher, deployController *deploy.DeployController) {
	healthGroup := fuego.Group(server, apiV1.Path+"/health", option.Tags("Health"))
	router.RegisterHealthRoutes(healthGroup)

	webhookGroup := fuego.Group(server, apiV1.Path+"/webhook", option.Tags("Webhooks"))
	fuego.Post(
		webhookGroup,
		"",
		deployController.HandleGithubWebhook,
		fuego.OptionSummary("Handle GitHub webhook"),
	)

	router.RegisterWebSocketRoutes(server, deployController, router.schedulers.HealthCheck)

	trailInternalController := machine_controller.NewTrailController(router.app.Store, router.app.Ctx, router.logger, router.cache)
	trailInternalGroup := fuego.Group(server, apiV1.Path+"/trail", option.Tags("Trial"))
	router.RegisterTrailInternalRoutes(trailInternalGroup, trailInternalController)

	authController := router.createAuthController()
	authGroup := fuego.Group(server, apiV1.Path+"/auth", option.Tags("Auth"))
	router.RegisterAuthRoutes(authGroup, authController)

	mcpPublicCtrl := mcpController.NewMCPController(router.app.Store, router.app.Ctx, router.logger)
	mcpPublicGroup := fuego.Group(server, apiV1.Path+"/mcp", option.Tags("MCP"))
	router.RegisterMCPPublicRoutes(mcpPublicGroup, mcpPublicCtrl)

	telemetryCtrl := telemetry.NewTelemetryController(router.app.Store.DB, router.app.Ctx, router.logger)
	telemetryGroup := fuego.Group(server, apiV1.Path+"/cli/telemetry", option.Tags("Telemetry"))
	// OS env wins over viper/app config (YAML "production" should not beat ENV=test from the shell).
	env := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	if env == "" {
		env = strings.ToLower(strings.TrimSpace(config.AppConfig.App.Environment))
	}
	// Burst-3 limiter trips integration tests (many POSTs, one client IP). Enable only in prod.
	if env == "production" || env == "prod" {
		fuego.Use(telemetryGroup, middleware.NewRateLimiterWithConfig(0.01, 3))
	}
	router.RegisterTelemetryRoutes(telemetryGroup, telemetryCtrl)
}

// registerProtectedRoutes registers routes that require authentication
func (router *Router) registerProtectedRoutes(server *fuego.Server, apiV1 api.Version, dispatcher *notification.Dispatcher, deployController *deploy.DeployController) {
	authController := router.createAuthController()
	authProtectedGroup := fuego.Group(server, apiV1.Path+"/auth", option.Tags("Auth"))
	router.applyMiddleware(authProtectedGroup, MiddlewareConfig{RBAC: false, Audit: false, ResourceName: "auth"})
	router.RegisterAuthProtectedRoutes(authProtectedGroup, authController)

	userController := user.NewUserController(router.app.Store, router.app.Ctx, router.logger, router.cache)
	userGroup := fuego.Group(server, apiV1.Path+"/user", option.Tags("User"))
	router.applyMiddleware(userGroup, MiddlewareConfig{RBAC: false, Audit: false, ResourceName: "user"})
	router.RegisterUserRoutes(userGroup, userController)

	domainController := domain.NewDomainsController(router.app.Store, router.app.Ctx, router.logger, dispatcher)
	domainGroup := fuego.Group(server, apiV1.Path+"/domain", option.Tags("Domains"))
	router.applyMiddleware(domainGroup, MiddlewareConfig{RBAC: true, FeatureFlag: "domain", Audit: true, ResourceName: "domain"})
	router.RegisterDomainRoutes(domainGroup, domainController)

	githubConnectorController := githubConnector.NewGithubConnectorController(router.app.Store, router.app.Ctx, router.logger, dispatcher)
	githubConnectorGroup := fuego.Group(server, apiV1.Path+"/github-connector", option.Tags("GitHub Connector"))
	router.applyMiddleware(githubConnectorGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "github_connector",
		Audit:        true,
		ResourceName: "github-connector",
	})
	router.RegisterGithubConnectorRoutes(githubConnectorGroup, githubConnectorController)

	notifController := notificationController.NewNotificationController(router.app.Store, router.app.Ctx, router.logger, dispatcher)
	notificationGroup := fuego.Group(server, apiV1.Path+"/notification", option.Tags("Notifications"))
	router.applyMiddleware(notificationGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "notifications",
		Audit:        true,
		ResourceName: "notification",
	})
	router.RegisterNotificationRoutes(notificationGroup, notifController)

	deployGroup := fuego.Group(server, apiV1.Path+"/deploy", option.Tags("Deploy"))
	router.applyMiddleware(deployGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "deploy",
		Audit:        true,
		ResourceName: "deploy",
	})
	router.RegisterDeployRoutes(deployGroup, deployController)

	auditController := audit.NewAuditController(router.app.Store.DB, router.app.Ctx, router.logger)
	auditGroup := fuego.Group(server, apiV1.Path+"/audit", option.Tags("Audit"))
	router.applyMiddleware(auditGroup, MiddlewareConfig{RBAC: true, FeatureFlag: "audit", Audit: true, ResourceName: "audit"})
	router.RegisterAuditRoutes(auditGroup, auditController)

	updateService := update_service.NewUpdateService(router.app, &router.logger, router.app.Ctx)
	updateController := update.NewUpdateController(updateService, &router.logger)
	updateGroup := fuego.Group(server, apiV1.Path+"/update", option.Tags("Update"))
	router.RegisterUpdateRoutes(updateGroup, updateController)

	featureFlagController := router.createFeatureFlagController()
	featureFlagReadGroup := fuego.Group(server, apiV1.Path+"/feature-flags", option.Tags("Feature Flags"))
	featureFlagWriteGroup := fuego.Group(server, apiV1.Path+"/feature-flags", option.Tags("Feature Flags"))
	featureFlagMiddleware := MiddlewareConfig{RBAC: true, Audit: true, ResourceName: "feature_flags"}
	router.applyMiddleware(featureFlagReadGroup, featureFlagMiddleware)
	router.applyMiddleware(featureFlagWriteGroup, featureFlagMiddleware)
	router.RegisterFeatureFlagRoutes(featureFlagReadGroup, featureFlagWriteGroup, featureFlagController)

	containerController, err := container.NewContainerController(router.app.Store, router.app.Ctx, router.logger, dispatcher)
	if err != nil {
		apilog.Fatalf("Failed to create container controller: %v", err)
	}
	containerGroup := fuego.Group(server, apiV1.Path+"/container", option.Tags("Containers"))
	fuego.Use(containerGroup, middleware.ServerIDMiddleware)
	router.applyMiddleware(containerGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "container",
		Audit:        true,
		ResourceName: "container",
	})
	router.RegisterContainerRoutes(containerGroup, containerController)

	healthCheckController := healthcheck.NewHealthCheckController(router.app.Store, router.app.Ctx, router.logger)
	healthCheckGroup := fuego.Group(server, apiV1.Path+"/healthcheck", option.Tags("Health Checks"))
	router.applyMiddleware(healthCheckGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "deploy",
		Audit:        true,
		ResourceName: "healthcheck",
	})
	router.RegisterHealthCheckRoutes(healthCheckGroup, healthCheckController)

	extensionController := extension.NewExtensionsController(router.app.Store, router.app.Ctx, router.logger, router.cache)
	extensionGroup := fuego.Group(server, apiV1.Path+"/extensions", option.Tags("Extensions"))
	router.applyMiddleware(extensionGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "extension",
		Audit:        true,
		ResourceName: "extension",
	})
	router.RegisterExtensionRoutes(extensionGroup, extensionController)

	machineTimescaleStore, _ := machine_storage.NewTimescaleStore(router.app.Ctx, config.AppConfig.Timescale.URL)
	machineController := machine_controller.NewMachineController(router.app.Store, router.app.Ctx, router.logger, machineTimescaleStore)

	machinesGroup := fuego.Group(server, apiV1.Path+"/machines", option.Tags("Machines"))
	router.applyMiddleware(machinesGroup, MiddlewareConfig{
		RBAC:         true,
		Audit:        true,
		ResourceName: "machine",
	})
	router.RegisterMachinesRoutes(machinesGroup, machineController)
	router.RegisterMachineRegistrationRoutes(machinesGroup, machineController)

	machinesOpsGroup := fuego.Group(server, apiV1.Path+"/machines", option.Tags("Machines"))
	fuego.Use(machinesOpsGroup, middleware.ServerIDMiddleware)
	router.applyMiddleware(machinesOpsGroup, MiddlewareConfig{
		RBAC:         true,
		Audit:        true,
		ResourceName: "machine",
	})
	router.RegisterMachineRoutes(machinesOpsGroup, machineController)

	machinesBillingGroup := fuego.Group(server, apiV1.Path+"/machines", option.Tags("Machines"))
	fuego.Use(machinesBillingGroup, middleware.ServerIDMiddleware)
	router.applyMiddleware(machinesBillingGroup, MiddlewareConfig{
		RBAC:         false,
		Audit:        true,
		ResourceName: "machine",
	})
	router.RegisterMachineBillingRoutes(machinesBillingGroup, machineController)

	trailController := machine_controller.NewTrailController(router.app.Store, router.app.Ctx, router.logger, router.cache)
	trailGroup := fuego.Group(server, apiV1.Path+"/machines/trial", option.Tags("Machines"))
	router.applyMiddleware(trailGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "trail",
		Audit:        true,
		ResourceName: "trail",
	})
	router.RegisterTrailRoutes(trailGroup, trailController)

	mcpCtrl := mcpController.NewMCPController(router.app.Store, router.app.Ctx, router.logger)
	mcpGroup := fuego.Group(server, apiV1.Path+"/mcp", option.Tags("MCP"))
	router.applyMiddleware(mcpGroup, MiddlewareConfig{
		RBAC:         true,
		FeatureFlag:  "mcp",
		Audit:        true,
		ResourceName: "mcp",
	})
	router.RegisterMCPRoutes(mcpGroup, mcpCtrl)
}

func (router *Router) createAuthController() *auth.AuthController {
	userStorage := &user_storage.UserStorage{DB: router.app.Store.DB, Ctx: router.app.Ctx, Logger: &router.logger}
	authService := auth_service.NewAuthService(userStorage, userStorage.DB, router.logger, router.app.Ctx, config.AppConfig.Redis.URL)
	return auth.NewAuthController(router.app.Ctx, router.logger, authService)
}

func (router *Router) createFeatureFlagController() *feature_flags_controller.FeatureFlagController {
	featureFlagStorage := &feature_flags_storage.FeatureFlagStorage{DB: router.app.Store.DB, Ctx: router.app.Ctx, Logger: &router.logger}
	featureFlagService := feature_flags_service.NewFeatureFlagService(featureFlagStorage, router.logger, router.app.Ctx)
	return feature_flags_controller.NewFeatureFlagController(featureFlagService, router.logger, router.app.Ctx, router.cache)
}
