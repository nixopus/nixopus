package config

import (
	"os"
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigLoading(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedConfig types.Config
		description    string
	}{
		{
			name: "Load from environment variables",
			envVars: map[string]string{
				"PORT":           "9090",
				"HOST_NAME":      "test-db-host",
				"DB_PORT":        "5433",
				"USERNAME":       "testuser",
				"PASSWORD":       "testpass",
				"DB_NAME":        "testdb",
				"SSL_MODE":       "require",
				"REDIS_URL":      "redis://test-redis:6379",
				"SSH_HOST":       "test-ssh-host",
				"SSH_PORT":       "2222",
				"SSH_USER":       "sshuser",
				"CADDY_ENDPOINT": "http://test-caddy:2019",
				"ALLOWED_ORIGIN": "http://test-frontend:3000",
				"ENV":            "test",
				"APP_VERSION":    "1.0.0-test",
				"LOGS_PATH":      "/test/logs",
			},
			expectedConfig: types.Config{
				Server: types.ServerConfig{
					Port: "9090",
				},
				Database: types.DatabaseConfig{
					Host:     "test-db-host",
					Port:     "5433",
					Username: "testuser",
					Password: "testpass",
					Name:     "testdb",
					SSLMode:  "require",
				},
				Redis: types.RedisConfig{
					URL: "redis://test-redis:6379",
				},
				Proxy: types.ProxyConfig{
					CaddyEndpoint: "http://test-caddy:2019",
				},
				CORS: types.CORSConfig{
					AllowedOrigin: "http://test-frontend:3000",
				},
				App: types.AppConfig{
					Environment: "test",
					Version:     "1.0.0-test",
					LogsPath:    "/test/logs",
				},
			},
			description: "Should load configuration from environment variables",
		},
		{
			name: "Load with missing environment variables",
			envVars: map[string]string{
				"PORT":      "8080",
				"HOST_NAME": "localhost",
				"DB_PORT":   "5432",
			},
			expectedConfig: types.Config{
				Server: types.ServerConfig{
					Port: "8080",
				},
				Database: types.DatabaseConfig{
					Host: "localhost",
					Port: "5432",
				},
			},
			description: "Should load partial configuration from environment variables",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			viper.Reset()

			initViper()

			var config types.Config
			err := viper.Unmarshal(&config)
			require.NoError(t, err, "Failed to unmarshal config")

			assert.Equal(t, tt.expectedConfig.Server.Port, config.Server.Port, "Server port mismatch")
			assert.Equal(t, tt.expectedConfig.Database.Host, config.Database.Host, "Database host mismatch")
			assert.Equal(t, tt.expectedConfig.Database.Port, config.Database.Port, "Database port mismatch")
			assert.Equal(t, tt.expectedConfig.Database.Username, config.Database.Username, "Database username mismatch")
			assert.Equal(t, tt.expectedConfig.Database.Password, config.Database.Password, "Database password mismatch")
			assert.Equal(t, tt.expectedConfig.Database.Name, config.Database.Name, "Database name mismatch")
			assert.Equal(t, tt.expectedConfig.Database.SSLMode, config.Database.SSLMode, "Database SSL mode mismatch")
			assert.Equal(t, tt.expectedConfig.Redis.URL, config.Redis.URL, "Redis URL mismatch")
			assert.Equal(t, tt.expectedConfig.Proxy.CaddyEndpoint, config.Proxy.CaddyEndpoint, "Caddy endpoint mismatch")
			assert.Equal(t, tt.expectedConfig.CORS.AllowedOrigin, config.CORS.AllowedOrigin, "Allowed origin mismatch")
			assert.Equal(t, tt.expectedConfig.App.Environment, config.App.Environment, "Environment mismatch")
			assert.Equal(t, tt.expectedConfig.App.Version, config.App.Version, "Version mismatch")
			assert.Equal(t, tt.expectedConfig.App.LogsPath, config.App.LogsPath, "Logs path mismatch")
		})
	}
}

func TestConfigPathResolution(t *testing.T) {
	t.Run("Test config path resolution", func(t *testing.T) {
		viper.Reset()

		initViper()

		assert.NotPanics(t, func() {
			initViper()
		}, "initViper should not panic even if config file is not found")
	})
}

