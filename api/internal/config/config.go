package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/raghavyuva/nixopus-api/internal/secrets"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/spf13/viper"
)

var (
	AppConfig   types.Config
	GlobalStore *storage.Store // Global storage instance, set during Init()
)

// getMigrationsPath returns the migrations path from environment variable or defaults to path relative to executable
func getMigrationsPath() string {
	// Use MIGRATIONS_PATH environment variable if set
	if migrationsPath := os.Getenv("MIGRATIONS_PATH"); migrationsPath != "" {
		return migrationsPath
	}

	// Default: use migrations directory relative to executable location
	// If executable is at api/nixopus-mcp-server or api/nixopus-api, migrations are at api/migrations
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		migrationsPath := filepath.Join(execDir, "migrations")
		if absPath, err := filepath.Abs(migrationsPath); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				return absPath
			}
		}
	}

	// Final fallback: relative to current working directory
	return "./migrations"
}

// Init initializes the app configuration using Viper to load values from config files,
// environment variables, and defaults. It then creates a new PostgreSQL client using
// the loaded configuration and initializes the storage.Store.
func Init() *storage.Store {
	// Load secrets from secret manager first (if enabled)
	// This allows secrets to override .env file values
	secretConfig := secrets.LoadSecretManagerConfig("api")
	if secretConfig.Enabled {
		secretManager, err := secrets.NewSecretManager(secretConfig)
		if err != nil {
			log.Printf("Warning: Failed to initialize secret manager: %v. Falling back to .env files", err)
		} else {
			// Load secrets with service-specific prefix (e.g., "API_", "NIXOPUS_API_")
			prefixes := []string{"API_", "NIXOPUS_API_", ""}
			for _, prefix := range prefixes {
				if err := secrets.LoadSecretsIntoEnv(secretManager, prefix); err != nil {
					log.Printf("Warning: Failed to load secrets with prefix %s: %v", prefix, err)
				}
			}
		}
	}

	// Load .env file (will be overridden by secrets if they exist)
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	initViper()

	AppConfig = types.Config{}
	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	log.Printf("Configuration loaded successfully for environment: %s", AppConfig.App.Environment)

	if err := validateConfig(AppConfig); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Log key configuration values (without sensitive data)
	log.Printf("Server will start on port: %s", AppConfig.Server.Port)
	log.Printf("Database host: %s:%s", AppConfig.Database.Host, AppConfig.Database.Port)
	log.Printf("Redis URL configured: %t", AppConfig.Redis.URL != "")

	migrationsPath := getMigrationsPath()

	storage_config := storage.Config{
		Host:           AppConfig.Database.Host,
		Port:           AppConfig.Database.Port,
		Username:       AppConfig.Database.Username,
		Password:       AppConfig.Database.Password,
		DBName:         AppConfig.Database.Name,
		SSLMode:        AppConfig.Database.SSLMode,
		MaxOpenConn:    AppConfig.Database.MaxOpenConn,
		Debug:          AppConfig.Database.Debug,
		MaxIdleConn:    AppConfig.Database.MaxIdleConn,
		MigrationsPath: migrationsPath,
	}

	store, err := storage.NewDB(&storage_config)
	if err != nil {
		log.Fatal(err)
	}

	if AppConfig.Server.Port == "" {
		AppConfig.Server.Port = "8080"
	}

	storageInstance := storage.NewStore(store)

	err = storageInstance.Init(context.Background())

	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Set global store for use throughout the application
	GlobalStore = storageInstance

	return storageInstance
}

func initViper() {
	configName := getConfigFileName()
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")

	// Check for custom config path from environment variable first
	configPaths := []string{}
	if customConfigPath := os.Getenv("NIXOPUS_CONFIG_PATH"); customConfigPath != "" {
		configPaths = append(configPaths, customConfigPath)
		log.Printf("Using custom config path from NIXOPUS_CONFIG_PATH: %s", customConfigPath)
	}

	// default fallback paths
	defaultPaths := []string{
		"../helpers",                  // Relative to api directory
		"/etc/nixopus/source/helpers", // Docker mounted path
		"./helpers",                   // Current directory helpers
		".",                           // Current directory
	}

	configPaths = append(configPaths, defaultPaths...)

	for _, path := range configPaths {
		viper.AddConfigPath(path)
	}

	viper.AutomaticEnv()

	setupEnvVarMappings()

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Printf("Info: Config file '%s' not found, using environment variables", configName)
		} else {
			log.Printf("Warning: Error reading config file: %v", err)
		}
		log.Println("Using environment variables and defaults")
	} else {
		log.Printf("Successfully loaded config file: %s", viper.ConfigFileUsed())
	}
}

