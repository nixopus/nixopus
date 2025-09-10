package main

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/raghavyuva/nixopus-api/internal"
	"github.com/raghavyuva/nixopus-api/internal/config"
	_ "github.com/raghavyuva/nixopus-api/internal/log"
	"github.com/raghavyuva/nixopus-api/internal/queue"
	"github.com/raghavyuva/nixopus-api/internal/redisclient"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/vmihailenco/taskq/v3"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	store := config.Init()
	ctx := context.Background()
	app := storage.NewApp(&types.Config{}, store, ctx)

	// Initialize task queue (Redis) and start consumers alongside the server
	redisClient, err := redisclient.New(config.AppConfig.Redis.URL)
	if err != nil {
		log.Fatalf("failed to create redis client for queue due to %v", err)
	}
	taskq.SetLogger(log.New(io.Discard, "", 0))
	queue.Init(redisClient)
	router := internal.NewRouter(app)
	router.Routes()
	log.Printf("Server starting on port %s", config.AppConfig.Server.Port)
	log.Fatal(http.ListenAndServe(":"+config.AppConfig.Server.Port, nil))
}
