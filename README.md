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

Nixopus is a powerful platform designed to simplify VPS management. Whether you're a DevOps engineer, system administrator, or developer, Nixopus streamlines your workflow with comprehensive tools for deployment, monitoring, and maintenance.

> ⚠️ **Important Note**: Nixopus is currently in alpha/pre-release stage and is not yet ready for production use. While you're welcome to try it out, we recommend waiting for the beta or stable release before using it in production environments. The platform is still undergoing testing and development.

## Features

- **Simplified VPS management**
  - *1 Click Application Deployment*: Deploy applications effortlessly with a single click.
  - *Integrated Web-Based Terminal*: Access your server's terminal directly from the browser.
  - *Intuitive File Manager*: Navigate and manage server files through a user-friendly interface.
  - *Real Time Monitoring*: Monitor your server's CPU, RAM, containers usage in real-time.
  - *Built in TLS Management*: Configure & manage TLS certificates for your domains.
  - *GitHub Integration for CI/CD*: Seamlessly integrate GitHub repositories.
  - *Proxy Management via Caddy*: Configure and manage reverse proxies.
  - *Notification Integration*: Configure to send real-time alerts to channels including Slack, Discord, or Email.
- **Comprehensive deployment tools**
- **User-friendly interface**
- **Customizable installation options**
- **Self Host Deployment**

## Table of Contents

