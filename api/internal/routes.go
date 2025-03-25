package internal

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	auth "github.com/raghavyuva/nixopus-api/internal/features/auth/controller"
	deploy "github.com/raghavyuva/nixopus-api/internal/features/deploy/controller"
	domain "github.com/raghavyuva/nixopus-api/internal/features/domain/controller"
	githubConnector "github.com/raghavyuva/nixopus-api/internal/features/github-connector/controller"
	health "github.com/raghavyuva/nixopus-api/internal/features/health"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	notificationController "github.com/raghavyuva/nixopus-api/internal/features/notification/controller"
	organization "github.com/raghavyuva/nixopus-api/internal/features/organization/controller"
	permission "github.com/raghavyuva/nixopus-api/internal/features/permission/controller"
	role "github.com/raghavyuva/nixopus-api/internal/features/role/controller"
	user "github.com/raghavyuva/nixopus-api/internal/features/user/controller"
	"github.com/raghavyuva/nixopus-api/internal/middleware"
	"github.com/raghavyuva/nixopus-api/internal/realtime"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Router struct {
	app *storage.App
}

func NewRouter(app *storage.App) *Router {
	return &Router{
		app: app,
	}
}

// @title Nixopus Documentation
// @version 1.0
// @description Api for Nixopus
// @termsOfService http://nixopus.com/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email raghav@nixopus.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your bearer token in the format **Bearer &lt;token&gt;**
func (router *Router) Routes() *mux.Router {
	r := mux.NewRouter()
	l := logger.NewLogger()
	r.Use(middleware.CorsMiddleware)
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.RateLimiter)

	router.setupBasicRoutes(r)

	router.setupSwaggerRoutes(r)

	notificationManager := notification.NewNotificationManager(notification.NewNotificationChannels(), router.app.Store.DB)
	notificationManager.Start()

	deployController := deploy.NewDeployController(router.app.Store, router.app.Ctx, l, notificationManager)

	router.setupWebSocketServer(r, deployController)

	authController := auth.NewAuthController(router.app.Store, router.app.Ctx, l, notificationManager)
	router.setupAuthRoutes(r, authController)

	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(func(next http.Handler) http.Handler {
		return middleware.AuthMiddleware(next, router.app)
	})

	router.setupAuthenticatedAuthRoutes(api, authController)

	roleController := role.NewRolesController(router.app.Store, router.app.Ctx, l)
	router.setupRoleRoutes(api, roleController)

	organizationController := organization.NewOrganizationsController(router.app.Store, router.app.Ctx, l, notificationManager)
	router.setupOrganizationRoutes(api, organizationController)

	permissionController := permission.NewPermissionController(router.app.Store, router.app.Ctx, l)
	router.setupPermissionRoutes(api, permissionController)
	router.setupRolePermissionRoutes(api, permissionController)

	userController := user.NewUserController(router.app.Store, router.app.Ctx, l)
	router.setupUserRoutes(api, userController)

	notifController := notificationController.NewNotificationController(router.app.Store, router.app.Ctx, l, notificationManager)
	router.setupNotificationRoutes(api, notifController)

	domainController := domain.NewDomainsController(router.app.Store, router.app.Ctx, l, notificationManager)
	router.setupDomainRoutes(api, domainController)

	githubConnectorController := githubConnector.NewGithubConnectorController(router.app.Store, router.app.Ctx, l, notificationManager)
	router.setupGithubConnectorRoutes(api, githubConnectorController)

	router.setupDeployRoutes(api, deployController)

	return r
}

func (router *Router) setupBasicRoutes(r *mux.Router) {
	r.HandleFunc("/health", health.HealthCheck).Methods("GET", "OPTIONS")
}

func (router *Router) setupSwaggerRoutes(r *mux.Router) {
	r.PathPrefix("/docs/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	r.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "./docs/swagger.json")
	})
}

func (router *Router) setupWebSocketServer(r *mux.Router, deployController *deploy.DeployController) {
	wsServer, err := realtime.NewSocketServer(deployController, router.app.Store.DB, router.app.Ctx)
	if err != nil {
		log.Fatal(err)
	}

	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsServer.HandleHTTP(w, r)
	})
}

func (router *Router) setupAuthRoutes(r *mux.Router, authController *auth.AuthController) {
	authCB := r.PathPrefix("/api/v1/auth").Subrouter()
	authCB.HandleFunc("/register", authController.Register).Methods("POST", "OPTIONS")
	authCB.HandleFunc("/login", authController.Login).Methods("POST", "OPTIONS")
	authCB.HandleFunc("/refresh-token", authController.RefreshToken).Methods("POST", "OPTIONS")
}

