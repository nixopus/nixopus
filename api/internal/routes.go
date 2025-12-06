package internal

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
	authService "github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	container "github.com/raghavyuva/nixopus-api/internal/features/container/controller"
	deploy "github.com/raghavyuva/nixopus-api/internal/features/deploy/controller"
	domain "github.com/raghavyuva/nixopus-api/internal/features/domain/controller"
	extension "github.com/raghavyuva/nixopus-api/internal/features/extension/controller"
	feature_flags_controller "github.com/raghavyuva/nixopus-api/internal/features/feature-flags/controller"
	feature_flags_service "github.com/raghavyuva/nixopus-api/internal/features/feature-flags/service"
	feature_flags_storage "github.com/raghavyuva/nixopus-api/internal/features/feature-flags/storage"
	file_manager "github.com/raghavyuva/nixopus-api/internal/features/file-manager/controller"
	githubConnector "github.com/raghavyuva/nixopus-api/internal/features/github-connector/controller"
	health "github.com/raghavyuva/nixopus-api/internal/features/health"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	notificationController "github.com/raghavyuva/nixopus-api/internal/features/notification/controller"
	organization "github.com/raghavyuva/nixopus-api/internal/features/organization/controller"
	organization_service "github.com/raghavyuva/nixopus-api/internal/features/organization/service"
	organization_storage "github.com/raghavyuva/nixopus-api/internal/features/organization/storage"

	"github.com/raghavyuva/nixopus-api/internal/features/supertokens"
	update "github.com/raghavyuva/nixopus-api/internal/features/update/controller"
	update_service "github.com/raghavyuva/nixopus-api/internal/features/update/service"
	user "github.com/raghavyuva/nixopus-api/internal/features/user/controller"
	"github.com/raghavyuva/nixopus-api/internal/middleware"
	"github.com/raghavyuva/nixopus-api/internal/realtime"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	api "github.com/raghavyuva/nixopus-api/internal/version-manager"
)

type Router struct {
	app   *storage.App
	cache *cache.Cache
}

func NewRouter(app *storage.App) *Router {
	// Initialize cache
	cache, err := cache.NewCache(config.AppConfig.Redis.URL)
	if err != nil {
		log.Fatal("Error creating redis client", err)
	}
	return &Router{
		app:   app,
		cache: cache,
	}
}

