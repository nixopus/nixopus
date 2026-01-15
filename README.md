<div align="center">

<a href="https://nixopus.com"><img width="1800" height="520" alt="Nixopus" src="https://github.com/user-attachments/assets/e103a9df-7abf-4f78-b75a-221331231247" /></a>

<h5 align="center">
  Open Source alternative to Vercel, Heroku, Netlify with Terminal integration, and Self Hosting capabilities
</h5>

<p align="center">
  <a href="https://nixopus.com"><b>Website</b></a> •
  <a href="https://docs.nixopus.com"><b>Documentation</b></a> •
  <a href="https://nixopus.com/blog"><b>Blog</b></a> •
  <a href="https://discord.gg/skdcq39Wpv"><b>Join Discord</b></a> •
  <a href="https://github.com/raghavyuva/nixopus/discussions/262"><b>Roadmap</b></a>
</p>

<p align="center">
  <a href="https://github.com/raghavyuva/nixopus/stargazers"><img src="https://img.shields.io/github/stars/raghavyuva/nixopus?style=social" alt="GitHub stars" /></a>
  <a href="https://github.com/raghavyuva/nixopus/issues"><img src="https://img.shields.io/github/issues/raghavyuva/nixopus" alt="GitHub issues" /></a>
  <a href="https://github.com/raghavyuva/nixopus/blob/master/LICENSE.md"><img src="https://img.shields.io/badge/license-FSL--1.1--ALv2-blue" alt="License" /></a>
  <br>
  <a href="https://trendshift.io/repositories/15336" target="_blank"><img src="https://trendshift.io/api/badge/repositories/15336" alt="Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/></a>
</p>

</div>

---

<details>
<summary><h2>Table of Contents</h2></summary>

- [Getting Started](#getting-started)
- [Features](#features)
  - [Core Capabilities](#core-capabilities)
  - [Additional Features](#additional-features)
- [Demo](#demo)
  - [Dashboard Overview](#dashboard-overview)
  - [Application Management](#application-management)
  - [Terminal Integration](#terminal-integration)
  - [File Manager](#file-manager)
- [Installation](#installation)
  - [Quick Install](#quick-install)
  - [Custom Installation](#custom-installation)
  - [Installation Options](#installation-options)
- [Links](#links)
- [Acknowledgments](#acknowledgments)
- [About the Name](#about-the-name)

</details>

---

## Getting Started

> **Important Note**: Nixopus is currently in **alpha/pre-release** stage and is not yet ready for production use. While you're welcome to try it out, we recommend waiting for the beta or stable release before using it in production environments.

Nixopus transforms your VPS into a complete application hosting environment. Deploy applications directly from GitHub, manage server files through a browser-based interface, and execute commands via an integrated terminal—all without leaving the dashboard.

---

## Features

### Core Capabilities

- **One-Click Deployments** - Deploy applications from GitHub repositories with automatic builds and zero configuration files
- **Built-in Terminal** - Execute server commands through a secure, web-based terminal with SSH integration
- **File Manager** - Browse, edit, upload, and organize files using drag-and-drop operations
- **Real-time Monitoring** - Track CPU, RAM, and disk usage with live system statistics
- **Automatic SSL** - Generate and manage HTTPS certificates for your domains automatically
- **CI/CD Integration** - Trigger deployments automatically when you push code to GitHub
- **Reverse Proxy** - Route traffic to your applications using the built-in Caddy reverse proxy
- **Multi-channel Notifications** - Receive deployment alerts via Slack, Discord, or email

### Additional Features

- **Docker Support** - Deploy Docker Compose applications, Dockerfiles, or static sites
- **Authentication** - Built-in user management with SuperTokens integration
- **Health Checks** - Monitor application health with customizable health check endpoints
- **Environment Variables** - Manage secrets and configuration variables per deployment
- **Domain Management** - Configure custom domains with automatic SSL certificate generation
- **Extensions** - Automate server tasks through a library of pre-built configurations

---

## Demo

<div align="center">

### Dashboard Overview

![Dashboard](assets/nixopus_dashboard.jpeg)

*Nixopus Dashboard - Manage all your deployments from one place*

</div>

<div align="center">

### Application Management

![Dashboard](assets/nixopus_dashboard.jpeg)

*Deploy and manage applications with ease*

</div>

<div align="center">

### Terminal Integration

![Dashboard](assets/nixopus_dashboard.jpeg)

*Access your server terminal directly from the browser*

</div>

<div align="center">

### File Manager

![Dashboard](assets/nixopus_dashboard.jpeg)

*Browse and edit files with a visual file manager*

</div>

---

## Installation

### Quick Install

Install Nixopus on your VPS with a single command:

```bash
curl -sSL https://install.nixopus.com | bash
```

### Custom Installation

**For custom IP setups:**

```bash
curl -sSL https://install.nixopus.com | bash -s -- --host-ip 10.0.0.154
```

**To install only the CLI tool:**

```bash
curl -sSL https://install.nixopus.com | bash -s -- --skip-nixopus-install
```

### Installation Options

You can customize your installation with the following optional parameters:

| Parameter | Short | Description | Example |
|-----------|-------|-------------|---------|
| `--api-domain` | `-ad` | Domain for Nixopus API | `nixopusapi.example.tld` |
| `--view-domain` | `-vd` | Domain for Nixopus dashboard | `nixopus.example.tld` |
| `--host-ip` | `-ip` | IP address of the server | `10.0.0.154` |
| `--verbose` | `-v` | Show detailed installation logs | - |
| `--timeout` | `-t` | Timeout for each step (default: 300s) | `600` |
| `--force` | `-f` | Replace existing files | - |
| `--dry-run` | `-d` | Preview changes without applying | - |
| `--config-file` | `-c` | Path to custom config file | `/path/to/config.yaml` |

**Example with custom domains:**

```bash
sudo nixopus install \
  --api-domain nixopusapi.example.tld \
  --view-domain nixopus.example.tld \
  --verbose \
  --timeout 600
```

> [!NOTE]
> Running `nixopus install` requires root privileges (sudo) to install system dependencies like Docker. If you encounter permission errors, make sure to run the command with `sudo`.

For more detailed installation instructions, visit our [Installation Guide](https://docs.nixopus.com/install/).

---

## Links

- **Website**: [https://nixopus.com](https://nixopus.com)
- **Documentation**: [https://docs.nixopus.com](https://docs.nixopus.com)
- **Discord Community**: [https://discord.gg/skdcq39Wpv](https://discord.gg/skdcq39Wpv)
- **Blog**: [https://nixopus.com/blog](https://nixopus.com/blog)
- **Roadmap**: [GitHub Discussions](https://github.com/raghavyuva/nixopus/discussions/262)
- **Report Issues**: [GitHub Issues](https://github.com/raghavyuva/nixopus/issues)
- **Feature Requests**: [GitHub Discussions](https://github.com/raghavyuva/nixopus/discussions)

---

## Acknowledgments

Thank you to all the contributors who help make Nixopus better!

<a href="https://github.com/raghavyuva/nixopus/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=raghavyuva/nixopus" alt="Nixopus project contributors" />
</a>

---

## About the Name

Nixopus is derived from the combination of "octopus" (representing flexibility and multi-tasking) and the Linux penguin mascot (Tux). While the name might suggest a connection to [NixOS](https://nixos.org/), Nixopus is an independent project with no direct relation to NixOS or its ecosystem.

---

<div align="center">

**Made with love by the Nixopus community**

[Back to Top](#table-of-contents)

</div>
