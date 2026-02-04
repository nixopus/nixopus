package engine

import (
	"context"
	"fmt"
	"strconv"

	"github.com/raghavyuva/caddygo"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type proxyModule struct{}

func (proxyModule) Type() string { return "proxy" }

func AddDomainToProxy(domain string, port string, upstreamHost string) error {
	client := tasks.GetCaddyClient()
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}
	if err := client.AddDomainWithAutoTLS(domain, upstreamHost, p, caddygo.DomainOptions{}); err != nil {
		return err
	}
	client.Reload()
	return nil
}

func UpdateDomainInProxy(domain string, port string, upstreamHost string) error {
	return AddDomainToProxy(domain, port, upstreamHost)
}

func RemoveDomainFromProxy(domain string) error {
	client := tasks.GetCaddyClient()
	if err := client.DeleteDomain(domain); err != nil {
		return err
	}
	client.Reload()
	return nil
}

func (proxyModule) Execute(ctx context.Context, sshClient *ssh.SSH, step types.SpecStep, vars map[string]interface{}) (string, func(), error) {
	action, _ := step.Properties["action"].(string)
	domain, _ := step.Properties["domain"].(string)
	port, _ := step.Properties["port"].(string)

	if domain != "" {
		domain = replaceVars(domain, vars)
	}
	if port != "" {
		port = replaceVars(port, vars)
	}

	switch action {
	case "add":
		if domain == "" || port == "" {
			return "proxy step skipped: domain and port are optional", nil, nil
		}
		if sshClient.Host == "" {
			return "", nil, fmt.Errorf("SSH host not configured")
		}
		if err := AddDomainToProxy(domain, port, sshClient.Host); err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("proxy added for %s -> %s:%s", domain, sshClient.Host, port), nil, nil
	case "update":
		if domain == "" || port == "" {
			return "proxy step skipped: domain and port are optional", nil, nil
		}
		if sshClient.Host == "" {
			return "", nil, fmt.Errorf("SSH host not configured")
		}
		if err := UpdateDomainInProxy(domain, port, sshClient.Host); err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("proxy updated for %s -> %s:%s", domain, sshClient.Host, port), nil, nil
	case "remove":
		if domain == "" {
			return "proxy step skipped: domain is optional", nil, nil
		}
		if err := RemoveDomainFromProxy(domain); err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("proxy removed for %s", domain), nil, nil
	default:
		return "", nil, fmt.Errorf("unsupported proxy action: %s", action)
	}
}

func init() {
	RegisterModule(proxyModule{})
}
