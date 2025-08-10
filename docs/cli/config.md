# CLI Configuration

The Nixopus CLI uses a built-in YAML configuration file that defines default values for all commands.

## Configuration File

The CLI reads configuration from:
```
helpers/config.prod.yaml
```

This file is bundled with the CLI and contains production-ready defaults that can be overridden through environment variables.

## Key Configuration Sections

### Service Defaults
```yaml
services:
  api:
    env:
      PORT: ${API_PORT:-8443}
      DB_NAME: ${DB_NAME:-postgres}
      USERNAME: ${USERNAME:-postgres}
      PASSWORD: ${PASSWORD:-changeme}
      # ... other API settings
  
  view:
    env:
      PORT: ${VIEW_PORT:-7443}
      NEXT_PUBLIC_PORT: ${NEXT_PUBLIC_PORT:-7443}
      # ... other view settings
  
  caddy:
    env:
      PROXY_PORT: ${PROXY_PORT:-2019}
      API_DOMAIN: ${API_DOMAIN:-}
      VIEW_DOMAIN: ${VIEW_DOMAIN:-}
      # ... other proxy settings
```

### System Dependencies
```yaml
deps:
  curl:           { package: "curl",           command: "curl" }
  python3:        { package: "python3",        command: "python3" }
  git:            { package: "git",            command: "git" }
  docker.io:      { package: "docker.io",      command: "docker" }
  openssl:        { package: "openssl",        command: "openssl" }
  openssh-client: { package: "openssh-client", command: "ssh" }
```

### Network Ports
```yaml
ports: [2019, 80, 443, 7443, 8443, 6379, 5432]
```

### Repository Settings
```yaml
clone:
  repo: "https://github.com/raghavyuva/nixopus"
  branch: "master"
  source-path: source
```

### SSH Configuration
```yaml
ssh_key_size: 4096
ssh_key_type: rsa
ssh_file_path: ssh/id_rsa
```

### File Paths
```yaml
nixopus-config-dir: /etc/nixopus
compose-file-path: source/docker-compose.yml
```

## Environment Variable Overrides

All configuration values use environment variable expansion:
```yaml
PORT: ${API_PORT:-8443}  # Uses API_PORT if set, otherwise 8443
```

**Common overrides:**
```bash
# Override API domain
export API_DOMAIN=api.example.com

# Override database credentials  
export USERNAME=myuser
export PASSWORD=mypassword

# Override ports
export API_PORT=9443
export VIEW_PORT=8443
```

## Command Usage

Commands read specific configuration sections:

| Command | Configuration Used |
|---------|-------------------|
| **preflight** | `ports`, `deps` |
| **install** | Service defaults, paths, SSH settings |
| **service** | Service environment variables |
| **conf** | Service environment configurations |
| **proxy** | `services.caddy.env` settings |
| **clone** | `clone` repository settings |

## Configuration Access

Commands access configuration through the CLI's config system - users don't need to manage the configuration file directly. Use command-line options and environment variables to customize behavior.