func (router *Router) setupAuthenticatedAuthRoutes(api *mux.Router, authController *auth.AuthController) {
	authApi := api.PathPrefix("/auth").Subrouter()
	authApi.HandleFunc("/request-password-reset", authController.GeneratePasswordResetLink).Methods("POST", "OPTIONS")
	authApi.HandleFunc("/reset-password", authController.ResetPassword).Methods("POST", "OPTIONS")
	authApi.HandleFunc("/logout", authController.Logout).Methods("POST", "OPTIONS")
	authApi.HandleFunc("/send-verification-email", authController.SendVerificationEmail).Methods("POST", "OPTIONS")
	authApi.HandleFunc("/verify-email", authController.VerifyEmail).Methods("POST", "OPTIONS")
}

func (router *Router) setupRoleRoutes(api *mux.Router, roleController *role.RolesController) {
	roleApi := api.PathPrefix("/roles").Subrouter()
	roleApi.Use(middleware.IsAdmin)
	roleApi.HandleFunc("", roleController.CreateRole).Methods("POST", "OPTIONS")
	roleApi.HandleFunc("", roleController.GetRole).Methods("GET", "OPTIONS")
	roleApi.HandleFunc("", roleController.UpdateRole).Methods("PUT", "OPTIONS")
	roleApi.HandleFunc("", roleController.DeleteRole).Methods("DELETE", "OPTIONS")
	roleApi.HandleFunc("/all", roleController.GetRoles).Methods("GET", "OPTIONS")
}

func (router *Router) setupOrganizationRoutes(api *mux.Router, organizationController *organization.OrganizationsController) {
	orgApi := api.PathPrefix("/organizations").Subrouter()
	orgApi.Use(middleware.IsAdmin)
	orgApi.HandleFunc("", organizationController.CreateOrganization).Methods("POST", "OPTIONS")
	orgApi.HandleFunc("", organizationController.GetOrganization).Methods("GET", "OPTIONS")
	orgApi.HandleFunc("", organizationController.UpdateOrganization).Methods("PUT", "OPTIONS")
	orgApi.HandleFunc("", organizationController.DeleteOrganization).Methods("DELETE", "OPTIONS")
	orgApi.HandleFunc("/all", organizationController.GetOrganizations).Methods("GET", "OPTIONS")
	orgApi.HandleFunc("/user", organizationController.AddUserToOrganization).Methods("POST", "OPTIONS")
	orgApi.HandleFunc("/user", organizationController.RemoveUserFromOrganization).Methods("DELETE", "OPTIONS")
	orgApi.HandleFunc("/users", organizationController.GetOrganizationUsers).Methods("GET", "OPTIONS")
}

func (router *Router) setupPermissionRoutes(api *mux.Router, permissionController *permission.PermissionController) {
	permApi := api.PathPrefix("/permissions").Subrouter()
	permApi.Use(middleware.IsAdmin)
	permApi.HandleFunc("", permissionController.CreatePermission).Methods("POST", "OPTIONS")
	permApi.HandleFunc("", permissionController.GetPermission).Methods("GET", "OPTIONS")
	permApi.HandleFunc("", permissionController.UpdatePermission).Methods("PUT", "OPTIONS")
	permApi.HandleFunc("", permissionController.DeletePermission).Methods("DELETE", "OPTIONS")
	permApi.HandleFunc("/all", permissionController.GetPermissions).Methods("GET", "OPTIONS")
}

func (router *Router) setupRolePermissionRoutes(api *mux.Router, permissionController *permission.PermissionController) {
	rolePermApi := api.PathPrefix("/roles/permission").Subrouter()
	rolePermApi.Use(middleware.IsAdmin)
	rolePermApi.HandleFunc("", permissionController.AddPermissionToRole).Methods("POST", "OPTIONS")
	rolePermApi.HandleFunc("", permissionController.RemovePermissionFromRole).Methods("DELETE", "OPTIONS")
	rolePermApi.HandleFunc("", permissionController.GetPermissionsByRole).Methods("GET", "OPTIONS")
}

func (router *Router) setupUserRoutes(api *mux.Router, userController *user.UserController) {
	userApi := api.PathPrefix("/user").Subrouter()
	userApi.HandleFunc("", userController.GetUserDetails).Methods("GET", "OPTIONS")
	userApi.HandleFunc("/name", userController.UpdateUserName).Methods("PATCH", "OPTIONS")
	userApi.HandleFunc("/organizations", userController.GetUserOrganizations).Methods("GET", "OPTIONS")
}

