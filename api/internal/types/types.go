package types

type Config struct {
	Port        string
	HostName    string
	Password    string
	DBName      string
	Username    string
	SSLMode     string
	MaxOpenConn int
	Debug       bool
	MaxIdleConn int
	DB_PORT     string
	RedisURL    string
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
