package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/spf13/viper"
)

var (
	AppConfig types.Config
)

// Init initializes the app configuration using Viper to load values from config files,
// environment variables, and defaults. It then creates a new PostgreSQL client using
// the loaded configuration and initializes the storage.Store.
func Init() *storage.Store {
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
		MigrationsPath: "./migrations",
	}

	store, err := storage.NewDB(&storage_config)
	if err != nil {
		log.Fatal(err)
	}
	err = storage.RunMigrations(store, storage_config.MigrationsPath)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations completed successfully")
	if AppConfig.Server.Port == "" {
		AppConfig.Server.Port = "8080"
	}
	if err != nil {
		log.Fatalf("Failed to initialize postgres client: %v", err)
	}

	storage := storage.NewStore(store)

	err = storage.Init(context.Background())

	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	return storage
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

	// SSH
	viper.BindEnv("ssh.host", "SSH_HOST")
	viper.BindEnv("ssh.port", "SSH_PORT")
	viper.BindEnv("ssh.user", "SSH_USER")
	viper.BindEnv("ssh.password", "SSH_PASSWORD")
	viper.BindEnv("ssh.private_key", "SSH_PRIVATE_KEY")
	viper.BindEnv("ssh.private_key_protected", "SSH_PRIVATE_KEY_PROTECTED")

	// Deployment
	viper.BindEnv("deployment.mount_path", "MOUNT_PATH")

	// Docker
	viper.BindEnv("docker.host", "DOCKER_HOST")
	viper.BindEnv("docker.port", "DOCKER_PORT")
	viper.BindEnv("docker.context", "DOCKER_CONTEXT")

	// Proxy
	viper.BindEnv("proxy.caddy_endpoint", "CADDY_ENDPOINT")

	// CORS
	viper.BindEnv("cors.allowed_origin", "ALLOWED_ORIGIN")

	// App
	viper.BindEnv("app.environment", "ENV")
	viper.BindEnv("app.version", "APP_VERSION")
	viper.BindEnv("app.logs_path", "LOGS_PATH")

	// SuperTokens
	viper.BindEnv("supertokens.api_key", "SUPERTOKENS_API_KEY")
	viper.BindEnv("supertokens.api_domain", "SUPERTOKENS_API_DOMAIN")
	viper.BindEnv("supertokens.website_domain", "SUPERTOKENS_WEBSITE_DOMAIN")
	viper.BindEnv("supertokens.connection_uri", "SUPERTOKENS_CONNECTION_URI")
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

	if config.SSH.Host == "" {
		errors = append(errors, "SSH host is required")
	}
	if config.SSH.User == "" {
		errors = append(errors, "SSH user is required")
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
	if config.Supertokens.APIKey == "" {
		errors = append(errors, "SuperTokens API key is required")
	}
	if config.Supertokens.APIDomain == "" {
		errors = append(errors, "SuperTokens API domain is required")
	}
	if config.Supertokens.WebsiteDomain == "" {
		errors = append(errors, "SuperTokens website domain is required")
	}
	if config.Supertokens.ConnectionURI == "" {
		errors = append(errors, "SuperTokens connection URI is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %v", errors)
	}

	return nil
}
