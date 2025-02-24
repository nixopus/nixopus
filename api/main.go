package main

import (
	"context"
	"log"
	"net/http"

	"github.com/raghavyuva/nixopus-api/docs"
	"github.com/raghavyuva/nixopus-api/internal"
	"github.com/raghavyuva/nixopus-api/internal/config"
	_ "github.com/raghavyuva/nixopus-api/internal/log"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

func main() {
	docs.SwaggerInfo.Title = "Nixopus API"
	docs.SwaggerInfo.Description = "Api for Nixopus"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = ""
	docs.SwaggerInfo.Schemes = []string{"http"}
	store := config.Init()
	ctx := context.Background()
	app := storage.NewApp(&types.Config{}, store, ctx)
	router := internal.NewRouter(app)
	r := router.Routes()
	log.Printf("Server starting on port %s", config.AppConfig.Port)
	log.Fatal(http.ListenAndServe(":"+config.AppConfig.Port, r))
}
