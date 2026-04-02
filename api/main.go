package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/nixopus/nixopus/api/internal/config"
	_ "github.com/nixopus/nixopus/api/internal/log"
	"github.com/nixopus/nixopus/api/internal/queue"
	"github.com/nixopus/nixopus/api/internal/redisclient"
	"github.com/nixopus/nixopus/api/internal/routes"
	"github.com/nixopus/nixopus/api/internal/scheduler"
	"github.com/nixopus/nixopus/api/internal/storage"
	"github.com/nixopus/nixopus/api/internal/types"
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
	// Load .env file if it exists (optional when using secret manager)
	if err := godotenv.Load(); err != nil {
		// .env file is optional when using secret manager, so we just log a warning
		log.Println("Info: .env file not found, using environment variables and secret manager")
	}

	types.InitJWTSecret()
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

	queue.SetupProvisionQueue()
	queue.SetupCustomDomainQueue()
	queue.SetupResourceUpdateQueue()
	queue.SetupMachineLifecycleQueue(ctx)
	queue.SetupMachineBackupQueue(ctx)
	queue.SetupVMDeleteQueue()

	router := routes.NewRouter(app)

	// Initialize schedulers
	schedulers := scheduler.InitSchedulers(store, ctx)
	router.SetSchedulers(schedulers)

	// Start schedulers
	if err := schedulers.Main.Start(); err != nil {
		log.Printf("Warning: failed to start scheduler: %v", err)
	} else {
		log.Println("Scheduler started successfully")
	}
	schedulers.HealthCheck.Start()
	log.Println("Health check scheduler started successfully")
	schedulers.Billing.Start()
	log.Println("Billing scheduler started successfully")
	schedulers.Backup.Start()
	log.Println("Backup scheduler started successfully")
	schedulers.TrialExpiry.Start()
	log.Println("Trial expiry scheduler started successfully")

	router.SetupRoutes()

	// Setup graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		queue.Close()
		schedulers.Main.Stop()
		schedulers.HealthCheck.Stop()
		schedulers.Billing.Stop()
		schedulers.Backup.Stop()
		schedulers.TrialExpiry.Stop()
		os.Exit(0)
	}()
	log.Printf("Server starting on port %s", config.AppConfig.Server.Port)
	log.Fatal(http.ListenAndServe(":"+config.AppConfig.Server.Port, nil))
}
