package routes

import (
	"log"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-fuego/fuego"
	"github.com/joho/godotenv"
	"github.com/raghavyuva/nixopus-api/internal/cache"
	"github.com/raghavyuva/nixopus-api/internal/config"
	audit "github.com/raghavyuva/nixopus-api/internal/features/audit/controller"
	auth "github.com/raghavyuva/nixopus-api/internal/features/auth/controller"
	auth_service "github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	auth_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	billing "github.com/raghavyuva/nixopus-api/internal/features/billing/controller"
	container "github.com/raghavyuva/nixopus-api/internal/features/container/controller"
	deploy "github.com/raghavyuva/nixopus-api/internal/features/deploy/controller"
	domain "github.com/raghavyuva/nixopus-api/internal/features/domain/controller"
	extension "github.com/raghavyuva/nixopus-api/internal/features/extension/controller"
	feature_flags_controller "github.com/raghavyuva/nixopus-api/internal/features/feature-flags/controller"
	feature_flags_service "github.com/raghavyuva/nixopus-api/internal/features/feature-flags/service"
	feature_flags_storage "github.com/raghavyuva/nixopus-api/internal/features/feature-flags/storage"
	file_manager "github.com/raghavyuva/nixopus-api/internal/features/file-manager/controller"
	githubConnector "github.com/raghavyuva/nixopus-api/internal/features/github-connector/controller"
	healthcheck "github.com/raghavyuva/nixopus-api/internal/features/healthcheck/controller"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	notificationController "github.com/raghavyuva/nixopus-api/internal/features/notification/controller"

	// Organization packages removed - migrated to Better Auth
	update "github.com/raghavyuva/nixopus-api/internal/features/update/controller"
	update_service "github.com/raghavyuva/nixopus-api/internal/features/update/service"
	user "github.com/raghavyuva/nixopus-api/internal/features/user/controller"
	"github.com/raghavyuva/nixopus-api/internal/middleware"
	"github.com/raghavyuva/nixopus-api/internal/realtime"
	"github.com/raghavyuva/nixopus-api/internal/scheduler"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	api "github.com/raghavyuva/nixopus-api/internal/version"
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
	)
}

