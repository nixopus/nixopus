# Installation Guide

Welcome to the Nixopus installation guide. This section will help you set up Nixopus on your VPS quickly.

## Prerequisites

Before you begin, ensure you have:

::: info Prerequisites Checklist
- **VPS with sudo access** - Required for system-level installation
- **Internet connection** - For downloading dependencies and updates
:::

## System Requirements

Make sure your system meets these requirements:

| Requirement | Minimum | Recommended (Production) |
|-------------|---------|--------------------------|
| **Operating System** | Linux (Ubuntu 20.04+, Debian 11+), macOS (CLI only) | Same |
| **CPU** | 2 cores | 4+ cores |
| **RAM** | 2GB | 4GB+ |
| **Storage** | 5GB free | 10GB+ free |
| **Network** | Internet connection | Stable connection |

## Quick Installation

The easiest way to install Nixopus is using the one-liner installation script, which automatically installs the CLI and sets up Nixopus on your VPS:

::: code-group

```bash [Standard Install]
curl -sSL https://install.nixopus.com | bash
```

```bash [With Sudo]
curl -sSL https://install.nixopus.com | sudo bash
```

:::

::: info What the Script Does
This single command will:
1. **Detect** your system architecture and operating system
2. **Download and install** the appropriate Nixopus CLI package
3. **Automatically run** `nixopus install` to set up Nixopus on your server
:::

## Two-Step Installation

If you prefer to install the CLI separately first, you can use the two-step process:

### Step 1: Install the Nixopus CLI Only

```bash
curl -sSL https://install.nixopus.com | bash -s -- --skip-nixopus-install
```

### Step 2: Install Nixopus on your VPS

Once the CLI is installed, you can install Nixopus on your VPS:

```bash
sudo nixopus install
```

::: warning Sudo Required
Running `nixopus install` requires root privileges to install system dependencies (like Docker). Always use `sudo` when running the install command. If you encounter "exit status 100" or permission errors, ensure you're using sudo.
:::

::: info CLI Verification
Before proceeding, verify the CLI is working:
```bash
nixopus --version
nixopus --help
```
:::

## Installation Options

You can customize your installation by providing optional parameters. Options are organized by category:

### Domain Configuration

Configure how Nixopus will be accessed:

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| `--api-domain` | `-ad` | Domain for the Nixopus API | `api.nixopus.com` |
| `--view-domain` | `-vd` | Domain for the Nixopus app | `nixopus.com` |
| `--host-ip` | `-ip` | Server IP when no domains provided | `10.0.0.154` |

::: tip Domain vs IP
- **Domains**: Recommended for production. Enables HTTPS and proper SSL certificates
- **IP Address**: Useful for testing or internal deployments. Falls back to IP if domains not provided
:::

### Port Configuration

Customize service ports if defaults conflict with existing services:

| Service | Option | Default (Production) | Default (Development) |
|---------|--------|---------------------|----------------------|
| API | `--api-port PORT` | 8443 | 8080 |
| Frontend | `--view-port PORT` | 7443 | 3000 |
| PostgreSQL | `--db-port PORT` | 5432 | 5432 |
| Redis | `--redis-port PORT` | 6379 | 6379 |
| Caddy Admin | `--caddy-admin-port PORT` | 2019 | 2019 |
| Caddy HTTP | `--caddy-http-port PORT` | 80 | 80 |
| Caddy HTTPS | `--caddy-https-port PORT` | 443 | 443 |
| SuperTokens | `--supertokens-port PORT` | 3567 | 3567 |

::: warning Port Conflicts
Ensure all specified ports are available. The installer will check, but conflicts can cause installation failures.
:::

### General Options

Common installation options:

| Option | Short | Description |
|-------|-------|-------------|
| `--verbose` | `-v` | Show detailed installation logs |
| `--timeout` | `-t` | Set timeout per step (default: 300s) |
| `--force` | `-f` | Replace existing files |
| `--dry-run` | `-d` | Preview changes without installing |
| `--config-file` | `-c` | Path to custom config file |

::: tip Dry Run
Always test with `--dry-run` first to see what changes will be made without actually installing:
```bash
curl -sSL https://install.nixopus.com | bash -s -- --dry-run
```
:::

::: details Custom Config File
The `--config-file` option allows you to use a custom configuration file instead of the default [`config.prod.yaml`](https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/helpers/config.prod.yaml).

This is useful for:
- Custom deployment configurations
- Environment-specific settings
- Testing different configurations
:::

