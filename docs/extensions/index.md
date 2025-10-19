# Nixopus Extensions Specification

Nixopus Extensions provide a powerful way to automate server operations through YAML based configuration files. Extensions support various step types for file operations, service management, Docker containers, user management, and more.

## Extension Structure

Extensions are defined in YAML files with three main sections:

```yaml
metadata:
  # Extension metadata and identification
variables:
  # Input variables for the extension
execution:
  # Steps to execute
```

## Metadata

The `metadata` section defines the extension's identity and properties:

### Required Fields

- **id**: Extension identifier (3-50 characters, lowercase letters, numbers, hyphens only)
- **name**: Human-readable name
- **description**: Brief description of the extension's purpose
- **author**: Extension author
- **icon**: Icon identifier or URL
- **category**: One of: `Security`, `Containers`, `Database`, `Web Server`, `Maintenance`, `Monitoring`, `Storage`, `Network`, `Development`, `Other`
- **type**: Either `install` or `run`

### Optional Fields

- **version**: Semantic version (format: `x.y.z` where x, y, z are numbers)
- **isVerified**: Boolean indicating if the extension is verified

### Example

```yaml
metadata:
  id: "nginx-setup"
  name: "Nginx Web Server Setup"
  description: "Installs and configures Nginx web server"
  author: "Nixopus Team"
  icon: "nginx"
  category: "Web Server"
  type: "install"
  version: "1.0.0"
  isVerified: true
```

## Variables

Variables allow users to customize extension behavior. They are referenced in step properties using `{{ variable_name }}` syntax.

### Variable Properties

- **type**: Variable type (`string`, `integer`, `boolean`, `array`)
- **description**: Human-readable description
- **default**: Default value (optional)
- **is_required**: Whether the variable is mandatory
- **validation_pattern**: Regex pattern for validation (optional)

### Example

```yaml
variables:
  domain_name:
    type: "string"
    description: "Domain name for the website"
    is_required: true
    validation_pattern: "^[a-zA-Z0-9.-]+$"
  
  port:
    type: "integer"
    description: "Port number for the service"
    default: 80
    is_required: false
  
  enable_ssl:
    type: "boolean"
    description: "Enable SSL/TLS"
    default: true
    is_required: false
```

## Execution Phases

Extensions support two execution phases:

- **run**: Main execution steps
- **validate**: Post-execution validation steps

Both phases support rollback through compensation functions that are automatically executed if a step fails.

## Step Types

### Command

Executes shell commands over SSH.

**Properties:**
- `cmd` (required): Shell command to execute

**Example:**
```yaml
- name: "Update package list"
  type: "command"
  properties:
    cmd: "apt update"
```

### File

Performs file operations via SFTP.

**Properties:**
- `action` (required): Operation type (`move`, `copy`, `upload`, `delete`, `mkdir`)
- `src`: Source path (required for move/copy/upload)
- `dest`: Destination path (required for move/copy/upload; target for delete/mkdir)

**Actions:**
- `move`: SFTP rename operation
- `copy`: Remote `cp -r` command
- `upload`: SFTP upload from local API host to remote destination
- `delete`: SFTP remove operation
- `mkdir`: SFTP mkdir -p operation

**Example:**
```yaml
- name: "Create application directory"
  type: "file"
  properties:
    action: "mkdir"
    dest: "/var/www/myapp"
```

### Service

Manages system services using systemctl (with service fallback).

**Properties:**
- `name` (required): Service name
- `action` (required): Action (`start`, `stop`, `restart`, `enable`, `disable`)

**Example:**
```yaml
- name: "Start Nginx service"
  type: "service"
  properties:
    name: "nginx"
    action: "start"
```

### User

Manages system users and groups.

**Properties:**
- `username` (required): Username
- `action` (required): Action (`ensure`, `modify`, `delete`, `check`, `add_groups`, `remove_groups`)
- `shell`: Shell path (e.g., `/bin/bash`)
- `home`: Home directory path (e.g., `/home/deploy`)
- `groups`: Comma-separated groups (e.g., `sudo,docker`)

