package internal

import (
	"log"
	"net/http"
	"os"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/go-fuego/fuego/param"
	"github.com/joho/godotenv"
	auth "github.com/raghavyuva/nixopus-api/internal/features/auth/controller"
	authService "github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	deploy "github.com/raghavyuva/nixopus-api/internal/features/deploy/controller"
	domain "github.com/raghavyuva/nixopus-api/internal/features/domain/controller"
	file_manager "github.com/raghavyuva/nixopus-api/internal/features/file-manager/controller"
	githubConnector "github.com/raghavyuva/nixopus-api/internal/features/github-connector/controller"
	health "github.com/raghavyuva/nixopus-api/internal/features/health"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	notificationController "github.com/raghavyuva/nixopus-api/internal/features/notification/controller"
	organization "github.com/raghavyuva/nixopus-api/internal/features/organization/controller"
	organization_service "github.com/raghavyuva/nixopus-api/internal/features/organization/service"
	organization_storage "github.com/raghavyuva/nixopus-api/internal/features/organization/storage"
	permissions_service "github.com/raghavyuva/nixopus-api/internal/features/permission/service"
	permissions_storage "github.com/raghavyuva/nixopus-api/internal/features/permission/storage"
	role_service "github.com/raghavyuva/nixopus-api/internal/features/role/service"
	role_storage "github.com/raghavyuva/nixopus-api/internal/features/role/storage"
	user "github.com/raghavyuva/nixopus-api/internal/features/user/controller"
	"github.com/raghavyuva/nixopus-api/internal/middleware"
	"github.com/raghavyuva/nixopus-api/internal/realtime"
	"github.com/raghavyuva/nixopus-api/internal/storage"
)

type Router struct {
	app *storage.App
}

func NewRouter(app *storage.App) *Router {
	return &Router{
		app: app,
	}
}

func (router *Router) Routes() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	PORT := os.Getenv("PORT")

	log.Printf("open port %s", PORT)

	l := logger.NewLogger()
	server := fuego.NewServer(
		fuego.WithGlobalMiddlewares(
			middleware.CorsMiddleware,
			middleware.LoggingMiddleware,
			// middleware.RateLimiter
		),
		fuego.WithAddr(":"+PORT),
	)

	healthGroup := fuego.Group(server, "/api/v1/health")
	router.BasicRoutes(healthGroup)

	notificationManager := notification.NewNotificationManager(notification.NewNotificationChannels(), router.app.Store.DB)
	notificationManager.Start()
	deployController := deploy.NewDeployController(router.app.Store, router.app.Ctx, l, notificationManager)
	router.WebSocketServer(server, deployController)

	userStorage := &user_storage.UserStorage{DB: router.app.Store.DB, Ctx: router.app.Ctx}
	permStorage := &permissions_storage.PermissionStorage{DB: router.app.Store.DB, Ctx: router.app.Ctx}
	roleStorage := &role_storage.RoleStorage{DB: router.app.Store.DB, Ctx: router.app.Ctx}
	orgStorage := &organization_storage.OrganizationStore{DB: router.app.Store.DB, Ctx: router.app.Ctx}
	permService := permissions_service.NewPermissionService(router.app.Store, router.app.Ctx, l, permStorage)
	roleService := role_service.NewRoleService(router.app.Store, router.app.Ctx, l, roleStorage)
	orgService := organization_service.NewOrganizationService(router.app.Store, router.app.Ctx, l, orgStorage)
	authService := authService.NewAuthService(userStorage, l, permService, roleService, orgService, router.app.Ctx)
	authController := auth.NewAuthController(router.app.Ctx, l, notificationManager, *authService)
	authGroup := fuego.Group(server, "/api/v1/auth")
	router.AuthRoutes(authController, authGroup)

	fuego.Use(server, func(next http.Handler) http.Handler {
		return middleware.AuthMiddleware(next, router.app)
	})

	s := fuego.Group(server, "/api/v1", option.Header("Authorization", "Bearer token", param.Required()))

	authProtectedGroup := fuego.Group(server, "/api/v1/auth")
	router.AuthenticatedAuthRoutes(authProtectedGroup, authController)

	userController := user.NewUserController(router.app.Store, router.app.Ctx, l)
	userGroup := fuego.Group(s, "/user")

	router.UserRoutes(userGroup, userController)

	domainController := domain.NewDomainsController(router.app.Store, router.app.Ctx, l, notificationManager)
	domainGroup := fuego.Group(s, "/domain")
	domainsAllGroup := fuego.Group(s, "/domains")
	router.DomainRoutes(domainGroup, domainsAllGroup, domainController)

	githubConnectorController := githubConnector.NewGithubConnectorController(router.app.Store, router.app.Ctx, l, notificationManager)
	githubConnectorGroup := fuego.Group(s, "/github-connector")
	router.GithubConnectorRoutes(githubConnectorGroup, githubConnectorController)

	notifController := notificationController.NewNotificationController(router.app.Store, router.app.Ctx, l, notificationManager)
	notificationGroup := fuego.Group(s, "/notification")
	router.NotificationRoutes(notificationGroup, notifController)

	organizationController := organization.NewOrganizationsController(router.app.Store, router.app.Ctx, l, notificationManager)
	organizationGroup := fuego.Group(s, "/organizations")
	router.OrganizationRoutes(organizationGroup, organizationController)

	fileManagerController := file_manager.NewFileManagerController(router.app.Ctx, l, notificationManager)
	fileManagerGroup := fuego.Group(s, "/file-manager")
	fuego.Use(fileManagerGroup, middleware.IsAdmin)
	router.FileManagerRoutes(fileManagerGroup, fileManagerController)

	deployGroup := fuego.Group(s, "/deploy")
	router.DeployRoutes(deployGroup, deployController)

	server.Run()
}

