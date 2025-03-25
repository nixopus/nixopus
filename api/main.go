package main

import (
	"context"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/raghavyuva/nixopus-api/docs"
	"github.com/raghavyuva/nixopus-api/internal"
	"github.com/raghavyuva/nixopus-api/internal/config"
	_ "github.com/raghavyuva/nixopus-api/internal/log"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
)


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
func main() {
	docs.SwaggerInfo.Title = "Nixopus API"
	docs.SwaggerInfo.Description = "Api Nixopus"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http"}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	store := config.Init()
	ctx := context.Background()
	app := storage.NewApp(&types.Config{}, store, ctx)
	router := internal.NewRouter(app)
	r := router.Routes()
	log.Printf("Server starting on port %s", config.AppConfig.Port)
	log.Fatal(http.ListenAndServe(":"+config.AppConfig.Port, r))
}