func (router *Router) Routes() {
	// Initialize SuperTokens authentication system
	supertokens.Init(router.app)
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	PORT := config.AppConfig.Server.Port

	docs := api.NewVersionDocumentation()
	if err := docs.Save("api/versions.json"); err != nil {
		log.Printf("Warning: Failed to save version documentation: %v", err)
	}

	l := logger.NewLogger()
	server := fuego.NewServer(
		fuego.WithGlobalMiddlewares(
			middleware.SupertokensCorsMiddleware,
			middleware.RecoveryMiddleware,
			middleware.CorsMiddleware,
			middleware.LoggingMiddleware,
			api.VersionMiddleware,
			api.MigrationMiddleware,
			// middleware.RateLimiter
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
		fuego.WithAddr(":"+PORT),
	)

	apiV1 := api.NewVersion(api.CurrentVersion)

	healthGroup := fuego.Group(server, apiV1.Path+"/health")
	router.BasicRoutes(healthGroup)

	notificationManager := notification.NewNotificationManager(router.app.Store.DB)
	notificationManager.Start()
	deployController := deploy.NewDeployController(router.app.Store, router.app.Ctx, l, notificationManager)

	webhookGroup := fuego.Group(server, apiV1.Path+"/webhook")
	fuego.Post(webhookGroup, "", deployController.HandleGithubWebhook)

	router.WebSocketServer(server, deployController)

	userStorage := &user_storage.UserStorage{DB: router.app.Store.DB, Ctx: router.app.Ctx}
	orgStorage := &organization_storage.OrganizationStore{DB: router.app.Store.DB, Ctx: router.app.Ctx}
	orgService := organization_service.NewOrganizationService(router.app.Store, router.app.Ctx, l, orgStorage, router.cache)
	authService := authService.NewAuthService(userStorage, l, orgService, router.app.Ctx)
	authController := auth.NewAuthController(router.app.Ctx, l, notificationManager, *authService)
	authGroup := fuego.Group(server, apiV1.Path+"/auth")
	router.AuthRoutes(authController, authGroup)

	// Auth middleware for development environment will be bypassed for swagger UI
	fuego.Use(server, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.AppConfig.App.Environment == "development" && strings.HasPrefix(r.URL.Path, "/swagger") {
				next.ServeHTTP(w, r)
				return
			}
			middleware.AuthMiddleware(next, router.app, router.cache).ServeHTTP(w, r)
		})
	})

	// Remove generic audit middleware - will be applied per route group with explicit resource types

	authProtectedGroup := fuego.Group(server, apiV1.Path+"/auth")
	router.AuthenticatedAuthRoutes(authProtectedGroup, authController)

	userController := user.NewUserController(router.app.Store, router.app.Ctx, l, router.cache)
	userGroup := fuego.Group(server, apiV1.Path+"/user")
	fuego.Use(userGroup, func(next http.Handler) http.Handler {
		return middleware.AuditMiddleware(next, router.app, l, "user")
	})
	router.UserRoutes(userGroup, userController)

	domainController := domain.NewDomainsController(router.app.Store, router.app.Ctx, l, notificationManager)
	domainGroup := fuego.Group(server, apiV1.Path+"/domain")
	domainsAllGroup := fuego.Group(server, apiV1.Path+"/domains")
	fuego.Use(domainGroup, func(next http.Handler) http.Handler {
		return middleware.RBACMiddleware(next, router.app, "domain")
	})
	fuego.Use(domainsAllGroup, func(next http.Handler) http.Handler {
		return middleware.RBACMiddleware(next, router.app, "domain")
	})
	fuego.Use(domainGroup, func(next http.Handler) http.Handler {
		return middleware.FeatureFlagMiddleware(next, router.app, "domain", router.cache)
	})
	fuego.Use(domainsAllGroup, func(next http.Handler) http.Handler {
		return middleware.FeatureFlagMiddleware(next, router.app, "domain", router.cache)
	})
	fuego.Use(domainGroup, func(next http.Handler) http.Handler {
		return middleware.AuditMiddleware(next, router.app, l, "domain")
	})
	fuego.Use(domainsAllGroup, func(next http.Handler) http.Handler {
		return middleware.AuditMiddleware(next, router.app, l, "domain")
	})
	router.DomainRoutes(domainGroup, domainsAllGroup, domainController)

	githubConnectorController := githubConnector.NewGithubConnectorController(router.app.Store, router.app.Ctx, l, notificationManager)
	githubConnectorGroup := fuego.Group(server, apiV1.Path+"/github-connector")
	fuego.Use(githubConnectorGroup, func(next http.Handler) http.Handler {
		return middleware.RBACMiddleware(next, router.app, "github-connector")
	})
	fuego.Use(githubConnectorGroup, func(next http.Handler) http.Handler {
		return middleware.FeatureFlagMiddleware(next, router.app, "github_connector", router.cache)
	})
	fuego.Use(githubConnectorGroup, func(next http.Handler) http.Handler {
		return middleware.AuditMiddleware(next, router.app, l, "github-connector")
	})
	router.GithubConnectorRoutes(githubConnectorGroup, githubConnectorController)

	notifController := notificationController.NewNotificationController(router.app.Store, router.app.Ctx, l, notificationManager)
	notificationGroup := fuego.Group(server, apiV1.Path+"/notification")
	fuego.Use(notificationGroup, func(next http.Handler) http.Handler {
		return middleware.RBACMiddleware(next, router.app, "notification")
	})
	fuego.Use(notificationGroup, func(next http.Handler) http.Handler {
		return middleware.FeatureFlagMiddleware(next, router.app, "notifications", router.cache)
	})
	fuego.Use(notificationGroup, func(next http.Handler) http.Handler {
		return middleware.AuditMiddleware(next, router.app, l, "notification")
	})
	router.NotificationRoutes(notificationGroup, notifController)

	organizationController := organization.NewOrganizationsController(router.app.Store, router.app.Ctx, l, notificationManager, router.cache)
	organizationGroup := fuego.Group(server, apiV1.Path+"/organizations")
	fuego.Use(organizationGroup, func(next http.Handler) http.Handler {
		return middleware.RBACMiddleware(next, router.app, "organization")
	})
	fuego.Use(organizationGroup, func(next http.Handler) http.Handler {
		return middleware.AuditMiddleware(next, router.app, l, "organization")
	})
	router.OrganizationRoutes(organizationGroup, organizationController)

	fileManagerController := file_manager.NewFileManagerController(router.app.Ctx, l, notificationManager)
	fileManagerGroup := fuego.Group(server, apiV1.Path+"/file-manager")
	fuego.Use(fileManagerGroup, func(next http.Handler) http.Handler {
		return middleware.RBACMiddleware(next, router.app, "file-manager")
	})
	fuego.Use(fileManagerGroup, func(next http.Handler) http.Handler {
		return middleware.FeatureFlagMiddleware(next, router.app, "file_manager", router.cache)
	})
	fuego.Use(fileManagerGroup, func(next http.Handler) http.Handler {
		return middleware.AuditMiddleware(next, router.app, l, "file-manager")
	})
	router.FileManagerRoutes(fileManagerGroup, fileManagerController)

	deployGroup := fuego.Group(server, apiV1.Path+"/deploy")
	fuego.Use(deployGroup, func(next http.Handler) http.Handler {
		return middleware.RBACMiddleware(next, router.app, "deploy")
	})
	fuego.Use(deployGroup, func(next http.Handler) http.Handler {
		return middleware.FeatureFlagMiddleware(next, router.app, "deploy", router.cache)
	})
	fuego.Use(deployGroup, func(next http.Handler) http.Handler {
		return middleware.AuditMiddleware(next, router.app, l, "deploy")
	})
	router.DeployRoutes(deployGroup, deployController)

	auditController := audit.NewAuditController(router.app.Store.DB, router.app.Ctx, l)
	auditGroup := fuego.Group(server, apiV1.Path+"/audit")
	fuego.Use(auditGroup, func(next http.Handler) http.Handler {
		return middleware.RBACMiddleware(next, router.app, "audit")
	})
	fuego.Use(auditGroup, func(next http.Handler) http.Handler {
		return middleware.FeatureFlagMiddleware(next, router.app, "audit", router.cache)
	})
	router.AuditRoutes(auditGroup, auditController)

	updateService := update_service.NewUpdateService(router.app, &l, router.app.Ctx)
	updateController := update.NewUpdateController(updateService, &l)
	updateGroup := fuego.Group(server, apiV1.Path+"/update")
	router.UpdateRoutes(updateGroup, updateController)

	featureFlagStorage := &feature_flags_storage.FeatureFlagStorage{DB: router.app.Store.DB, Ctx: router.app.Ctx}
	featureFlagService := feature_flags_service.NewFeatureFlagService(featureFlagStorage, l, router.app.Ctx)
	featureFlagController := feature_flags_controller.NewFeatureFlagController(featureFlagService, l, router.app.Ctx, router.cache)

	// We need to allow everyone to read feature flags
	featureFlagReadGroup := fuego.Group(server, apiV1.Path+"/feature-flags")
	featureFlagWriteGroup := fuego.Group(server, apiV1.Path+"/feature-flags")

	// Apply RBAC middleware only to write operations (update feature flags)
	fuego.Use(featureFlagWriteGroup, func(next http.Handler) http.Handler {
		return middleware.RBACMiddleware(next, router.app, "feature_flags")
	})

	router.FeatureFlagRoutes(featureFlagReadGroup, featureFlagWriteGroup, featureFlagController)

	containerController := container.NewContainerController(router.app.Store, router.app.Ctx, l, notificationManager)
	containerGroup := fuego.Group(server, apiV1.Path+"/container")
	fuego.Use(containerGroup, func(next http.Handler) http.Handler {
		return middleware.RBACMiddleware(next, router.app, "container")
	})
	fuego.Use(containerGroup, func(next http.Handler) http.Handler {
		return middleware.FeatureFlagMiddleware(next, router.app, "container", router.cache)
	})
	fuego.Use(containerGroup, func(next http.Handler) http.Handler {
		return middleware.AuditMiddleware(next, router.app, l, "container")
	})
	router.ContainerRoutes(containerGroup, containerController)

	extensionController := extension.NewExtensionsController(router.app.Store, router.app.Ctx, l)
	extensionGroup := fuego.Group(server, apiV1.Path+"/extensions")
	fuego.Use(extensionGroup, func(next http.Handler) http.Handler {
		return middleware.RBACMiddleware(next, router.app, "extension")
	})
	fuego.Use(extensionGroup, func(next http.Handler) http.Handler {
		return middleware.AuditMiddleware(next, router.app, l, "extension")
	})
	router.ExtensionRoutes(extensionGroup, extensionController)

	log.Printf("Server starting on port %s", PORT)
	log.Printf("Swagger UI available at: http://localhost:%s/swagger/", PORT)
	server.Run()
}

