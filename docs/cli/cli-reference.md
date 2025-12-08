# Nixopus CLI Reference

Nixopus CLI - A powerful deployment and management tool

**Usage**:

```console
$ nixopus [OPTIONS] COMMAND [ARGS]...
```

**Options**:

* `-v, --version`: Show version information
* `--help`: Show this message and exit.

**Commands**:

* `install`: Install Nixopus
* `uninstall`: Uninstall Nixopus
* `update`: Update Nixopus
* `version`: Show version information

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
* `-D, --development`: Use development workflow (local setup, dev compose, dev env)
* `--dev-path TEXT`: Installation directory for development workflow (defaults to current directory)
* `-ad, --api-domain TEXT`: The domain where the nixopus api will be accessible (e.g. api.nixopus.com), if not provided you can use the ip address of the server and the port (e.g. 192.168.1.100:8443)
* `-vd, --view-domain TEXT`: The domain where the nixopus view will be accessible (e.g. nixopus.com), if not provided you can use the ip address of the server and the port (e.g. 192.168.1.100:80)
* `-ip, --host-ip TEXT`: The IP address of the server to use when no domains are provided (e.g. 10.0.0.154 or 192.168.1.100). If not provided, the public IP will be automatically detected.
* `--api-port INTEGER`: Port for the API service (default: 8443 for production, 8080 for development)
* `--view-port INTEGER`: Port for the View/Frontend service (default: 7443 for production, 3000 for development)
* `--db-port INTEGER`: Port for the PostgreSQL database (default: 5432)
* `--redis-port INTEGER`: Port for the Redis service (default: 6379)
* `--caddy-admin-port INTEGER`: Port for Caddy admin API (default: 2019)
* `--caddy-http-port INTEGER`: Port for Caddy HTTP traffic (default: 80)
* `--caddy-https-port INTEGER`: Port for Caddy HTTPS traffic (default: 443)
* `--supertokens-port INTEGER`: Port for SuperTokens service (default: 3567)
* `-r, --repo TEXT`: GitHub repository URL to clone (defaults to config value)
* `-b, --branch TEXT`: Git branch to clone (defaults to config value)
* `--help`: Show this message and exit.

**Commands**:

* `development`: Install Nixopus for local development in...
* `ssh`: Generate an SSH key pair with proper...
* `deps`: Install dependencies

### `nixopus install development`

Install Nixopus for local development in specified or current directory

**Usage**:

```console
$ nixopus install development [OPTIONS]
```

**Options**:

* `-p, --path TEXT`: Installation directory (defaults to current directory)
* `-v, --verbose`: Show more details while installing
* `-t, --timeout INTEGER`: How long to wait for each step (in seconds)  [default: 1800]
* `-f, --force`: Replace files if they already exist
* `-d, --dry-run`: See what would happen, but don&#x27;t make changes
* `-c, --config-file TEXT`: Path to custom config file (defaults to config.dev.yaml)
* `-r, --repo TEXT`: GitHub repository URL to clone (defaults to config value)
* `-b, --branch TEXT`: Git branch to clone (defaults to config value)
* `--api-port INTEGER`: Port for the API service (default: 8080)
* `--view-port INTEGER`: Port for the View/Frontend service (default: 3000)
* `--db-port INTEGER`: Port for the PostgreSQL database (default: 5432)
* `--redis-port INTEGER`: Port for the Redis service (default: 6379)
* `--caddy-admin-port INTEGER`: Port for Caddy admin API (default: 2019)
* `--caddy-http-port INTEGER`: Port for Caddy HTTP traffic (default: 80)
* `--caddy-https-port INTEGER`: Port for Caddy HTTPS traffic (default: 443)
* `--supertokens-port INTEGER`: Port for SuperTokens service (default: 3567)
* `--help`: Show this message and exit.

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

## `nixopus update`

Update Nixopus

**Usage**:

```console
$ nixopus update [OPTIONS] COMMAND [ARGS]...
```

**Options**:

* `-v, --verbose`: Show more details while updating
* `--help`: Show this message and exit.

**Commands**:

* `cli`: Update CLI tool

### `nixopus update cli`

Update CLI tool

**Usage**:

```console
$ nixopus update cli [OPTIONS]
```

**Options**:

* `-v, --verbose`: Show more details while updating
* `--help`: Show this message and exit.

## `nixopus version`

Show version information

**Usage**:

```console
$ nixopus version [OPTIONS] COMMAND [ARGS]...
```

**Options**:

* `--help`: Show this message and exit.