### Advanced Options

::: details Advanced Configuration
For advanced users and custom deployments:

| Option | Description | Default |
|--------|-------------|---------|
| `--repo REPOSITORY` | GitHub repository URL | `https://github.com/raghavyuva/nixopus` |
| `--branch BRANCH` | Git branch to use | `master` |

::: warning Custom Repository/Branch
When using a custom repository or branch, the installer will use `docker-compose-staging.yml` instead of `docker-compose.yml`. This is intended for development and testing.
:::


## Installation Examples

Common installation scenarios with ready-to-use commands:

### Basic Installation with Domains

Recommended for production deployments with SSL certificates:

```bash
curl -sSL https://install.nixopus.com | bash -s -- \
  --api-domain api.example.com \
  --view-domain example.com
```

::: tip Domain Setup
Before installation, ensure your domains point to your VPS IP address:
- `api.example.com` → Your VPS IP
- `example.com` → Your VPS IP

The installer will automatically configure SSL certificates via Caddy.
:::

### Installation with IP Address

For testing or internal deployments without domains:

```bash
curl -sSL https://install.nixopus.com | bash -s -- \
  --host-ip 10.0.0.154
```

::: info IP Detection
If `--host-ip` is not provided, the installer will automatically detect your public IP address.
:::

### Installation with Custom Ports

When default ports conflict with existing services:

```bash
curl -sSL https://install.nixopus.com | bash -s -- \
  --api-port 9000 \
  --view-port 9001
```

### Verbose Installation with Custom Timeout

For detailed logs and longer timeout (useful for slow connections):

```bash
curl -sSL https://install.nixopus.com | bash -s -- \
  --verbose \
  --timeout 600
```

::: tip Verbose Mode
Use `--verbose` to see detailed installation progress, which is helpful for troubleshooting.
:::

### Dry Run

Preview what will happen without installing:

```bash
curl -sSL https://install.nixopus.com | bash -s -- --dry-run
```

::: info Dry Run Benefits
The dry run shows:
- What files will be created
- What services will be configured
- What ports will be used
- Any potential conflicts

Perfect for testing before actual installation!
:::

### Using Custom Repository and Branch

For development or testing with custom code:

```bash
curl -sSL https://install.nixopus.com | bash -s -- \
  --repo https://github.com/user/fork \
  --branch develop
```

::: warning Development Use Only
Custom repositories and branches use `docker-compose-staging.yml` and are intended for development/testing, not production.
:::

### Manual Installation After CLI Setup

If you've already installed the CLI separately, you can run `nixopus install` directly:

::: code-group

```bash [With Domains]
sudo nixopus install \
  --api-domain api.example.com \
  --view-domain example.com \
  --verbose
```

```bash [With IP]
sudo nixopus install \
  --host-ip 192.168.1.100 \
  --verbose
```

```bash [Custom Ports]
sudo nixopus install \
  --api-port 9000 \
  --view-port 9001 \
  --timeout 600
```

:::

::: tip Why Sudo?
The `nixopus install` command needs root privileges to:

- Install Docker and other system dependencies
- Configure system-level services
- Set up network configurations

If you're already running as root user, you can omit `sudo`.
:::

## Accessing Nixopus

After successful installation, access your Nixopus instance:

::: info Access URLs
**With Domain Configuration:**
- Frontend: `https://your-view-domain.com` (e.g., `https://nixopus.example.com`)
- API: `https://your-api-domain.com` (e.g., `https://api.example.com`)

**With IP Configuration:**
- Frontend: `http://YOUR_IP:PORT` (e.g., `http://192.168.1.100:80`)
- API: `http://YOUR_IP:API_PORT` (e.g., `http://192.168.1.100:8443`)
:::

::: tip First Login
After installation, you'll need to:
1. Visit the frontend URL
2. Complete the initial setup
3. Create your admin account

Check the installation logs for any setup instructions specific to your deployment.
:::

## Troubleshooting

If you encounter issues during installation:

::: warning Installation Issues
The installation script has not been tested in all distributions and operating systems. If you encounter any issues:

1. **Check the logs**: Use `--verbose` flag to see detailed error messages
2. **Verify prerequisites**: Ensure your system meets all requirements
3. **Check ports**: Make sure required ports are available
4. **Report issues**: Create an issue on our [GitHub repository](https://github.com/raghavyuva/nixopus/issues) with:
   - Your operating system and version
   - Installation command used
   - Full error message/output
   - System requirements (CPU, RAM, storage)
:::
