<div align="center"><a name="readme-top"></a>

![][image-overview]

An open-source, AI-powered platform that deploys, monitors, and fixes your apps, autonomously.<br/>
Self-host on your own infrastructure or use [Nixopus Cloud][dashboard-link] to go live in minutes.

<p align="center">
  <a href="https://nixopus.com"><b>Website</b></a> •
  <a href="https://docs.nixopus.com"><b>Documentation</b></a> •
  <a href="https://nixopus.com/blog"><b>Blog</b></a> •
  <a href="https://discord.gg/skdcq39Wpv"><b>Discord</b></a>
</p>

<!-- SHIELD GROUP -->

[![][github-stars-shield]][github-stars-link]
[![][github-forks-shield]][github-forks-link]
[![][github-issues-shield]][github-issues-link]
[![][github-license-shield]][github-license-link]
[![][discord-shield]][discord-link]

[![][github-trending-shield]][github-trending-url]

</div>

<details>
<summary><kbd>Table of contents</kbd></summary>

#### TOC

- [Getting Started \& Join Our Community](#getting-started--join-our-community)
  - [How It Works](#how-it-works)
- [Features](#features)
  - [AI-Powered Lifecycle](#ai-powered-lifecycle)
  - [Chat Interface](#chat-interface)
  - [Editor Extension](#editor-extension)
  - [Multi-Server Orchestration](#multi-server-orchestration)
  - [Full Machine Access](#full-machine-access)
  - [Framework Agnostic](#framework-agnostic)
  - [`*` What's more](#-whats-more)
- [Demo](#demo)
- [Self Hosting](#self-hosting)
  - [Quick Install](#quick-install)
  - [Configuration](#configuration)
  - [Requirements](#requirements)
- [Contributing](#contributing)
- [About the Name](#about-the-name)


####

<br/>

</details>

## Getting Started & Join Our Community

Nixopus is the deployment platform where an AI agent handles your entire deploy lifecycle, from analyzing your codebase and generating configs to shipping your app and fixing failures. Connect your repo, tell the agent to deploy, and go live. Learn more in the [introduction][docs-introduction] or jump to the [quickstart][docs-quickstart].

| [![][go-live-shield-badge]][dashboard-link] | No installation required! Sign up and deploy your first app on Nixopus Cloud.  |
| :------------------------------------------ | :----------------------------------------------------------------------------- |
| [![][discord-shield-badge]][discord-link]    | Join our Discord community! Connect with developers and other Nixopus users.   |

> [!IMPORTANT]
>
> **Star Us**, You will receive all release notifications from GitHub without any delay ~ ⭐️

### How It Works

1. **Connect your repo** - [link your GitHub account][docs-github-integration] and select a repository.
2. **Tell the agent to deploy** - from the [dashboard][docs-ai-chat] or your [editor][docs-extension-deploying], the agent analyzes your codebase, generates the right config, and deploys.
3. **Go live** - your app gets a URL with HTTPS. Automatic SSL, routing, and domain management.
4. **Agent keeps watching** *(in development)* - if something fails, the agent reads the logs, creates a PR with a fix, and redeploys.

<div align="right">

[![][back-to-top]](#readme-top)

</div>

## Features


### [AI-Powered Lifecycle][docs-introduction]
![][image-feat-ai-deploy]
The agent generates configs, deploys, and fixes failures, autonomously. On other platforms, a failed deployment means reading logs, copying the error to an AI tool, getting a fix, pushing, redeploying. Repeat. Nixopus closes that loop. The agent detects the failure, diagnoses it, raises a PR with the fix, and redeploys without you.

[![][back-to-top]](#readme-top)

### [Chat Interface][docs-ai-chat]
![][image-feat-chat]
Deploy, add domains, check logs, roll back, troubleshoot. One conversational interface. Talk to the agent in natural language from the [dashboard][docs-ai-chat] or your [editor][docs-extension-deploying]. Tag resources with `@App`, `@Container`, or `@Domain` to give it focus.

[![][back-to-top]](#readme-top)

### [Editor Extension][docs-extension]
![][image-feat-editor]
Deploy from VS Code or Cursor without opening a browser. The [extension][docs-extension] puts the same agent in your sidebar. Chat, deploy, and manage your apps without leaving the editor.

[![][back-to-top]](#readme-top)

### [Multi-Server Orchestration][docs-introduction]
![][image-feat-multi-server]
One dashboard. Every server. Connect multiple servers and manage them all from one place. Monitor CPU, RAM, and running apps across your entire fleet. Multi-machine deployments, load balancing, and automated scaling are on the roadmap.

[![][back-to-top]](#readme-top)

### [Full Machine Access][docs-terminal]
![][image-feat-terminal]
Terminal and containers right in the dashboard. SSH into your server, inspect running [containers][docs-containers], stream logs, and debug from the browser. No separate [terminal][docs-terminal] or jump host required.

[![][back-to-top]](#readme-top)

### [Framework Agnostic][docs-introduction]
![][image-feat-framework]
Next.js, Django, Rails, Go, FastAPI, Compose stacks. Anything that runs in a container. If it fits a container image, it [ships on Nixopus][docs-deploying-apps]. The agent detects your stack and generates the right build configuration automatically.

<div align="right">

[![][back-to-top]](#readme-top)

</div>

### `*` What's more

Beyond these features, Nixopus also includes:

- [x] **Auto TLS**: Every deployment gets TLS via Caddy. SSL provisioned and renewed automatically via Let's Encrypt. See [configuration][docs-configuration].
- [x] **[Custom Domains][docs-domains]**: Point your domain at Nixopus with automatic DNS verification and SSL. Tell the agent "add domain app.mysite.com."
- [x] **[Instant Rollbacks][docs-deployments]**: Roll back to any previous deployment. Previous images are retained, so rollbacks don't require a full rebuild.
- [x] **[Open Source & Self-Hostable][docs-installation]**: Your code, your data, your infrastructure. Self-hosting is free forever with full feature parity. No lock-in.

> More features are being added as Nixopus evolves.

---

<div align="right">

[![][back-to-top]](#readme-top)

</div>

## Demo

See Nixopus in action, from deploying apps to day-to-day operations on your infrastructure.

<https://github.com/user-attachments/assets/6d6f24ef-47d5-4fe2-ab63-65f0ed5f7782>

<div align="right">

[![][back-to-top]](#readme-top)

</div>

## Self Hosting

Run the full Nixopus stack on a machine you control.

> [!TIP]
>
> Learn more in the [Self-Hosting Installation Guide][docs-installation].

### Quick Install

```bash
curl -fsSL install.nixopus.com | sudo bash
```

Or with a custom domain and admin email:

```bash
curl -fsSL install.nixopus.com | sudo DOMAIN=panel.example.com ADMIN_EMAIL=admin@example.com OPENROUTER_API_KEY=sk-or-xxxxx bash
```

Or use a different LLM provider:

```bash
curl -fsSL install.nixopus.com | sudo LLM_PROVIDER=anthropic ANTHROPIC_API_KEY=sk-ant-xxxxx bash
```

<br/>

### Configuration

All parameters are optional. Pass them as environment variables before the install command:

| Environment Variable | Required | Description                  | Default                |
| -------------------- | -------- | ---------------------------- | ---------------------- |
| `DOMAIN`             | No       | Domain for automatic HTTPS   | *(empty, IP mode)*   |
| `HOST_IP`            | No       | Public IP of the machine     | *(auto-detected)*      |
| `ADMIN_EMAIL`        | No       | Admin account email          | *(empty)*              |
| `CADDY_HTTP_PORT`    | No       | HTTP port                    | `80`                   |
| `CADDY_HTTPS_PORT`   | No       | HTTPS port                   | `443`                  |
| `NIXOPUS_HOME`       | No       | Installation directory       | `/opt/nixopus`         |
| `LLM_PROVIDER`       | No       | LLM provider for the AI agent: `openrouter`, `openai`, `anthropic`, `google`, `deepseek`, or `groq` | `openrouter` |
| `OPENROUTER_API_KEY`  | No       | [OpenRouter](https://openrouter.ai) API key (default provider, gives access to Claude, GPT-4, Gemini, and more via one key) | *(empty)* |
| `OPENAI_API_KEY`     | No       | OpenAI API key               | *(empty)*              |
| `ANTHROPIC_API_KEY`  | No       | Anthropic API key            | *(empty)*              |
| `GOOGLE_GENERATIVE_AI_API_KEY` | No | Google AI API key       | *(empty)*              |
| `DEEPSEEK_API_KEY`   | No       | DeepSeek API key             | *(empty)*              |
| `GROQ_API_KEY`       | No       | Groq API key                 | *(empty)*              |

> [!NOTE]
>
> The complete list of environment variables can be found in the [Configuration Guide][docs-configuration].

<br/>

### Requirements

| Requirement  | Minimum                                                                       |
| ------------ | ----------------------------------------------------------------------------- |
| Machine      | Fresh VPS from any cloud provider (Hetzner, DigitalOcean, AWS, etc.)          |
| Architecture | x86_64 (amd64) or aarch64 (arm64)                                            |
| RAM          | 1 GB minimum (2 GB+ recommended)                                             |
| Disk         | 2 GB free minimum                                                             |
| Access       | Root (the installer must run as root)                                         |
| Docker       | Installed automatically if not present (Docker Engine + Compose V2)           |

> [!NOTE]
>
> Use a fresh, dedicated VPS, not a machine already running other production services. The first user to sign up becomes the admin. After that, registration is closed and you invite users manually.

<div align="right">

[![][back-to-top]](#readme-top)

</div>

## Contributing

Contributions of all types are more than welcome; if you are interested in contributing code, feel free to check out our GitHub [Issues][github-issues-link] to get stuck in to show us what you're made of.

> [!TIP]
>
> We welcome all contributions that help make Nixopus better. Whether it's bug fixes, new features, documentation, or feedback, every contribution counts.

<a href="https://github.com/raghavyuva/nixopus/graphs/contributors" target="_blank">
  <table>
    <tr>
      <th colspan="2">
        <br><img src="https://contrib.rocks/image?repo=raghavyuva/nixopus"><br><br>
      </th>
    </tr>
  </table>
</a>

<div align="right">

[![][back-to-top]](#readme-top)

</div>

## About the Name

Nixopus is derived from the combination of "octopus" (representing flexibility and multi-tasking) and the Linux penguin mascot (Tux). While the name might suggest a connection to [NixOS](https://nixos.org/), Nixopus is an independent project with no direct relation to NixOS or its ecosystem.

<div align="right">

[![][back-to-top]](#readme-top)

</div>

---

<details><summary><h4>License</h4></summary>

Distributed under the FSL-1.1-ALv2 License. Visit [LICENSE.md](./LICENSE.md) or the [docs][docs-license] for more information.

</details>

Copyright © 2025 [Nixopus][website-link]. <br />
This project is [FSL-1.1-ALv2](./LICENSE.md) licensed.

<!-- LINK GROUP -->

[back-to-top]: https://img.shields.io/badge/-BACK_TO_TOP-151515?style=flat-square
[dashboard-link]: https://dashboard.nixopus.com
[discord-link]: https://discord.gg/skdcq39Wpv
[discord-shield]: https://img.shields.io/badge/Discord-Join-5865F2?labelColor=black&logo=discord&logoColor=white&style=flat-square
[discord-shield-badge]: https://img.shields.io/badge/Discord-Join-5865F2?labelColor=black&logo=discord&logoColor=white&style=for-the-badge
[docs-ai-chat]: https://docs.nixopus.com/guides/ai-chat
[docs-configuration]: https://docs.nixopus.com/self-hosting/configuration
[docs-containers]: https://docs.nixopus.com/guides/containers
[docs-deploying-apps]: https://docs.nixopus.com/guides/deploying-apps
[docs-deployments]: https://docs.nixopus.com/concepts/deployments
[docs-domains]: https://docs.nixopus.com/concepts/domains
[docs-extension]: https://docs.nixopus.com/extension/overview
[docs-extension-deploying]: https://docs.nixopus.com/extension/deploying
[docs-github-integration]: https://docs.nixopus.com/guides/github-integration
[docs-installation]: https://docs.nixopus.com/self-hosting/installation
[docs-introduction]: https://docs.nixopus.com/getting-started/introduction
[docs-license]: https://docs.nixopus.com/license
[docs-quickstart]: https://docs.nixopus.com/getting-started/quickstart
[docs-terminal]: https://docs.nixopus.com/guides/terminal
[github-forks-link]: https://github.com/raghavyuva/nixopus/network/members
[github-forks-shield]: https://img.shields.io/github/forks/raghavyuva/nixopus?color=8ae8ff&labelColor=black&style=flat-square
[github-issues-link]: https://github.com/raghavyuva/nixopus/issues
[github-issues-shield]: https://img.shields.io/github/issues/raghavyuva/nixopus?color=ff80eb&labelColor=black&style=flat-square
[github-license-link]: https://github.com/raghavyuva/nixopus/blob/master/LICENSE.md
[github-license-shield]: https://img.shields.io/badge/license-FSL--1.1--ALv2-white?labelColor=black&style=flat-square
[github-stars-link]: https://github.com/raghavyuva/nixopus/stargazers
[github-stars-shield]: https://img.shields.io/github/stars/raghavyuva/nixopus?color=ffcb47&labelColor=black&style=flat-square
[github-trending-shield]: https://trendshift.io/api/badge/repositories/15336
[github-trending-url]: https://trendshift.io/repositories/15336
[go-live-shield-badge]: https://img.shields.io/badge/TRY_NIXOPUS-CLOUD-55b467?labelColor=black&style=for-the-badge
[image-feat-ai-deploy]: assets/graphics/2.png
[image-feat-chat]: assets/graphics/3.png
[image-feat-editor]: assets/graphics/4.png
[image-feat-framework]: assets/graphics/7.png
[image-feat-multi-server]: assets/graphics/5.png
[image-feat-terminal]: assets/graphics/6.png
[image-overview]: assets/graphics/8.png
[website-link]: https://nixopus.com