func getConfigFileName() string {
	env := strings.ToLower(os.Getenv("ENV"))
	switch env {
	case "development", "dev":
		return "config.dev"
	case "staging", "stage":
		return "config.staging"
	case "production", "prod":
		return "config.prod"
	default:
		// Default to production config if environment is not specified
		return "config.prod"
	}
}

func setupEnvVarMappings() {
	// Server
	viper.BindEnv("server.port", "PORT")

	// Database
	viper.BindEnv("database.host", "HOST_NAME")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.username", "USERNAME")
	viper.BindEnv("database.password", "PASSWORD")
	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("database.ssl_mode", "SSL_MODE")

	// Redis
	viper.BindEnv("redis.url", "REDIS_URL")

	// Deployment
	viper.BindEnv("deployment.mount_path", "MOUNT_PATH")

	// Proxy
	viper.BindEnv("proxy.caddy_endpoint", "CADDY_ENDPOINT")

	// CORS
	viper.BindEnv("cors.allowed_origin", "ALLOWED_ORIGIN")

	// App
	viper.BindEnv("app.environment", "ENV")
	viper.BindEnv("app.version", "APP_VERSION")
	viper.BindEnv("app.logs_path", "LOGS_PATH")

	// GitHub App (shared credentials)
	viper.BindEnv("github.app_id", "GITHUB_APP_ID")
	viper.BindEnv("github.slug", "GITHUB_APP_SLUG")
	viper.BindEnv("github.pem", "GITHUB_APP_PEM")
	viper.BindEnv("github.client_id", "GITHUB_APP_CLIENT_ID")
	viper.BindEnv("github.client_secret", "GITHUB_APP_CLIENT_SECRET")
	viper.BindEnv("github.webhook_secret", "GITHUB_APP_WEBHOOK_SECRET")
	// Better Auth
	viper.BindEnv("betterauth.url", "BETTER_AUTH_URL")
	viper.BindEnv("betterauth.secret", "BETTER_AUTH_SECRET")

	// Stripe
	viper.BindEnv("stripe.secret_key", "STRIPE_SECRET_KEY")
	viper.BindEnv("stripe.webhook_secret", "STRIPE_WEBHOOK_SECRET")
	viper.BindEnv("stripe.price_id", "STRIPE_PRICE_ID")
	viper.BindEnv("stripe.free_deployments_limit", "FREE_DEPLOYMENTS_LIMIT")

	// Set default for free deployments limit
	viper.SetDefault("stripe.free_deployments_limit", 1)
}

func validateConfig(config types.Config) error {
	var errors []string

	if config.Server.Port == "" {
		errors = append(errors, "server port is required")
	}

	if config.Database.Host == "" {
		errors = append(errors, "database host is required")
	}
	if config.Database.Port == "" {
		errors = append(errors, "database port is required")
	}
	if config.Database.Username == "" {
		errors = append(errors, "database username is required")
	}
	if config.Database.Password == "" {
		errors = append(errors, "database password is required")
	}
	if config.Database.Name == "" {
		errors = append(errors, "database name is required")
	}

	if config.Redis.URL == "" {
		errors = append(errors, "redis URL is required")
	}
	if config.Deployment.MountPath == "" {
		errors = append(errors, "deployment mount path is required")
	}

	if config.Proxy.CaddyEndpoint == "" {
		errors = append(errors, "proxy caddy endpoint is required")
	}

	if config.CORS.AllowedOrigin == "" {
		errors = append(errors, "CORS allowed origin is required")
	}
	if config.BetterAuth.URL == "" {
		errors = append(errors, "Better Auth URL is required")
	}
	if config.BetterAuth.Secret == "" {
		errors = append(errors, "Better Auth secret is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %v", errors)
	}

	return nil
}
