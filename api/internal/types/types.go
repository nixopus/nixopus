package types

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Deployment DeploymentConfig `mapstructure:"deployment"`
	Docker     DockerConfig     `mapstructure:"docker"`
	Proxy      ProxyConfig      `mapstructure:"proxy"`
	CORS       CORSConfig       `mapstructure:"cors"`
	App        AppConfig        `mapstructure:"app"`
	GitHub     GitHubConfig     `mapstructure:"github"`
	BetterAuth BetterAuthConfig `mapstructure:"betterauth"`
	Stripe     StripeConfig     `mapstructure:"stripe"`
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

type DeploymentConfig struct {
	MountPath string `mapstructure:"mount_path" validate:"required"`
}

type DockerConfig struct {
	Host    string `mapstructure:"host"`
	Port    string `mapstructure:"port"`
	Context string `mapstructure:"context"`
}

type ProxyConfig struct {
	CaddyEndpoint string `mapstructure:"caddy_endpoint" validate:"required"`
}

type CORSConfig struct {
	AllowedOrigin string `mapstructure:"allowed_origin" validate:"required"`
}

type AppConfig struct {
	Environment string `mapstructure:"environment"`
	Version     string `mapstructure:"version"`
	LogsPath    string `mapstructure:"logs_path"`
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

var JWTSecretKey = []byte("secret-key")

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