func (s *Router) BasicRoutes(fs *fuego.Server) {
	fuego.Get(fs, "", health.HealthCheck)
	versionGroup := fuego.Group(fs, "/versions")
	fuego.Get(versionGroup, "", func(c fuego.ContextNoBody) (interface{}, error) {
		docs := api.NewVersionDocumentation()
		if err := docs.Load("api/versions.json"); err != nil {
			return nil, err
		}
		return docs, nil
	})
}

func (router *Router) WebSocketServer(f *fuego.Server, deployController *deploy.DeployController) {
	wsServer, err := realtime.NewSocketServer(deployController, router.app.Store.DB, router.app.Ctx)
	if err != nil {
		log.Fatal(err)
	}
	wsHandler := func(c fuego.ContextNoBody) (interface{}, error) {
		log.Printf("WebSocket connection attempt from: %s", c.Request().RemoteAddr)

		wsServer.HandleHTTP(c.Response(), c.Request())
		return nil, nil
	}

	fuego.Get(f, "/ws", wsHandler)
}

func (router *Router) AuthRoutes(authController *auth.AuthController, s *fuego.Server) {
	fuego.Get(s, "/is-admin-registered", authController.IsAdminRegistered)
}

func (router *Router) AuthenticatedAuthRoutes(s *fuego.Server, authController *auth.AuthController) {
	fuego.Post(s, "/logout", authController.Logout)
	fuego.Post(s, "/send-verification-email", authController.SendVerificationEmail)
	fuego.Get(s, "/verify-email", authController.VerifyEmail)
	fuego.Post(s, "/create-user", authController.CreateUser)
	fuego.Post(s, "/setup-2fa", authController.SetupTwoFactor)
	fuego.Post(s, "/verify-2fa", authController.VerifyTwoFactor)
	fuego.Post(s, "/disable-2fa", authController.DisableTwoFactor)
	fuego.Post(s, "/2fa-login", authController.TwoFactorLogin)
}

