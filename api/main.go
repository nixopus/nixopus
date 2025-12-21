package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
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

// testRedisConnection tests the Redis connection and fails early in production if connection fails
func testRedisConnection(ctx context.Context, redisClient *redis.Client) {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		env := strings.ToLower(config.AppConfig.App.Environment)
		isProduction := env == "production" || env == "prod"
		if isProduction {
			log.Fatalf("failed to connect to Redis in production: %v", err)
		}
		log.Printf("Warning: failed to connect to Redis: %v (continuing in non-production environment)", err)
	} else {
		log.Println("Successfully connected to Redis")
	}
}

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

	// Test Redis connection - fail early in production if connection fails
	testRedisConnection(ctx, redisClient)

	taskq.SetLogger(log.New(io.Discard, "", 0))
	queue.Init(redisClient)
	router := internal.NewRouter(app)
	router.Routes()
	log.Printf("Server starting on port %s", config.AppConfig.Server.Port)
	log.Fatal(http.ListenAndServe(":"+config.AppConfig.Server.Port, nil))
}
