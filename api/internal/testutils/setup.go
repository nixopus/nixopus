package testutils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/raghavyuva/nixopus-api/internal/cache"
	authService "github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	authTypes "github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	organization_service "github.com/raghavyuva/nixopus-api/internal/features/organization/service"
	organization_storage "github.com/raghavyuva/nixopus-api/internal/features/organization/storage"
	dbstorage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var (
	testDB *bun.DB
	ctx    context.Context
)

// TestSetup holds all the common test dependencies
type TestSetup struct {
	DB          *bun.DB
	Ctx         context.Context
	Store       *dbstorage.Store
	Logger      logger.Logger
	UserStorage *user_storage.UserStorage
	OrgStorage  *organization_storage.OrganizationStore
	OrgService  *organization_service.OrganizationService
	AuthService *authService.AuthService
}

func init() {
	ctx = context.Background()

	envFiles := []string{
		"../../../../env.test",
		"../../../../../env.test",
		"env.test",
	}

	envLoaded := false
	for _, file := range envFiles {
		if err := godotenv.Load(file); err == nil {
			envLoaded = true
			break
		}
	}

	if !envLoaded {
		fmt.Println("Warning: Could not load env.test file from any location")
	}

	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbPort := getEnvOrDefault("DB_PORT", "5433")
	dbUser := getEnvOrDefault("DB_USER", "nixopus")
	dbPassword := getEnvOrDefault("DB_PASSWORD", "nixopus")
	dbName := getEnvOrDefault("DB_NAME", "nixopus_test")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	fmt.Printf("Connecting to test database: %s\n", connStr)

	config, err := pgx.ParseConfig(connStr)
	if err != nil {
		fmt.Printf("Failed to parse config: %v\n", err)
		os.Exit(1)
	}

	sqldb := stdlib.OpenDB(*config)
	testDB = bun.NewDB(sqldb, pgdialect.New())

	if err := testDB.Ping(); err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully connected to test database")

	store := dbstorage.NewStore(testDB)
	if err := store.Init(ctx); err != nil {
		fmt.Printf("Failed to initialize store: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully ran migrations")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func findMigrationsPath() string {
	paths := []string{
		"../../../migrations",
		"../../../../migrations",
		"migrations",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			absPath, err := filepath.Abs(path)
			if err == nil {
				return absPath
			}
		}
	}

	return "migrations"
}

func cleanDatabase() error {
	// Drop all tables with CASCADE to ensure all dependencies are removed
	_, err := testDB.ExecContext(ctx, `
		DO $$ DECLARE
			r RECORD;
		BEGIN
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
				EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
			END LOOP;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to drop all tables: %w", err)
	}

	// Reset migrations
	if err := dbstorage.ResetMigrations(testDB); err != nil {
		return fmt.Errorf("failed to reset migrations: %w", err)
	}

	// Run migrations
	migrationsPath := findMigrationsPath()
	if err := dbstorage.RunMigrations(testDB, migrationsPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// NewTestSetup creates a new test setup with all common dependencies
func NewTestSetup() *TestSetup {
	if testDB == nil {
		panic("testDB is nil - database not initialized")
	}
	if ctx == nil {
		panic("ctx is nil - context not initialized")
	}

	// Clean database before each test
	if err := cleanDatabase(); err != nil {
		panic(fmt.Sprintf("failed to clean database: %v", err))
	}

	l := logger.NewLogger()
	store := dbstorage.NewStore(testDB)

	// Create repositories
	userStorage := &user_storage.UserStorage{DB: testDB, Ctx: ctx}
	orgStorage := &organization_storage.OrganizationStore{DB: testDB, Ctx: ctx}
	cache, err := cache.NewCache(getEnvOrDefault("REDIS_URL", "redis://localhost:6379"))
	if err != nil {
		panic(fmt.Sprintf("failed to create redis client: %v", err))
	}
	// Create services
	orgService := organization_service.NewOrganizationService(store, ctx, l, orgStorage, cache)
	authService := authService.NewAuthService(userStorage, l, orgService, ctx)

	return &TestSetup{
		DB:          testDB,
		Ctx:         ctx,
		Store:       store,
		Logger:      l,
		UserStorage: userStorage,
		OrgStorage:  orgStorage,
		OrgService:  orgService,
		AuthService: authService,
	}
}

// CreateTestUserAndOrg creates a test user and organization
// This should be called by individual test cases when needed
func (s *TestSetup) CreateTestUserAndOrg() (*types.User, *types.Organization, error) {
	authResponse, org, err := s.RegistrationHelper("test@example.com", "Password123@", "testuser", "test-org", "Test organization", "admin")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create test user: %w", err)
	}

	return &authResponse.User, org, nil
}

func (s *TestSetup) GetTestAuthResponse() (*authTypes.AuthResponse, *types.Organization, error) {
	authResponse, org, err := s.RegistrationHelper("test@example.com", "Password123@", "testuser", "test-org", "Test organization", "admin")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create test user: %w", err)
	}

	return authResponse, org, nil
}

func (s *TestSetup) RegistrationHelper(email, password, username, orgName, orgDescription string, userType string) (*authTypes.AuthResponse, *types.Organization, error) {
	// Create test user
	registrationRequest := authTypes.RegisterRequest{
		Email:    email,
		Password: password,
		Username: username,
		Type:     userType,
	}

	authResponse, err := s.AuthService.Register(registrationRequest, "admin")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create test user: %w", err)
	}

	// Create test organization
	org := &types.Organization{
		ID:          uuid.New(),
		Name:        "test-org",
		Description: "Test organization",
	}

	if err := s.OrgStorage.CreateOrganization(*org); err != nil {
		return nil, nil, fmt.Errorf("failed to create test organization: %w", err)
	}

	orgUser := &types.OrganizationUsers{
		ID:             uuid.New(),
		UserID:         authResponse.User.ID,
		OrganizationID: org.ID,
	}

	if err := s.OrgStorage.AddUserToOrganization(*orgUser); err != nil {
		return nil, nil, fmt.Errorf("failed to add user to organization: %w", err)
	}

	return &authResponse, org, nil
}
