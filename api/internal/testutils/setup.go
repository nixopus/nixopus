package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	authService "github.com/nixopus/nixopus/api/internal/features/auth/service"
	user_storage "github.com/nixopus/nixopus/api/internal/features/auth/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	dbstorage "github.com/nixopus/nixopus/api/internal/storage"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var (
	testDB         *bun.DB
	ctx            context.Context
	baseURL        = "http://localhost:8080"
	authServiceURL = "http://localhost:9090"
)

// TestAuthResponse holds the authentication response from Better Auth test utils.
type TestAuthResponse struct {
	Cookies        []*http.Cookie
	AccessToken    string
	User           *types.User
	OrganizationID string
}

// GetAuthCookiesHeader returns a formatted cookie header string for use in HTTP requests.
func (r *TestAuthResponse) GetAuthCookiesHeader() string {
	var cookieStrs []string
	for _, cookie := range r.Cookies {
		cookieStrs = append(cookieStrs, cookie.Name+"="+cookie.Value)
	}
	return strings.Join(cookieStrs, "; ")
}

// TestSetup holds all the common test dependencies.
type TestSetup struct {
	DB          *bun.DB
	Ctx         context.Context
	Store       *dbstorage.Store
	Logger      logger.Logger
	UserStorage *user_storage.UserStorage
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

	if url := os.Getenv("AUTH_SERVICE_URL"); url != "" {
		authServiceURL = url
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
		fmt.Printf("Warning: store.Init failed (non-fatal for tests): %v\n", err)
	}

	fmt.Println("Successfully connected and initialized test database")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func cleanDatabase() error {
	// Truncate all public tables instead of dropping them.
	// Schema is managed by the auth service's drizzle migrations,
	// so we only clear data between tests.
	_, err := testDB.ExecContext(ctx, `
		DO $$ DECLARE
			r RECORD;
		BEGIN
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
				EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE';
			END LOOP;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	// Flush Redis to clear any cached state (e.g. admin_registered).
	redisURL := getEnvOrDefault("REDIS_URL", "redis://localhost:6379")
	opt, err := redis.ParseURL(redisURL)
	if err == nil {
		rdb := redis.NewClient(opt)
		defer rdb.Close()
		rdb.FlushDB(ctx)
	}

	return nil
}

// NewTestSetup creates a new test setup with all common dependencies.
func NewTestSetup() *TestSetup {
	if testDB == nil {
		panic("testDB is nil - database not initialized")
	}
	if ctx == nil {
		panic("ctx is nil - context not initialized")
	}

	if err := cleanDatabase(); err != nil {
		panic(fmt.Sprintf("failed to clean database: %v", err))
	}

	l := logger.NewLogger()
	store := dbstorage.NewStore(testDB)

	userStorage := &user_storage.UserStorage{DB: testDB, Ctx: ctx}
	authSvc := authService.NewAuthService(userStorage, l, ctx, "")

	return &TestSetup{
		DB:          testDB,
		Ctx:         ctx,
		Store:       store,
		Logger:      l,
		UserStorage: userStorage,
		AuthService: authSvc,
	}
}

// postJSON sends a POST request to the auth service test utils API.
func postJSON(url string, payload interface{}) ([]byte, int, error) {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}

	return body, resp.StatusCode, nil
}

// CreateTestUserViaAuth creates a user through the Better Auth test utils API
// and returns a session with cookies for authenticated requests.
//
// The OrganizationID in the response is sourced from the session's
// activeOrganizationId (set by nixopus-auth's setupNewUser hook), which is the
// same org the API middleware resolves on every request. This ensures test data
// inserted using OrganizationID will be found by API handlers.
func (s *TestSetup) CreateTestUserViaAuth(email, name string) (*TestAuthResponse, error) {
	saveUserURL := authServiceURL + "/api/test/save-user"
	body, status, err := postJSON(saveUserURL, map[string]interface{}{
		"email":         email,
		"name":          name,
		"emailVerified": true,
	})
	if err != nil {
		return nil, fmt.Errorf("save-user request failed: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("save-user failed with status %d: %s", status, string(body))
	}

	var savedUser struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &savedUser); err != nil {
		return nil, fmt.Errorf("failed to parse save-user response: %w (body: %s)", err, string(body))
	}

	loginURL := authServiceURL + "/api/test/login"
	body, status, err = postJSON(loginURL, map[string]interface{}{
		"userId": savedUser.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("login failed with status %d: %s", status, string(body))
	}

	var loginResp struct {
		Session struct {
			Token string `json:"token"`
		} `json:"session"`
		Token   string `json:"token"`
		Cookies []struct {
			Name     string `json:"name"`
			Value    string `json:"value"`
			Domain   string `json:"domain"`
			Path     string `json:"path"`
			HTTPOnly bool   `json:"httpOnly"`
			Secure   bool   `json:"secure"`
			SameSite string `json:"sameSite"`
		} `json:"cookies"`
	}
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return nil, fmt.Errorf("failed to parse login response: %w (body: %s)", err, string(body))
	}

	var cookies []*http.Cookie
	var accessToken string
	for _, c := range loginResp.Cookies {
		cookie := &http.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			HttpOnly: c.HTTPOnly,
			Secure:   c.Secure,
		}
		cookies = append(cookies, cookie)
		if strings.Contains(c.Name, "session_token") {
			accessToken = c.Value
		}
	}

	if accessToken == "" {
		accessToken = loginResp.Token
	}
	if accessToken == "" {
		accessToken = loginResp.Session.Token
	}

	// Resolve the org ID that the session actually uses. Better Auth's
	// setupNewUser hook creates a default org and sets activeOrganizationId on
	// the session; using anything else causes a mismatch with the API middleware.
	var orgID string
	_ = s.DB.NewRaw(
		"SELECT active_organization_id FROM session WHERE token = ?",
		accessToken,
	).Scan(ctx, &orgID)

	if orgID == "" {
		// Fallback: query the user's earliest membership (matches getInitialOrganization).
		_ = s.DB.NewRaw(
			"SELECT organization_id FROM member WHERE user_id = ? ORDER BY created_at ASC LIMIT 1",
			savedUser.ID,
		).Scan(ctx, &orgID)
	}

	user, err := s.UserStorage.FindUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user after creation: %w", err)
	}

	return &TestAuthResponse{
		Cookies:        cookies,
		AccessToken:    accessToken,
		User:           user,
		OrganizationID: orgID,
	}, nil
}

// LoginTestUser creates a session for an existing user via Better Auth test utils.
func (s *TestSetup) LoginTestUser(userID string) (*TestAuthResponse, error) {
	loginURL := authServiceURL + "/api/test/login"
	body, status, err := postJSON(loginURL, map[string]interface{}{
		"userId": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("login failed with status %d: %s", status, string(body))
	}

	var loginResp struct {
		Session struct {
			Token string `json:"token"`
		} `json:"session"`
		User struct {
			Email string `json:"email"`
		} `json:"user"`
		Token   string `json:"token"`
		Cookies []struct {
			Name     string `json:"name"`
			Value    string `json:"value"`
			Domain   string `json:"domain"`
			Path     string `json:"path"`
			HTTPOnly bool   `json:"httpOnly"`
			Secure   bool   `json:"secure"`
		} `json:"cookies"`
	}
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return nil, fmt.Errorf("failed to parse login response: %w (body: %s)", err, string(body))
	}

	var cookies []*http.Cookie
	var accessToken string
	for _, c := range loginResp.Cookies {
		cookie := &http.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			HttpOnly: c.HTTPOnly,
			Secure:   c.Secure,
		}
		cookies = append(cookies, cookie)
		if strings.Contains(c.Name, "session_token") {
			accessToken = c.Value
		}
	}
	if accessToken == "" {
		accessToken = loginResp.Token
	}
	if accessToken == "" {
		accessToken = loginResp.Session.Token
	}

	user, err := s.UserStorage.FindUserByEmail(loginResp.User.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	var orgID string
	if len(user.OrganizationUsers) > 0 && user.OrganizationUsers[0].Organization != nil {
		orgID = user.OrganizationUsers[0].Organization.ID.String()
	}

	return &TestAuthResponse{
		Cookies:        cookies,
		AccessToken:    accessToken,
		User:           user,
		OrganizationID: orgID,
	}, nil
}

// CreateTestUserAndOrg creates a test user and organization via Better Auth test utils.
func (s *TestSetup) CreateTestUserAndOrg() (*types.User, *types.Organization, error) {
	authResponse, err := s.CreateTestUserViaAuth("test@example.com", "Test User")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create test user: %w", err)
	}

	var org *types.Organization
	if len(authResponse.User.OrganizationUsers) > 0 && authResponse.User.OrganizationUsers[0].Organization != nil {
		org = authResponse.User.OrganizationUsers[0].Organization
	}

	return authResponse.User, org, nil
}

// GetAuthResponse creates a user via Better Auth test utils and returns authentication info.
// Creates a new user with a unique email to avoid conflicts.
func (s *TestSetup) GetAuthResponse() (*TestAuthResponse, error) {
	uniqueEmail := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
	return s.CreateTestUserViaAuth(uniqueEmail, "Test User")
}

// SeedCredentialAccount inserts a row into the `account` table with a dummy password hash.
// This makes the user discoverable by IsAdminRegistered, which checks for
// account rows where password IS NOT NULL.
func (s *TestSetup) SeedCredentialAccount(userID string) error {
	_, err := s.DB.NewRaw(
		"INSERT INTO account (id, account_id, provider_id, user_id, password, created_at, updated_at) "+
			"VALUES (gen_random_uuid(), ?, 'credential', ?, '$2a$10$dummyhashfortest', NOW(), NOW()) "+
			"ON CONFLICT DO NOTHING",
		userID, userID,
	).Exec(ctx)
	return err
}

// SeedDeploymentWithArtifact inserts a deployment row with S3 artifact metadata
// for an existing application. Returns the deployment ID.
func (s *TestSetup) SeedDeploymentWithArtifact(appID, s3Key string, imageSize int64) (string, error) {
	var deploymentID string
	err := s.DB.NewRaw(
		`INSERT INTO application_deployment
			(id, application_id, commit_hash, image_s3_key, image_size, created_at, updated_at)
		 VALUES
			(gen_random_uuid(), ?, 'abc123', ?, ?, NOW(), NOW())
		 RETURNING id`,
		appID, s3Key, imageSize,
	).Scan(ctx, &deploymentID)
	if err != nil {
		return "", fmt.Errorf("failed to seed deployment with artifact: %w", err)
	}
	return deploymentID, nil
}

// SeedDeploymentWithoutArtifact inserts a deployment row without S3 artifact metadata.
func (s *TestSetup) SeedDeploymentWithoutArtifact(appID string) (string, error) {
	var deploymentID string
	err := s.DB.NewRaw(
		`INSERT INTO application_deployment
			(id, application_id, commit_hash, created_at, updated_at)
		 VALUES
			(gen_random_uuid(), ?, 'def456', NOW(), NOW())
		 RETURNING id`,
		appID,
	).Scan(ctx, &deploymentID)
	if err != nil {
		return "", fmt.Errorf("failed to seed deployment without artifact: %w", err)
	}
	return deploymentID, nil
}

// SeedOrganizationSettings inserts or updates organization settings with the given data.
func (s *TestSetup) SeedOrganizationSettings(organizationID string, settingsJSON string) error {
	_, err := s.DB.NewRaw(
		`INSERT INTO organization_settings (id, organization_id, settings, created_at, updated_at)
		 VALUES (gen_random_uuid(), ?, ?::jsonb, NOW(), NOW())
		 ON CONFLICT (organization_id)
		 DO UPDATE SET settings = ?::jsonb, updated_at = NOW()`,
		organizationID, settingsJSON, settingsJSON,
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to seed organization settings: %w", err)
	}
	return nil
}

// SeedApplication inserts a minimal application record so that healthcheck
// (and similar) tests can reference a valid application_id without requiring
// the full deploy pipeline.
func (s *TestSetup) SeedApplication(userID, organizationID string) (string, error) {
	appID := "00000000-0000-0000-0000-000000000001"
	_, err := s.DB.NewRaw(
		`INSERT INTO applications
			(id, name, port, environment, build_variables, environment_variables,
			 build_pack, repository, branch, pre_run_command, post_run_command,
			 user_id, organization_id)
		 VALUES
			(?, 'test-app', 3000, 'production', '', '',
			 'dockerfile', 'https://github.com/test/repo', 'main', '', '',
			 ?, ?)
		 ON CONFLICT (id) DO NOTHING`,
		appID, userID, organizationID,
	).Exec(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to seed application: %w", err)
	}
	return appID, nil
}