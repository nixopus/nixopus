package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/nixopus/nixopus/api/internal/secrets"
	"github.com/nixopus/nixopus/api/internal/storage"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

type errSecretsManager struct{}

func (e *errSecretsManager) GetSecret(ctx context.Context, key string) (string, error) {
	return "", errors.New("no secret")
}

func (e *errSecretsManager) GetSecrets(ctx context.Context, prefix string) (map[string]string, error) {
	return nil, errors.New("no secrets")
}

func resetInitHooks(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		initNewDB = storage.NewDB
		initStoreInit = defaultInitStoreInit
		initLogFatalf = log.Fatalf
		initLogFatal = log.Fatal
		initNewSecretManager = secrets.NewSecretManager
		initAfterViperHook = nil
		GlobalStore = nil
		AppConfig = types.Config{}
		viper.Reset()
	})
}

func minimalInitEnv(t *testing.T) {
	t.Helper()
	t.Setenv("HOST_NAME", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("USERNAME", "u")
	t.Setenv("PASSWORD", "p")
	t.Setenv("DB_NAME", "db")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("ALLOWED_ORIGIN", "http://localhost:3000")
	t.Setenv("AUTH_SERVICE_URL", "http://localhost:3000/api/auth")
	t.Setenv("AUTH_SERVICE_SECRET", "secret")
	t.Setenv("PORT", "9000")
	_ = os.Unsetenv("DATABASE_URL")
	_ = os.Unsetenv("NIXOPUS_CONFIG_PATH")
}

func TestDefaultInitStoreInit_nilStorePanics(t *testing.T) {
	defer func() {
		require.NotNil(t, recover(), "expected panic from Init on nil store")
	}()
	_ = defaultInitStoreInit(nil, context.Background())
}

func TestGetDeployDomain_and_BuildDeployDomainURL(t *testing.T) {
	resetInitHooks(t)
	_ = os.Unsetenv("DEPLOY_DOMAIN")
	AppConfig = types.Config{}
	AppConfig.App.DeployDomain = "custom.example"
	assert.Equal(t, "custom.example", GetDeployDomain())
	assert.Equal(t, "", BuildDeployDomainURL(""))
	assert.Equal(t, "", BuildDeployDomainURL("short"))
	assert.Equal(t, "https://12345678.custom.example", BuildDeployDomainURL("123456789012"))

	AppConfig.App.DeployDomain = ""
	t.Setenv("DEPLOY_DOMAIN", "from.env")
	assert.Equal(t, "from.env", GetDeployDomain())
	assert.Equal(t, "https://abcdefgh.from.env", BuildDeployDomainURL("abcdefghxyz"))

	_ = os.Unsetenv("DEPLOY_DOMAIN")
	assert.Equal(t, "nixopus.com", GetDeployDomain())
}

func TestParseDatabaseURL_invalidURL(t *testing.T) {
	db := &types.DatabaseConfig{URL: "http://["}
	err := parseDatabaseURL(db)
	require.Error(t, err)
}

func TestParseDatabaseURL_sslFromQueryOnly(t *testing.T) {
	db := &types.DatabaseConfig{
		URL: "postgresql://h/db?sslmode=prefer",
	}
	require.NoError(t, parseDatabaseURL(db))
	assert.Equal(t, "prefer", db.SSLMode)
}

func TestParseDatabaseURL_passwordWhenUserHasPassword(t *testing.T) {
	db := &types.DatabaseConfig{
		URL: "postgresql://u:pw@host:5432/dbname",
	}
	require.NoError(t, parseDatabaseURL(db))
	assert.Equal(t, "pw", db.Password)
}

func TestValidateConfig_individualErrors(t *testing.T) {
	base := types.Config{
		Server: types.ServerConfig{Port: "1"},
		Database: types.DatabaseConfig{
			Host: "h", Port: "5432", Username: "u", Password: "p", Name: "n",
		},
		Redis:      types.RedisConfig{URL: "redis://x"},
		CORS:       types.CORSConfig{AllowedOrigin: "http://x"},
		BetterAuth: types.BetterAuthConfig{URL: "http://a", Secret: "s"},
	}
	assert.ErrorContains(t, validateConfig(mutate(base, func(c *types.Config) { c.Server.Port = "" })), "server port")
	assert.ErrorContains(t, validateConfig(mutate(base, func(c *types.Config) { c.Database.Host = "" })), "database host")
	assert.ErrorContains(t, validateConfig(mutate(base, func(c *types.Config) { c.Database.Port = "" })), "database port")
	assert.ErrorContains(t, validateConfig(mutate(base, func(c *types.Config) { c.Database.Username = "" })), "database username")
	assert.ErrorContains(t, validateConfig(mutate(base, func(c *types.Config) { c.Database.Password = "" })), "database password")
	assert.ErrorContains(t, validateConfig(mutate(base, func(c *types.Config) { c.Database.Name = "" })), "database name")
	assert.ErrorContains(t, validateConfig(mutate(base, func(c *types.Config) { c.Redis.URL = "" })), "redis URL")
	assert.ErrorContains(t, validateConfig(mutate(base, func(c *types.Config) { c.CORS.AllowedOrigin = "" })), "CORS allowed origin")
	assert.ErrorContains(t, validateConfig(mutate(base, func(c *types.Config) { c.BetterAuth.URL = "" })), "Better Auth URL")
	assert.ErrorContains(t, validateConfig(mutate(base, func(c *types.Config) { c.BetterAuth.Secret = "" })), "Better Auth secret")
}

func mutate(c types.Config, fn func(*types.Config)) types.Config {
	fn(&c)
	return c
}

func TestGetConfigFileName_variants(t *testing.T) {
	for _, tc := range []struct{ env, want string }{
		{"DEV", "config.dev"},
		{"STAGE", "config.staging"},
		{"PROD", "config.prod"},
	} {
		t.Run(tc.env, func(t *testing.T) {
			t.Setenv("ENV", tc.env)
			assert.Equal(t, tc.want, getConfigFileName())
		})
	}
}

func TestInitViper_customPathAndReadErrors(t *testing.T) {
	resetInitHooks(t)
	dir := t.TempDir()
	t.Setenv("NIXOPUS_CONFIG_PATH", dir)
	t.Setenv("ENV", "production")
	_ = os.Unsetenv("PORT")
	viper.Reset()
	initViper()
	assert.NotPanics(t, func() { initViper() })

	viper.Reset()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.prod.yaml"), []byte("bad: [\n"), 0o600))
	initViper()
}

func TestInitViper_readsExistingFile(t *testing.T) {
	resetInitHooks(t)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.prod.yaml"), []byte("server:\n  port: \"7777\"\n"), 0o600))
	t.Setenv("NIXOPUS_CONFIG_PATH", dir)
	t.Setenv("ENV", "production")
	viper.Reset()
	initViper()
	assert.Equal(t, "7777", viper.GetString("server.port"))
}