func (router *Router) setupNotificationRoutes(api *mux.Router, notificationController *notificationController.NotificationController) {
	notificationApi := api.PathPrefix("/notification").Subrouter()
	notificationApi.HandleFunc("/smtp", notificationController.AddSmtp).Methods("POST", "OPTIONS")
	notificationApi.HandleFunc("/smtp", notificationController.GetSmtp).Methods("GET", "OPTIONS")
	notificationApi.HandleFunc("/smtp", notificationController.UpdateSmtp).Methods("PUT", "OPTIONS")
	notificationApi.HandleFunc("/smtp", notificationController.DeleteSmtp).Methods("DELETE", "OPTIONS")
	notificationApi.HandleFunc("/preferences", notificationController.UpdatePreference).Methods("POST", "OPTIONS")
	notificationApi.HandleFunc("/preferences", notificationController.GetPreferences).Methods("GET", "OPTIONS")
}

func (router *Router) setupDomainRoutes(api *mux.Router, domainController *domain.DomainsController) {
	domainApi := api.PathPrefix("/domain").Subrouter()
	domainApi.HandleFunc("", domainController.CreateDomain).Methods("POST", "OPTIONS")
	domainApi.HandleFunc("", domainController.UpdateDomain).Methods("PUT", "OPTIONS")
	domainApi.HandleFunc("", domainController.DeleteDomain).Methods("DELETE", "OPTIONS")
	domainApi.HandleFunc("/all", domainController.GetDomains).Methods("GET", "OPTIONS")
	domainApi.HandleFunc("/generate", domainController.GenerateRandomSubDomain).Methods("GET", "OPTIONS")
}

func (router *Router) setupGithubConnectorRoutes(api *mux.Router, githubConnectorController *githubConnector.GithubConnectorController) {
	githubConnectorApi := api.PathPrefix("/github-connector").Subrouter()
	githubConnectorApi.Use(middleware.IsAdmin)
	githubConnectorApi.HandleFunc("", githubConnectorController.CreateGithubConnector).Methods("POST", "OPTIONS")
	githubConnectorApi.HandleFunc("", githubConnectorController.UpdateGithubConnectorRequest).Methods("PUT", "OPTIONS")
	githubConnectorApi.HandleFunc("/all", githubConnectorController.GetGithubConnectors).Methods("GET", "OPTIONS")
	githubConnectorApi.HandleFunc("/repositories", githubConnectorController.GetGithubRepositories).Methods("GET", "OPTIONS")
}

func (router *Router) setupDeployRoutes(api *mux.Router, deployController *deploy.DeployController) {
	deployApi := api.PathPrefix("/deploy").Subrouter()
	deployApi.Use(middleware.IsAdmin)
	router.setupDeployValidatorRoutes(deployApi, deployController)
	deployApi.HandleFunc("/applications", deployController.GetApplications).Methods("GET", "OPTIONS")
	router.setupDeployApplicationRoutes(deployApi, deployController)
}

func (router *Router) setupDeployValidatorRoutes(deployApi *mux.Router, deployController *deploy.DeployController) {
	deployApiValidator := deployApi.PathPrefix("/validate").Subrouter()
	deployApiValidator.HandleFunc("/name", deployController.IsNameAlreadyTaken).Methods("POST", "OPTIONS")
	deployApiValidator.HandleFunc("/domain", deployController.IsDomainAlreadyTaken).Methods("POST", "OPTIONS")
	deployApiValidator.HandleFunc("/port", deployController.IsPortAlreadyTaken).Methods("POST", "OPTIONS")
}

func (router *Router) setupDeployApplicationRoutes(deployApi *mux.Router, deployController *deploy.DeployController) {
	deployApplicationApi := deployApi.PathPrefix("/application").Subrouter()
	deployApplicationApi.HandleFunc("", deployController.HandleDeploy).Methods("POST", "OPTIONS")
	deployApplicationApi.HandleFunc("", deployController.GetApplicationById).Methods("GET", "OPTIONS")
	deployApplicationApi.HandleFunc("", deployController.DeleteApplication).Methods("DELETE", "OPTIONS")
	deployApplicationApi.HandleFunc("", deployController.UpdateApplication).Methods("PUT", "OPTIONS")
	deployApplicationApi.HandleFunc("/redeploy", deployController.ReDeployApplication).Methods("POST", "OPTIONS")
	deployApplicationApi.HandleFunc("/deployments/{deployment_id}", deployController.GetDeploymentById).Methods("GET", "OPTIONS")
	deployApplicationApi.HandleFunc("/rollback", deployController.HandleRollback).Methods("POST", "OPTIONS")
	deployApplicationApi.HandleFunc("/restart", deployController.HandleRestart).Methods("POST", "OPTIONS")
}