// setupAuthentication configures the authentication middleware
func (router *Router) setupAuthentication(server *fuego.Server) {
	fuego.Use(server, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for public routes
			if config.AppConfig.App.Environment == "development" && strings.HasPrefix(r.URL.Path, "/swagger") {
				next.ServeHTTP(w, r)
				return
			}
			// Skip auth for live deploy routes
			if strings.HasPrefix(r.URL.Path, "/api/v1/live") || strings.HasPrefix(r.URL.Path, "/ws/live") {
				next.ServeHTTP(w, r)
				return
			}
			// Skip global auth for API key routes (they have their own APIKeyAuthMiddleware)
			// API key routes use APIKeyAuthMiddleware which is applied to their route group
			// We detect API key routes by checking if they're under /deploy/application/project/add-to-family
			if strings.HasPrefix(r.URL.Path, "/api/v1/deploy/application/project/add-to-family") {
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

// SetupRoutes initializes and configures all application routes
func (router *Router) SetupRoutes() {
	// Load .env file if it exists (optional when using secret manager)
	if err := godotenv.Load(); err != nil {
		// .env file is optional when using secret manager
		log.Println("Info: .env file not found, using environment variables and secret manager")
	}

	// Save version documentation
	// Commented out - version manager creating version.json every time causing troubles
	// docs := api.NewVersionDocumentation()
	// if err := docs.Save("api/versions.json"); err != nil {
	// 	log.Printf("Warning: Failed to save version documentation: %v", err)
	// }

	// Initialize notification manager
	notificationManager := notification.NewNotificationManager(router.app.Store.DB)
	notificationManager.Start()

	// Create and configure server
	PORT := config.AppConfig.Server.Port
	server := router.createServer(PORT)
	apiV1 := api.NewVersion(api.CurrentVersion)

	// Register public routes (no auth required)
	router.registerPublicRoutes(server, apiV1, notificationManager)

	// Apply authentication middleware
	router.setupAuthentication(server)

	// Register protected routes
	router.registerProtectedRoutes(server, apiV1, notificationManager)

	log.Printf("Server starting on port %s", PORT)
	log.Printf("Swagger UI available at: http://localhost:%s/swagger/", PORT)
	server.Run()
}

// registerPublicRoutes registers routes that don't require authentication
func (router *Router) registerPublicRoutes(server *fuego.Server, apiV1 api.Version, notificationManager *notification.NotificationManager) {
	// Health routes
	healthGroup := fuego.Group(server, apiV1.Path+"/health")
	router.RegisterHealthRoutes(healthGroup)

	// Webhook routes
	deployController, err := deploy.NewDeployController(router.app.Store, router.app.Ctx, router.logger, notificationManager)
	if err != nil {
		log.Fatalf("Failed to create deploy controller: %v", err)
	}
	webhookGroup := fuego.Group(server, apiV1.Path+"/webhook")
	fuego.Post(webhookGroup, "", deployController.HandleGithubWebhook)

	// WebSocket routes
	router.RegisterWebSocketRoutes(server, deployController, router.schedulers.HealthCheck)

	router.RegisterLiveDeployRoutes(server, apiV1)

	// Stripe webhook route (public, uses signature verification)
	billingController := billing.NewBillingController(router.app.Store, router.app.Ctx, router.logger, config.AppConfig.Stripe)
	stripeWebhookGroup := fuego.Group(server, apiV1.Path+"/stripe")
	fuego.Post(stripeWebhookGroup, "/webhook", billingController.HandleWebhook)

	// Public auth routes
	authController := router.createAuthController(notificationManager)
	authGroup := fuego.Group(server, apiV1.Path+"/auth")
	router.RegisterAuthRoutes(authGroup, authController)
}

// registerProtectedRoutes registers routes that require authentication
func (router *Router) registerProtectedRoutes(server *fuego.Server, apiV1 api.Version, notificationManager *notification.NotificationManager) {
	// Protected auth routes
	authController := router.createAuthController(notificationManager)
	authProtectedGroup := fuego.Group(server, apiV1.Path+"/auth")
	router.RegisterAuthenticatedAuthRoutes(authProtectedGroup, authController)

	// User routes
	userController := user.NewUserController(router.app.Store, router.app.Ctx, router.logger, router.cache)
	userGroup := fuego.Group(server, apiV1.Path+"/user")
	router.applyMiddleware(userGroup, MiddlewareConfig{RBAC: false, Audit: false, ResourceName: "user"})
	router.RegisterUserRoutes(userGroup, userController)

	// Domain routes
	domainController := domain.NewDomainsController(router.app.Store, router.app.Ctx, router.logger, notificationManager)
	domainGroup := fuego.Group(server, apiV1.Path+"/domain")
	domainsAllGroup := fuego.Group(server, apiV1.Path+"/domains")
	domainMiddleware := MiddlewareConfig{RBAC: false, FeatureFlag: "domain", Audit: false, ResourceName: "domain"}
	router.applyMiddleware(domainGroup, domainMiddleware)
	router.applyMiddleware(domainsAllGroup, domainMiddleware)
	router.RegisterDomainRoutes(domainGroup, domainsAllGroup, domainController)

	// GitHub connector routes
	githubConnectorController := githubConnector.NewGithubConnectorController(router.app.Store, router.app.Ctx, router.logger, notificationManager)
	githubConnectorGroup := fuego.Group(server, apiV1.Path+"/github-connector")
	router.applyMiddleware(githubConnectorGroup, MiddlewareConfig{
		RBAC:         false,
		FeatureFlag:  "github_connector",
		Audit:        false,
		ResourceName: "github-connector",
	})
	router.RegisterGithubConnectorRoutes(githubConnectorGroup, githubConnectorController)

	// Notification routes
	notifController := notificationController.NewNotificationController(router.app.Store, router.app.Ctx, router.logger, notificationManager)
	notificationGroup := fuego.Group(server, apiV1.Path+"/notification")
	router.applyMiddleware(notificationGroup, MiddlewareConfig{
		RBAC:         false,
		FeatureFlag:  "notifications",
		Audit:        false,
		ResourceName: "notification",
	})
	router.RegisterNotificationRoutes(notificationGroup, notifController)

	// Organization routes - migrated to Better Auth
	// Organization management is now handled by Better Auth in octoagent

	// File manager routes
	fileManagerController := file_manager.NewFileManagerController(router.app.Ctx, router.logger, notificationManager)
	fileManagerGroup := fuego.Group(server, apiV1.Path+"/file-manager")
	router.applyMiddleware(fileManagerGroup, MiddlewareConfig{
		RBAC:         false,
		FeatureFlag:  "file_manager",
		Audit:        false,
		ResourceName: "file-manager",
	})
	router.RegisterFileManagerRoutes(fileManagerGroup, fileManagerController)

	// Deploy routes
	deployController, err := deploy.NewDeployController(router.app.Store, router.app.Ctx, router.logger, notificationManager)
	if err != nil {
		log.Fatalf("Failed to create deploy controller: %v", err)
	}
	deployGroup := fuego.Group(server, apiV1.Path+"/deploy")
	router.applyMiddleware(deployGroup, MiddlewareConfig{
		RBAC:         false,
		FeatureFlag:  "deploy",
		Audit:        false,
		ResourceName: "deploy",
	})
	router.RegisterDeployRoutes(deployGroup, deployController)

	// Deploy API key routes (for CLI - uses API key authentication instead of session)
	deployAPIKeyGroup := fuego.Group(server, apiV1.Path+"/deploy")
	fuego.Use(deployAPIKeyGroup, func(next http.Handler) http.Handler {
		return middleware.APIKeyAuthMiddleware(next, router.app)
	})
	router.applyMiddleware(deployAPIKeyGroup, MiddlewareConfig{
		RBAC:         false,
		FeatureFlag:  "deploy",
		Audit:        false,
		ResourceName: "deploy",
	})
	router.RegisterDeployAPIKeyRoutes(deployAPIKeyGroup, deployController)

	// Audit routes
	auditController := audit.NewAuditController(router.app.Store.DB, router.app.Ctx, router.logger)
	auditGroup := fuego.Group(server, apiV1.Path+"/audit")
	router.applyMiddleware(auditGroup, MiddlewareConfig{RBAC: false, FeatureFlag: "audit", Audit: false, ResourceName: "audit"})
	router.RegisterAuditRoutes(auditGroup, auditController)

	// Update routes
	updateService := update_service.NewUpdateService(router.app, &router.logger, router.app.Ctx)
	updateController := update.NewUpdateController(updateService, &router.logger)
	updateGroup := fuego.Group(server, apiV1.Path+"/update")
	router.RegisterUpdateRoutes(updateGroup, updateController)

	// Feature flag routes
	featureFlagController := router.createFeatureFlagController()
	featureFlagReadGroup := fuego.Group(server, apiV1.Path+"/feature-flags")
	featureFlagWriteGroup := fuego.Group(server, apiV1.Path+"/feature-flags")
	router.applyMiddleware(featureFlagWriteGroup, MiddlewareConfig{RBAC: false, Audit: false, ResourceName: "feature_flags"})
	router.RegisterFeatureFlagRoutes(featureFlagReadGroup, featureFlagWriteGroup, featureFlagController)

	// Container routes
	containerController, err := container.NewContainerController(router.app.Store, router.app.Ctx, router.logger, notificationManager)
	if err != nil {
		log.Fatalf("Failed to create container controller: %v", err)
	}
	containerGroup := fuego.Group(server, apiV1.Path+"/container")
	router.applyMiddleware(containerGroup, MiddlewareConfig{
		RBAC:         false,
		FeatureFlag:  "container",
		Audit:        false,
		ResourceName: "container",
	})
	router.RegisterContainerRoutes(containerGroup, containerController)

	// Health check routes
	healthCheckController := healthcheck.NewHealthCheckController(router.app.Store, router.app.Ctx, router.logger)
	healthCheckGroup := fuego.Group(server, apiV1.Path+"/healthcheck")
	router.applyMiddleware(healthCheckGroup, MiddlewareConfig{
		RBAC:         false,
		FeatureFlag:  "deploy",
		Audit:        false,
		ResourceName: "healthcheck",
	})
	router.RegisterHealthCheckRoutes(healthCheckGroup, healthCheckController)

	// Billing routes
	billingController := billing.NewBillingController(router.app.Store, router.app.Ctx, router.logger, config.AppConfig.Stripe)
	billingGroup := fuego.Group(server, apiV1.Path+"/billing")
	router.applyMiddleware(billingGroup, MiddlewareConfig{
		RBAC:         false,
		Audit:        false,
		ResourceName: "billing",
	})
	router.RegisterBillingRoutes(billingGroup, billingController)

	// Extension routes
	extensionController := extension.NewExtensionsController(router.app.Store, router.app.Ctx, router.logger)
	extensionGroup := fuego.Group(server, apiV1.Path+"/extensions")
	router.applyMiddleware(extensionGroup, MiddlewareConfig{
		RBAC:         false,
		FeatureFlag:  "extension",
		Audit:        false,
		ResourceName: "extension",
	})
	router.RegisterExtensionRoutes(extensionGroup, extensionController)
}

// createAuthController creates and returns an auth controller
// Only used for API key management - Better Auth handles authentication
func (router *Router) createAuthController(notificationManager *notification.NotificationManager) *auth.AuthController {
	userStorage := &user_storage.UserStorage{DB: router.app.Store.DB, Ctx: router.app.Ctx}
	// Create minimal auth service (only GetUserByEmail is used)
	authService := auth_service.NewAuthService(userStorage, router.logger, router.app.Ctx)

	// Create API key service
	apiKeyStorage := auth_storage.APIKeyStorage{DB: router.app.Store.DB, Ctx: router.app.Ctx}
	apiKeyService := auth_service.NewAPIKeyService(apiKeyStorage, router.logger)

	return auth.NewAuthController(router.app.Ctx, router.logger, notificationManager, *authService, apiKeyService, router.app.Store)
}

// createFeatureFlagController creates and returns a feature flag controller
func (router *Router) createFeatureFlagController() *feature_flags_controller.FeatureFlagController {
	featureFlagStorage := &feature_flags_storage.FeatureFlagStorage{DB: router.app.Store.DB, Ctx: router.app.Ctx}
	featureFlagService := feature_flags_service.NewFeatureFlagService(featureFlagStorage, router.logger, router.app.Ctx)
	return feature_flags_controller.NewFeatureFlagController(featureFlagService, router.logger, router.app.Ctx, router.cache)
}