func TestEnvironmentVariablePrecedence(t *testing.T) {
	t.Run("Environment variables override config file", func(t *testing.T) {
		os.Setenv("PORT", "9999")
		defer os.Unsetenv("PORT")

		viper.Reset()

		initViper()

		port := viper.GetString("server.port")
		assert.Equal(t, "9999", port, "Environment variable should override config file")
	})
}

func TestConfigValidation(t *testing.T) {
	t.Run("Required fields validation", func(t *testing.T) {
		os.Setenv("PORT", "8080")
		os.Setenv("HOST_NAME", "localhost")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("USERNAME", "postgres")
		os.Setenv("PASSWORD", "password")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("REDIS_URL", "redis://localhost:6379")
		os.Setenv("SSH_HOST", "localhost")
		os.Setenv("SSH_USER", "root")
		os.Setenv("CADDY_ENDPOINT", "http://localhost:2019")
		os.Setenv("ALLOWED_ORIGIN", "http://localhost:3000")

		defer func() {
			os.Unsetenv("PORT")
			os.Unsetenv("HOST_NAME")
			os.Unsetenv("DB_PORT")
			os.Unsetenv("USERNAME")
			os.Unsetenv("PASSWORD")
			os.Unsetenv("DB_NAME")
			os.Unsetenv("REDIS_URL")
			os.Unsetenv("SSH_HOST")
			os.Unsetenv("SSH_USER")
			os.Unsetenv("CADDY_ENDPOINT")
			os.Unsetenv("ALLOWED_ORIGIN")
		}()

		viper.Reset()

		initViper()

		var config types.Config
		err := viper.Unmarshal(&config)
		require.NoError(t, err, "Failed to unmarshal config")

		assert.NotEmpty(t, config.Server.Port, "Server port should not be empty")
		assert.NotEmpty(t, config.Database.Host, "Database host should not be empty")
		assert.NotEmpty(t, config.Database.Port, "Database port should not be empty")
		assert.NotEmpty(t, config.Database.Username, "Database username should not be empty")
		assert.NotEmpty(t, config.Database.Password, "Database password should not be empty")
		assert.NotEmpty(t, config.Database.Name, "Database name should not be empty")
		assert.NotEmpty(t, config.Redis.URL, "Redis URL should not be empty")
		assert.NotEmpty(t, config.Proxy.CaddyEndpoint, "Caddy endpoint should not be empty")
		assert.NotEmpty(t, config.CORS.AllowedOrigin, "Allowed origin should not be empty")
	})
}

func TestProductionEnvironmentSimulation(t *testing.T) {
	t.Run("Simulate production with mounted configs", func(t *testing.T) {
		prodEnvVars := map[string]string{
			"PORT":            "8443",
			"HOST_NAME":       "nixopus-db",
			"DB_PORT":         "5432",
			"USERNAME":        "postgres",
			"PASSWORD":        "production-password",
			"DB_NAME":         "nixopus",
			"SSL_MODE":        "require",
			"REDIS_URL":       "redis://nixopus-redis:6379",
			"SSH_HOST":        "production-host",
			"SSH_PORT":        "22",
			"SSH_USER":        "root",
			"SSH_PRIVATE_KEY": "/etc/nixopus/ssh/id_rsa",
			"CADDY_ENDPOINT":  "http://nixopus-caddy:2019",
			"ALLOWED_ORIGIN":  "https://app.nixopus.com",
			"ENV":             "production",
			"APP_VERSION":     "1.0.0",
			"LOGS_PATH":       "/var/log/nixopus",
		}

		for key, value := range prodEnvVars {
			os.Setenv(key, value)
			defer os.Unsetenv(key)
		}

		viper.Reset()

		initViper()

		var config types.Config
		err := viper.Unmarshal(&config)
		require.NoError(t, err, "Failed to unmarshal production config")

		assert.Equal(t, "8443", config.Server.Port, "Production port should be 8443")
		assert.Equal(t, "nixopus-db", config.Database.Host, "Production DB host should be nixopus-db")
		assert.Equal(t, "production", config.App.Environment, "Environment should be production")
		assert.Equal(t, "https://app.nixopus.com", config.CORS.AllowedOrigin, "Production allowed origin should be HTTPS")
		assert.Equal(t, "/var/log/nixopus", config.App.LogsPath, "Production logs path should be /var/log/nixopus")
	})
}

