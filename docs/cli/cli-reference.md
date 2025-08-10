# `nixopus`

Nixopus CLI - A powerful deployment and management tool

**Usage**:

```console
$ nixopus [OPTIONS] COMMAND [ARGS]...
```

**Options**:

* `-v, --version`: Show version information
* `--help`: Show this message and exit.

**Commands**:

* `preflight`: Preflight checks for system compatibility
* `clone`: Clone a repository
* `conf`: Manage configuration
* `service`: Manage Nixopus services
* `proxy`: Manage Nixopus proxy (Caddy) configuration
* `install`: Install Nixopus
* `uninstall`: Uninstall Nixopus
* `version`: Show version information
* `test`: Run tests (only in DEVELOPMENT environment)

## `nixopus preflight`

Preflight checks for system compatibility

**Usage**:

```console
$ nixopus preflight [OPTIONS] COMMAND [ARGS]...
```

**Options**:

* `--help`: Show this message and exit.

**Commands**:

* `check`: Run all preflight checks
* `ports`: Check if list of ports are available on a...
* `deps`: Check if list of dependencies are...

### `nixopus preflight check`

Run all preflight checks

**Usage**:

```console
$ nixopus preflight check [OPTIONS]
```

**Options**:

* `-v, --verbose`: Verbose output
* `-o, --output TEXT`: Output format, text,json  [default: text]
* `-t, --timeout INTEGER`: Timeout in seconds  [default: 10]
* `--help`: Show this message and exit.

### `nixopus preflight ports`

Check if list of ports are available on a host

**Usage**:

```console
$ nixopus preflight ports [OPTIONS] PORTS...
```

**Arguments**:

* `PORTS...`: The list of ports to check  [required]

**Options**:

* `-h, --host TEXT`: The host to check  [default: localhost]
* `-v, --verbose`: Verbose output
* `-o, --output TEXT`: Output format, text, json  [default: text]
* `-t, --timeout INTEGER`: Timeout in seconds  [default: 10]
* `--help`: Show this message and exit.

### `nixopus preflight deps`

Check if list of dependencies are available on the system

**Usage**:

```console
$ nixopus preflight deps [OPTIONS] DEPS...
```

**Arguments**:

* `DEPS...`: The list of dependencies to check  [required]

**Options**:

* `-v, --verbose`: Verbose output
* `-o, --output TEXT`: Output format, text, json  [default: text]
* `-t, --timeout INTEGER`: Timeout in seconds  [default: 10]
* `--help`: Show this message and exit.

## `nixopus clone`

Clone a repository

**Usage**:

```console
$ nixopus clone [OPTIONS] COMMAND [ARGS]...
```

**Options**:

