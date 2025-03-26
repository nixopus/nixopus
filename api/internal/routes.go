package internal

import (
	"log"
	"net/http"

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

func (router *Router) Routes() *mux.Router {
	r := mux.NewRouter()
	l := logger.NewLogger()
	r.Use(middleware.CorsMiddleware)
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.RateLimiter)

	router.BasicRoutes(r)

	router.SwaggerRoutes(r)

	notificationManager := notification.NewNotificationManager(notification.NewNotificationChannels(), router.app.Store.DB)
	notificationManager.Start()

	deployController := deploy.NewDeployController(router.app.Store, router.app.Ctx, l, notificationManager)

	router.WebSocketServer(r, deployController)

	authController := auth.NewAuthController(router.app.Store, router.app.Ctx, l, notificationManager)
	router.AuthRoutes(r, authController)

	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(func(next http.Handler) http.Handler {
		return middleware.AuthMiddleware(next, router.app)
	})

	router.AuthenticatedAuthRoutes(api, authController)
	organizationController := organization.NewOrganizationsController(router.app.Store, router.app.Ctx, l, notificationManager)
	router.OrganizationRoutes(api, organizationController)

	userController := user.NewUserController(router.app.Store, router.app.Ctx, l)
	router.UserRoutes(api, userController)

	notifController := notificationController.NewNotificationController(router.app.Store, router.app.Ctx, l, notificationManager)
	router.NotificationRoutes(api, notifController)

	domainController := domain.NewDomainsController(router.app.Store, router.app.Ctx, l, notificationManager)
	router.DomainRoutes(api, domainController)

	githubConnectorController := githubConnector.NewGithubConnectorController(router.app.Store, router.app.Ctx, l, notificationManager)
	router.GithubConnectorRoutes(api, githubConnectorController)

	file_managerController := file_manager.NewFileManagerController(router.app.Ctx, l, notificationManager)
	router.FileManagerRoutes(api, file_managerController)

	router.DeployRoutes(api, deployController)
	return r
}

func (router *Router) BasicRoutes(r *mux.Router) {
	r.HandleFunc("/health", health.HealthCheck).Methods("GET", "OPTIONS")
}