func (router *Router) UserRoutes(s *fuego.Server, userController *user.UserController) {
	fuego.Get(s, "", userController.GetUserDetails)
	fuego.Patch(s, "/name", userController.UpdateUserName)
	fuego.Get(s, "/organizations", userController.GetUserOrganizations)
	fuego.Get(s, "/settings", userController.GetSettings)
	fuego.Patch(s, "/settings/font", userController.UpdateFont)
	fuego.Patch(s, "/settings/theme", userController.UpdateTheme)
	fuego.Patch(s, "/settings/language", userController.UpdateLanguage)
	fuego.Patch(s, "/settings/auto-update", userController.UpdateAutoUpdate)
	fuego.Patch(s, "/avatar", userController.UpdateAvatar)
}

func (router *Router) NotificationRoutes(s *fuego.Server, notificationController *notificationController.NotificationController) {
	smtpGroup := fuego.Group(s, "/smtp")
	fuego.Post(smtpGroup, "", notificationController.AddSmtp)
	fuego.Get(smtpGroup, "", notificationController.GetSmtp)
	fuego.Put(smtpGroup, "", notificationController.UpdateSmtp)
	fuego.Delete(smtpGroup, "", notificationController.DeleteSmtp)

	preferenceGroup := fuego.Group(s, "/preferences")
	fuego.Post(preferenceGroup, "", notificationController.UpdatePreference)
	fuego.Get(preferenceGroup, "", notificationController.GetPreferences)

	webhookGroup := fuego.Group(s, "/webhook")
	fuego.Post(webhookGroup, "", notificationController.CreateWebhookConfig)
	fuego.Get(webhookGroup, "/{type}", notificationController.GetWebhookConfig)
	fuego.Put(webhookGroup, "", notificationController.UpdateWebhookConfig)
	fuego.Delete(webhookGroup, "", notificationController.DeleteWebhookConfig)
}

