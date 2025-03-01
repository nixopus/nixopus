package internal

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	auth "github.com/raghavyuva/nixopus-api/internal/features/auth/controller"
	health "github.com/raghavyuva/nixopus-api/internal/features/health"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	notificationController "github.com/raghavyuva/nixopus-api/internal/features/notification/controller"
	organization "github.com/raghavyuva/nixopus-api/internal/features/organization/controller"
	permission "github.com/raghavyuva/nixopus-api/internal/features/permission/controller"
	role "github.com/raghavyuva/nixopus-api/internal/features/role/controller"
	user "github.com/raghavyuva/nixopus-api/internal/features/user/controller"
	"github.com/raghavyuva/nixopus-api/internal/middleware"
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

	r.HandleFunc("/health", health.HealthCheck).Methods("GET", "OPTIONS")
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

	wsServer, err := NewSocketServer()
	if err != nil {
		log.Fatal(err)
	}

	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsServer.HandleHTTP(w, r)
	})

	notificationManager := notification.NewNotificationManager(notification.NewNotificationChannels(), router.app.Store.DB)
	notificationManager.Start()

	u := r.PathPrefix("/api/v1").Subrouter()

	authCB := u.PathPrefix("/auth").Subrouter()

	authController := auth.NewAuthController(router.app.Store, router.app.Ctx, l, notificationManager)
	authCB.HandleFunc("/register", authController.Register).Methods("POST", "OPTIONS")
	authCB.HandleFunc("/login", authController.Login).Methods("POST", "OPTIONS")
	authCB.HandleFunc("/refresh-token", authController.RefreshToken).Methods("POST", "OPTIONS")

	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(func(next http.Handler) http.Handler {
		return middleware.AuthMiddleware(next, router.app)
	})

	authApi := api.PathPrefix("/auth").Subrouter()
	authApi.HandleFunc("/request-password-reset", authController.GeneratePasswordResetLink).Methods("POST", "OPTIONS")
	authApi.HandleFunc("/reset-password", authController.ResetPassword).Methods("POST", "OPTIONS")
	authApi.HandleFunc("/logout", authController.Logout).Methods("POST", "OPTIONS")
	authApi.HandleFunc("/send-verification-email", authController.SendVerificationEmail).Methods("POST", "OPTIONS")
	authApi.HandleFunc("/verify-email", authController.VerifyEmail).Methods("POST", "OPTIONS")

	roleApi := api.PathPrefix("/roles").Subrouter()
	roleApi.Use(middleware.IsAdmin)
	roleController := role.NewRolesController(router.app.Store, router.app.Ctx, l)
	roleApi.HandleFunc("", roleController.CreateRole).Methods("POST", "OPTIONS")
	roleApi.HandleFunc("", roleController.GetRole).Methods("GET", "OPTIONS")
	roleApi.HandleFunc("", roleController.UpdateRole).Methods("PUT", "OPTIONS")
	roleApi.HandleFunc("", roleController.DeleteRole).Methods("DELETE", "OPTIONS")
	roleApi.HandleFunc("/all", roleController.GetRoles).Methods("GET", "OPTIONS")

	orgApi := api.PathPrefix("/organizations").Subrouter()
	orgApi.Use(middleware.IsAdmin)
	organizationController := organization.NewOrganizationsController(router.app.Store, router.app.Ctx, l, notificationManager)
	orgApi.HandleFunc("", organizationController.CreateOrganization).Methods("POST", "OPTIONS")
	orgApi.HandleFunc("", organizationController.GetOrganization).Methods("GET", "OPTIONS")
	orgApi.HandleFunc("", organizationController.UpdateOrganization).Methods("PUT", "OPTIONS")
	orgApi.HandleFunc("", organizationController.DeleteOrganization).Methods("DELETE", "OPTIONS")
	orgApi.HandleFunc("/all", organizationController.GetOrganizations).Methods("GET", "OPTIONS")
	orgApi.HandleFunc("/user", organizationController.AddUserToOrganization).Methods("POST", "OPTIONS")
	orgApi.HandleFunc("/user", organizationController.RemoveUserFromOrganization).Methods("DELETE", "OPTIONS")
	orgApi.HandleFunc("/users", organizationController.GetOrganizationUsers).Methods("GET", "OPTIONS")

	permApi := api.PathPrefix("/permissions").Subrouter()
	permApi.Use(middleware.IsAdmin)
	permissionController := permission.NewPermissionController(router.app.Store, router.app.Ctx, l)
	permApi.HandleFunc("", permissionController.CreatePermission).Methods("POST", "OPTIONS")
	permApi.HandleFunc("", permissionController.GetPermission).Methods("GET", "OPTIONS")
	permApi.HandleFunc("", permissionController.UpdatePermission).Methods("PUT", "OPTIONS")
	permApi.HandleFunc("", permissionController.DeletePermission).Methods("DELETE", "OPTIONS")
	permApi.HandleFunc("/all", permissionController.GetPermissions).Methods("GET", "OPTIONS")

	rolePermApi := api.PathPrefix("/roles/permission").Subrouter()
	rolePermApi.Use(middleware.IsAdmin)
	rolePermApi.HandleFunc("", permissionController.AddPermissionToRole).Methods("POST", "OPTIONS")
	rolePermApi.HandleFunc("", permissionController.RemovePermissionFromRole).Methods("DELETE", "OPTIONS")
	rolePermApi.HandleFunc("", permissionController.GetPermissionsByRole).Methods("GET", "OPTIONS")

	userApi := api.PathPrefix("/user").Subrouter()
	userController := user.NewUserController(router.app.Store, router.app.Ctx, l)
	userApi.HandleFunc("", userController.GetUserDetails).Methods("GET", "OPTIONS")
	userApi.HandleFunc("/name", userController.UpdateUserName).Methods("PATCH", "OPTIONS")
	userApi.HandleFunc("/organizations", userController.GetUserOrganizations).Methods("GET", "OPTIONS")

	notificationApi := api.PathPrefix("/notification").Subrouter()
	notificationController := notificationController.NewNotificationController(router.app.Store, router.app.Ctx, l, notificationManager)
	notificationApi.HandleFunc("/smtp", notificationController.AddSmtp).Methods("POST", "OPTIONS")
	notificationApi.HandleFunc("/smtp", notificationController.GetSmtp).Methods("GET", "OPTIONS")
	notificationApi.HandleFunc("/smtp", notificationController.UpdateSmtp).Methods("PUT", "OPTIONS")
	notificationApi.HandleFunc("/smtp", notificationController.DeleteSmtp).Methods("DELETE", "OPTIONS")

	return r
}
