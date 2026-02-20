package types

import (
	"log"
	"os"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Proxy      ProxyConfig      `mapstructure:"proxy"`
	CORS       CORSConfig       `mapstructure:"cors"`
	App        AppConfig        `mapstructure:"app"`
	GitHub     GitHubConfig     `mapstructure:"github"`
	BetterAuth BetterAuthConfig `mapstructure:"betterauth"`
	Stripe     StripeConfig     `mapstructure:"stripe"`
	Agent      AgentConfig      `mapstructure:"agent"`
	Trail      TrailConfig      `mapstructure:"trail"`
	S3         S3Config         `mapstructure:"s3"`
}

type S3Config struct {
	Endpoint  string `mapstructure:"endpoint"`
	Bucket    string `mapstructure:"bucket"`
	Region    string `mapstructure:"region"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

type AgentConfig struct {
	Endpoint string `mapstructure:"endpoint"`
}

type BetterAuthConfig struct {
	URL    string `mapstructure:"url" validate:"required"`
	Secret string `mapstructure:"secret" validate:"required"`
}

type ServerConfig struct {
	Port    string `mapstructure:"port" validate:"required"`
	MCPPort string `mapstructure:"mcp_port"`
}

type DatabaseConfig struct {
	Host        string `mapstructure:"host" validate:"required"`
	Port        string `mapstructure:"port" validate:"required"`
	Username    string `mapstructure:"username" validate:"required"`
	Password    string `mapstructure:"password" validate:"required"`
	Name        string `mapstructure:"name" validate:"required"`
	SSLMode     string `mapstructure:"ssl_mode"`
	MaxOpenConn int    `mapstructure:"max_open_conn"`
	MaxIdleConn int    `mapstructure:"max_idle_conn"`
	Debug       bool   `mapstructure:"debug"`
}

type RedisConfig struct {
	URL string `mapstructure:"url" validate:"required"`
}

type SSHConfig struct {
	Host                string `mapstructure:"host" validate:"required"`
	Port                uint   `mapstructure:"port"`
	User                string `mapstructure:"user" validate:"required"`
	Password            string `mapstructure:"password"`
	PrivateKey          string `mapstructure:"private_key"`
	PrivateKeyProtected string `mapstructure:"private_key_protected"`
}

type ProxyConfig struct {
	CaddyEndpoint string `mapstructure:"caddy_endpoint"` // Caddy admin API (tunneled via SSH). Defaults to http://localhost:2019
}

type CORSConfig struct {
	AllowedOrigin string `mapstructure:"allowed_origin" validate:"required"`
}

type AppConfig struct {
	Environment  string `mapstructure:"environment"`
	Version      string `mapstructure:"version"`
	LogsPath     string `mapstructure:"logs_path"`
	DeployDomain string `mapstructure:"deploy_domain"` // Base domain for generated app URLs (e.g. nixopus.com)
}

type GitHubConfig struct {
	AppID         string `mapstructure:"app_id"`
	Slug          string `mapstructure:"slug"`
	Pem           string `mapstructure:"pem"`
	ClientID      string `mapstructure:"client_id"`
	ClientSecret  string `mapstructure:"client_secret"`
	WebhookSecret string `mapstructure:"webhook_secret"`
}

type StripeConfig struct {
	SecretKey            string `mapstructure:"secret_key"`
	WebhookSecret        string `mapstructure:"webhook_secret"`
	PriceID              string `mapstructure:"price_id"`
	FreeDeploymentsLimit int    `mapstructure:"free_deployments_limit"`
}

// TrailConfig holds configuration for trail provisioning.
type TrailConfig struct {
	MaxConcurrentTrails int      `mapstructure:"max_concurrent_trails"`
	DefaultImage        string   `mapstructure:"default_image"`
	AllowedImages       []string `mapstructure:"allowed_images"`
	TrailDomain         string   `mapstructure:"trail_domain"`
}

type ClientContext string
type contextKey string

const UserContextKey contextKey = "user"
const AuthTokenKey ClientContext = "auth_token"
const DBContextKey ClientContext = "db"
const OrganizationIDKey contextKey = "organization_id"

type AvailableActions string

const (
	PING                   AvailableActions = "ping"
	SUBSCRIBE              AvailableActions = "subscribe"
	UNSUBSCRIBE            AvailableActions = "unsubscribe"
	AUTHENTICATE           AvailableActions = "authenticate"
	TERMINAL               AvailableActions = "terminal"
	TERMINAL_RESIZE        AvailableActions = "terminal_resize"
	DASHBOARD_MONITOR      AvailableActions = "dashboard_monitor"
	STOP_DASHBOARD_MONITOR AvailableActions = "stop_dashboard_monitor"
	MONITOR_APPLICATION    AvailableActions = "monitor_application"
)

type Payload struct {
	Action AvailableActions `json:"action"`
	Data   interface{}      `json:"data"`
	Topic  string           `json:"topic"`
}

var JWTSecretKey []byte

const defaultJWTSecret = "a3f1b9c7e2d4a6f8b0c1d3e5f7a9b2c4d6e8f0a1b3c5d7e9f0a2b4c6d8e0f1"

func InitJWTSecret() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" || len(secret) < 32 {
		log.Println("WARNING: JWT_SECRET is not set or too short, falling back to default secret.")
		secret = defaultJWTSecret
	}
	JWTSecretKey = []byte(secret)
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