func TestInit_secretManagerInitError(t *testing.T) {
	resetInitHooks(t)
	resetMoverHooks(t)
	minimalInitEnv(t)
	t.Setenv("SECRET_MANAGER_ENABLED", "true")
	t.Setenv("SECRET_MANAGER_TYPE", "infisical")
	_ = os.Unsetenv("INFISICAL_TOKEN")
	viper.Reset()
	initNewDB = func(*storage.Config) (*bun.DB, error) { return nil, nil }
	initStoreInit = func(*storage.Store, context.Context) error { return nil }
	_ = Init()
}

func TestInit_secretManagerLoadSecretsWarning(t *testing.T) {
	resetInitHooks(t)
	resetMoverHooks(t)
	minimalInitEnv(t)
	t.Setenv("SECRET_MANAGER_ENABLED", "true")
	t.Setenv("SECRET_MANAGER_TYPE", "infisical")
	t.Setenv("INFISICAL_TOKEN", "tok")
	initNewSecretManager = func(*secrets.SecretManagerConfig) (secrets.SecretManager, error) {
		return &errSecretsManager{}, nil
	}
	viper.Reset()
	initNewDB = func(*storage.Config) (*bun.DB, error) { return nil, nil }
	initStoreInit = func(*storage.Store, context.Context) error { return nil }
	_ = Init()
}