**Actions:**
- `ensure`: Creates user if missing, then applies shell/home/groups
- `modify`: Applies provided shell/home/groups to existing user
- `delete`: Removes user and home directory (`userdel -r`)
- `check`: Prints `exists` or `missing` to step output
- `add_groups`: Adds user to each group listed in `groups`
- `remove_groups`: Removes user from each group listed in `groups`

**Example:**
```yaml
- name: "Create deploy user"
  type: "user"
  properties:
    username: "deploy"
    action: "ensure"
    shell: "/bin/bash"
    home: "/home/deploy"
    groups: "sudo,docker"
```

### Docker

Manages Docker containers and images.

**Properties:**
- `action` (required): Action (`pull`, `run`, `stop`, `start`, `rm`)
- `name`: Container name (required for run/stop/start/rm)
- `image`: Image name (required for pull/run)
- `tag`: Image tag (optional)
- `ports`: Port mappings in format "host:container,host:container"
- `restart`: Restart policy (`no`, `on-failure`, `always`, `unless-stopped`)
- `cmd`: Command to run in container
- `env`: Environment variables (object, array, or comma-separated string)
- `volumes`: Volume mounts (array or comma-separated string)
- `networks`: Network names (array or comma-separated string)

**Example:**
```yaml
- name: "Run Redis container"
  type: "docker"
  properties:
    action: "run"
    name: "redis"
    image: "redis"
    tag: "alpine"
    ports: "6379:6379"
    restart: "always"
```

### Docker Compose

Manages Docker Compose stacks.

**Properties:**
- `action` (required): Action (`up`, `down`, `build`)
- `file` (required): Path to docker-compose.yml file

**Example:**
```yaml
- name: "Deploy application stack"
  type: "docker_compose"
  properties:
    action: "up"
    file: "/opt/myapp/docker-compose.yml"
```

### Proxy

Manages reverse proxy configurations using Caddy.

**Properties:**
- `action` (required): Action (`add`, `update`, `remove`)
- `domain` (required): Domain name
- `port` (required for add/update): Backend port

**Example:**
```yaml
- name: "Add domain to proxy"
  type: "proxy"
  properties:
    action: "add"
    domain: "{{ domain_name }}"
    port: "{{ port }}"
```

## Common Step Options

- **ignore_errors**: Boolean (optional) — Continue to next step on failure
- **timeout**: Number (optional, seconds) — Step execution timeout
- **conditions**: Array of strings (optional) — Conditional execution

## Special Variables

- `{{ uploaded_file_path }}`: Path to file uploaded via multipart form to the run endpoint

## Complete Example

```yaml
metadata:
  id: "web-app-deploy"
  name: "Web Application Deployment"
  description: "Deploys a web application with Nginx and Docker"
  author: "Nixopus Team"
  icon: "web"
  category: "Web Server"
  type: "install"
  version: "1.0.0"

variables:
  domain_name:
    type: "string"
    description: "Domain name for the application"
    is_required: true
    validation_pattern: "^[a-zA-Z0-9.-]+$"
  
  app_port:
    type: "integer"
    description: "Application port"
    default: 3000
    is_required: false

execution:
  run:
    - name: "Create application directory"
      type: "file"
      properties:
        action: "mkdir"
        dest: "/var/www/{{ domain_name }}"

    - name: "Deploy application container"
      type: "docker"
      properties:
        action: "run"
        name: "webapp"
        image: "myapp"
        ports: "{{ app_port }}:3000"
        restart: "always"

    - name: "Add domain to proxy"
      type: "proxy"
      properties:
        action: "add"
        domain: "{{ domain_name }}"
        port: "{{ app_port }}"

  validate:
    - name: "Check container is running"
      type: "command"
      properties:
        cmd: "docker ps | grep webapp"
      ignore_errors: false

    - name: "Test application response"
      type: "command"
      properties:
        cmd: "curl -f http://localhost:{{ app_port }}"
      timeout: 30
```
