# Features

Everything Nixopus can do, organized by area. from AI-assisted deployments and in-browser terminals to multi-server routing, automatic TLS, and a full self-host CLI. If it's built, it's listed here.

## AI Agent

- **Natural language infra control** — Deploy, rollback, or diagnose issues by describing what you want; the agent figures out the steps
- **Domain-specific agents** — Separate agent profiles for deployments, diagnostics, GitHub, notifications, machines, and billing so the agent responds with the right expertise for the task
- **@-mention context** — Type `@` in chat to attach a specific app or resource as structured context so the agent operates on exactly what you mean
- **Bring your own LLM** — Point the agent at any OpenAI-compatible endpoint (OpenRouter, self-hosted models, etc.) and switch models per conversation
- **Per-conversation model picker** — Choose a different model for each chat and the selection persists across reloads
- **Auto-run tools toggle** — Control whether the agent executes tool calls automatically or pauses for your confirmation before taking action
- **Incident auto-analysis** — Deploy failures and critical health check alerts are automatically analyzed and surfaced as incident threads in chat
- **Repo → chat deep-link** — Link directly to chat with a repo pre-loaded as context so the agent starts a deploy conversation already knowing the target
- **Scheduled automation** — Have the agent run recurring tasks on a schedule so you stop doing the same thing manually every week
- **Schedule run history** — See every past run of a scheduled task with its outcome so you know if automation is working or silently failing
- **MCP tool integration** — The agent can call your connected tools mid-conversation; say "open a Linear issue for this deploy failure" and it creates the ticket with context already filled in, or "check Sentry for errors since the last release" and it returns the stack traces inline
- **AI usage tracking** — See prompt tokens, completion tokens, latency, and cost per request; filter by model (e.g. how much did claude-3-5 cost this week?), by user (who ran the most agent sessions?), or by day to spot usage spikes

## Deployments

- **Framework-agnostic deployment** — Deploy any language or framework that runs in Docker; Node, Python, Go, Ruby, PHP, Rust — if it has a Dockerfile it deploys
- **One-click deploy from GitHub** — Connect a repo and deploy any branch without writing pipeline config
- **Multiple GitHub accounts** — Connect more than one GitHub account or org so teams with separate repos can all deploy from the same dashboard
- **GitHub App install wizard** — Set up a GitHub App through a guided manifest flow instead of manually creating one in GitHub settings
- **Auto-deploy on push** — Every push to a branch goes live automatically; no CI setup, no extra config
- **Build pack selection** — Choose Dockerfile, Docker Compose, or static site as the build strategy per app
- **Monorepo support** — Set a custom Dockerfile path and base directory so you can deploy any sub-project from a monorepo
- **Environment variables** — Set runtime environment variables per app from the dashboard without touching code or SSH
- **Build variables** — Inject variables that are only available during the build phase, separate from runtime secrets
- **Pre/post-run commands** — Run shell commands before the app starts or after it stops; useful for migrations, cache warm-ups, or cleanup scripts
- **Multi-server routing** — Deploy the same app to multiple machines simultaneously with round-robin load balancing or primary-failover strategy
- **Application labels** — Tag apps with custom labels to filter, search, and organize large fleets
- **Docker Compose support** — Deploy multi-service apps defined as compose stacks without any extra tooling
- **Compose per-service domain routing** — Map a different hostname to each service in a compose stack so every container gets its own URL
- **Redeploy without cache** — Force a full rebuild that bypasses Docker's layer cache when a stale layer is causing problems
- **Draft & deploy later** — Save an app's full configuration as a draft and trigger the first deployment only when you're ready
- **App library** — Browse all apps with label filters, text search, sort controls, and a grid/list view toggle
- **Instant rollback** — Go back to any previous deployment in seconds when something breaks in production
- **Restart without redeploy** — Bounce a stuck app without triggering a full rebuild and losing time
- **Cancel in-flight deploy** — Stop a deployment that's going wrong before it finishes and makes things worse
- **App recovery** — Bring a broken app back to a known-good state from the dashboard without manual SSH intervention
- **Deployment history** — Browse every deploy ever made for an app to understand what changed and when
- **Build logs** — See the full build output for any specific deployment to pinpoint exactly where it failed
- **Runtime logs** — Read live application logs with time range and keyword filters without touching the server
- **Build artifact downloads** — Download the built image from any past deployment for auditing or local testing
- **Template library** — Deploy services like Postgres, n8n, MinIO, or Uptime Kuma from a ready-made template instead of writing compose files from scratch
- **Project groups** — Group related apps into a project and deploy or manage them all together with one action
- **Duplicate project** — Clone a project's full config to spin up a new environment instantly without reconfiguring everything from scratch
- **Multi-environment families** — Link apps across staging, canary, and production into a family and see all environments side by side
- **Move app between machines** — Reassign an app to a different server without deleting and recreating it
- **Per-application health checks** — Attach HTTP or TCP probes directly to a deployed app so failures surface in the same place you manage deployments
- **Proactive failure detection** — Get alerted the moment a check fails so you find out about downtime before your users do
- **Uptime history** — Know the uptime percentage and response time trend for every monitored endpoint over the last hour, day, week, or month

## Domains & TLS

- **Custom domains** — Point your own domain at any app and have it live immediately
- **Auto TLS** — HTTPS is provisioned and renewed automatically; you never touch a certificate
- **Instant subdomain** — Get a shareable public URL for a new app before you have a real domain ready

## Machines & Fleet