func (router *Router) DomainRoutes(s *fuego.Server, domainsGroup *fuego.Server, domainController *domain.DomainsController) {
	fuego.Post(s, "", domainController.CreateDomain)
	fuego.Put(s, "", domainController.UpdateDomain)
	fuego.Delete(s, "", domainController.DeleteDomain)
	fuego.Get(s, "/generate", domainController.GenerateRandomSubDomain)
	fuego.Get(domainsGroup, "", domainController.GetDomains)
}

func (router *Router) GithubConnectorRoutes(s *fuego.Server, githubConnectorController *githubConnector.GithubConnectorController) {
	fuego.Post(s, "", githubConnectorController.CreateGithubConnector)
	fuego.Put(s, "", githubConnectorController.UpdateGithubConnectorRequest)
	fuego.Delete(s, "", githubConnectorController.DeleteGithubConnector)
	fuego.Get(s, "/all", githubConnectorController.GetGithubConnectors)
	fuego.Get(s, "/repositories", githubConnectorController.GetGithubRepositories)
	fuego.Post(s, "/repository/branches", githubConnectorController.GetGithubRepositoryBranches)
}

func (router *Router) DeployRoutes(f *fuego.Server, deployController *deploy.DeployController) {
	fuego.Get(f, "/applications", deployController.GetApplications)
	deploy_application_group := fuego.Group(f, "/application")
	router.DeployApplicationRoutes(deploy_application_group, deployController)
}

func (router *Router) DeployApplicationRoutes(f *fuego.Server, deployController *deploy.DeployController) {
	fuego.Post(f, "", deployController.HandleDeploy)
	fuego.Get(f, "", deployController.GetApplicationById)
	fuego.Delete(f, "", deployController.DeleteApplication)
	fuego.Put(f, "", deployController.UpdateApplication)
	fuego.Post(f, "/redeploy", deployController.ReDeployApplication)
	fuego.Get(f, "/deployments/{deployment_id}", deployController.GetDeploymentById)
	fuego.Post(f, "/rollback", deployController.HandleRollback)
	fuego.Post(f, "/restart", deployController.HandleRestart)
	fuego.Get(f, "/logs/{application_id}", deployController.GetLogs)
	fuego.Get(f, "/deployments/{deployment_id}/logs", deployController.GetDeploymentLogs)
	fuego.Get(f, "/deployments", deployController.GetApplicationDeployments)
}

