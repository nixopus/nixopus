package engine

import (
	"fmt"
	"strconv"

	"github.com/raghavyuva/caddygo"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type proxyModule struct{}

func (proxyModule) Type() string { return "proxy" }

func AddDomainToProxy(domain string, port string) error {
	client := tasks.GetCaddyClient()
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}
	upstreamHost := config.AppConfig.SSH.Host
	if err := client.AddDomainWithAutoTLS(domain, upstreamHost, p, caddygo.DomainOptions{}); err != nil {
		return err
	}
	client.Reload()
	return nil
}

func UpdateDomainInProxy(domain string, port string) error {
	return AddDomainToProxy(domain, port)
}

func RemoveDomainFromProxy(domain string) error {
	client := tasks.GetCaddyClient()
	if err := client.DeleteDomain(domain); err != nil {
		return err
	}
	client.Reload()
	return nil
}

func (proxyModule) Execute(sshClient *ssh.SSH, step types.SpecStep, vars map[string]interface{}) (string, func(), error) {
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
		if err := AddDomainToProxy(domain, port); err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("proxy added for %s -> %s:%s", domain, config.AppConfig.SSH.Host, port), nil, nil
	case "update":
		if domain == "" || port == "" {
			return "proxy step skipped: domain and port are optional", nil, nil
		}
		if err := UpdateDomainInProxy(domain, port); err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("proxy updated for %s -> %s:%s", domain, config.AppConfig.SSH.Host, port), nil, nil
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
