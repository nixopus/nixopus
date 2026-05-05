# Nixopus Self-Host Installer

One-line installer for self-hosting Nixopus.

```bash
curl -fsSL install.nixopus.com | sudo bash
```

> **Nixopus is in active development.** Self-hosted images are pulled as `latest` and are not pinned to a specific version. Use in production workloads at your own risk.

## When to self-host

**Self-host** when you need full control over your data, have specific compliance requirements, or want to run on your own infrastructure. Great for indie hackers and developers experimenting with hobby projects on their own VPS.

**Use managed** for production workloads. The managed version includes security patches handled by the Nixopus team, pinned and tested image versions, automatic backups, and priority support — so you can focus on shipping instead of maintaining infrastructure.

**Just want to try Nixopus?** Skip the self-host setup entirely. Sign up at [dashboard.nixopus.com](https://dashboard.nixopus.com) and get free allocated machine resources on signup to explore the platform — including agentic AI assistance for deployments. No VPS required.

## Requirements

Nixopus is a deployment platform that manages Docker, binds ports 80/443, and SSH-es into the host. **Use a fresh, dedicated VPS** — not a server already running other production services. Shared servers will have port conflicts, permission issues, and risk interfering with existing workloads.

- **Server:** Fresh VPS from any cloud provider (Hetzner, DigitalOcean, AWS, etc.)
- **Arch:** x86_64 (amd64) or aarch64 (arm64)
- **RAM:** 1 GB minimum (2 GB+ recommended)
- **Disk:** 2 GB free minimum
- **Access:** Root (the installer must run as root)
- **Docker:** Installed automatically if not present (Docker Engine + Compose V2)

### What the installer modifies on your system

The installer will ask for confirmation in interactive mode before proceeding. Here is everything it touches outside of `$NIXOPUS_HOME` (`/opt/nixopus` by default):

| Change | Path | Notes |
|---|---|---|
| Installs prereqs if missing | System packages | `curl`, `openssl`, `openssh-client` via apt/dnf/apk |
| Installs Docker if missing | System packages | Docker Engine + Compose V2, enabled on boot |
| Management CLI | `/usr/local/bin/nixopus` | Overwritten on each install |
| SSH public key | `~/.ssh/authorized_keys` | Appended once (skips if already present). Required for deployments via SSH. |

Everything else (config, compose files, SSH keys, Caddyfile) is contained in `$NIXOPUS_HOME`.

### Tested distributions

These are tested in CI on every release:

| Distribution | Version |
|---|---|
| Ubuntu | 22.04, 24.04 |
| Debian | 12 |
| Rocky Linux | 9 |
| Alpine | 3.20 |

### Should also work

The installer has support paths for these but they are not tested in CI:

| Distribution | Notes |
|---|---|
| Alma Linux | Uses the same install path as Rocky |
| CentOS / RHEL | Uses the same install path as Rocky |
| Fedora | Uses `dnf`, same as Rocky/Alma |

Other Linux distributions may work if Docker and Compose V2 are already installed. The installer requires `/etc/os-release` to be present.

## Configuration

All parameters are optional. Pass them as environment variables to the installer.

```bash
curl -fsSL install.nixopus.com | sudo DOMAIN=panel.example.com ADMIN_EMAIL=admin@example.com bash
```

> **Note:** Variables must be placed after `sudo`, not before `curl`. Placing them before `curl` sets them for the download process, not the installer.

| Variable | Default | Description |
|---|---|---|
| `DOMAIN` | *(empty — IP mode)* | Domain for automatic HTTPS |
| `HOST_IP` | *(auto-detected)* | Public IP of the server |
| `CADDY_HTTP_PORT` | `80` | HTTP port |
| `CADDY_HTTPS_PORT` | `443` | HTTPS port |
| `ADMIN_EMAIL` | *(empty)* | Admin account email. When set, the installer pre-creates the admin via the auth API after services start. |
| `ADMIN_PASSWORD` | *(auto-generated)* | Admin password. Used only when `ADMIN_EMAIL` is set. Auto-generated to satisfy the password rules below if omitted; printed in the install summary and persisted to `.env` (mode `600`). |
| `ADMIN_BOOTSTRAP_TIMEOUT` | `60` | Seconds the installer waits for the admin sign-up call to succeed. Retry later with `nixopus admin-bootstrap`. |
| `SSH_HOST` | `$HOST_IP` | SSH host the API connects to |
| `SSH_PORT` | `22` | SSH port (auto-detected from sshd_config if non-standard) |
| `SSH_USER` | `root` | SSH user |
| `DB_PASSWORD` | *(random)* | Postgres password |
| `REDIS_PASSWORD` | *(random)* | Redis password |
| `DATABASE_URL` | `postgres://nixopus:$DB_PASSWORD@nixopus-db:5432/nixopus` | Full DB connection string. Set to an external URL to skip the bundled DB |
| `REDIS_URL` | `redis://default:$REDIS_PASSWORD@nixopus-redis:6379` | Full Redis connection string. Set to an external URL to skip bundled Redis |
| `AUTH_SERVICE_SECRET` | *(random)* | Auth service secret |
| `JWT_SECRET` | *(random)* | JWT signing secret |
| `NIXOPUS_HOME` | `/opt/nixopus` | Installation directory |
| `LLM_PROVIDER` | `openrouter` | LLM provider: `openrouter`, `openai`, `anthropic`, `google`, `deepseek`, or `groq` |
| `OPENROUTER_API_KEY` | *(empty)* | OpenRouter API key ([openrouter.ai](https://openrouter.ai)) |
| `OPENAI_API_KEY` | *(empty)* | OpenAI API key |
| `ANTHROPIC_API_KEY` | *(empty)* | Anthropic API key |
| `GOOGLE_GENERATIVE_AI_API_KEY` | *(empty)* | Google AI API key |
| `DEEPSEEK_API_KEY` | *(empty)* | DeepSeek API key |
| `GROQ_API_KEY` | *(empty)* | Groq API key |
| `NIXOPUS_TELEMETRY` | `on` | Set to `off` to disable anonymous install telemetry |
| `DO_NOT_TRACK` | `0` | Set to `1` to disable telemetry ([consented.dev](https://consented.dev)) |
| `LOG_LEVEL` | `debug` | Log level |

## Admin account

When you pass `ADMIN_EMAIL`, the installer creates the initial admin user for you by calling the auth service after the stack is healthy — no need to visit `/register` manually.

```bash
curl -fsSL install.nixopus.com | sudo \
  DOMAIN=panel.example.com \
  ADMIN_EMAIL=admin@example.com \
  ADMIN_PASSWORD='ChangeMe!23' bash
```

If you omit `ADMIN_PASSWORD`, the installer generates a 16-character password that satisfies all rules below and prints it once at the end of the install. The same value is written to `/opt/nixopus/.env` so re-runs preserve it.

### Password rules

The auth service enforces these on every credential, including the auto-generated one:

- At least 8 characters
- At least one uppercase letter (A–Z)
- At least one lowercase letter (a–z)
- At least one digit (0–9)
- At least one special character: `!@#$%^&*(),.?":{}|<>`

### Skipping the bootstrap

If you don't pass `ADMIN_EMAIL`, the installer doesn't create any account. On first visit, the dashboard sends you to `/register` and the first person to sign up becomes the admin.

### Retrying the bootstrap

If the auth service was still warming up when the installer gave up, or you need to recreate the admin, retry without reinstalling:

```bash
sudo nixopus admin-bootstrap
# or override the stored values:
sudo nixopus admin-bootstrap admin@example.com 'NewPass!23'
```

The retry uses the values stored in `.env` by default, treats `USER_ALREADY_EXISTS_USE_ANOTHER_EMAIL` as success, and surfaces the auth service's `code` field on validation failures (`PASSWORD_TOO_SHORT`, `INVALID_ORIGIN`, `EMAIL_PASSWORD_SIGN_UP_DISABLED`, etc.).

## Ports

Nixopus binds the following ports on the host:

| Port | Service | Configurable | Notes |
|---|---|---|---|
| `80` | Caddy (HTTP) | `CADDY_HTTP_PORT` | Required for Let's Encrypt HTTPS challenges |
| `443` | Caddy (HTTPS) | `CADDY_HTTPS_PORT` | TLS termination |
| `2019` | Caddy admin API | No | Bound to `127.0.0.1` only (not exposed externally) |

Internal services (Docker network only, not exposed to host):

| Port | Service |
|---|---|
| `9090` | nixopus-auth |
| `8443` | nixopus-api |
| `7443` | nixopus-view |
| `4090` | nixopus-agent |
| `5432` | nixopus-db (bundled Postgres) |
| `6379` | nixopus-redis (bundled Redis) |

The SSH port on your host (default `22`) must also be accessible from the Docker network — the API connects back to the host via SSH for deployments.

If ports 80/443 are already in use (Apache, Nginx, another container), either stop the conflicting service or install with custom ports:

```bash
curl -fsSL install.nixopus.com | sudo CADDY_HTTP_PORT=8080 CADDY_HTTPS_PORT=8443 bash
```

Use `docker ps --format '{{.Ports}} {{.Names}}'` to find what's using a port.

### Firewall

The installer warns about `ufw` and `firewalld` but does not modify firewall rules. You must open the HTTP/HTTPS ports yourself.

**ufw (Ubuntu/Debian):**

```bash
sudo ufw allow 80/tcp && sudo ufw allow 443/tcp && sudo ufw reload
```

**firewalld (RHEL/Rocky/Alma/Fedora):**

```bash
sudo firewall-cmd --permanent --add-port=80/tcp && sudo firewall-cmd --permanent --add-port=443/tcp && sudo firewall-cmd --reload
```

**iptables (manual):**

```bash
sudo iptables -A INPUT -p tcp --dport 80 -j ACCEPT && sudo iptables -A INPUT -p tcp --dport 443 -j ACCEPT
```

**Cloud providers:** Also open ports in your cloud firewall (AWS Security Groups, GCP Firewall Rules, Azure NSG, etc.). These are separate from the OS-level firewall.

If using custom ports, replace `80`/`443` with your values in all commands above.

## AI Agent

The installer includes an AI agent that assists with deployments, diagnostics, and infrastructure management. It runs as the `nixopus-agent` service.

### LLM Provider

During installation, you'll be prompted to choose an LLM provider. The agent supports:

| Provider | Env Variable | Models |
|---|---|---|
| **OpenRouter** (default) | `OPENROUTER_API_KEY` | Claude, GPT-4, Gemini, and more via one key |
| **OpenAI** | `OPENAI_API_KEY` | GPT-4o, GPT-4o Mini |
| **Anthropic** | `ANTHROPIC_API_KEY` | Claude Sonnet 4, Claude Haiku 3.5 |
| **Google** | `GOOGLE_GENERATIVE_AI_API_KEY` | Gemini 2.5 Flash, Gemini 2.0 Flash |
| **DeepSeek** | `DEEPSEEK_API_KEY` | DeepSeek Chat |
| **Groq** | `GROQ_API_KEY` | Llama 3.3 70B |

You can also set the provider non-interactively:

```bash
curl -fsSL install.nixopus.com | sudo LLM_PROVIDER=anthropic ANTHROPIC_API_KEY=sk-ant-xxxxx bash
```

Or change the provider after installation:

```bash
nixopus config set ANTHROPIC_API_KEY=sk-ant-xxxxx
nixopus config set AGENT_MODEL=anthropic/claude-sonnet-4
nixopus config set AGENT_LIGHT_MODEL=anthropic/claude-haiku-3.5
nixopus restart
```

Users can also switch models per-chat from the model dropdown in the UI.

## Branch Preview (Testing PRs before merge)

Preview images are automatically built for every pull request. This lets you test a PR on a real VPS before merging.

### Testing a PR

When a PR is opened, CI builds `pr-<number>` tagged images and comments install instructions on the PR. Use the one-liner from the comment, or:

```bash
curl -fsSL https://raw.githubusercontent.com/nixopus/nixopus/<branch>/installer/get.sh | \
  sudo bash -s -- --preview <pr-number>
```

This pulls the API and View images tagged `pr-<number>` and uses the installer files from that PR's branch. Auth and Agent images fall back to `:latest` since they live in separate repos.

### Testing a specific branch

```bash
curl -fsSL https://raw.githubusercontent.com/nixopus/nixopus/<branch>/installer/get.sh | \
  sudo bash -s -- --branch <branch-name> --preview <pr-number>
```

### Testing a fork

```bash
curl -fsSL https://raw.githubusercontent.com/<owner>/<repo>/<branch>/installer/get.sh | \
  sudo bash -s -- --fork <owner>/<repo> --branch <branch-name>
```

For forks, you need to build and push images to your own GHCR first (the preview CI only runs on the main repo). You can also override individual images:

```bash
curl -fsSL install.nixopus.com | \
  sudo NIXOPUS_API_IMAGE=ghcr.io/youruser/nixopus-api:my-branch \
       NIXOPUS_VIEW_IMAGE=ghcr.io/youruser/nixopus-view:my-branch \
       bash
```

### Image override variables

| Variable | Default | Description |
|---|---|---|
| `NIXOPUS_API_IMAGE` | `ghcr.io/nixopus/nixopus-api:latest` | API container image |
| `NIXOPUS_VIEW_IMAGE` | `ghcr.io/nixopus/nixopus-view:latest` | View container image |
| `NIXOPUS_AUTH_IMAGE` | `ghcr.io/nixopus/auth:latest` | Auth container image |
| `NIXOPUS_AGENT_IMAGE` | `ghcr.io/nixopus/agent:latest` | Agent container image |

### Preview image cleanup

Preview images are automatically deleted when the PR closes (merged or not). A weekly sweep also removes any orphaned preview images older than 7 days. You can also trigger cleanup manually from the Actions tab.

### Reverting to stable

Re-run the installer without `--preview` to go back to the latest stable images:

```bash
curl -fsSL install.nixopus.com | sudo bash
```

## HTTPS

When you provide a `DOMAIN`, Caddy automatically obtains and renews TLS certificates from Let's Encrypt. For this to work:

1. **DNS A record** must point to your server's public IP *before* installing.
2. **Port 80 must be open** — Let's Encrypt uses HTTP-01 challenges on port 80, even if you only serve on 443.
3. **Not behind a proxy** — If using Cloudflare, set to "DNS only" (grey cloud) during initial setup so the challenge can reach your server directly. You can re-enable proxying after the first certificate is issued.

Without a `DOMAIN`, Nixopus runs in IP mode over plain HTTP.

## Management CLI

After installation, the `nixopus` command is available. All commands require root (`sudo`):

```bash
sudo nixopus status
```

| Command | Description |
|---|---|
| `nixopus status` | Show service health |
| `nixopus logs [service]` | Tail logs (services: `nixopus-api`, `nixopus-auth`, `nixopus-view`, `nixopus-caddy`, `nixopus-agent`, `nixopus-db`, `nixopus-redis`) |
| `nixopus update` | Pull latest images and restart |
| `nixopus restart [service]` | Restart all or a specific service |
| `nixopus stop` | Stop all services |
| `nixopus config` | Show current configuration |
| `nixopus config set KEY=VALUE` | Update a config value (restart required) |
| `nixopus domain add <domain>` | Switch to domain-based HTTPS (ensure DNS is configured first) |
| `nixopus domain remove` | Switch back to IP-based HTTP |
| `nixopus ip set <ip>` | Change host IP |
| `nixopus port set <http\|https\|ssh> <port>` | Change a port |
| `nixopus backup` | Backup database and config |
| `nixopus admin-bootstrap [email] [password]` | Create the initial admin via the auth API (uses `ADMIN_EMAIL`/`ADMIN_PASSWORD` from `.env` unless overridden) |
| `nixopus uninstall` | Remove containers (keeps data) |
| `nixopus uninstall --purge` | Remove everything including data |

## Updates

```bash
nixopus update
```

This pulls the latest Docker images and restarts services. It does **not** update Docker Compose files, the Caddyfile, or the CLI itself.

For a full upgrade (new compose files, CLI, etc.), re-run the installer — secrets and config are preserved automatically:

```bash
curl -fsSL install.nixopus.com | sudo bash
```

## Backup & Restore

### Backup

```bash
nixopus backup
```

Saves a database dump and a copy of `.env` to `/opt/nixopus/backups/<timestamp>/`.

### Restore

```bash
# 1. Restore the .env
cp /opt/nixopus/backups/<timestamp>/env.bak /opt/nixopus/.env

# 2. Restart services so the DB container is running
nixopus restart

# 3. Restore the database dump
docker exec -i nixopus-db psql -U nixopus nixopus < /opt/nixopus/backups/<timestamp>/db.sql
```

## Troubleshooting

### Reporting install or runtime issues

The installer writes a full transcript to `/opt/nixopus/install.log` (and shows the path on failure). Attach that file when asking for help, or generate a paste-friendly bundle with **`sudo nixopus report`** (redacts the installer log tail and `.env`; compose logs may still contain secrets — review before sharing):

```bash
sudo nixopus report
# or save to a file:
sudo nixopus report > /tmp/nixopus-report.txt
```

Installer logs may still contain API keys or passwords echoed during the run — review before sharing.

### Services fail to start after reinstall

**Symptom:** `nixopus-auth` or `nixopus-api` crash-loop with database authentication errors.

**Cause:** Containers were removed but Docker volumes still hold the old database with the original password. The reinstall generated new credentials that don't match.

**Fix:**

```bash
# Option 1: Check the backup for the original password
cat /opt/nixopus/.env.bak | grep DB_PASSWORD
# Then reinstall with it
curl -fsSL install.nixopus.com | sudo DB_PASSWORD=<original_password> bash

# Option 2: Start fresh (destroys all data)
docker volume rm $(docker volume ls -q --filter name=nixopus)
curl -fsSL install.nixopus.com | sudo bash
```

### Health check timeout during install

**Symptom:** Installer hangs at "Waiting for services to start..." and times out after 180s.

**Fix:** Check which service is unhealthy:

```bash
nixopus status
nixopus logs
```

Common causes: port conflict (see [Ports](#ports)), DNS not configured (see [HTTPS](#https)), or insufficient resources (see [Requirements](#requirements)).

### Admin bootstrap failed during install

**Symptom:** Install summary shows `Could not auto-create admin (HTTP …)` or `Auth service rejected request: …`.

The installer creates the admin via the auth service after the stack starts. If the auth service is slow, the origin is misconfigured, or the credentials don't pass validation, the bootstrap is skipped — but the install itself succeeds.

**Fix:** Retry without reinstalling.

```bash
sudo nixopus admin-bootstrap
```

Common error codes printed by the auth service:

| Code | Meaning | Fix |
|---|---|---|
| `USER_ALREADY_EXISTS_USE_ANOTHER_EMAIL` | Admin already exists | Nothing — log in with the existing credentials |
| `EMAIL_PASSWORD_SIGN_UP_DISABLED` | Email/password sign-up turned off in the auth service | Enable it in the auth service config or seed via DB |
| `INVALID_ORIGIN` / `MISSING_OR_NULL_ORIGIN` | Origin header rejected | Check `ALLOWED_ORIGIN` matches your `BASE_URL` in `/opt/nixopus/.env` |
| `PASSWORD_TOO_SHORT` / `PASSWORD_TOO_LONG` / `INVALID_PASSWORD` | Password rules failed | Pick a password matching the [rules](#password-rules) and re-run with `nixopus admin-bootstrap <email> <password>` |
| `INVALID_EMAIL` | Email malformed | Re-run with a valid email |

If the auth service is just slow to come up, increase the timeout:

```bash
sudo ADMIN_BOOTSTRAP_TIMEOUT=120 nixopus admin-bootstrap
```

### Lost the admin password

The auto-generated admin password is stored in `/opt/nixopus/.env` as `ADMIN_PASSWORD`:

```bash
sudo grep '^ADMIN_PASSWORD=' /opt/nixopus/.env
```

If that's missing or the account is locked out, recreate the admin by deleting the existing credential row and re-running the bootstrap:

```bash
docker exec -it nixopus-db psql -U nixopus -d nixopus -c \
  "DELETE FROM account WHERE user_id = (SELECT id FROM \"user\" WHERE email='admin@example.com');
   DELETE FROM \"user\" WHERE email='admin@example.com';"

sudo nixopus admin-bootstrap admin@example.com 'NewPass!23'
```

> **Warning:** Deleting the user row drops anything tied to that user (org memberships, API keys, audit ownership). Only do this if you understand the impact.

### Cannot access the dashboard

**Symptom:** Browser shows connection refused or timeout.

**Fix:** Verify services are running with `nixopus status`, check ports with `nixopus port`, and ensure firewall rules are in place (see [Firewall](#firewall)). If behind a cloud provider, also check the security group / firewall rules in your cloud console.

### Deployments failing (SSH connection errors)

**Symptom:** Deploys fail with SSH connection refused or permission denied.

The API container SSH-es back into the host to manage deployments. This requires:

1. **SSH service running** on the host on the configured port (`SSH_PORT`, default `22`).
2. **The Nixopus SSH public key** must be in the host's `~/.ssh/authorized_keys` (the installer adds this automatically, but it can be lost if `authorized_keys` is regenerated).
3. **The host must be reachable** from the Docker network.

**Fix:**

```bash
# Verify the key is in authorized_keys
grep -q "$(cat /opt/nixopus/ssh/id_rsa.pub)" ~/.ssh/authorized_keys || \
  cat /opt/nixopus/ssh/id_rsa.pub >> ~/.ssh/authorized_keys

# Test SSH from the API container
docker exec nixopus-api ssh -i /etc/nixopus/ssh/id_rsa -p ${SSH_PORT:-22} -o StrictHostKeyChecking=no ${SSH_USER:-root}@${SSH_HOST} echo ok
```

### SELinux blocking services (RHEL/Rocky/Alma)

**Symptom:** Containers fail to start or can't access mounted volumes on RHEL-based systems.

**Fix:**

```bash
# Check if SELinux is enforcing
getenforce

# Option 1: Allow Docker to access the volume (recommended)
chcon -Rt svirt_sandbox_file_t /opt/nixopus

# Option 2: Set SELinux to permissive (less secure)
setenforce 0
# To persist: edit /etc/selinux/config and set SELINUX=permissive
```

### Disk space running out

Container logs are capped at 10MB per service (30MB with rotation). If disk still fills up, check:

```bash
# Docker disk usage
docker system df

# Clean unused images and build cache
docker system prune -f

# Check Postgres data size
docker exec nixopus-db psql -U nixopus -c "SELECT pg_size_pretty(pg_database_size('nixopus'));"
```

### Services not starting after server reboot

Services use `restart: unless-stopped`, so they start automatically with Docker. If they don't:

```bash
# Ensure Docker starts on boot
sudo systemctl enable docker

# Start services manually
nixopus restart
```

### Viewing secrets

```bash
sudo cat /opt/nixopus/.env
```

### Resetting to a clean state

```bash
nixopus uninstall --purge
curl -fsSL install.nixopus.com | sudo bash
```

## Telemetry

The installer sends anonymous telemetry events to help us understand adoption and improve the install experience. **No personal data is collected** — only OS name, CPU architecture, installer version, and install duration.

Three events are sent during installation:

| Event | When |
|---|---|
| `install_started` | Installer begins |
| `install_success` | Installation completes successfully |
| `install_failure` | Installation fails (includes a 200-char error snippet) |

IP addresses are hashed with a server-side salt before storage; raw IPs are never persisted.

### Opting out

Disable telemetry with either method before running the installer:

```bash
# Method 1
export NIXOPUS_TELEMETRY=off

# Method 2 (DO_NOT_TRACK standard)
export DO_NOT_TRACK=1
```

### Installer CLI flags

The installer accepts flags when invoked via `bash -s --`:

| Flag | Description |
|---|---|
| `--preview <pr>` | Use preview images built for a PR |
| `--branch <name>` | Fetch installer files from a specific branch |
| `--fork <owner/repo>` | Fetch installer files and images from a fork |
| `--help` | Show installer flags help |

## Contents

- `get.sh` - Installer script
- `nixopus.sh` - Management CLI (installed to `/usr/local/bin/nixopus`)
- `selfhost/` - Docker Compose files (base, db, redis, agent overlays)
- `test/` - Cross-distro test suite
