package caddy

import (
	"github.com/raghavyuva/caddygo"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
)

// Client is the Caddy client type.
// This is a re-export of the caddygo.Client type.
type Client = caddygo.Client

// GetDefaultClient returns the default Caddy client.
func GetDefaultClient() *Client {
	return tasks.GetCaddyClient()
}

// AddDomainWithTLS adds a domain to the Caddy proxy with automatic TLS.
// This wraps the lower-level Caddy API for convenience.
func AddDomainWithTLS(domain string, upstreamHost string, port int, options caddygo.DomainOptions) error {
	client := tasks.GetCaddyClient()
	return client.AddDomainWithAutoTLS(domain, upstreamHost, port, options)
}
