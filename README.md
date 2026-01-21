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
  <a href="https://github.com/raghavyuva/nixopus/stargazers"><img src="https://img.shields.io/github/stars/raghavyuva/nixopus?style=flat-square" alt="GitHub stars" /></a>
  <a href="https://github.com/raghavyuva/nixopus/network/members"><img src="https://img.shields.io/github/forks/raghavyuva/nixopus?style=flat-square" alt="GitHub forks" /></a>
  <a href="https://github.com/raghavyuva/nixopus/issues"><img src="https://img.shields.io/github/issues/raghavyuva/nixopus?style=flat-square" alt="GitHub issues" /></a>
  <a href="https://github.com/raghavyuva/nixopus/blob/master/LICENSE.md"><img src="https://img.shields.io/badge/license-FSL--1.1--ALv2-blue?style=flat-square" alt="License" /></a>
  <br>
  <a href="https://trendshift.io/repositories/15336" target="_blank"><img src="https://trendshift.io/api/badge/repositories/15336" alt="Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/></a>
</p>

</div>

---

<details>
<summary><h2>Table of Contents</h2></summary>

- [About the name](#about-the-name)
- [Getting Started](#getting-started)
- [Features](#features)
  - [Extensions](#extensions)
  - [Hosting Projects](#hosting-projects)
  - [Terminal](#terminal)
  - [File Manager](#file-manager)
- [Demo](#demo)
- [Installation](#installation)
- [Links](#links)
- [Contributing](#contributing)
- [License](#license)

</details>

---

<a id="about-the-name"></a>
## About the Name

Nixopus is derived from the combination of "octopus" (representing flexibility and multi-tasking) and the Linux penguin mascot (Tux). While the name might suggest a connection to [NixOS](https://nixos.org/), Nixopus is an independent project with no direct relation to NixOS or its ecosystem.

---

<a id="getting-started"></a>
## Getting Started

> **Important Note**: Nixopus is currently in **alpha/pre-release** stage and is not yet ready for production use. While you're welcome to try it out, we recommend waiting for the beta or stable release before using it in production environments.

Nixopus transforms your VPS into a complete application hosting environment. Deploy applications directly from GitHub, manage server files through a browser-based interface, and execute commands via an integrated terminal—all without leaving the dashboard.

### Quick Start

1. **Install Nixopus** on your VPS:
   ```bash
   curl -sSL https://install.nixopus.com | bash
   ```

2. **Access the Dashboard** at `http://your-server-ip` or your configured domain

3. **Deploy Your First Project** by connecting your GitHub repository

For detailed installation instructions, visit our [Installation Guide](https://docs.nixopus.com/install/).

> [!IMPORTANT]
> Star us, you’ll receive all release notifications from GitHub without any delay.

[![Star](assets/star.png)](https://github.com/raghavyuva/nixopus)


---


<a id="features"></a>
## Features

Nixopus transforms your VPS into a complete application hosting environment. Deploy applications directly from GitHub, manage server files through a browser based interface, and execute commands via an integrated terminal all without leaving the dashboard.

<a id="extensions"></a>
### [Extensions](https://docs.nixopus.com/extensions)

Automate server tasks through a library of pre-built configurations. Extensions allow you to extend Nixopus functionality with modular components that integrate seamlessly with your workflow.

**Key Capabilities:**
- Pre-built extension library
- Custom extension support
- Easy integration with existing deployments
- Automated server task management


![Extensions](assets/nixopus_dashboard.jpeg)

<a id="hosting-projects"></a>
### [Hosting Projects](https://docs.nixopus.com/self-host)

Deploy applications directly from GitHub repositories with automatic builds and zero configuration files. Nixopus handles the entire deployment pipeline from code to production.

**Key Capabilities:**
- One-click deployments from GitHub
- Automatic builds and zero configuration
- CI/CD integration with automatic deployments on push
- Docker support for Compose, Dockerfiles, and static sites
- Automatic SSL certificate generation
- Reverse proxy routing with built-in Caddy
- Health checks and monitoring
- Environment variable management
- Custom domain configuration

![Hosting Projects](assets/nixopus_dashboard.jpeg)

<a id="terminal"></a>
### [Terminal](https://docs.nixopus.com/terminal)

Execute server commands through a secure, web-based terminal with SSH integration. Access your server directly from the browser without additional tools.

**Key Capabilities:**
- Web-based terminal interface
- Full SSH integration
- Secure command execution
- Real-time command output
- Multi-session support

![Terminal](assets/nixopus_dashboard.jpeg)

<a id="file-manager"></a>
### [File Manager](https://docs.nixopus.com/file-manager)

Browse, edit, upload, and organize files using drag-and-drop operations. Manage your server files through an intuitive visual interface.

**Key Capabilities:**
- Visual file browser
- Drag-and-drop file uploads
- In-browser file editing
- File organization and management
- Multi-file operations

![File Manager](assets/nixopus_dashboard.jpeg)

### Additional Features

- **Real-time Monitoring** - Track CPU, RAM, and disk usage with live system statistics
- **Smart Notifications** - Receive deployment alerts via Slack, Discord, or email
- **Authentication** - Built-in user management with SuperTokens integration
- **Domain Management** - Configure custom domains with automatic SSL certificate generation

---

<a id="demo"></a>
## Demo

See Nixopus in action, from deployments to day-to-day operations on your VPS.

### What you’ll see

- **Deploy from GitHub**: connect a repo and ship updates with a streamlined flow
- **Operate from the dashboard**: manage apps, domains, env vars, and deployments in one place
- **Built-in tools**: use the web terminal + file manager without leaving your browser

### Screenshots (placeholders)

> Replace the image paths below when you’re ready (screenshots / gifs / videos will be added later).

<div align="center">

| Dashboard | Hosting Projects |
| --- | --- |
| ![Dashboard Overview](assets/nixopus_dashboard.jpeg) | ![Hosting Projects](assets/nixopus_dashboard.jpeg) |
| *Overview of deployments, health, and activity* | *Connect a repo, deploy, and manage releases* |

| Terminal | File Manager |
| --- | --- |
| ![Terminal](assets/nixopus_dashboard.jpeg) | ![File Manager](assets/nixopus_dashboard.jpeg) |
| *Run commands securely from the browser* | *Upload, edit, and organize server files visually* |

</div>

### Video (placeholder)

- **Demo video**: (add YouTube link / MP4 later)

### Live Preview (optional)

- (add hosted demo URL later)

---

<a id="installation"></a>
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

<a id="links"></a>
## 🔗 Links

- **Website**: [https://nixopus.com](https://nixopus.com)
- **Documentation**: [https://docs.nixopus.com](https://docs.nixopus.com)
- **Discord Community**: [https://discord.gg/skdcq39Wpv](https://discord.gg/skdcq39Wpv)
- **Blog**: [https://nixopus.com/blog](https://nixopus.com/blog)
- **Roadmap**: [GitHub Discussions](https://github.com/raghavyuva/nixopus/discussions/262)
- **Report Issues**: [GitHub Issues](https://github.com/raghavyuva/nixopus/issues)
- **Feature Requests**: [GitHub Discussions](https://github.com/raghavyuva/nixopus/discussions)

---

<a id="contributing"></a>
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**, feel free to check out our [Github Issues](https://github.com/raghavyuva/nixopus/issues)

Thank you to all the contributors who help make Nixopus better!

<a href="https://github.com/raghavyuva/nixopus/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=raghavyuva/nixopus" alt="Nixopus project contributors" />
</a>

---

<a id="license"></a>
## License

Distributed under the FSL-1.1-ALv2 License. Visit [LICENSE.md](./LICENSE.md) for more information.

---


<div align="center">

**Made with ❤️ by the Nixopus community**

[Back to Top](#table-of-contents)

</div>
