package internal

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/raghavyuva/nixopus-api/internal/controller"
	"github.com/raghavyuva/nixopus-api/internal/middleware"
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

// Routes returns a new router that handles all API routes, including
// unauthenticated and authenticated routes.
// The following middleware is used:
//
// - middleware.CorsMiddleware: enables CORS
// - middleware.LoggingMiddleware: logs all requests
// - middleware.AuthMiddleware: checks if the request is authenticated
func (router *Router) Routes() *mux.Router {
	r := mux.NewRouter()

	r.Use(middleware.CorsMiddleware)
	r.Use(middleware.LoggingMiddleware)

	r.HandleFunc("/health", controller.HealthCheck).Methods("GET", "OPTIONS")

	wsServer, err := NewSocketServer()
	if err != nil {
		log.Fatal(err)
	}

	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsServer.HandleHTTP(w, r)
	})

	u := r.PathPrefix("/api/v1").Subrouter()

	// Unauthenticated routes
	authController := controller.NewAuthController(router.app)
	u.HandleFunc("/auth/register", authController.Register).Methods("POST", "OPTIONS")
	u.HandleFunc("/auth/login", authController.Login).Methods("POST", "OPTIONS")

	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(func(next http.Handler) http.Handler {
		return middleware.AuthMiddleware(next, router.app)
	})

	// Authenticated routes
	api.HandleFunc("/auth/request-password-reset", authController.GeneratePasswordResetLink).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/reset-password", authController.ResetPassword).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/refresh-token", authController.RefreshToken).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/logout", authController.Logout).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/send-verification-email", authController.SendVerificationEmail).Methods("POST", "OPTIONS")

	return r
}