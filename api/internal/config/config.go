package config

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

var (
	AppConfig types.Config
)

// Init initializes the app configuration by loading values from the .env file,
// and creating a new PostgreSQL client using the loaded configuration.
// It then initializes the storage.Store using the new client and checks if the
// users table exists in the database. If the table does not exist, it creates it.
// Finally, it sets a default value of "8080" for AppConfig.Port if it is empty.
func Init() *storage.Store {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	AppConfig = types.Config{
		DB_PORT:     os.Getenv("DB_PORT"),
		Port:        os.Getenv("PORT"),
		HostName:    os.Getenv("HOST_NAME"),
		Password:    os.Getenv("PASSWORD"),
		DBName:      os.Getenv("DB_NAME"),
		Username:    os.Getenv("USERNAME"),
		SSLMode:     os.Getenv("SSL_MODE"),
		MaxOpenConn: 10,
		Debug:       true,
		MaxIdleConn: 5,
	}

	storage_config := storage.Config{
		Host:           AppConfig.HostName,
		Port:           AppConfig.DB_PORT,
		Username:       AppConfig.Username,
		Password:       AppConfig.Password,
		DBName:         AppConfig.DBName,
		SSLMode:        AppConfig.SSLMode,
		MaxOpenConn:    AppConfig.MaxOpenConn,
		Debug:          AppConfig.Debug,
		MaxIdleConn:    AppConfig.MaxIdleConn,
		MigrationsPath: "migrations",
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
	if AppConfig.Port == "" {
		AppConfig.Port = "8080"
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
