---
name: nixopus-docs
description: Look up Nixopus platform documentation at runtime. Use when the user asks about Nixopus features, configuration, concepts, self-hosting, API reference, guides, or anything where accurate product information is needed. Prevents hallucination by fetching the latest docs.
metadata:
  version: "2.0"
---

# Nixopus Documentation Lookup

When you need to answer questions about Nixopus — features, setup, configuration, API, guides, concepts, self-hosting, extensions, domains, deployments, or any product-level detail — **always fetch the latest documentation** instead of relying on memory.

## When to Use

- User asks "how does X work in Nixopus?"
- User asks about configuration, environment variables, or setup
- User references a Nixopus feature you're unsure about
- User asks about self-hosting, cloud, billing, teams, or extensions
- User needs API reference or endpoint details
- You need to explain a Nixopus concept (deployments, domains, organizations, auth)
- Whenever you're about to describe Nixopus behavior and aren't 100% certain from tool results

## IMPORTANT — How to Fetch Docs

These are **runtime tools**, NOT files inside this skill. Do NOT use `skill_read` or `skill_search` to look up docs.

To fetch a doc page:
1. Run `search_tools("nixopus docs")` to find the tool.
2. Run `load_tool("fetch_nixopus_docs_page")` to load it.
3. Call `fetch_nixopus_docs_page({ path: "concepts/deployments.md" })` with the page path from the index below.

You do NOT need to call `fetch_nixopus_docs_index` — the full index is embedded below.

## Docs Index

Source: https://docs.nixopus.com/llms.txt

### Getting Started
- [Introduction](getting-started/introduction.md) — How Nixopus works and what you can do with it.
- [Quickstart](getting-started/quickstart.md) — Deploy your first app on Nixopus Cloud.

### Concepts
- [Authentication](concepts/authentication.md) — How users and services authenticate with Nixopus.
- [Deployments](concepts/deployments.md) — How Nixopus builds and deploys your apps.
- [Domains](concepts/domains.md) — How Nixopus handles routing, SSL, and custom domains.
- [Organizations](concepts/organizations.md) — Manage teams and access control.

### Guides
- [AI chat](guides/ai-chat.md) — Deploy, configure, and troubleshoot with the Nixopus AI agent.
- [Charts](guides/charts.md) — Monitor system resources and running services.
- [Containers](guides/containers.md) — Manage your running containers.
- [Deploy your first app](guides/deploying-apps.md) — Go from repo to live in a few clicks.
- [Multi-service apps](guides/docker-compose.md) — Deploy multi-service apps with Docker Compose.
- [Environment variables](guides/environment-variables.md) — Configure build-time and runtime variables.
- [Extensions](guides/extensions.md) — Extend Nixopus with the extension marketplace.
- [GitHub integration](guides/github-integration.md) — Connect your repos and enable auto-deploy.
- [Health checks](guides/health-checks.md) — Monitor your apps and catch issues early.
- [Notifications](guides/notifications.md) — Get alerted about deployments and events.
- [Terminal](guides/terminal.md) — Access a built-in terminal from the dashboard.

### Cloud
- [API keys](cloud/api-keys.md) — Create and manage API keys for programmatic access.
- [Credits](cloud/credits.md) — How credit-based billing works on Nixopus Cloud.
- [Custom domains](cloud/custom-domains.md) — Connect your own domain to Nixopus Cloud.
- [Machines](cloud/machines.md) — How machines are provisioned, managed, and scaled.
- [Teams](cloud/teams.md) — Collaborate with your team on Nixopus Cloud.

### Self-Hosting
- [Installation](self-hosting/installation.md) — Install Nixopus on your own machine.
- [Configuration](self-hosting/configuration.md) — Environment variables, ports, firewall, and HTTPS.
- [Updating](self-hosting/updating.md) — Keep your self-hosted instance up to date.
- [Backup & Restore](self-hosting/backup-and-restore.md) — Back up and restore your data.
- [Branch Preview](self-hosting/branch-preview.md) — Test PRs and branches on a real VPS.
- [Management CLI](self-hosting/management-cli.md) — Manage from the command line.
- [Troubleshooting](self-hosting/troubleshooting.md) — Common issues and fixes.

### Editor Extension
- [Overview](extension/overview.md) — Deploy from VS Code or Cursor.
- [Installation](extension/installation.md) — Install the Nixopus extension.
- [Authentication](extension/authentication.md) — Sign in from your editor.
- [Deploying](extension/deploying.md) — Go live without leaving your editor.
- [Workspace sync](extension/workspace-sync.md) — Link your workspace to a project.
- [Chat interface](extension/chat-interface.md) — Talk to the deploy agent.

### API Reference
- [Introduction](api-reference/introduction.md) — Interact with Nixopus programmatically.
- [OpenAPI spec](api-reference/openapi.json) — Full OpenAPI specification.
- Deploy: delete, deploy, get, update application.
- Domains: list, add, remove, custom domains, DNS check.
- GitHub: connectors, repositories, branches, webhooks.
- Containers: list, get, start, stop, restart, remove, logs, resources.
- MCP: add, delete, list, update, test servers; discover/invoke tools.
- Notifications: SMTP, webhook config; send; preferences.
- Health checks: create, get, update, delete, toggle, stats, results.
- Machines: list, register BYOS, pause, resume, restart, remove, events, metrics, billing, backup.
- Extensions: list, get, fork, delete, run, executions, categories.
- File manager: list, upload, create/copy/move/delete directory.
- User: profile, settings, preferences, avatar, name.
- Other: feature flags, bootstrap, onboarding, updates, audit logs, compose, projects, trail.

### Other
- [Changelog](changelog.md) — Product updates, new features, and bug fixes.
- [Contributing](contributing.md) — Set up your development environment and contribute.
- [Code of Conduct](code-of-conduct.md) — Community standards.
- [License](license.md) — FSL-1.1-ALv2.

## Workflow

1. **Find the right page** — scan the index above for the page matching the user's question.
2. **Fetch it** — `search_tools("nixopus docs")` → `load_tool("fetch_nixopus_docs_page")` → call with the path.
3. **Answer from the content** — base your response on the fetched documentation. If one page isn't enough, fetch more.

## Rules

- Never fabricate Nixopus features, settings, or API endpoints. If you can't find it in the docs, say so.
- For API reference questions, fetch the specific endpoint page from `api-reference/`.
- If the user's question spans multiple topics, fetch multiple pages.
- Keep fetched content as context — don't re-fetch the same page in the same conversation.
