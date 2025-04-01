package main

import (
	"context"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/raghavyuva/nixopus-api/internal"
	"github.com/raghavyuva/nixopus-api/internal/config"
	_ "github.com/raghavyuva/nixopus-api/internal/log"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	store := config.Init()
	ctx := context.Background()
	app := storage.NewApp(&types.Config{}, store, ctx)

	// cacheClient, err := cache.NewCache(config.AppConfig.RedisURL)
	// if err != nil {
	// 	log.Fatalf("Failed to initialize cache: %v", err)
	// }

	router := internal.NewRouter(app)
	router.Routes()
	log.Printf("Server starting on port %s", config.AppConfig.Port)
	log.Fatal(http.ListenAndServe(":"+config.AppConfig.Port, nil))
}
