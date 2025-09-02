<div id="user-content-toc">
  <ul style="list-style: none;">
    <summary>
      <h1><samp>Nixopus</samp></h1><br>
      <h6>Streamline Your Entire Server Workflow — ServerOps with No Fuss</h6>
      <a href="https://nixopus.com"><img align="right" src="./assets/nixopus_logo_transparent.png" alt="Nixopus Logo" width="250" /></a>
    </summary>
  </ul>
</div>

<samp>
  <table>  
    <tr>
      <td>
        <a href="https://github.com/raghavyuva/nixopus/actions/workflows/security.yml">
          <img src="https://github.com/raghavyuva/nixopus/actions/workflows/security.yml/badge.svg" alt="Security Scan" />
        </a>
        <a href="https://github.com/raghavyuva/nixopus/actions/workflows/build_container.yml">
          <img src="https://github.com/raghavyuva/nixopus/actions/workflows/build_container.yml/badge.svg" alt="Package Manager" />
        </a>
        <a href="https://github.com/raghavyuva/nixopus/actions/workflows/release.yml">
          <img src="https://github.com/raghavyuva/nixopus/actions/workflows/release.yml/badge.svg" alt="Release" />
        </a>
        <br />
        <a href="https://www.youtube.com/watch?v=DrDGWNq4JM4">
          <img src="https://img.shields.io/youtube/views/DrDGWNq4JM4?style=social&label=View%20Demo" alt="YouTube Video Views" />
        </a>
       <img alt="GitHub commit activity" src="https://img.shields.io/github/commit-activity/y/raghavyuva/nixopus">
        <img src="https://madewithlove.now.sh/in?heart=true&colorA=%23ff671f&colorB=%23046a38&text=India" alt="Made with love in India" />
        <br><br>
        <div align="center">
          <strong>
            <a href="https://nixopus.com"> Website</a> |
            <a href="https://docs.nixopus.com"> Documentation</a> |
            <a href="https://docs.nixopus.com/blog/"> Blogs</a>
          </strong>
        </div>
        <br>
        <p align="center">
          <a href="https://discord.gg/skdcq39Wpv" target="_blank">
            <img src="https://user-images.githubusercontent.com/31022056/158916278-4504b838-7ecb-4ab9-a900-7dc002aade78.png" alt="Join our Discord Community" width="200" style="border-radius: 12px; box-shadow: 0px 4px 12px rgba(0,0,0,0.15);" />
          </a>
        </p>
      </td>
    </tr>
  </table>
</samp>

## Project Overview

Nixopus streamlines your workflow with comprehensive tools for deployment, monitoring, and maintenance.

> ⚠️ **Important Note**: Nixopus is currently in alpha/pre-release stage and is not yet ready for production use. While you're welcome to try it out, we recommend waiting for the beta or stable release before using it in production environments. The platform is still undergoing testing and development.

## Demo / Screenshots

| Self Host Stats | Team Display | File Manager |
| :-: | :-: | :-: |
| <a href="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/28nkmy49nm7oi5tq1t8c.webp"><img src="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/28nkmy49nm7oi5tq1t8c.webp" alt="Self Host Stats" /></a> | <a href="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/gd5wei3oorzo6nwz96ro.webp"><img src="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/gd5wei3oorzo6nwz96ro.webp" alt="Team Display" /></a> | <a href="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/ikku6lr6cuqvv4ap5532.webp"><img src="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/ikku6lr6cuqvv4ap5532.webp" alt="File Manager" /></a> |

| Self Host Logs | Dashboard Overview |  Notification Preferences |
| :-: | :-: | :-: |
| <a href="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/quinawz7qvb6b5czi7u9.webp"><img src="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/quinawz7qvb6b5czi7u9.webp" alt="Self Host Logs" /></a> | <a href="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/iu7s99nj347eb24b2sdz.webp"><img src="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/iu7s99nj347eb24b2sdz.webp" alt="Dashboard Overview" /></a> |  <a href="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/jtcayilnk5oeyy3qmcrp.webp"><img src="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/jtcayilnk5oeyy3qmcrp.webp" alt="Notification Preferences" /></a> |

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

### Self Host Nixopus on your VPS

Once the CLI is installed, you can self host Nixopus on your VPS:

```bash
nixopus install
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

#### Installation with Options

You can also install the CLI and run `nixopus install` with options in a single command, refer [installation documentation](https://docs.nixopus.com/install/#installation-options) for more details on options:

```bash
# Basic installation with verbose output
curl -sSL https://install.nixopus.com | bash -s -- --verbose

# Installation with custom domains
curl -sSL https://install.nixopus.com | bash -s -- --api-domain api.example.com --view-domain app.example.com

# Dry-run to see what would happen
curl -sSL https://install.nixopus.com | bash -s -- --dry-run

# Install CLI only, skip nixopus install
curl -sSL https://install.nixopus.com | bash -s -- --skip-nixopus-install
```

Run `nixopus --help` to see all available commands.

## About the Name

Nixopus is derived from the combination of "octopus" and the Linux penguin (Tux). While the name might suggest a connection to [NixOS](https://nixos.org/), Nixopus is an independent project with no direct relation to NixOS or its ecosystem.

## Contributors

<a href="https://github.com/raghavyuva/nixopus/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=raghavyuva/nixopus" alt="Nixopus project contributors" />
</a>

<!-- sponsors-start -->
## 🎗️ Sponsors

| Avatar | Sponsor |
| ------ | ------- |
| [![](https://avatars.githubusercontent.com/u/47430686?u=4185ecc1ab0fb92dd3f722f0d3a34ed044de0aec&v=4&s=150)](https://github.com/shravan20) | [shravan20](https://github.com/shravan20) |

❤️ Thank you for your support!
<!-- sponsors-end -->
