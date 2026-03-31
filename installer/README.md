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

All parameters are optional. Pass them as environment variables before the install command.

```bash
DOMAIN=panel.example.com ADMIN_EMAIL=admin@example.com curl -fsSL install.nixopus.com | sudo bash
```

| Variable | Default | Description |
|---|---|---|
| `DOMAIN` | *(empty — IP mode)* | Domain for automatic HTTPS |
| `HOST_IP` | *(auto-detected)* | Public IP of the server |
| `CADDY_HTTP_PORT` | `80` | HTTP port |
| `CADDY_HTTPS_PORT` | `443` | HTTPS port |
| `ADMIN_EMAIL` | *(empty)* | Admin account email |
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
| `OPENROUTER_API_KEY` | *(empty — Ollama)* | OpenRouter API key for cloud LLM models. Leave blank to use Ollama (local). |
| `NIXOPUS_TELEMETRY` | `on` | Set to `off` to disable anonymous telemetry |
| `LOG_LEVEL` | `debug` | Log level |

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
| `11434` | nixopus-ollama (when using local LLM) |

The SSH port on your host (default `22`) must also be accessible from the Docker network — the API connects back to the host via SSH for deployments.

If ports 80/443 are already in use (Apache, Nginx, another container), either stop the conflicting service or install with custom ports:

```bash
CADDY_HTTP_PORT=8080 CADDY_HTTPS_PORT=8443 curl -fsSL install.nixopus.com | sudo bash
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

During installation, you'll be prompted for an **OpenRouter API key**. This determines the LLM backend:

- **Blank (default):** The agent runs with **Ollama** for fully local inference. An Ollama container is started automatically — no API keys or external services needed.
- **API key provided:** The agent uses **OpenRouter** to access cloud models (Claude, GPT-4o, etc.).

You can change the provider after installation:

```bash
# Switch to OpenRouter
nixopus config set OPENROUTER_API_KEY=sk-or-v1-xxxxx
nixopus config set USE_OLLAMA=false
nixopus restart

# Switch to Ollama
nixopus config set OPENROUTER_API_KEY=
nixopus config set USE_OLLAMA=true
nixopus restart
```

### Resource requirements with Ollama

When using Ollama (local LLM), the server needs additional resources:

- **RAM:** 4 GB minimum (8 GB+ recommended for better models)
- **Disk:** 5 GB+ additional for model weights

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
| `nixopus logs [service]` | Tail logs (services: `nixopus-api`, `nixopus-auth`, `nixopus-view`, `nixopus-caddy`, `nixopus-agent`, `nixopus-ollama`, `nixopus-db`, `nixopus-redis`) |
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

### Services fail to start after reinstall

**Symptom:** `nixopus-auth` or `nixopus-api` crash-loop with database authentication errors.

**Cause:** Containers were removed but Docker volumes still hold the old database with the original password. The reinstall generated new credentials that don't match.

**Fix:**

```bash
# Option 1: Check the backup for the original password
cat /opt/nixopus/.env.bak | grep DB_PASSWORD
# Then reinstall with it
DB_PASSWORD=<original_password> curl -fsSL install.nixopus.com | sudo bash

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

## Contents

- `get.sh` - Installer script
- `nixopus.sh` - Management CLI (installed to `/usr/local/bin/nixopus`)
- `selfhost/` - Docker Compose files (base, db, redis, agent, ollama overlays)
- `test/` - Cross-distro test suite
