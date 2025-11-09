# Installation Guide

Welcome to the Nixopus installation guide. This section will help you set up Nixopus on your VPS quickly.

## Prerequisites

- **VPS with sudo access**
- **Nixopus CLI installed** (see [CLI Installation Guide](../cli/installation.md))

## System Requirements

- **Operating System**: Linux (Ubuntu 20.04+, Debian 11+, or similar distributions)
- **CPU**: Minimum 2 cores (4+ cores recommended for production)
- **RAM**: Minimum 2GB (4GB+ recommended for production)
- **Storage**: Minimum 5GB free disk space 
- **Network**: Internet connection for downloading dependencies and updates

## Quick Installation

### Step 1: Install the Nixopus CLI

First, install the Nixopus CLI tool:

```bash
curl -sSL https://install.nixopus.com | bash -s -- --skip-nixopus-install
```

### Step 2: Install Nixopus on your VPS

Once the CLI is installed, you can install Nixopus on your VPS:

```bash
nixopus install
```

## Installation Options

You can customize your installation by providing the following optional parameters:

- `--api-domain` or `-ad`: Specify the domain where the Nixopus API will be accessible (e.g., `nixopusapi.example.tld`)
- `--view-domain` or `-vd`: Specify the domain where the Nixopus app will be accessible (e.g., `nixopus.example.tld`)
- `--verbose` or `-v`: Show more details while installing
- `--timeout` or `-t`: Set timeout for each step (default: 300 seconds)
- `--force` or `-f`: Replace files if they already exist
- `--dry-run` or `-d`: See what would happen without making changes
- `--config-file` or `-c`: Path to custom config file (defaults to built-in [`config.prod.yaml`](https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/helpers/config.prod.yaml))

Example with optional parameters:

```bash
nixopus install \
  --api-domain nixopusapi.example.tld \
  --view-domain nixopus.example.tld \
  --verbose \
  --timeout 600
```

## Accessing Nixopus

After successful installation, you can access the Nixopus dashboard by visiting the URL you specified in the `--app-domain` parameter (e.g., `https://nixopus.example.tld`). Use the email and password you provided during installation to log in.

> **Note**: The installation script has not been tested in all distributions and different operating systems. If you encounter any issues during installation, please create an issue on our [GitHub repository](https://github.com/raghavyuva/nixopus/issues) with details about your environment and the error message you received.