- **BYOS** — Bring any VM you already own and deploy to it the same way as a cloud machine
- **Multi-machine fleet** — Manage every server from one place and deploy across machines without separate SSH sessions
- **Default machine** — Mark one machine as the org default so new apps land there without you having to pick every time
- **In-browser terminal** — Open a full interactive SSH session on any machine directly in the browser; no local SSH client or key setup needed
- **Multi-tab terminal sessions** — Run up to 5 independent shell sessions simultaneously and switch between them
- **Split panes** — Divide the terminal into up to 4 side-by-side panes with drag-to-resize so you can watch logs in one pane while running commands in another
- **Terminal appearance settings** — Customize font family, font size, cursor style, line height, letter spacing, and scrollback buffer size
- **Remote exec** — Run a one-off command on any machine from the dashboard without opening a terminal
- **Live machine metrics** — See CPU, RAM, disk, and bandwidth for every machine in real time with historical time-range queries
- **Anomaly detection** — Get notified when unusual network or resource activity appears on a machine before it becomes an incident
- **Machine plans** — Choose a compute plan for a provisioned machine and switch plans as your needs change

## Containers

- **Container management** — Start, stop, restart, and read logs for any Docker container without opening a terminal
- **Container terminal** — Open an interactive shell inside any running container directly from the browser
- **Resource visibility** — See which containers are consuming CPU or memory so you can act before things break
- **Disk cleanup** — Reclaim space from unused images and stale build cache with one click

## Backups

- **Scheduled backups** — Set a recurring backup cadence and stop worrying about doing it manually
- **On-demand snapshots** — Capture a backup right before a risky change so you always have a restore point

## Monitoring

- **Live system dashboard** — CPU, RAM, disk, and container status across all machines on one screen, always current
- **System info card** — See hostname, OS, CPU model, core count, kernel version, and uptime at a glance without SSHing in
- **Deployment stats widget** — At-a-glance recent deploy activity and aggregate success/failure counts pinned to the dashboard
- **Personalized dashboard layout** — Drag widgets into any order, remove ones you don't need, and reset to default; layout is saved per browser

## Notifications

- **Slack / Discord alerts** — Get deploy completions, failures, and health events posted straight into your team channels
- **Outbound webhooks** — Send any Nixopus event to any URL so you can wire alerts into tools that aren't Slack or Discord
- **Email alerts** — Receive critical alerts and team invitations via your own SMTP server

## MCP Servers

Connect external tools so the AI agent can take real actions across your stack. Supported providers out of the box:

- **Supabase** — Query and manage your Supabase project from the agent
- **GitHub** — Create issues, review PRs, and interact with repos without leaving the chat
- **Linear** — Let the agent create and triage issues as part of an incident or deploy workflow
- **Sentry** — Pull error details and stack traces into the agent's context when debugging
- **Atlassian** — Access Jira issues and Confluence pages directly from the agent
- **Semgrep** — Run security scans and surface vulnerabilities through the agent
- **Neon** — Manage Neon Postgres databases via the agent
- **PlanetScale** — Query and manage PlanetScale MySQL databases via the agent
- **Custom** — Point the agent at any self-hosted MCP-compatible server

## Extensions

Browse and deploy a catalog of pre-packaged open-source services with one click. Current catalog includes:

- **Databases** — PostgreSQL, MariaDB, Redis, Qdrant
- **Storage** — MinIO (S3-compatible object storage), File Browser
- **Monitoring** — Netdata, Uptime Kuma, Statping-ng, Dozzle, Beszel, Healthchecks, Changedetection.io, Apprise API, Gotify
- **Automation & Dev tools** — n8n, Code Server (VS Code in browser), Webhook Tester, Ollama, Grav
- **Productivity** — BookStack, DokuWiki, HedgeDoc, Excalidraw, Kimai, Linkding
- **Social** — Mastodon, Postiz
- **Network** — DuckDNS

## Teams

- **Invite members** — Send a magic-link invite to any email address with a role pre-assigned (Admin, Member, or Viewer)
- **Role management** — Change a teammate's role at any time; Admins control everything, Members deploy and manage apps, Viewers can only read
- **Remove members** — Revoke access for a teammate instantly without affecting their work history
- **Team settings** — Update the org name and description from one place

## Audit & Access

- **Audit log** — See every change made in your org with who did it and when, so nothing goes unexplained
- **RBAC** — Every action in the dashboard is gated by role so access always matches responsibility
- **API keys** — Generate keys with configurable expiration; the dashboard shows each key's prefix, last-used date, and active/expiring status so you always know what's in circulation
- **Passkeys** — Log in with a fingerprint or device PIN instead of a password
- **Step-up authentication** — Sensitive actions require a fresh passkey verification so a hijacked session can't cause damage without physical device access
- **Two-factor authentication** — Add a TOTP authenticator app as a second layer so a stolen password alone can't get in

## Credits & Billing (Hosted)

- **Wallet balance** — See how many AI credits you have left before you run out mid-workflow
- **Transaction ledger** — Every credit purchase and deduction is logged so you can see exactly where credits went
- **Per-request usage logs** — See the token count and cost of every individual agent call, not just totals

## Settings

- **Container defaults** — Set org-wide defaults for log tail length, restart policy, stop timeout, and auto-prune behavior so every container starts with consistent behavior
- **Network tuning** — Adjust WebSocket reconnect attempts, API retry counts, and response caching to match your network environment
- **Troubleshooting mode** — Enable debug output and richer API error details in the UI when diagnosing issues

## Self-Host

- **One-liner install** — Get the full platform running on your own server with a single `curl` command, owning all your data
- **Host CLI** — The `nixopus` CLI on the server handles status checks, log tailing, restarts, config changes, domain management, port changes, database backups, and full uninstall — all without editing compose files by hand
- **Guided onboarding** — A first-run flow walks new users through connecting a machine and deploying their first app
- **Self-update** — Apply platform updates from the dashboard without touching the server
- **Feature flags** — Enable or disable individual platform capabilities per org without editing config files