func TestInit_unmarshalFatal(t *testing.T) {
	resetInitHooks(t)
	resetMoverHooks(t)
	minimalInitEnv(t)
	viper.Reset()
	initAfterViperHook = func() { viper.Set("database.max_open_conn", "not-an-int") }
	var got string
	initLogFatalf = func(format string, v ...interface{}) {
		got = fmt.Sprintf(format, v...)
		panic("fatal")
	}
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic")
		assert.Contains(t, got, "Failed to unmarshal config")
	}()
	Init()
}

func TestInit_parseDatabaseURLFatal(t *testing.T) {
	resetInitHooks(t)
	resetMoverHooks(t)
	minimalInitEnv(t)
	t.Setenv("DATABASE_URL", "http://[")
	viper.Reset()
	var got string
	initLogFatalf = func(format string, v ...interface{}) {
		got = fmt.Sprintf(format, v...)
		panic("fatal")
	}
	defer func() {
		require.NotNil(t, recover())
		assert.Contains(t, got, "Failed to parse DATABASE_URL")
	}()
	Init()
}

func TestInit_validateFatal(t *testing.T) {
	resetInitHooks(t)
	resetMoverHooks(t)
	minimalInitEnv(t)
	_ = os.Unsetenv("REDIS_URL")
	viper.Reset()
	var got string
	initLogFatalf = func(format string, v ...interface{}) {
		got = fmt.Sprintf(format, v...)
		panic("fatal")
	}
	defer func() {
		require.NotNil(t, recover())
		assert.Contains(t, got, "Configuration validation failed")
	}()
	Init()
}

func TestInit_newDBFatal(t *testing.T) {
	resetInitHooks(t)
	resetMoverHooks(t)
	minimalInitEnv(t)
	viper.Reset()
	initNewDB = func(*storage.Config) (*bun.DB, error) { return nil, errors.New("open db") }
	var got string
	initLogFatal = func(v ...interface{}) {
		got = fmt.Sprint(v...)
		panic("fatal")
	}
	defer func() {
		require.NotNil(t, recover())
		assert.Contains(t, got, "open db")
	}()
	Init()
}

func TestInit_storeInitFatal(t *testing.T) {
	resetInitHooks(t)
	resetMoverHooks(t)
	minimalInitEnv(t)
	viper.Reset()
	initNewDB = func(*storage.Config) (*bun.DB, error) { return nil, nil }
	initStoreInit = func(*storage.Store, context.Context) error { return errors.New("store init") }
	var got string
	initLogFatalf = func(format string, v ...interface{}) {
		got = fmt.Sprintf(format, v...)
		panic("fatal")
	}
	defer func() {
		require.NotNil(t, recover())
		assert.Contains(t, got, "Failed to initialize storage")
	}()
	Init()
}

func TestInit_success_stubbedDB(t *testing.T) {
	resetInitHooks(t)
	resetMoverHooks(t)
	minimalInitEnv(t)
	_ = os.Unsetenv("PORT")
	viper.Reset()
	initNewDB = func(*storage.Config) (*bun.DB, error) { return nil, nil }
	initStoreInit = func(*storage.Store, context.Context) error { return nil }
	store := Init()
	require.NotNil(t, store)
	assert.Equal(t, "8080", AppConfig.Server.Port)
	assert.NotNil(t, GlobalStore)
}