func (router *Router) FileManagerRoutes(f *fuego.Server, fileManagerController *file_manager.FileManagerController) {
	fuego.Get(f, "", fileManagerController.ListFiles)
	fuego.Post(f, "/create-directory", fileManagerController.CreateDirectory)
	fuego.Post(f, "/move-directory", fileManagerController.MoveDirectory)
	fuego.Post(f, "/copy-directory", fileManagerController.CopyDirectory)
	fuego.Post(f, "/upload", fileManagerController.UploadFile)
	fuego.Delete(f, "/delete-directory", fileManagerController.DeleteDirectory)
}

func (router *Router) AuditRoutes(s *fuego.Server, auditController *audit.AuditController) {
	fuego.Get(s, "/logs", auditController.GetRecentAuditLogs)
}

func (router *Router) UpdateRoutes(s *fuego.Server, updateController *update.UpdateController) {
	fuego.Get(s, "/check", updateController.CheckForUpdates)
	fuego.Post(s, "", updateController.PerformUpdate)
}

func (router *Router) FeatureFlagRoutes(readGroup *fuego.Server, writeGroup *fuego.Server, featureFlagController *feature_flags_controller.FeatureFlagController) {
	fuego.Get(readGroup, "", featureFlagController.GetFeatureFlags)
	fuego.Put(writeGroup, "", featureFlagController.UpdateFeatureFlag)
	fuego.Get(readGroup, "/check", featureFlagController.IsFeatureEnabled)
}

func (router *Router) ContainerRoutes(s *fuego.Server, containerController *container.ContainerController) {
	fuego.Get(s, "", containerController.ListContainers)
	fuego.Get(s, "/{container_id}", containerController.GetContainer)
	fuego.Delete(s, "/{container_id}", containerController.RemoveContainer)
	fuego.Post(s, "/{container_id}/start", containerController.StartContainer)
	fuego.Post(s, "/{container_id}/stop", containerController.StopContainer)
	fuego.Post(s, "/{container_id}/restart", containerController.RestartContainer)
	fuego.Post(s, "/{container_id}/logs", containerController.GetContainerLogs)
	fuego.Post(s, "/prune/build-cache", containerController.PruneBuildCache)
	fuego.Post(s, "/prune/images", containerController.PruneImages)
	fuego.Post(s, "/images", containerController.ListImages)
}

func (router *Router) ExtensionRoutes(s *fuego.Server, extensionController *extension.ExtensionsController) {
	fuego.Get(s, "", extensionController.GetExtensions)
	fuego.Get(s, "/categories", extensionController.GetCategories)
	fuego.Get(s, "/{id}", extensionController.GetExtension)
	fuego.Get(s, "/by-extension-id/{extension_id}", extensionController.GetExtensionByExtensionID)
	fuego.Get(s, "/by-extension-id/{extension_id}/executions", extensionController.ListExecutionsByExtensionID)
	fuego.Post(s, "/{extension_id}/run", extensionController.RunExtension)
	fuego.Post(s, "/execution/{execution_id}/cancel", extensionController.CancelExecution)
	fuego.Get(s, "/execution/{execution_id}", extensionController.GetExecution)
	fuego.Get(s, "/execution/{execution_id}/logs", extensionController.ListExecutionLogs)
	fuego.Post(s, "/{extension_id}/fork", extensionController.ForkExtension)
	fuego.Delete(s, "/{id}", extensionController.DeleteFork)
}

func (router *Router) OrganizationRoutes(f *fuego.Server, organizationController *organization.OrganizationsController) {
	fuego.Get(f, "/users", organizationController.GetOrganizationUsers)
	fuego.Post(f, "/remove-user", organizationController.RemoveUserFromOrganization)
	fuego.Post(f, "/update-user-role", organizationController.UpdateUserRole)
	fuego.Put(f, "", organizationController.UpdateOrganization)
	fuego.Post(f, "", organizationController.CreateOrganization)
	fuego.Delete(f, "", organizationController.DeleteOrganization)
	fuego.Get(f, "", organizationController.GetOrganization)
	fuego.Get(f, "/all", organizationController.GetOrganizations)
	fuego.Post(f, "/invite/send", organizationController.SendInvite)
	fuego.Post(f, "/invite/resend", organizationController.ResendInvite)
}
