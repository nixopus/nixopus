package proxy

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

type Caddy struct {
	Logger         *logger.Logger
	Endpoint       string
	RootDir        string
	Domain         string
	Port           string
	client         *http.Client
	FileServerType FileServerType
}

type FileServerType string

const (
	FileServer   FileServerType = "file_server"
	ReverseProxy FileServerType = "reverse_proxy"
)

type CaddyConfig struct {
	Apps struct {
		HTTP struct {
			Servers map[string]Server `json:"servers,omitempty"`
		} `json:"http,omitempty"`
	} `json:"apps,omitempty"`
}

type Server struct {
	Listen         []string       `json:"listen,omitempty"`
	Routes         []Route        `json:"routes,omitempty"`
	AutomaticHTTPS AutomaticHTTPS `json:"automatic_https,omitempty"`
}

type AutomaticHTTPS struct {
	Disable bool     `json:"disable,omitempty"`
	Skip    []string `json:"skip,omitempty"`
}

type Route struct {
	Match  []Match       `json:"match,omitempty"`
	Handle []interface{} `json:"handle,omitempty"`
}

type Match struct {
	Host []string `json:"host,omitempty"`
}

type SubrouteHandle struct {
	Handler string  `json:"handler,omitempty"`
	Routes  []Route `json:"routes,omitempty"`
}

type FileServerHandle struct {
	Handler string   `json:"handler,omitempty"`
	Root    string   `json:"root,omitempty"`
	Browse  struct{} `json:"browse,omitempty"`
}

type ReverseProxyHandle struct {
	Handler   string     `json:"handler,omitempty"`
	Upstreams []Upstream `json:"upstreams,omitempty"`
}

type Upstream struct {
	Dial string `json:"dial,omitempty"`
}