func (router *Router) SwaggerRoutes(r *mux.Router) {
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
func (router *Router) AuthRoutes(r *mux.Router, authController *auth.AuthController) {
	authApi := r.PathPrefix("/api/v1/auth").Subrouter()

	//register route is disabled for now (we do not have register seperately either the one who installs it, or the one who is added by admin)
	// authApi.HandleFunc("/register", authController.Register).Methods("POST", "OPTIONS")
	authApi.HandleFunc("/login", authController.Login).Methods("POST", "OPTIONS")
}

func (router *Router) AuthenticatedAuthRoutes(api *mux.Router, authController *auth.AuthController) {
	authApi := api.PathPrefix("/auth").Subrouter()
	authApi.HandleFunc("/request-password-reset", authController.GeneratePasswordResetLink).Methods("POST", "OPTIONS")
	authApi.HandleFunc("/reset-password", authController.ResetPassword).Methods("POST", "OPTIONS")
	authApi.HandleFunc("/logout", authController.Logout).Methods("POST", "OPTIONS")
	authApi.HandleFunc("/send-verification-email", authController.SendVerificationEmail).Methods("POST", "OPTIONS")
	authApi.HandleFunc("/verify-email", authController.VerifyEmail).Methods("POST", "OPTIONS")
	authApi.HandleFunc("/create-user", authController.CreateUser).Methods("POST", "OPTIONS")
	authApi.HandleFunc("/refresh-token", authController.RefreshToken).Methods("POST", "OPTIONS")
}

func (router *Router) UserRoutes(api *mux.Router, userController *user.UserController) {
	userApi := api.PathPrefix("/user").Subrouter()
	userApi.HandleFunc("", userController.GetUserDetails).Methods("GET", "OPTIONS")
	userApi.HandleFunc("/name", userController.UpdateUserName).Methods("PATCH", "OPTIONS")
	userApi.HandleFunc("/organizations", userController.GetUserOrganizations).Methods("GET", "OPTIONS")
}

func (router *Router) NotificationRoutes(api *mux.Router, notificationController *notificationController.NotificationController) {
	notificationApi := api.PathPrefix("/notification").Subrouter()
	notificationSmtpApi := notificationApi.PathPrefix("/smtp").Subrouter()
	notificationSmtpApi.HandleFunc("", notificationController.AddSmtp).Methods("POST", "OPTIONS")
	notificationSmtpApi.HandleFunc("", notificationController.GetSmtp).Methods("GET", "OPTIONS")
	notificationSmtpApi.HandleFunc("", notificationController.UpdateSmtp).Methods("PUT", "OPTIONS")
	notificationSmtpApi.HandleFunc("", notificationController.DeleteSmtp).Methods("DELETE", "OPTIONS")

	notificationPreferencesApi := notificationApi.PathPrefix("/preferences").Subrouter()
	notificationPreferencesApi.HandleFunc("", notificationController.UpdatePreference).Methods("POST", "OPTIONS")
	notificationPreferencesApi.HandleFunc("", notificationController.GetPreferences).Methods("GET", "OPTIONS")
}

func (router *Router) DomainRoutes(api *mux.Router, domainController *domain.DomainsController) {
	domainApi := api.PathPrefix("/domain").Subrouter()
	domainApi.HandleFunc("", domainController.CreateDomain).Methods("POST", "OPTIONS")
	domainApi.HandleFunc("", domainController.UpdateDomain).Methods("PUT", "OPTIONS")
	domainApi.HandleFunc("", domainController.DeleteDomain).Methods("DELETE", "OPTIONS")
	domainApi.HandleFunc("/generate", domainController.GenerateRandomSubDomain).Methods("GET", "OPTIONS")

	api.HandleFunc("/domains", domainController.GetDomains).Methods("GET", "OPTIONS")
}

func (router *Router) GithubConnectorRoutes(api *mux.Router, githubConnectorController *githubConnector.GithubConnectorController) {
	githubConnectorApi := api.PathPrefix("/github-connector").Subrouter()
	githubConnectorApi.HandleFunc("", githubConnectorController.CreateGithubConnector).Methods("POST", "OPTIONS")
	githubConnectorApi.HandleFunc("", githubConnectorController.UpdateGithubConnectorRequest).Methods("PUT", "OPTIONS")
	githubConnectorApi.HandleFunc("/all", githubConnectorController.GetGithubConnectors).Methods("GET", "OPTIONS")
	githubConnectorApi.HandleFunc("/repositories", githubConnectorController.GetGithubRepositories).Methods("GET", "OPTIONS")
}

func (router *Router) DeployRoutes(api *mux.Router, deployController *deploy.DeployController) {
	deployApi := api.PathPrefix("/deploy").Subrouter()
	router.DeployValidatorRoutes(deployApi, deployController)
	deployApi.HandleFunc("/applications", deployController.GetApplications).Methods("GET", "OPTIONS")
	router.DeployApplicationRoutes(deployApi, deployController)
}

func (router *Router) DeployValidatorRoutes(deployApi *mux.Router, deployController *deploy.DeployController) {
	deployApiValidator := deployApi.PathPrefix("/validate").Subrouter()
	deployApiValidator.HandleFunc("/name", deployController.IsNameAlreadyTaken).Methods("POST", "OPTIONS")
	deployApiValidator.HandleFunc("/domain", deployController.IsDomainAlreadyTaken).Methods("POST", "OPTIONS")
	deployApiValidator.HandleFunc("/port", deployController.IsPortAlreadyTaken).Methods("POST", "OPTIONS")
}

func (router *Router) DeployApplicationRoutes(deployApi *mux.Router, deployController *deploy.DeployController) {
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

func (router *Router) FileManagerRoutes(api *mux.Router, fileManagerController *file_manager.FileManagerController) {
	file_manager_api := api.PathPrefix("/file-manager").Subrouter()
	// we will expose these routes only for admin
	file_manager_api.Use(middleware.IsAdmin)
	file_manager_api.HandleFunc("", fileManagerController.ListFiles).Methods("GET", "OPTIONS")
	file_manager_api.HandleFunc("/create-directory", fileManagerController.CreateDirectory).Methods("POST", "OPTIONS")
}

func (router *Router) OrganizationRoutes(api *mux.Router, organizationController *organization.OrganizationsController) {
	orgApi := api.PathPrefix("/organizations").Subrouter()
	orgApi.HandleFunc("/users", organizationController.GetOrganizationUsers).Methods("GET", "OPTIONS")
}
