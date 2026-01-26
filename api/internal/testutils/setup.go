package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	authService "github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	authTypes "github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	// organization_service "github.com/raghavyuva/nixopus-api/internal/features/organization/service"
	// organization_storage "github.com/raghavyuva/nixopus-api/internal/features/organization/storage"
	dbstorage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var (
	testDB  *bun.DB
	ctx     context.Context
	baseURL = "http://localhost:8080"
	// SuperTokens endpoints are at /auth/* (not /api/v1/auth/*)
	// The SuperTokens middleware handles these endpoints directly
	supertokensBaseURL = "http://localhost:8080"
)

// SuperTokensAuthResponse holds the authentication response from SuperTokens
type SuperTokensAuthResponse struct {
	Cookies        []*http.Cookie
	AccessToken    string
	User           *types.User
	OrganizationID string
}

// GetAuthCookiesHeader returns a formatted cookie header string for use in HTTP requests
func (s *SuperTokensAuthResponse) GetAuthCookiesHeader() string {
	var cookieStrs []string
	for _, cookie := range s.Cookies {
		cookieStrs = append(cookieStrs, cookie.Name+"="+cookie.Value)
	}
	return strings.Join(cookieStrs, "; ")
}

// TestSetup holds all the common test dependencies
type TestSetup struct {
	DB          *bun.DB
	Ctx         context.Context
	Store       *dbstorage.Store
	Logger      logger.Logger
	UserStorage *user_storage.UserStorage
	// OrgStorage  *organization_storage.OrganizationStore
	// OrgService  *organization_service.OrganizationService
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
	// orgStorage := &organization_storage.OrganizationStore{DB: testDB, Ctx: ctx}
	// cache, err := cache.NewCache(getEnvOrDefault("REDIS_URL", "redis://localhost:6379"))
	// if err != nil {
	// 	panic(fmt.Sprintf("failed to create redis client: %v", err))
	// }
	// Create services
	// orgService := organization_service.NewOrganizationService(store, ctx, l, orgStorage, cache)
	authService := authService.NewAuthService(userStorage, l, ctx)

	return &TestSetup{
		DB:          testDB,
		Ctx:         ctx,
		Store:       store,
		Logger:      l,
		UserStorage: userStorage,
		// OrgStorage:  orgStorage,
		// OrgService:  orgService,
		AuthService: authService,
	}
}

// CreateTestUserAndOrg creates a test user and organization
// This should be called by individual test cases when needed
// Deprecated: Use SignupViaSupertokens or SigninViaSupertokens instead
func (s *TestSetup) CreateTestUserAndOrg() (*types.User, *types.Organization, error) {
	// Use SuperTokens for authentication instead
	authResponse, err := s.SignupViaSupertokens("test@example.com", "Password123@")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create test user: %w", err)
	}

	// Get organization from the user's OrganizationUsers relation
	var org *types.Organization
	if len(authResponse.User.OrganizationUsers) > 0 && authResponse.User.OrganizationUsers[0].Organization != nil {
		org = authResponse.User.OrganizationUsers[0].Organization
	}

	return authResponse.User, org, nil
}

// GetTestAuthResponse is deprecated - use GetSupertokensAuthResponse instead
func (s *TestSetup) GetTestAuthResponse() (*authTypes.AuthResponse, *types.Organization, error) {
	// This function is deprecated since RegistrationHelper no longer works
	// Use GetSupertokensAuthResponse instead
	return nil, nil, fmt.Errorf("GetTestAuthResponse is deprecated - use GetSupertokensAuthResponse instead")
}

// RegistrationHelper is deprecated - Better Auth handles authentication now
// Use SignupViaSupertokens or SigninViaSupertokens instead
func (s *TestSetup) RegistrationHelper(email, password, username, orgName, orgDescription string, userType string) (*authTypes.AuthResponse, *types.Organization, error) {
	// This function is deprecated since AuthService.Register() no longer exists
	// Better Auth handles authentication now
	// Use SignupViaSupertokens or SigninViaSupertokens for test authentication
	return nil, nil, fmt.Errorf("RegistrationHelper is deprecated - use SignupViaSupertokens or SigninViaSupertokens instead")
	
	// Old implementation (commented out):
	// registrationRequest := authTypes.RegisterRequest{
	// 	Email:    email,
	// 	Password: password,
	// 	Username: username,
	// 	Type:     userType,
	// }
	// authResponse, err := s.AuthService.Register(registrationRequest, "admin")
	// if err != nil {
	// 	return nil, nil, fmt.Errorf("failed to create test user: %w", err)
	// }
	// org := &types.Organization{
	// 	ID:          uuid.New(),
	// 	Name:        "test-org",
	// 	Description: "Test organization",
	// }
	// if err := s.OrgStorage.CreateOrganization(*org); err != nil {
	// 	return nil, nil, fmt.Errorf("failed to create test organization: %w", err)
	// }
	// orgUser := &types.OrganizationUsers{
	// 	ID:             uuid.New(),
	// 	UserID:         authResponse.User.ID,
	// 	OrganizationID: org.ID,
	// }
	// if err := s.OrgStorage.AddUserToOrganization(*orgUser); err != nil {
	// 	return nil, nil, fmt.Errorf("failed to add user to organization: %w", err)
	// }
	// return &authResponse, org, nil
}

