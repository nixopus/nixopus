# Installation

This guide walks you through installing Nixopus on your VPS. The installation process downloads the Nixopus CLI and uses it to set up all required services including Docker, PostgreSQL, Redis, and Caddy.

## Requirements

Before installing, ensure your server meets these requirements:

| Requirement | Minimum | Recommended |
|-------------|---------|-------------|
| **Operating System** | Ubuntu 20.04+, Debian 11+ | Ubuntu 22.04+ |
| **CPU** | 2 cores | 4+ cores |
| **RAM** | 2 GB | 4+ GB |
| **Storage** | 5 GB free | 10+ GB free |
| **Access** | Root or sudo privileges | |
| **Ports** | 80, 443, 8443, 7443 available | |
| **Network** | Internet connection | Stable connection |

## Quick Installation

Run this single command to install Nixopus:

```bash
curl -sSL https://install.nixopus.com | bash
```

This command performs the following steps:

1. Detects your system architecture (amd64, arm64) and operating system
2. Downloads the appropriate Nixopus CLI binary
3. Installs the CLI to `/usr/local/bin/nixopus`
4. Runs `nixopus install` to set up all services

## Generate Installation Command

Use the interactive generator below to customize your installation with domains, IP addresses, and other options.

<InstallGenerator />

## Installation Examples

::: code-group

```bash [Default (IP based)]
curl -sSL https://install.nixopus.com | bash
```

```bash [Custom IP]
curl -sSL https://install.nixopus.com | bash -s -- --host-ip 10.0.0.154
```

```bash [With Domains]
sudo nixopus install \
  --api-domain api.example.com \
  --view-domain app.example.com \
  --verbose
```

```bash [CLI Only]
curl -sSL https://install.nixopus.com | bash -s -- --skip-nixopus-install
```

:::

::: tip Domain Installation
When using domains, ensure your DNS records point to your server before running the install command. Caddy will automatically obtain SSL certificates.
:::

## Accessing Nixopus

After installation completes, open your browser and navigate to:

::: code-group

```txt [IP Configuration]
Dashboard: http://YOUR_IP:80
API:       http://YOUR_IP:8443
```

```txt [Domain Configuration]
Dashboard: https://your-view-domain.com
API:       https://your-api-domain.com
```

:::

On first access, you will be prompted to create an admin account.

## Troubleshooting

::: details Installation Fails
Run with verbose mode to see detailed errors:

```bash
sudo nixopus install --verbose
```

Check that required ports are not in use:

```bash
sudo lsof -i :80 -i :443 -i :8443 -i :7443
```

Verify Docker is installed and running:

```bash
docker --version
docker ps
```
:::

::: details Permission Errors
Ensure you're using sudo:

```bash
sudo nixopus install
```

Or run as root:

```bash
su -
nixopus install
```
:::

::: details Port Conflicts
If ports are already in use, stop the conflicting services or configure Nixopus to use different ports by editing the configuration file at `/etc/nixopus/source/helpers/config.prod.yaml`.
:::

::: details DNS Issues
For domain based installations, verify your DNS records resolve correctly:

```bash
dig +short api.example.com
dig +short app.example.com
```

Both should return your server's IP address.
:::

::: warning Getting Help
If you continue to experience issues, create an issue on [GitHub](https://github.com/raghavyuva/nixopus/issues) with your operating system version, installation command, and complete error output.
:::

## Next Steps

After installation, you can:

- [Deploy your first application](/apps/)
- [Configure notifications](/notifications/)
- [Learn CLI commands](/cli/cli-reference)