- [Project Overview](#project-overview)
- [Features](#features)
- [Table of Contents](#table-of-contents)
- [Demo / Screenshots](#demo--screenshots)
- [Installation \& Quick Start](#installation--quick-start)
    - [Optional Parameters](#optional-parameters)
    - [Accessing Nixopus](#accessing-nixopus)
- [Usage](#usage)
- [Architecture](#architecture)
- [Development Guide](#development-guide)
  - [Development Setup](#development-setup)
  - [Running the Application](#running-the-application)
  - [Making Changes](#making-changes)
  - [Submitting a Pull Request](#submitting-a-pull-request)
  - [Proposing New Features](#proposing-new-features)
  - [Extending Documentation](#extending-documentation)
- [Contribution Guidelines](#contribution-guidelines)
- [Sponsorship](#sponsorship)
- [Community \& Support](#community--support)
- [License](#license)
- [Code of Conduct](#code-of-conduct)
- [Acknowledgments](#acknowledgments)
- [About the Name](#about-the-name)
- [Contributors](#contributors)
- [🎗️ Sponsors](#️-sponsors)

## Demo / Screenshots

| Self Host Stats | Team Display | File Manager |
| :-: | :-: | :-: |
| <a href="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/28nkmy49nm7oi5tq1t8c.webp"><img src="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/28nkmy49nm7oi5tq1t8c.webp" alt="Self Host Stats" /></a> | <a href="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/gd5wei3oorzo6nwz96ro.webp"><img src="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/gd5wei3oorzo6nwz96ro.webp" alt="Team Display" /></a> | <a href="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/ikku6lr6cuqvv4ap5532.webp"><img src="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/ikku6lr6cuqvv4ap5532.webp" alt="File Manager" /></a> |

| Self Host Logs | Dashboard Overview |  Notification Preferences |
| :-: | :-: | :-: |
| <a href="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/quinawz7qvb6b5czi7u9.webp"><img src="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/quinawz7qvb6b5czi7u9.webp" alt="Self Host Logs" /></a> | <a href="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/iu7s99nj347eb24b2sdz.webp"><img src="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/iu7s99nj347eb24b2sdz.webp" alt="Dashboard Overview" /></a> |  <a href="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/jtcayilnk5oeyy3qmcrp.webp"><img src="https://dev-to-uploads.s3.amazonaws.com/uploads/articles/jtcayilnk5oeyy3qmcrp.webp" alt="Notification Preferences" /></a> |

## Installation & Quick Start

This section will help you set up Nixopus on your VPS quickly.

To install Nixopus on your VPS, ensure you have sudo access and run the following command:

```
sudo bash -c "$(curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install.sh)"
```

#### Optional Parameters

You can customize your installation by providing the following optional parameters:

- `--api-domain`: Specify the domain where the Nixopus API will be accessible (e.g., `nixopusapi.example.tld`)
- `--app-domain`: Specify the domain where the Nixopus app will be accessible (e.g., `nixopus.example.tld`)
- `--email` or `-e`: Set the email for the admin account
- `--password` or `-p`: Set the password for the admin account

Example with optional parameters:

```
sudo bash -c "$(curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install.sh)" -- \
  --api-domain nixopusapi.example.tld \
  --app-domain nixopus.example.tld \
  --email admin@example.tld \
  --password Adminpassword@123 \
  --env production
```

#### Accessing Nixopus

After successful installation, you can access the Nixopus dashboard by visiting the URL you specified in the `--app-domain` parameter (e.g., `https://nixopus.example.tld`). Use the email and password you provided during installation to log in.

> **Note**: The installation script has not been tested in all distributions and different operating systems. If you encounter any issues during installation, please create an issue on our [GitHub repository](https://github.com/raghavyuva/nixopus/issues) with details about your environment and the error message you received.

## Usage

Once installed, Nixopus provides a dashboard for managing your VPS. You can deploy applications, monitor performance, and perform maintenance tasks directly from the interface.

## Architecture

Nixopus is built using a microservices architecture, leveraging Go for backend services and React for the frontend. It uses PostgreSQL for data storage and is designed to be scalable and efficient. To learn more about the architecture, refer to the [Architecture Overview](https://docs.nixopus.com/architecture) section in the documentation.

## Development Guide

### Development Setup

1. Clone the repository:

```bash
git clone https://github.com/raghavyuva/nixopus.git
cd nixopus
```

2. Install Go (version 1.23.6 or newer), and PostgreSQL.

3. Set up PostgreSQL databases:

```bash
createdb postgres -U postgres

createdb nixopus_test -U postgres
```

4. Copy and configure environment variables:

```bash
cp .env.sample .env
```

5. Install project dependencies:

```bash
cd api
go mod download

cd ../view
yarn install
```

6. Load development fixtures (optional but recommended):

```bash
cd ../api
# Load fixtures without affecting existing data
make fixtures-load
```

The fixtures system provides sample data including users, organizations, roles, permissions, and feature flags to help you get started quickly with development. Make Changes to Fixtures data in `/api/fixtures/development` directory if required. [Read More About Fixtures](https://docs.nixopus.com/contributing/fixtures)

### Running the Application

1. Start the API service:

```bash
cd api
air
```

2. Start the view service:

```bash
cd view
yarn dev
```

### Making Changes

Nixopus follows [Forking-Workflow]([https://www.atlassian.com/continuous-delivery/continuous-integration/trunk-based-development](https://www.atlassian.com/git/tutorials/comparing-workflows/forking-workflow)) conventions.

1. Create a new branch:

```bash
git checkout -b feature/your-feature-name
```

2. Make your changes following the project structure:
   - Place new features under `api/internal/features/`
   - Add tests for new functionality
   - Update migrations if needed
   - Follow existing patterns for controllers, services, and storage
   - For frontend changes, follow the Next.js app directory structure

3. Run tests:

```bash
cd api
make test

# View linting
cd view
yarn lint
```

4. Commit your changes with clear messages.

---

### Submitting a Pull Request

1. Push your branch and create a pull request.
2. Ensure your code:
   - Follows the project structure
   - Includes tests
   - Updates documentation if needed
   - Passes all CI checks
3. Be prepared to address feedback.

### Proposing New Features

1. Check existing issues and pull requests.

2. Create a new issue with the `Feature request` template.

3. Include:
   - Feature description
   - Technical implementation details
   - Impact on existing code

### Extending Documentation

Documentation is located in the `docs/` directory. Follow the existing structure and style when adding new content.

## Contribution Guidelines

Thank you for your interest in contributing to Nixopus! This [guide](docs/contributing/README.md) will help you get started with the development setup and explain the contribution process.

## Sponsorship

We've dedicated significant time to making Nixopus free and accessible. Your support helps us continue our development and vision for open source. Consider becoming a sponsor and join our community of supporters.

- ![GitHub Sponsors](https://img.shields.io/github/sponsors/raghavyuva?label=Github%20Sponsor)
- <a href="https://liberapay.com/raghavyuva/donate"><img src="https://img.shields.io/liberapay/goal/raghavyuva.svg?logo=liberapay" alt="Donate to raghavyuva via Liberapay"></a>

## Community & Support

If you find Nixopus useful, please consider giving it a star and sharing it with your network!

## License

Nixopus is licensed under the MIT License. See the [LICENSE](LICENSE.md) file for more information.

## Code of Conduct

Before contributing, please review and agree to our [Code of Conduct](/docs/code-of-conduct/index.md). We're committed to maintaining a welcoming and inclusive community.

## Acknowledgments

We would like to thank all contributors and supporters of Nixopus. Your efforts and feedback are invaluable to the project's success.

## About the Name

Nixopus is derived from the combination of "octopus" and the Linux penguin (Tux). While the name might suggest a connection to [NixOS](https://nixos.org/), Nixopus is an independent project with no direct relation to NixOS or its ecosystem.

## Contributors

<a href="https://github.com/raghavyuva/nixopus/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=raghavyuva/nixopus" alt="Nixopus project contributors" />
</a>

Made with [contrib.rocks](https://contrib.rocks).

<!-- sponsors-start -->
## 🎗️ Sponsors

| Avatar | Sponsor |
| ------ | ------- |
| [![](https://avatars.githubusercontent.com/u/47430686?u=4185ecc1ab0fb92dd3f722f0d3a34ed044de0aec&v=4&s=150)](https://github.com/shravan20) | [shravan20](https://github.com/shravan20) |

❤️ Thank you for your support!
<!-- sponsors-end -->
