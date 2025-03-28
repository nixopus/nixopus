package internal

import (
	"log"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/go-fuego/fuego/param"
	"github.com/gorilla/mux"
	auth "github.com/raghavyuva/nixopus-api/internal/features/auth/controller"
	deploy "github.com/raghavyuva/nixopus-api/internal/features/deploy/controller"
	domain "github.com/raghavyuva/nixopus-api/internal/features/domain/controller"
	file_manager "github.com/raghavyuva/nixopus-api/internal/features/file-manager/controller"
	githubConnector "github.com/raghavyuva/nixopus-api/internal/features/github-connector/controller"
	health "github.com/raghavyuva/nixopus-api/internal/features/health"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	notificationController "github.com/raghavyuva/nixopus-api/internal/features/notification/controller"
	organization "github.com/raghavyuva/nixopus-api/internal/features/organization/controller"
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

func (router *Router) Routes() *mux.Router {
	r := mux.NewRouter()
	l := logger.NewLogger()
	server := fuego.NewServer(
		fuego.WithAddr("localhost:8080"),
		fuego.WithGlobalMiddlewares(
			middleware.CorsMiddleware,
			middleware.LoggingMiddleware,
			middleware.RateLimiter),
		// fuego.WithoutLogger(),
	)

	healthGroup := fuego.Group(server, "/api/v1/health")
	router.BasicRoutes(healthGroup)

	notificationManager := notification.NewNotificationManager(notification.NewNotificationChannels(), router.app.Store.DB)
	notificationManager.Start()

	authController := auth.NewAuthController(router.app.Store, router.app.Ctx, l, notificationManager)
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
	organizationGroup := fuego.Group(s, "/organization")
	router.OrganizationRoutes(organizationGroup, organizationController)

	fileManagerController := file_manager.NewFileManagerController(router.app.Ctx, l, notificationManager)
	fileManagerGroup := fuego.Group(s, "/file-manager")
	fuego.Use(fileManagerGroup, middleware.IsAdmin)
	router.FileManagerRoutes(fileManagerGroup, fileManagerController)

	deployController := deploy.NewDeployController(router.app.Store, router.app.Ctx, l, notificationManager)
	deployGroup := fuego.Group(s, "/deploy")
	router.DeployRoutes(deployGroup, deployController)

	router.WebSocketServer(r, deployController)

	server.Run()
	return r
}

func (s *Router) BasicRoutes(fs *fuego.Server) {
	fuego.Get(fs, "", health.HealthCheck)
}

func (router *Router) WebSocketServer(r *mux.Router, deployController *deploy.DeployController) {
	wsServer, err := realtime.NewSocketServer(deployController, router.app.Store.DB, router.app.Ctx)
	if err != nil {
		log.Fatal(err)
	}

	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsServer.HandleHTTP(w, r)
	})
}

// these routes are public routes
func (router *Router) AuthRoutes(authController *auth.AuthController, s *fuego.Server) {
	//register route is disabled for now (we do not have register seperately either the one who installs it, or the one who is added by admin)
	// authApi.HandleFunc("/register", authController.Register).Methods("POST", "OPTIONS")
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
}

func (router *Router) OrganizationRoutes(f *fuego.Server, organizationController *organization.OrganizationsController) {
	fuego.Get(f, "/users", organizationController.GetOrganizationUsers)
}