func (s *Router) BasicRoutes(fs *fuego.Server) {
	fuego.Get(fs, "", health.HealthCheck)
}

// This is a special adapter that allows using a raw http.Handler with Fuego
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

// these routes are public routes
func (router *Router) AuthRoutes(authController *auth.AuthController, s *fuego.Server) {
	//register route is disabled for now (we do not have register seperately either the one who installs it, or the one who is added by admin)
	fuego.Post(s, "/register", authController.Register)
	fuego.Post(s, "/login", authController.Login)
}

func (router *Router) AuthenticatedAuthRoutes(s *fuego.Server, authController *auth.AuthController) {
	fuego.Post(s, "/request-password-reset", authController.GeneratePasswordResetLink)
	fuego.Post(s, "/reset-password", authController.ResetPassword)
	fuego.Post(s, "/logout", authController.Logout)
	fuego.Post(s, "/send-verification-email", authController.SendVerificationEmail)
	fuego.Post(s, "/verify-email", authController.VerifyEmail)
	fuego.Post(s, "/create-user", authController.CreateUser)
	fuego.Post(s, "/refresh-token", authController.RefreshToken)
}

func (router *Router) UserRoutes(s *fuego.Server, userController *user.UserController) {
	fuego.Get(s, "", userController.GetUserDetails)
	fuego.Patch(s, "/name", userController.UpdateUserName)
	fuego.Get(s, "/organizations", userController.GetUserOrganizations)
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
	fuego.Get(s, "/all", githubConnectorController.GetGithubConnectors)
	fuego.Get(s, "/repositories", githubConnectorController.GetGithubRepositories)
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
}

func (router *Router) FileManagerRoutes(f *fuego.Server, fileManagerController *file_manager.FileManagerController) {
	fuego.Get(f, "", fileManagerController.ListFiles)
	fuego.Post(f, "/create-directory", fileManagerController.CreateDirectory)
	fuego.Post(f, "/move-directory", fileManagerController.MoveDirectory)
}

func (router *Router) OrganizationRoutes(f *fuego.Server, organizationController *organization.OrganizationsController) {
	fuego.Get(f, "/users", organizationController.GetOrganizationUsers)
}