// SignupViaSupertokens creates a user through SuperTokens HTTP API and returns session cookies.
// This is the preferred method for integration tests that need to authenticate with protected endpoints.
func (s *TestSetup) SignupViaSupertokens(email, password string) (*SuperTokensAuthResponse, error) {
	// SuperTokens endpoints are handled by the middleware at /auth/* path
	signupURL := supertokensBaseURL + "/auth/signup"

	// Prepare the signup request body (SuperTokens email-password format)
	requestBody := map[string]interface{}{
		"formFields": []map[string]string{
			{"id": "email", "value": email},
			{"id": "password", "value": password},
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal signup request: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", signupURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create signup request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Add SuperTokens required headers for emailpassword recipe
	req.Header.Set("rid", "emailpassword")
	req.Header.Set("st-auth-mode", "cookie")

	// Create HTTP client with cookie jar to automatically capture cookies (like Postman)
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}
	client := &http.Client{
		Jar: jar,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute signup request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, _ := io.ReadAll(resp.Body)

	// Parse response to check for SuperTokens errors
	var respData map[string]interface{}
	if err := json.Unmarshal(respBody, &respData); err == nil {
		// Check if SuperTokens returned an error status
		if status, ok := respData["status"].(string); ok && status != "OK" {
			return nil, fmt.Errorf("signup failed: status=%s, response=%s", status, string(respBody))
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("signup failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Extract session cookies from the cookie jar (automatically captures cookies from Set-Cookie headers)
	cookieURL := resp.Request.URL
	if cookieURL == nil {
		cookieURL = req.URL
	}
	cookies := jar.Cookies(cookieURL)

	if len(cookies) == 0 {
		return nil, fmt.Errorf("no session cookies returned from signup")
	}

	// Extract access token from cookies
	var accessToken string
	for _, cookie := range cookies {
		if cookie.Name == "sAccessToken" {
			accessToken = cookie.Value
			break
		}
	}

	// Find user in our database by email (this also loads OrganizationUsers)
	user, err := s.UserStorage.FindUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user after signup: %w", err)
	}

	// Get user's organization from the loaded OrganizationUsers relation
	var orgID string
	if len(user.OrganizationUsers) > 0 && user.OrganizationUsers[0].Organization != nil {
		orgID = user.OrganizationUsers[0].Organization.ID.String()
	}

	return &SuperTokensAuthResponse{
		Cookies:        cookies,
		AccessToken:    accessToken,
		User:           user,
		OrganizationID: orgID,
	}, nil
}

// SigninViaSupertokens logs in a user through SuperTokens HTTP API and returns session cookies.
func (s *TestSetup) SigninViaSupertokens(email, password string) (*SuperTokensAuthResponse, error) {
	// SuperTokens endpoints are handled by the middleware at /auth/* path
	signinURL := supertokensBaseURL + "/auth/signin"

	// Prepare the signin request body (SuperTokens email-password format)
	requestBody := map[string]interface{}{
		"formFields": []map[string]string{
			{"id": "email", "value": email},
			{"id": "password", "value": password},
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal signin request: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", signinURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create signin request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Add SuperTokens required headers for emailpassword recipe
	req.Header.Set("rid", "emailpassword")
	req.Header.Set("st-auth-mode", "cookie")

	// Create HTTP client with cookie jar to automatically capture cookies (like Postman)
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}
	client := &http.Client{
		Jar: jar,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute signin request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for error checking
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("signin failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Check for SuperTokens error status in response
	var respData map[string]interface{}
	if err := json.Unmarshal(respBody, &respData); err == nil {
		if status, ok := respData["status"].(string); ok && status != "OK" {
			return nil, fmt.Errorf("signin failed: status=%s, response=%s", status, string(respBody))
		}
	}

	// Extract session cookies from the cookie jar (automatically captures cookies from Set-Cookie headers)
	cookieURL := resp.Request.URL
	if cookieURL == nil {
		cookieURL = req.URL
	}
	cookies := jar.Cookies(cookieURL)

	if len(cookies) == 0 {
		return nil, fmt.Errorf("no session cookies returned from signin")
	}

	// Extract access token from cookies if available
	var accessToken string
	for _, cookie := range cookies {
		if cookie.Name == "sAccessToken" {
			accessToken = cookie.Value
			break
		}
	}

	// Find user in our database by email (this also loads OrganizationUsers)
	user, err := s.UserStorage.FindUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user after signin: %w", err)
	}

	// Get user's organization from the loaded OrganizationUsers relation
	var orgID string
	if len(user.OrganizationUsers) > 0 && user.OrganizationUsers[0].Organization != nil {
		orgID = user.OrganizationUsers[0].Organization.ID.String()
	}

	return &SuperTokensAuthResponse{
		Cookies:        cookies,
		AccessToken:    accessToken,
		User:           user,
		OrganizationID: orgID,
	}, nil
}

// GetSupertokensAuthResponse creates a user via SuperTokens and returns authentication info.
// This should be used for integration tests that need to call protected API endpoints.
// It creates a new user with a unique email to avoid conflicts with existing SuperTokens users.
func (s *TestSetup) GetSupertokensAuthResponse() (*SuperTokensAuthResponse, error) {
	// Generate unique email to avoid conflicts with existing SuperTokens users
	uniqueEmail := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
	password := "Password123@"

	return s.SignupViaSupertokens(uniqueEmail, password)
}
