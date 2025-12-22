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

::: info What the Script Does
This single command will:

1. **Detect** your system architecture and operating system
2. **Download and install** the appropriate Nixopus CLI package
3. **Automatically run** `nixopus install` to set up Nixopus on your server
:::

## Generate your installation command

Customize your installation with optional flags and configurations provided below and ensure there are not validation errors previewed before copying.

<InstallGenerator />

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
