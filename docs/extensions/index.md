# Extensions

Extensions automate server tasks through a library of pre-built configurations. Deploy databases, web servers, monitoring tools, and more with a few clicks instead of manual setup.

## Browsing Extensions

Navigate to the **Extensions** page to explore available extensions. Use the search bar to find specific tools or filter by category:

- **Database**: PostgreSQL, MariaDB, Redis, Qdrant
- **Web Server**: Nginx, Caddy configurations
- **Monitoring**: Netdata, Uptime Kuma, Dashdot
- **Development**: Code Server, n8n, Gitea
- **Media**: Jellyfin, Plex, Calibre
- **And more**: 130+ extensions across all categories

## Running an Extension

1. Click on any extension card to view its details
2. Review the description and required variables
3. Fill in the configuration form (domain, ports, passwords, etc.)
4. Click **Run** to execute

The extension runs each step in sequence and displays live logs. If any step fails, the extension attempts to roll back changes automatically.

## Extension Types

Extensions come in two types:

- **Install**: Sets up a service for the first time (creates directories, pulls images, configures proxies)
- **Run**: Performs an action on an existing setup (backups, updates, maintenance tasks)

## Creating Extensions

Extensions are YAML files with three sections: metadata, variables, and execution steps.

### Basic Structure

```yaml
metadata:
  id: "my-extension"
  name: "My Extension"
  description: "What this extension does"
  author: "Your Name"
  icon: "server"
  category: "Development"
  type: "install"
  version: "1.0.0"

variables:
  domain:
    type: "string"
    description: "Domain for the service"
    is_required: true

execution:
  run:
    - name: "Step name"
      type: "command"
      properties:
        cmd: "echo Hello"
```

### Variable Types

Variables let users customize the extension. Reference them with `{{ variable_name }}` syntax.

- **string**: Text input
- **integer**: Numeric input
- **boolean**: True/false toggle
- **array**: List of values

```yaml
variables:
  domain:
    type: "string"
    description: "Domain name"
    is_required: true
    validation_pattern: "^[a-zA-Z0-9.-]+$"
  
  port:
    type: "integer"
    description: "Port number"
    default: 8080
    is_required: false
```

### Step Types

#### Command

Executes shell commands via SSH.

```yaml
- name: "Update packages"
  type: "command"
  properties:
    cmd: "apt update && apt upgrade -y"
```

#### File

Performs file operations via SFTP.

```yaml
- name: "Create directory"
  type: "file"
  properties:
    action: "mkdir"  # mkdir, copy, move, delete, upload
    dest: "/var/www/myapp"
```

#### Service

Manages systemd services.

```yaml
- name: "Start nginx"
  type: "service"
  properties:
    name: "nginx"
    action: "start"  # start, stop, restart, enable, disable
```

#### Docker

Manages containers and images.

```yaml
- name: "Run Redis"
  type: "docker"
  properties:
    action: "run"  # pull, run, stop, start, rm
    name: "redis"
    image: "redis"
    tag: "alpine"
    ports: "6379:6379"
    restart: "always"
```

#### Docker Compose

Manages compose stacks.

```yaml
- name: "Deploy stack"
  type: "docker_compose"
  properties:
    action: "up"  # up, down, build
    file: "/opt/app/docker-compose.yml"
```

#### Proxy

Configures Caddy reverse proxy.

```yaml
- name: "Add domain"
  type: "proxy"
  properties:
    action: "add"  # add, update, remove
    domain: "{{ domain }}"
    port: "{{ port }}"
```

#### User

Manages system users.

```yaml
- name: "Create deploy user"
  type: "user"
  properties:
    action: "ensure"  # ensure, modify, delete, check
    username: "deploy"
    shell: "/bin/bash"
    groups: "sudo,docker"
```

### Step Options

All steps support these optional properties:

- **ignore_errors**: Continue if step fails
- **timeout**: Maximum seconds to wait
- **conditions**: Array of conditions for execution

### Validation Phase

Add a `validate` section to verify the deployment succeeded:

```yaml
execution:
  run:
    - name: "Deploy container"
      type: "docker"
      properties:
        action: "run"
        name: "myapp"
        image: "myapp:latest"

  validate:
    - name: "Check container running"
      type: "command"
      properties:
        cmd: "docker ps | grep myapp"
```

## Contributing Extensions

To add an extension to the Nixopus library, create a YAML file following the specification above and submit a pull request to the [nixopus repository](https://github.com/raghavyuva/nixopus) in the `api/templates/` directory.