func TestConfigAccessMethods(t *testing.T) {
	t.Run("Test direct config access", func(t *testing.T) {
		os.Setenv("PORT", "7070")
		os.Setenv("HOST_NAME", "test-host")
		defer func() {
			os.Unsetenv("PORT")
			os.Unsetenv("HOST_NAME")
		}()

		viper.Reset()

		initViper()

		port := viper.GetString("server.port")
		host := viper.GetString("database.host")

		assert.Equal(t, "7070", port, "Direct viper access should work")
		assert.Equal(t, "test-host", host, "Direct viper access should work")

		var config types.Config
		err := viper.Unmarshal(&config)
		require.NoError(t, err, "Failed to unmarshal config")

		assert.Equal(t, "7070", config.Server.Port, "Config struct access should work")
		assert.Equal(t, "test-host", config.Database.Host, "Config struct access should work")
	})
}

func TestDevConfigLoading(t *testing.T) {
	t.Run("Test config.dev.yaml loading", func(t *testing.T) {
		viper.Reset()

		viper.SetConfigName("config.dev")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("../../../helpers")
		viper.AddConfigPath("../../helpers")
		viper.AddConfigPath("../helpers")
		viper.AddConfigPath(".")

		err := viper.ReadInConfig()

		assert.NoError(t, err, "Should be able to read config.dev.yaml")

		version := viper.GetString("version")
		assert.Equal(t, "1", version, "Version should be 1")

		services := viper.Get("services")
		assert.NotNil(t, services, "Services section should exist")
	})
}

func TestEnvironmentBasedConfigSelection(t *testing.T) {
	tests := []struct {
		name           string
		envValue       string
		expectedConfig string
		description    string
	}{
		{
			name:           "Development environment",
			envValue:       "development",
			expectedConfig: "config.dev",
			description:    "Should select config.dev for development environment",
		},
		{
			name:           "Production environment",
			envValue:       "production",
			expectedConfig: "config.prod",
			description:    "Should select config.prod for production environment",
		},
		{
			name:           "Staging environment",
			envValue:       "staging",
			expectedConfig: "config.staging",
			description:    "Should select config.staging for staging environment",
		},
		{
			name:           "Default environment",
			envValue:       "",
			expectedConfig: "config.prod",
			description:    "Should default to config.prod when no environment is specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("ENV", tt.envValue)
				defer os.Unsetenv("ENV")
			}

			configName := getConfigFileName()
			assert.Equal(t, tt.expectedConfig, configName, tt.description)
		})
	}
}

func TestConfigurationValidation(t *testing.T) {
	t.Run("Valid configuration should pass validation", func(t *testing.T) {
		validConfig := types.Config{
			Server: types.ServerConfig{
				Port: "8080",
			},
			Database: types.DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				Username: "postgres",
				Password: "password",
				Name:     "testdb",
			},
			Redis: types.RedisConfig{
				URL: "redis://localhost:6379",
			},
			Proxy: types.ProxyConfig{
				CaddyEndpoint: "http://localhost:2019",
			},
			CORS: types.CORSConfig{
				AllowedOrigin: "http://localhost:3000",
			},
		}

		err := validateConfig(validConfig)
		assert.NoError(t, err, "Valid configuration should pass validation")
	})

	t.Run("Invalid configuration should fail validation", func(t *testing.T) {
		invalidConfig := types.Config{
			Server: types.ServerConfig{
				Port: "", // Missing port
			},
			Database: types.DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				Username: "", // Missing username
				Password: "password",
				Name:     "testdb",
			},
			Redis: types.RedisConfig{
				URL: "", // Missing Redis URL
			},
			Proxy: types.ProxyConfig{
				CaddyEndpoint: "http://localhost:2019",
			},
			CORS: types.CORSConfig{
				AllowedOrigin: "http://localhost:3000",
			},
		}

		err := validateConfig(invalidConfig)
		assert.Error(t, err, "Invalid configuration should fail validation")

		errorMsg := err.Error()
		assert.Contains(t, errorMsg, "server port is required")
		assert.Contains(t, errorMsg, "database username is required")
		assert.Contains(t, errorMsg, "redis URL is required")
		assert.Contains(t, errorMsg, "SSH user is required")
	})
}

func TestBetterErrorHandling(t *testing.T) {
	t.Run("Test config file not found error handling", func(t *testing.T) {
		viper.Reset()

		viper.SetConfigName("non-existent-config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")

		assert.NotPanics(t, func() {
			err := viper.ReadInConfig()
			assert.Error(t, err, "Should return error for non-existent config file")
		}, "Should handle config file not found gracefully")
	})
}
