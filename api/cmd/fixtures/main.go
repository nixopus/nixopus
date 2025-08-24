package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/fixtures/loader"
)

func main() {
	var (
		fixturePath = flag.String("fixture", "fixtures/development/complete.yml", "Path to fixture file")
		recreate    = flag.Bool("recreate", false, "Recreate tables before loading fixtures")
		truncate    = flag.Bool("truncate", false, "Truncate tables before loading fixtures")
	)
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize config
	config.Init()

	host := config.AppConfig.Database.Host
	port := config.AppConfig.Database.Port
	username := config.AppConfig.Database.Username
	password := config.AppConfig.Database.Password
	dbName := config.AppConfig.Database.Name
	sslMode := config.AppConfig.Database.SSLMode

	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		username, password, host, port, dbName, sslMode)

	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("Failed to parse database config: %v", err)
	}

	sqldb := stdlib.OpenDB(*config)
	defer sqldb.Close()

	db := bun.NewDB(sqldb, pgdialect.New())

	fixtureLoader := loader.NewFixtureLoader(db)

	ctx := context.Background()

	if *recreate {
		fmt.Printf("Loading fixtures with table recreation: %s\n", *fixturePath)
		err = fixtureLoader.LoadFixturesWithRecreate(ctx, *fixturePath)
	} else if *truncate {
		fmt.Printf("Loading fixtures with table truncation: %s\n", *fixturePath)
		err = fixtureLoader.LoadFixturesWithTruncate(ctx, *fixturePath)
	} else {
		fmt.Printf("Loading fixtures: %s\n", *fixturePath)
		err = fixtureLoader.LoadFixtures(ctx, *fixturePath)
	}

	if err != nil {
		log.Fatalf("Failed to load fixtures: %v", err)
	}

	fmt.Println("Fixtures loaded successfully!")
}