* `-r, --repo TEXT`: The repository to clone  [default: https://github.com/raghavyuva/nixopus]
* `-b, --branch TEXT`: The branch to clone  [default: master]
* `-p, --path TEXT`: The path to clone the repository to  [default: /etc/nixopus/source]
* `-f, --force`: Force the clone
* `-v, --verbose`: Verbose output
* `-o, --output TEXT`: Output format, text, json  [default: text]
* `-d, --dry-run`: Dry run
* `-t, --timeout INTEGER`: Timeout in seconds  [default: 10]
* `--help`: Show this message and exit.

## `nixopus conf`

Manage configuration

**Usage**:

```console
$ nixopus conf [OPTIONS] COMMAND [ARGS]...
```

**Options**:

* `--help`: Show this message and exit.

**Commands**:

* `list`: List all configuration
* `delete`: Delete a configuration
* `set`: Set a configuration

### `nixopus conf list`

List all configuration

**Usage**:

```console
$ nixopus conf list [OPTIONS]
```

**Options**:

* `-s, --service TEXT`: The name of the service to list configuration for, e.g api,view  [default: api]
* `-v, --verbose`: Verbose output
* `-o, --output TEXT`: Output format, text, json  [default: text]
* `-d, --dry-run`: Dry run
* `-e, --env-file TEXT`: Path to the environment file
* `-t, --timeout INTEGER`: Timeout in seconds  [default: 10]
* `--help`: Show this message and exit.

### `nixopus conf delete`

Delete a configuration

**Usage**:

```console
$ nixopus conf delete [OPTIONS] KEY
```

**Arguments**:

* `KEY`: The key of the configuration to delete  [required]

**Options**:

* `-s, --service TEXT`: The name of the service to delete configuration for, e.g api,view  [default: api]
* `-v, --verbose`: Verbose output
* `-o, --output TEXT`: Output format, text, json  [default: text]
* `-d, --dry-run`: Dry run
* `-e, --env-file TEXT`: Path to the environment file
* `-t, --timeout INTEGER`: Timeout in seconds  [default: 10]
* `--help`: Show this message and exit.

### `nixopus conf set`

Set a configuration

**Usage**:

```console
$ nixopus conf set [OPTIONS] KEY_VALUE
```

**Arguments**:

* `KEY_VALUE`: Configuration in the form KEY=VALUE  [required]

**Options**:

* `-s, --service TEXT`: The name of the service to set configuration for, e.g api,view  [default: api]
* `-v, --verbose`: Verbose output
* `-o, --output TEXT`: Output format, text, json  [default: text]
* `-d, --dry-run`: Dry run
* `-e, --env-file TEXT`: Path to the environment file
* `-t, --timeout INTEGER`: Timeout in seconds  [default: 10]
* `--help`: Show this message and exit.

## `nixopus service`

Manage Nixopus services

**Usage**:

```console
$ nixopus service [OPTIONS] COMMAND [ARGS]...
```

**Options**:

* `--help`: Show this message and exit.

**Commands**:

* `up`: Start Nixopus services
* `down`: Stop Nixopus services
* `ps`: Show status of Nixopus services
* `restart`: Restart Nixopus services

### `nixopus service up`

Start Nixopus services

**Usage**:

```console
$ nixopus service up [OPTIONS]
```

**Options**:

* `-n, --name TEXT`: The name of the service to start, defaults to all  [default: all]
* `-v, --verbose`: Verbose output
* `-o, --output TEXT`: Output format, text, json  [default: text]
* `--dry-run`: Dry run
* `-d, --detach`: Detach from the service and run in the background
* `-e, --env-file TEXT`: Path to the environment file
* `-f, --compose-file TEXT`: Path to the compose file  [default: /etc/nixopus/source/docker-compose.yml]
* `-t, --timeout INTEGER`: Timeout in seconds  [default: 10]
* `--help`: Show this message and exit.

### `nixopus service down`

Stop Nixopus services

**Usage**:

```console
$ nixopus service down [OPTIONS]
```

**Options**:

* `-n, --name TEXT`: The name of the service to stop, defaults to all  [default: all]
* `-v, --verbose`: Verbose output
* `-o, --output TEXT`: Output format, text, json  [default: text]
* `--dry-run`: Dry run
* `-e, --env-file TEXT`: Path to the environment file
* `-f, --compose-file TEXT`: Path to the compose file  [default: /etc/nixopus/source/docker-compose.yml]
* `-t, --timeout INTEGER`: Timeout in seconds  [default: 10]
* `--help`: Show this message and exit.

### `nixopus service ps`

Show status of Nixopus services

**Usage**:

```console
$ nixopus service ps [OPTIONS]
```

**Options**:

* `-n, --name TEXT`: The name of the service to show, defaults to all  [default: all]
* `-v, --verbose`: Verbose output
* `-o, --output TEXT`: Output format, text, json  [default: text]
* `-d, --dry-run`: Dry run
* `-e, --env-file TEXT`: Path to the environment file
* `-f, --compose-file TEXT`: Path to the compose file  [default: /etc/nixopus/source/docker-compose.yml]
* `-t, --timeout INTEGER`: Timeout in seconds  [default: 10]
* `--help`: Show this message and exit.

### `nixopus service restart`

Restart Nixopus services

**Usage**:

```console
$ nixopus service restart [OPTIONS]
```

**Options**:

* `-n, --name TEXT`: The name of the service to restart, defaults to all  [default: all]
* `-v, --verbose`: Verbose output
* `-o, --output TEXT`: Output format, text, json  [default: text]
* `-d, --dry-run`: Dry run
* `-e, --env-file TEXT`: Path to the environment file
* `-f, --compose-file TEXT`: Path to the compose file  [default: /etc/nixopus/source/docker-compose.yml]
* `-t, --timeout INTEGER`: Timeout in seconds  [default: 10]
* `--help`: Show this message and exit.

## `nixopus proxy`

Manage Nixopus proxy (Caddy) configuration

**Usage**:

```console
$ nixopus proxy [OPTIONS] COMMAND [ARGS]...
```

**Options**:

* `--help`: Show this message and exit.

**Commands**:

* `load`: Load Caddy proxy configuration
* `status`: Check Caddy proxy status
* `stop`: Stop Caddy proxy

### `nixopus proxy load`

Load Caddy proxy configuration

**Usage**:

```console
$ nixopus proxy load [OPTIONS]
```

**Options**:

* `-p, --proxy-port INTEGER`: Caddy admin port  [default: 2019]
* `-v, --verbose`: Verbose output
* `-o, --output TEXT`: Output format: text, json  [default: text]
* `--dry-run`: Dry run
* `-c, --config-file TEXT`: Path to Caddy config file
* `-t, --timeout INTEGER`: Timeout in seconds  [default: 10]
* `--help`: Show this message and exit.

### `nixopus proxy status`

Check Caddy proxy status

**Usage**:

```console
$ nixopus proxy status [OPTIONS]
```

**Options**:

* `-p, --proxy-port INTEGER`: Caddy admin port  [default: 2019]
* `-v, --verbose`: Verbose output
* `-o, --output TEXT`: Output format: text, json  [default: text]
* `--dry-run`: Dry run
* `-t, --timeout INTEGER`: Timeout in seconds  [default: 10]
* `--help`: Show this message and exit.

### `nixopus proxy stop`

Stop Caddy proxy

**Usage**:

```console
$ nixopus proxy stop [OPTIONS]
```

**Options**:

* `-p, --proxy-port INTEGER`: Caddy admin port  [default: 2019]
* `-v, --verbose`: Verbose output
* `-o, --output TEXT`: Output format: text, json  [default: text]
* `--dry-run`: Dry run
* `-t, --timeout INTEGER`: Timeout in seconds  [default: 10]
* `--help`: Show this message and exit.

## `nixopus install`

Install Nixopus

**Usage**:

```console
$ nixopus install [OPTIONS] COMMAND [ARGS]...
```

**Options**:

* `-v, --verbose`: Show more details while installing
* `-t, --timeout INTEGER`: How long to wait for each step (in seconds)  [default: 300]
* `-f, --force`: Replace files if they already exist
* `-d, --dry-run`: See what would happen, but don&#x27;t make changes
* `-c, --config-file TEXT`: Path to custom config file (defaults to built-in config)
* `-ad, --api-domain TEXT`: The domain where the nixopus api will be accessible (e.g. api.nixopus.com), if not provided you can use the ip address of the server and the port (e.g. 192.168.1.100:8443)
* `-vd, --view-domain TEXT`: The domain where the nixopus view will be accessible (e.g. nixopus.com), if not provided you can use the ip address of the server and the port (e.g. 192.168.1.100:80)
* `--help`: Show this message and exit.

**Commands**:

* `ssh`: Generate an SSH key pair with proper...
* `deps`: Install dependencies

### `nixopus install ssh`

Generate an SSH key pair with proper permissions and optional authorized_keys integration

**Usage**:

```console
$ nixopus install ssh [OPTIONS]
```

**Options**:

* `-p, --path TEXT`: The SSH key path to generate  [default: ~/.ssh/nixopus_rsa]
* `-t, --key-type TEXT`: The SSH key type (rsa, ed25519, ecdsa)  [default: rsa]
* `-s, --key-size INTEGER`: The SSH key size  [default: 4096]
* `-P, --passphrase TEXT`: The passphrase to use for the SSH key
* `-v, --verbose`: Verbose output
* `-o, --output TEXT`: Output format, text, json  [default: text]
* `-d, --dry-run`: Dry run
* `-f, --force`: Force overwrite existing SSH key
* `-S, --set-permissions`: Set proper file permissions  [default: True]
* `-a, --add-to-authorized-keys`: Add public key to authorized_keys
* `-c, --create-ssh-directory`: Create .ssh directory if it doesn&#x27;t exist  [default: True]
* `-T, --timeout INTEGER`: Timeout in seconds  [default: 10]
* `--help`: Show this message and exit.

### `nixopus install deps`

Install dependencies

**Usage**:

```console
$ nixopus install deps [OPTIONS]
```

**Options**:

* `-v, --verbose`: Verbose output
* `-o, --output TEXT`: Output format, text, json  [default: text]
* `-d, --dry-run`: Dry run
* `-t, --timeout INTEGER`: Timeout in seconds  [default: 10]
* `--help`: Show this message and exit.

## `nixopus uninstall`

Uninstall Nixopus

**Usage**:

```console
$ nixopus uninstall [OPTIONS] COMMAND [ARGS]...
```

**Options**:

* `-v, --verbose`: Show more details while uninstalling
* `-t, --timeout INTEGER`: How long to wait for each step (in seconds)  [default: 300]
* `-d, --dry-run`: See what would happen, but don&#x27;t make changes
* `-f, --force`: Remove files without confirmation prompts
* `--help`: Show this message and exit.

## `nixopus version`

Show version information

**Usage**:

```console
$ nixopus version [OPTIONS] COMMAND [ARGS]...
```

**Options**:

* `--help`: Show this message and exit.

## `nixopus test`

Run tests (only in DEVELOPMENT environment)

**Usage**:

```console
$ nixopus test [OPTIONS] [TARGET] COMMAND [ARGS]...
```

**Arguments**:

* `[TARGET]`: Test target (e.g., version)

**Options**:

* `--help`: Show this message and exit.
