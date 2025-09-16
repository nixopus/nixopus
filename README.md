<div align="center">
<a href="https://nixopus.com"><img width="1800" height="520" alt="Heading(4)" src="https://github.com/user-attachments/assets/e103a9df-7abf-4f78-b75a-221331231247" /></a>
</div>


<p align="center">
 Open Source Server management platform with Terminal integration, and Self Hosting capabilities.
</p>

<p align="center">
  <a href="https://nixopus.com"><b>Website</b></a> •
  <a href="https://docs.nixopus.com"><b>Documentation</b></a> •
  <a href="https://docs.nixopus.com/blog/"><b>Blog</b></a> •
  <a href="https://discord.gg/skdcq39Wpv"><b>Discord</b></a> •
  <a href="https://github.com/raghavyuva/nixopus/discussions/262"><b>Roadmap</b></a>
</p>

<img width="1210" height="764" alt="image" src="https://github.com/user-attachments/assets/3f1dc1e0-956d-4785-8745-ed59d0390afd" />


> ⚠️ **Important Note**: Nixopus is currently in alpha/pre-release stage and is not yet ready for production use. While you're welcome to try it out, we recommend waiting for the beta or stable release before using it in production environments. The platform is still undergoing testing and development.

# Features

- **Deploy apps with one click.** No config files, no SSH commands.
- **Manage files in your browser.** Drag, drop, edit. Like any file manager.
- **Built-in terminal.** Access your server without leaving the page.
- **Real-time monitoring.** See CPU, RAM, disk usage at a glance.
- **Auto SSL certificates.** Your domains get HTTPS automatically.
- **GitHub integration.** Push code → auto deploy.
- **Proxy management.** Route traffic with Caddy reverse proxy.
- **Smart alerts.** Get notified via Slack, Discord, or email when something's wrong.

## Installation & Quick Start

This section will help you set up Nixopus on your VPS quickly.

### Install Nixopus:

**To get started without domain names, and to try out over ip:port deployment:**
```bash
curl -sSL https://install.nixopus.com | bash
```

**To install only the CLI tool without running `nixopus install`:**

```bash
curl -sSL https://install.nixopus.com | bash -s -- --skip-nixopus-install
```

#### Optional Parameters

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

You can also install the CLI and run `nixopus install` with options in a single command, refer [installation documentation](https://docs.nixopus.com/install/#installation-options) for more details on options

## About the Name

Nixopus is derived from the combination of "octopus" and the Linux penguin (Tux). While the name might suggest a connection to [NixOS](https://nixos.org/), Nixopus is an independent project with no direct relation to NixOS or its ecosystem.

## Contributors

<a href="https://github.com/raghavyuva/nixopus/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=raghavyuva/nixopus" alt="Nixopus project contributors" />
</a>

