#!/bin/bash
set -euo pipefail

NIXOPUS_VERSION="0.3.0"
NIXOPUS_HOME="${NIXOPUS_HOME:-/opt/nixopus}"
TELEMETRY_URL="${NIXOPUS_TELEMETRY_URL:-https://nixopus-api.nixopus.com/api/cli/installations}"
REPO_RAW="${NIXOPUS_REPO_RAW:-https://raw.githubusercontent.com/nixopus/nixopus/master/installer}"
INSTALL_START=$(date +%s)

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

log_info()  { echo -e "  ${BLUE}INFO${NC}  $1"; }
log_ok()    { echo -e "  ${GREEN} OK ${NC}  $1"; }
log_warn()  { echo -e "  ${YELLOW}WARN${NC}  $1"; }
log_error() { echo -e "  ${RED}FAIL${NC}  $1"; }
log_step()  { echo -e "\n${CYAN}${BOLD}[$1/$TOTAL_STEPS]${NC} ${BOLD}$2${NC}"; }

TOTAL_STEPS=6

fail() {
    log_error "$1"
    send_telemetry "install_failure" "$1" || true
    exit 1
}

# ── Step 1: Requirements ─────────────────────────────────────────────────────

check_root() {
    if [ "$(id -u)" -ne 0 ]; then
        fail "Must be run as root. Use: curl -fsSL ... | sudo bash"
    fi
}

detect_os() {
    if [ -f /etc/os-release ]; then
        # shellcheck disable=SC1091
        . /etc/os-release
        OS_ID="${ID:-unknown}"
        OS_NAME="${PRETTY_NAME:-${ID:-unknown}}"
    elif [ -f /etc/redhat-release ]; then
        OS_ID="rhel"
        OS_NAME=$(cat /etc/redhat-release)
    else
        fail "Unsupported OS: /etc/os-release not found"
    fi

    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) fail "Unsupported architecture: $ARCH" ;;
    esac

    log_ok "$OS_NAME ($ARCH)"
}

install_prereqs() {
    local need_install=false
    for cmd in curl openssl ssh-keygen; do
        command -v "$cmd" &>/dev/null || need_install=true
    done
    $need_install || return 0

    case "$OS_ID" in
        ubuntu|debian)
            apt-get update -qq >/dev/null 2>&1
            apt-get install -y -qq curl openssl openssh-client >/dev/null 2>&1
            ;;
        rocky|alma|centos|rhel|fedora)
            dnf install -y -q --allowerasing curl openssl openssh-clients >/dev/null 2>&1 \
                || yum install -y -q curl openssl openssh-clients >/dev/null 2>&1
            ;;
        alpine)
            apk add --no-cache curl openssl openssh-keygen >/dev/null 2>&1
            ;;
        *)
            for cmd in curl openssl ssh-keygen; do
                command -v "$cmd" &>/dev/null || fail "Required command not found: $cmd"
            done
            ;;
    esac
}

# ── Step 2: Docker ───────────────────────────────────────────────────────────

install_docker() {
    if command -v docker &>/dev/null; then
        log_ok "Docker $(docker --version | awk '{print $3}' | tr -d ',')"
    else
        log_info "Installing Docker..."
        case "$OS_ID" in
            alpine)
                apk add --no-cache docker docker-cli-compose >/dev/null 2>&1
                rc-update add docker default 2>/dev/null || true
                service docker start 2>/dev/null || true
                ;;
            rocky|alma|centos|rhel)
                dnf install -y -q dnf-plugins-core >/dev/null 2>&1 || yum install -y -q yum-utils >/dev/null 2>&1
                dnf config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo >/dev/null 2>&1 \
                    || yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo >/dev/null 2>&1
                dnf install -y -q docker-ce docker-ce-cli containerd.io docker-compose-plugin >/dev/null 2>&1 \
                    || yum install -y -q docker-ce docker-ce-cli containerd.io docker-compose-plugin >/dev/null 2>&1 \
                    || fail "Docker installation failed. Install manually: https://docs.docker.com/engine/install/"
                ;;
            *)
                curl -fsSL https://get.docker.com | sh >/dev/null 2>&1 \
                    || fail "Docker installation failed. Install manually: https://docs.docker.com/engine/install/"
                ;;
        esac
        systemctl enable docker 2>/dev/null || true
        systemctl start docker 2>/dev/null || true
        log_ok "Docker installed"
    fi

    if ! docker compose version &>/dev/null; then
        fail "Docker Compose v2 not found. Install: https://docs.docker.com/compose/install/"
    fi
}

# ── Step 3: Configuration ────────────────────────────────────────────────────

is_ipv6() { [[ "$1" == *:* ]]; }

format_ip_for_url() {
    local ip="$1"
    if is_ipv6 "$ip"; then
        echo "[$ip]"
    else
        echo "$ip"
    fi
}

detect_ip() {
    local ip=""
    for svc in "https://api64.ipify.org" "https://ifconfig.me" "https://icanhazip.com"; do
        ip=$(curl -4 -fsSL --connect-timeout 5 "$svc" 2>/dev/null | tr -d '[:space:]') && [ -n "$ip" ] && break
    done
    if [ -z "$ip" ]; then
        for svc in "https://api64.ipify.org" "https://ifconfig.me" "https://icanhazip.com"; do
            ip=$(curl -6 -fsSL --connect-timeout 5 "$svc" 2>/dev/null | tr -d '[:space:]') && [ -n "$ip" ] && break
        done
    fi
    if [ -z "$ip" ]; then
        ip=$(hostname -I 2>/dev/null | awk '{print $1}')
    fi
    echo "$ip"
}

is_private_ip() {
    local ip="$1"
    [[ "$ip" =~ ^10\. ]] || [[ "$ip" =~ ^172\.(1[6-9]|2[0-9]|3[01])\. ]] || \
    [[ "$ip" =~ ^192\.168\. ]] || [[ "$ip" =~ ^fc ]] || [[ "$ip" =~ ^fd ]]
}

check_port_available() {
    local port="$1"
    if command -v ss &>/dev/null; then
        ss -tlnp 2>/dev/null | grep -q ":${port} " && return 1
    elif command -v netstat &>/dev/null; then
        netstat -tlnp 2>/dev/null | grep -q ":${port} " && return 1
    fi
    return 0
}

check_resources() {
    local mem_kb
    mem_kb=$(grep MemTotal /proc/meminfo 2>/dev/null | awk '{print $2}') || return 0
    local mem_mb=$((mem_kb / 1024))
    if [ "$mem_mb" -lt 1024 ]; then
        log_warn "Low memory: ${mem_mb}MB detected (1024MB+ recommended)"
    fi

    local disk_avail_kb
    disk_avail_kb=$(df "$NIXOPUS_HOME" 2>/dev/null | tail -1 | awk '{print $4}') || return 0
    local disk_avail_mb=$((disk_avail_kb / 1024))
    if [ "$disk_avail_mb" -lt 2048 ]; then
        log_warn "Low disk space: ${disk_avail_mb}MB available (2048MB+ recommended)"
    fi
}

check_firewall() {
    local http_port="${CADDY_HTTP_PORT:-80}"
    local https_port="${CADDY_HTTPS_PORT:-443}"

    if command -v ufw &>/dev/null && ufw status 2>/dev/null | grep -q "active"; then
        local needs_open=false
        ufw status 2>/dev/null | grep -q "$http_port" || needs_open=true
        if [ "$needs_open" = true ]; then
            log_warn "ufw is active — run: ufw allow ${http_port}/tcp && ufw allow ${https_port}/tcp"
        fi
    elif command -v firewall-cmd &>/dev/null && firewall-cmd --state 2>/dev/null | grep -q "running"; then
        if ! firewall-cmd --list-ports 2>/dev/null | grep -q "${http_port}/tcp"; then
            log_warn "firewalld is active — run: firewall-cmd --permanent --add-port={${http_port},${https_port}}/tcp && firewall-cmd --reload"
        fi
    fi
}

gen_secret() { openssl rand -hex 32; }

prompt_if_tty() {
    local var_name="$1" prompt="$2" default="${3:-}"
    local current_val="${!var_name:-}"
    if [ -n "$current_val" ]; then return; fi

    if [ -t 0 ]; then
        local input=""
        if [ -n "$default" ]; then
            read -rp "  $prompt [$default]: " input
        else
            read -rp "  $prompt: " input
        fi
        eval "$var_name=\"\${input:-$default}\""
    else
        eval "$var_name=\"$default\""
    fi
}

load_existing_config() {
    if [ -f "$NIXOPUS_HOME/.env" ]; then
        log_info "Existing installation detected at $NIXOPUS_HOME"
        log_info "Preserving secrets from previous install"
        local saved_version="$NIXOPUS_VERSION"
        local saved_home="$NIXOPUS_HOME"
        set -a
        # shellcheck disable=SC1091
        . "$NIXOPUS_HOME/.env"
        set +a
        NIXOPUS_VERSION="$saved_version"
        NIXOPUS_HOME="$saved_home"
        # Re-detect network config since the server IP may have changed
        unset HOST_IP SSH_HOST
        return 0
    fi

    if command -v docker &>/dev/null; then
        local orphaned_volumes=""
        orphaned_volumes=$(docker volume ls --quiet --filter "name=nixopus" 2>/dev/null | grep "nixopus-db-data\|nixopus-redis-data" || true)
        if [ -n "$orphaned_volumes" ]; then
            log_warn "Found data volumes without config:"
            echo "$orphaned_volumes" | while read -r vol; do log_warn "  $vol"; done
            log_warn "Containers were removed without 'nixopus uninstall --purge'"
            log_warn "New credentials will NOT match the existing database"
            if [ -t 0 ]; then
                echo ""
                echo -e "  ${BOLD}Options:${NC}"
                echo "  1) Remove old volumes and start fresh:"
                echo "     docker volume rm \$(docker volume ls -q --filter name=nixopus)"
                echo "  2) Provide the original DB password:"
                echo "     DB_PASSWORD=<old_password> sudo bash get.sh"
                echo ""
                read -rp "  Continue with new credentials anyway? [y/N] " confirm
                if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
                    fail "Aborted. Remove orphaned volumes or provide original credentials."
                fi
            else
                log_warn "Non-interactive: proceeding, but services may fail to authenticate"
                log_warn "Fix: remove volumes and reinstall"
            fi
        fi
    fi
}

gather_config() {
    check_resources

    if [ -t 0 ] && [ -z "${DOMAIN:-}" ]; then
        echo ""
        echo -e "  ${BOLD}Domain Setup${NC}"
        echo -e "  ${DIM}Enter a domain for automatic HTTPS, or leave empty for IP-based HTTP.${NC}"
    fi
    prompt_if_tty DOMAIN "Domain (e.g. nixopus.example.com)" ""

    HOST_IP="${HOST_IP:-$(detect_ip)}"
    if [ -z "$HOST_IP" ] && [ -z "${DOMAIN:-}" ]; then
        fail "Cannot detect public IP and no domain set. Pass HOST_IP=x.x.x.x"
    fi

    if [ -n "${HOST_IP:-}" ] && is_private_ip "$HOST_IP" && [ -t 0 ]; then
        log_warn "Detected private IP: $HOST_IP (behind NAT?)"
        prompt_if_tty HOST_IP "Public IP (or keep private for LAN-only)" "$HOST_IP"
    fi

    CADDY_HTTP_PORT="${CADDY_HTTP_PORT:-80}"
    CADDY_HTTPS_PORT="${CADDY_HTTPS_PORT:-443}"

    if ! check_port_available "$CADDY_HTTP_PORT"; then
        if [ -t 0 ]; then
            log_warn "Port $CADDY_HTTP_PORT is in use"
            prompt_if_tty CADDY_HTTP_PORT "HTTP port" "8080"
        else
            fail "Port $CADDY_HTTP_PORT is in use. Set CADDY_HTTP_PORT=<port>"
        fi
    fi
    if ! check_port_available "$CADDY_HTTPS_PORT"; then
        if [ -t 0 ]; then
            log_warn "Port $CADDY_HTTPS_PORT is in use"
            prompt_if_tty CADDY_HTTPS_PORT "HTTPS port" "8443"
        else
            fail "Port $CADDY_HTTPS_PORT is in use. Set CADDY_HTTPS_PORT=<port>"
        fi
    fi

    local ip_for_url
    ip_for_url=$(format_ip_for_url "${HOST_IP:-127.0.0.1}")

    if [ -n "${DOMAIN:-}" ]; then
        SITE_ADDRESS="$DOMAIN"
        BASE_URL="https://$DOMAIN"
        AUTH_SECURE_COOKIES="true"
        AUTH_COOKIE_DOMAIN=".$(echo "$DOMAIN" | awk -F. '{print $(NF-1)"."$NF}')"
    else
        SITE_ADDRESS=":80"
        if [ "$CADDY_HTTP_PORT" = "80" ]; then
            BASE_URL="http://$ip_for_url"
        else
            BASE_URL="http://$ip_for_url:$CADDY_HTTP_PORT"
        fi
        AUTH_SECURE_COOKIES="false"
        AUTH_COOKIE_DOMAIN=""
    fi

    ALLOWED_ORIGIN="$BASE_URL"
    API_URL="$BASE_URL/api"
    AUTH_PUBLIC_URL="$BASE_URL"

    if [ -t 0 ] && [ -z "${ADMIN_EMAIL:-}" ]; then
        echo ""
        echo -e "  ${BOLD}Admin Account${NC}"
    fi
    prompt_if_tty ADMIN_EMAIL "Admin email" ""

    SSH_HOST="${SSH_HOST:-${HOST_IP:-host.docker.internal}}"
    SSH_USER="${SSH_USER:-root}"

    if [ -z "${SSH_PORT:-}" ] && [ -t 0 ]; then
        local current_ssh_port
        current_ssh_port=$(grep -E "^Port " /etc/ssh/sshd_config 2>/dev/null | awk '{print $2}') || true
        if [ -n "$current_ssh_port" ] && [ "$current_ssh_port" != "22" ]; then
            log_warn "SSH running on port $current_ssh_port (not 22)"
            SSH_PORT="$current_ssh_port"
        fi
    fi
    SSH_PORT="${SSH_PORT:-22}"

    DB_PASSWORD="${DB_PASSWORD:-$(gen_secret)}"
    REDIS_PASSWORD="${REDIS_PASSWORD:-$(gen_secret)}"
    DATABASE_URL="${DATABASE_URL:-postgres://nixopus:${DB_PASSWORD}@nixopus-db:5432/nixopus?sslmode=disable}"
    REDIS_URL="${REDIS_URL:-redis://default:${REDIS_PASSWORD}@nixopus-redis:6379}"

    USE_BUNDLED_DB=true
    USE_BUNDLED_REDIS=true
    [[ "$DATABASE_URL" != *"nixopus-db"* ]] && USE_BUNDLED_DB=false
    [[ "$REDIS_URL" != *"nixopus-redis"* ]] && USE_BUNDLED_REDIS=false
    AUTH_SERVICE_SECRET="${AUTH_SERVICE_SECRET:-$(gen_secret)}"
    JWT_SECRET="${JWT_SECRET:-$(gen_secret)}"

    # ── Agent / LLM configuration ──
    USE_AGENT="${USE_AGENT:-true}"
    if [ -t 0 ] && [ -z "${OPENROUTER_API_KEY:-}" ]; then
        echo ""
        echo -e "  ${BOLD}AI Agent${NC}"
        echo -e "  ${DIM}The agent uses an LLM for deployments and diagnostics.${NC}"
        echo -e "  ${DIM}Leave blank to use Ollama (local, no API key needed).${NC}"
        prompt_if_tty OPENROUTER_API_KEY "OpenRouter API key (or blank for Ollama)" ""
    fi

    USE_OLLAMA=false
    if [ -z "${OPENROUTER_API_KEY:-}" ]; then
        USE_OLLAMA=true
        OLLAMA_BASE_URL="${OLLAMA_BASE_URL:-http://nixopus-ollama:11434}"
    fi

    check_firewall

    log_ok "Configuration ready (${DOMAIN:-IP: $HOST_IP})"
}

# ── Step 4: Files ────────────────────────────────────────────────────────────

setup_directories() {
    mkdir -p "$NIXOPUS_HOME"/{ssh,configs,caddy}
    chmod 755 "$NIXOPUS_HOME"
}

setup_ssh() {
    local key_path="$NIXOPUS_HOME/ssh/id_rsa"
    if [ -f "$key_path" ]; then
        chmod 755 "$NIXOPUS_HOME/ssh"
        chmod 644 "$key_path"
        log_ok "SSH key exists"
    else
        ssh-keygen -t rsa -b 4096 -f "$key_path" -N "" -q
        chmod 755 "$NIXOPUS_HOME/ssh"
        chmod 644 "$key_path"
        chmod 644 "$key_path.pub"
        log_ok "SSH key generated"
    fi

    local auth_keys="${HOME}/.ssh/authorized_keys"
    local pubkey
    pubkey=$(cat "$key_path.pub")

    mkdir -p "$(dirname "$auth_keys")"
    touch "$auth_keys"
    chmod 600 "$auth_keys"

    if grep -qF "$pubkey" "$auth_keys" 2>/dev/null; then
        log_ok "SSH key already in authorized_keys"
        return
    fi

    echo "$pubkey" >> "$auth_keys"
    log_ok "SSH public key added to authorized_keys"
}

write_env() {
    if [ -f "$NIXOPUS_HOME/.env" ]; then
        cp "$NIXOPUS_HOME/.env" "$NIXOPUS_HOME/.env.bak"
        log_info "Previous .env backed up to $NIXOPUS_HOME/.env.bak"
    fi
    cat > "$NIXOPUS_HOME/.env" << EOF
NIXOPUS_VERSION=${NIXOPUS_VERSION}
NIXOPUS_HOME=${NIXOPUS_HOME}

DOMAIN=${DOMAIN:-}
SITE_ADDRESS=${SITE_ADDRESS}
HOST_IP=${HOST_IP:-}
CADDY_HTTP_PORT=${CADDY_HTTP_PORT}
CADDY_HTTPS_PORT=${CADDY_HTTPS_PORT}

SSH_HOST=${SSH_HOST}
SSH_PORT=${SSH_PORT}
SSH_USER=${SSH_USER}

DATABASE_URL=${DATABASE_URL}
REDIS_URL=${REDIS_URL}
DB_PASSWORD=${DB_PASSWORD}
REDIS_PASSWORD=${REDIS_PASSWORD}

AUTH_SERVICE_SECRET=${AUTH_SERVICE_SECRET}
JWT_SECRET=${JWT_SECRET}

ALLOWED_ORIGIN=${ALLOWED_ORIGIN}
API_URL=${API_URL}
AUTH_PUBLIC_URL=${AUTH_PUBLIC_URL}
AUTH_COOKIE_DOMAIN=${AUTH_COOKIE_DOMAIN}
AUTH_SECURE_COOKIES=${AUTH_SECURE_COOKIES}

USE_BUNDLED_DB=${USE_BUNDLED_DB}
USE_BUNDLED_REDIS=${USE_BUNDLED_REDIS}

ADMIN_EMAIL=${ADMIN_EMAIL:-}
SELF_HOSTED=true
NIXOPUS_TELEMETRY=${NIXOPUS_TELEMETRY:-on}
LOG_LEVEL=${LOG_LEVEL:-debug}

USE_AGENT=${USE_AGENT}
OPENROUTER_API_KEY=${OPENROUTER_API_KEY:-}
AGENT_MODEL=${AGENT_MODEL:-}
AGENT_LIGHT_MODEL=${AGENT_LIGHT_MODEL:-}
USE_OLLAMA=${USE_OLLAMA}
OLLAMA_BASE_URL=${OLLAMA_BASE_URL:-}
EOF
    chmod 600 "$NIXOPUS_HOME/.env"
}

copy_compose() {
    local src="${NIXOPUS_INSTALLER_DIR:-}/selfhost"
    if [ -n "${NIXOPUS_INSTALLER_DIR:-}" ] && [ -f "$src/docker-compose.yml" ]; then
        cp "$src/docker-compose.yml" "$NIXOPUS_HOME/"
        cp "$src/docker-compose.db.yml" "$NIXOPUS_HOME/"
        cp "$src/docker-compose.redis.yml" "$NIXOPUS_HOME/"
        [ -f "$src/docker-compose.agent.yml" ] && cp "$src/docker-compose.agent.yml" "$NIXOPUS_HOME/"
        [ -f "$src/docker-compose.ollama.yml" ] && cp "$src/docker-compose.ollama.yml" "$NIXOPUS_HOME/"
    else
        curl -fsSL "$REPO_RAW/selfhost/docker-compose.yml" -o "$NIXOPUS_HOME/docker-compose.yml"
        curl -fsSL "$REPO_RAW/selfhost/docker-compose.db.yml" -o "$NIXOPUS_HOME/docker-compose.db.yml"
        curl -fsSL "$REPO_RAW/selfhost/docker-compose.redis.yml" -o "$NIXOPUS_HOME/docker-compose.redis.yml"
        curl -fsSL "$REPO_RAW/selfhost/docker-compose.agent.yml" -o "$NIXOPUS_HOME/docker-compose.agent.yml"
        curl -fsSL "$REPO_RAW/selfhost/docker-compose.ollama.yml" -o "$NIXOPUS_HOME/docker-compose.ollama.yml"
    fi
}

write_caddyfile() {
    cat > "$NIXOPUS_HOME/Caddyfile" << 'CADDY'
{
    admin 0.0.0.0:2019
}

{$SITE_ADDRESS} {
    handle /api/v1/* {
        reverse_proxy nixopus-api:8443
    }

    handle /ws {
        reverse_proxy nixopus-api:8443
    }

    handle /ws/* {
        reverse_proxy nixopus-api:8443
    }

    handle /agent/* {
        reverse_proxy nixopus-agent:4090
    }

    handle {
        reverse_proxy nixopus-view:7443 {
            flush_interval -1
        }
    }
}
CADDY
}

write_files() {
    write_env
    copy_compose
    write_caddyfile
    log_ok "Config, Compose, and Caddyfile written to $NIXOPUS_HOME"
}

# ── Step 5: Services ─────────────────────────────────────────────────────────

compose_files() {
    local args="-f $NIXOPUS_HOME/docker-compose.yml"
    [ "$USE_BUNDLED_DB" = true ] && args="$args -f $NIXOPUS_HOME/docker-compose.db.yml"
    [ "$USE_BUNDLED_REDIS" = true ] && args="$args -f $NIXOPUS_HOME/docker-compose.redis.yml"
    [ "${USE_AGENT:-true}" = true ] && [ -f "$NIXOPUS_HOME/docker-compose.agent.yml" ] && args="$args -f $NIXOPUS_HOME/docker-compose.agent.yml"
    [ "${USE_OLLAMA:-false}" = true ] && [ -f "$NIXOPUS_HOME/docker-compose.ollama.yml" ] && args="$args -f $NIXOPUS_HOME/docker-compose.ollama.yml"
    echo "$args"
}

dc() { docker compose $(compose_files) --env-file "$NIXOPUS_HOME/.env" "$@"; }

pull_ollama_model() {
    [ "${USE_OLLAMA:-false}" = true ] || return 0
    local model="llama3.2"
    log_info "Pulling Ollama model '${model}' (this may take a few minutes on first install)..."
    if docker exec nixopus-ollama ollama pull "$model" 2>&1 | tail -1; then
        log_ok "Model '${model}' ready"
    else
        log_warn "Model pull failed — the agent will auto-download it on first request"
    fi
}

start_services() {
    cd "$NIXOPUS_HOME"

    local expected=4
    [ "$USE_BUNDLED_DB" = true ] && expected=$((expected + 1))
    [ "$USE_BUNDLED_REDIS" = true ] && expected=$((expected + 1))
    [ "${USE_AGENT:-true}" = true ] && expected=$((expected + 1))
    [ "${USE_OLLAMA:-false}" = true ] && expected=$((expected + 1))

    if [ "$USE_BUNDLED_DB" = false ]; then
        log_info "Using external database"
    fi
    if [ "$USE_BUNDLED_REDIS" = false ]; then
        log_info "Using external Redis"
    fi
    if [ "${USE_AGENT:-true}" = true ]; then
        if [ "${USE_OLLAMA:-false}" = true ]; then
            log_info "Agent enabled with Ollama (local LLM)"
        else
            log_info "Agent enabled with OpenRouter"
        fi
    fi

    dc pull 2>/dev/null || true
    dc up -d --remove-orphans

    log_info "Waiting for services to start..."
    local timeout=180 elapsed=0 interval=5

    while [ $elapsed -lt $timeout ]; do
        local healthy
        healthy=$(dc ps 2>/dev/null | grep -c "(healthy)" || echo "0")

        if [ "$healthy" -ge "$expected" ]; then
            log_ok "All services healthy"
            pull_ollama_model
            return
        fi

        printf "\r  ${DIM}%ds / %ds (%s/%s healthy)${NC}  " "$elapsed" "$timeout" "$healthy" "$expected"
        sleep $interval
        elapsed=$((elapsed + interval))
    done

    echo ""
    log_warn "Health check timed out. Services may still be starting."
    log_info "Check status: nixopus status"
    log_info "Check logs:   nixopus logs"
}

# ── Step 6: Management Script ────────────────────────────────────────────────

install_management_script() {
    cat > /usr/local/bin/nixopus << 'MGMT'
#!/bin/bash
set -euo pipefail

NIXOPUS_HOME="${NIXOPUS_HOME:-/opt/nixopus}"

if [ "$(id -u)" -ne 0 ]; then
    echo "nixopus requires root. Run: sudo nixopus $*" >&2
    exit 1
fi

if [ ! -f "$NIXOPUS_HOME/.env" ]; then
    echo "Nixopus not found at $NIXOPUS_HOME. Is it installed?" >&2
    exit 1
fi

load_env() {
    set -a
    # shellcheck disable=SC1091
    . "$NIXOPUS_HOME/.env"
    set +a
}

compose_files() {
    local args="-f $NIXOPUS_HOME/docker-compose.yml"
    [ "${USE_BUNDLED_DB:-true}" = true ] && [ -f "$NIXOPUS_HOME/docker-compose.db.yml" ] && args="$args -f $NIXOPUS_HOME/docker-compose.db.yml"
    [ "${USE_BUNDLED_REDIS:-true}" = true ] && [ -f "$NIXOPUS_HOME/docker-compose.redis.yml" ] && args="$args -f $NIXOPUS_HOME/docker-compose.redis.yml"
    [ "${USE_AGENT:-true}" = true ] && [ -f "$NIXOPUS_HOME/docker-compose.agent.yml" ] && args="$args -f $NIXOPUS_HOME/docker-compose.agent.yml"
    [ "${USE_OLLAMA:-false}" = true ] && [ -f "$NIXOPUS_HOME/docker-compose.ollama.yml" ] && args="$args -f $NIXOPUS_HOME/docker-compose.ollama.yml"
    echo "$args"
}

dc() { load_env; docker compose $(compose_files) --env-file "$NIXOPUS_HOME/.env" "$@"; }

sedi() {
    if sed --version 2>/dev/null | grep -q GNU; then
        sed -i "$@"
    else
        sed -i '' "$@"
    fi
}

redact() { echo "${1:0:4}****${1: -4}"; }

format_ip_for_url() {
    if [[ "$1" == *:* ]]; then
        echo "[$1]"
    else
        echo "$1"
    fi
}

cmd_status() {
    dc ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
}

cmd_logs() {
    local service="${1:-}"
    if [ -n "$service" ]; then
        dc logs -f --tail 100 "$service"
    else
        dc logs -f --tail 100
    fi
}

cmd_update() {
    echo "Pulling latest images..."
    dc pull --quiet 2>/dev/null || dc pull
    echo "Restarting services..."
    dc up -d --remove-orphans
    echo "Update complete."
    cmd_status
}

cmd_restart() {
    local service="${1:-}"
    if [ -n "$service" ]; then
        dc restart "$service"
    else
        dc restart
    fi
}

cmd_stop() {
    dc stop
    echo "Services stopped."
}

cmd_uninstall() {
    echo "This will stop all Nixopus services."
    if [ "${1:-}" = "--purge" ]; then
        echo "WARNING: --purge will also delete all data (database, redis, configs)."
    fi
    read -rp "Continue? [y/N] " confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        echo "Cancelled."
        exit 0
    fi

    if [ "${1:-}" = "--purge" ]; then
        dc down -v
        rm -rf "$NIXOPUS_HOME"
        echo "All data removed."
    else
        dc down
    fi

    rm -f /usr/local/bin/nixopus
    echo "Nixopus uninstalled."
}

cmd_config() {
    load_env
    if [ "${1:-}" = "set" ]; then
        shift
        for pair in "$@"; do
            local key="${pair%%=*}"
            local val="${pair#*=}"
            if grep -q "^${key}=" "$NIXOPUS_HOME/.env"; then
                sedi "s|^${key}=.*|${key}=${val}|" "$NIXOPUS_HOME/.env"
            else
                echo "${key}=${val}" >> "$NIXOPUS_HOME/.env"
            fi
            echo "Set $key"
        done
        echo "Restart services for changes to take effect: nixopus restart"
        return
    fi

    echo "── Nixopus Configuration ──"
    echo "Home:         $NIXOPUS_HOME"
    echo "Version:      ${NIXOPUS_VERSION:-unknown}"
    echo "Domain:       ${DOMAIN:-<none, IP mode>}"
    echo "Host IP:      ${HOST_IP:-<unknown>}"
    echo "Access:       ${ALLOWED_ORIGIN:-unknown}"
    echo "HTTP Port:    ${CADDY_HTTP_PORT:-80}"
    echo "HTTPS Port:   ${CADDY_HTTPS_PORT:-443}"
    echo "SSH Host:     ${SSH_HOST:-${HOST_IP:-<unknown>}}"
    echo "SSH Port:     ${SSH_PORT:-22}"
    echo "SSH User:     ${SSH_USER:-root}"
    echo ""
    echo "Database URL: $(echo "${DATABASE_URL:-}" | sed 's|://[^@]*@|://****@|')"
    echo "Redis URL:    $(echo "${REDIS_URL:-}" | sed 's|://[^@]*@|://****@|')"
    echo "Auth Secret:  $(redact "${AUTH_SERVICE_SECRET:-}")"
    echo "JWT Secret:   $(redact "${JWT_SECRET:-}")"
    echo ""
    echo "Agent:        ${USE_AGENT:-true}"
    if [ "${USE_AGENT:-true}" = true ]; then
        if [ -n "${OPENROUTER_API_KEY:-}" ]; then
            echo "LLM:          OpenRouter ($(redact "${OPENROUTER_API_KEY}"))"
        else
            echo "LLM:          Ollama (local)"
        fi
    fi
}

cmd_domain() {
    local action="${1:-}" domain="${2:-}"
    if [ "$action" != "add" ] && [ "$action" != "remove" ]; then
        echo "Usage: nixopus domain add <domain>"
        echo "       nixopus domain remove"
        return 1
    fi

    load_env

    if [ "$action" = "add" ]; then
        if [ -z "$domain" ]; then
            echo "Usage: nixopus domain add <domain>" >&2
            return 1
        fi
        local cookie_domain
        cookie_domain=".$(echo "$domain" | awk -F. '{print $(NF-1)"."$NF}')"

        sedi "s|^DOMAIN=.*|DOMAIN=${domain}|" "$NIXOPUS_HOME/.env"
        sedi "s|^SITE_ADDRESS=.*|SITE_ADDRESS=${domain}|" "$NIXOPUS_HOME/.env"
        sedi "s|^ALLOWED_ORIGIN=.*|ALLOWED_ORIGIN=https://${domain}|" "$NIXOPUS_HOME/.env"
        sedi "s|^API_URL=.*|API_URL=https://${domain}/api|" "$NIXOPUS_HOME/.env"
        sedi "s|^AUTH_PUBLIC_URL=.*|AUTH_PUBLIC_URL=https://${domain}|" "$NIXOPUS_HOME/.env"
        sedi "s|^AUTH_COOKIE_DOMAIN=.*|AUTH_COOKIE_DOMAIN=${cookie_domain}|" "$NIXOPUS_HOME/.env"
        sedi "s|^AUTH_SECURE_COOKIES=.*|AUTH_SECURE_COOKIES=true|" "$NIXOPUS_HOME/.env"

        echo "Domain set to $domain"
        echo "Ensure DNS A record for $domain points to ${HOST_IP:-your server IP}"
        echo "Restarting services for HTTPS..."
        dc up -d --remove-orphans
    fi

    if [ "$action" = "remove" ]; then
        local host_ip http_port base_url
        host_ip=$(grep "^HOST_IP=" "$NIXOPUS_HOME/.env" | cut -d= -f2)
        http_port=$(grep "^CADDY_HTTP_PORT=" "$NIXOPUS_HOME/.env" | cut -d= -f2)
        local ip_for_url
        ip_for_url=$(format_ip_for_url "$host_ip")
        if [ "${http_port:-80}" = "80" ]; then
            base_url="http://${ip_for_url}"
        else
            base_url="http://${ip_for_url}:${http_port}"
        fi

        sedi "s|^DOMAIN=.*|DOMAIN=|" "$NIXOPUS_HOME/.env"
        sedi "s|^SITE_ADDRESS=.*|SITE_ADDRESS=:80|" "$NIXOPUS_HOME/.env"
        sedi "s|^ALLOWED_ORIGIN=.*|ALLOWED_ORIGIN=${base_url}|" "$NIXOPUS_HOME/.env"
        sedi "s|^API_URL=.*|API_URL=${base_url}/api|" "$NIXOPUS_HOME/.env"
        sedi "s|^AUTH_PUBLIC_URL=.*|AUTH_PUBLIC_URL=${base_url}|" "$NIXOPUS_HOME/.env"
        sedi "s|^AUTH_COOKIE_DOMAIN=.*|AUTH_COOKIE_DOMAIN=|" "$NIXOPUS_HOME/.env"
        sedi "s|^AUTH_SECURE_COOKIES=.*|AUTH_SECURE_COOKIES=false|" "$NIXOPUS_HOME/.env"

        echo "Switched to IP-based mode (${base_url})"
        echo "Restarting services..."
        dc up -d --remove-orphans
    fi
}

cmd_port() {
    local action="${1:-}" port_type="${2:-}" port_val="${3:-}"
    if [ "$action" != "set" ] || [ -z "$port_type" ] || [ -z "$port_val" ]; then
        load_env
        echo "── Port Configuration ──"
        echo "HTTP:   ${CADDY_HTTP_PORT:-80}"
        echo "HTTPS:  ${CADDY_HTTPS_PORT:-443}"
        echo "SSH:    ${SSH_PORT:-22}"
        echo ""
        echo "Usage: nixopus port set <http|https|ssh> <port>"
        return
    fi

    case "$port_type" in
        http)  sedi "s|^CADDY_HTTP_PORT=.*|CADDY_HTTP_PORT=${port_val}|" "$NIXOPUS_HOME/.env" ;;
        https) sedi "s|^CADDY_HTTPS_PORT=.*|CADDY_HTTPS_PORT=${port_val}|" "$NIXOPUS_HOME/.env" ;;
        ssh)   sedi "s|^SSH_PORT=.*|SSH_PORT=${port_val}|" "$NIXOPUS_HOME/.env" ;;
        *)     echo "Unknown port type: $port_type (use http, https, or ssh)" >&2; return 1 ;;
    esac

    echo "Set $port_type port to $port_val"
    echo "Restarting services..."
    dc up -d --remove-orphans
}

cmd_ip() {
    local action="${1:-}" new_ip="${2:-}"
    load_env
    if [ "$action" != "set" ] || [ -z "$new_ip" ]; then
        echo "── IP Configuration ──"
        echo "Host IP:  ${HOST_IP:-<unknown>}"
        echo "Access:   ${ALLOWED_ORIGIN:-unknown}"
        echo ""
        echo "Usage: nixopus ip set <ip>"
        return
    fi

    sedi "s|^HOST_IP=.*|HOST_IP=${new_ip}|" "$NIXOPUS_HOME/.env"
    if [ -z "${DOMAIN:-}" ]; then
        local ip_for_url
        ip_for_url=$(format_ip_for_url "$new_ip")
        local http_port
        http_port=$(grep "^CADDY_HTTP_PORT=" "$NIXOPUS_HOME/.env" | cut -d= -f2)
        local base_url
        if [ "${http_port:-80}" = "80" ]; then
            base_url="http://${ip_for_url}"
        else
            base_url="http://${ip_for_url}:${http_port}"
        fi
        sedi "s|^ALLOWED_ORIGIN=.*|ALLOWED_ORIGIN=${base_url}|" "$NIXOPUS_HOME/.env"
        sedi "s|^API_URL=.*|API_URL=${base_url}/api|" "$NIXOPUS_HOME/.env"
        sedi "s|^AUTH_PUBLIC_URL=.*|AUTH_PUBLIC_URL=${base_url}|" "$NIXOPUS_HOME/.env"
    fi
    echo "Set IP to $new_ip"
    echo "Restarting services..."
    dc up -d --remove-orphans
}

cmd_backup() {
    local backup_dir="$NIXOPUS_HOME/backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"

    load_env

    echo "Backing up database..."
    docker exec nixopus-db pg_dump -U nixopus nixopus \
        > "$backup_dir/db.sql" 2>/dev/null \
        || echo "Warning: DB backup failed"

    echo "Backing up config..."
    cp "$NIXOPUS_HOME/.env" "$backup_dir/env.bak"

    echo "Backup saved to $backup_dir"
    ls -lh "$backup_dir"
}

cmd_info() {
    load_env
    echo "Nixopus v${NIXOPUS_VERSION:-unknown}"
    echo "Home:   $NIXOPUS_HOME"
    echo "Domain: ${DOMAIN:-<none>}"
    echo "Access: ${ALLOWED_ORIGIN:-unknown}"
    echo ""
    cmd_status
}

cmd_help() {
    echo "Usage: nixopus <command> [args]"
    echo ""
    echo "Commands:"
    echo "  status              Show service health"
    echo "  logs [service]      Tail service logs"
    echo "  update              Pull latest images and restart"
    echo "  restart [service]   Restart services"
    echo "  stop                Stop all services"
    echo "  uninstall [--purge] Remove Nixopus (--purge deletes data)"
    echo "  config              Show current configuration"
    echo "  config set K=V      Update configuration value"
    echo "  domain add <domain> Switch to domain-based HTTPS"
    echo "  domain remove       Switch back to IP-based HTTP"
    echo "  ip set <ip>         Change host IP (IP mode only)"
    echo "  port                Show port configuration"
    echo "  port set <type> <n> Change port (http, https, ssh)"
    echo "  backup              Backup database and config"
    echo "  info                Show install info and status"
    echo "  help                Show this help"
}

case "${1:-help}" in
    status)    cmd_status ;;
    logs)      shift; cmd_logs "${1:-}" ;;
    update)    cmd_update ;;
    restart)   shift; cmd_restart "${1:-}" ;;
    stop)      cmd_stop ;;
    uninstall) shift; cmd_uninstall "${1:-}" ;;
    config)    shift; cmd_config "$@" ;;
    domain)    shift; cmd_domain "$@" ;;
    ip)        shift; cmd_ip "$@" ;;
    port)      shift; cmd_port "$@" ;;
    backup)    cmd_backup ;;
    info)      cmd_info ;;
    help|--help|-h) cmd_help ;;
    *)         echo "Unknown command: $1"; cmd_help; exit 1 ;;
esac
MGMT
    chmod +x /usr/local/bin/nixopus
    log_ok "Management CLI installed: nixopus"
}

# ── Telemetry ─────────────────────────────────────────────────────────────────

send_telemetry() {
    local event="${1:-install_success}" error="${2:-}"
    [ "${NIXOPUS_TELEMETRY:-on}" = "off" ] && return 0

    local duration=$(( $(date +%s) - INSTALL_START ))
    local payload="{\"event_type\":\"$event\",\"os\":\"${OS_ID:-unknown}\",\"arch\":\"${ARCH:-unknown}\",\"version\":\"$NIXOPUS_VERSION\",\"duration\":$duration"
    [ -n "$error" ] && payload="$payload,\"error\":\"$(echo "$error" | head -c 200)\""
    payload="$payload}"

    curl -fsSL --connect-timeout 5 --max-time 10 \
        -X POST "$TELEMETRY_URL" \
        -H "Content-Type: application/json" \
        -d "$payload" &>/dev/null &
    return 0
}

# ── Main ──────────────────────────────────────────────────────────────────────

show_banner() {
    echo ""
    echo -e "${BOLD}  Nixopus Self-Host Installer v${NIXOPUS_VERSION}${NC}"
    echo -e "  ${DIM}https://nixopus.com${NC}"
    echo ""
}

show_complete() {
    local duration=$(( $(date +%s) - INSTALL_START ))
    echo ""
    echo -e "  ${GREEN}${BOLD}Nixopus is running!${NC}"
    echo ""
    echo -e "  ${BOLD}Access:${NC}    $BASE_URL"
    echo -e "  ${BOLD}Config:${NC}    $NIXOPUS_HOME/.env"
    echo -e "  ${BOLD}Installed:${NC} ${duration}s"
    echo ""
    echo -e "  ${BOLD}Commands:${NC}"
    echo "    nixopus status     Show service health"
    echo "    nixopus logs       View logs"
    echo "    nixopus update     Update to latest version"
    echo "    nixopus --help     All commands"
    echo ""
    if [ -n "${DOMAIN:-}" ]; then
        echo -e "  ${YELLOW}Ensure DNS A record for ${DOMAIN} → ${HOST_IP:-your-ip}${NC}"
        echo ""
    fi
    echo -e "  ${DIM}Docs: https://docs.nixopus.com | Discord: https://discord.gg/skdcq39Wpv${NC}"

    if [ "${NIXOPUS_TELEMETRY:-on}" != "off" ]; then
        echo -e "  ${DIM}Anonymous telemetry enabled. Disable: NIXOPUS_TELEMETRY=off${NC}"
    fi
    echo ""
}

main() {
    show_banner

    log_step 1 "Checking requirements"
    check_root
    detect_os
    install_prereqs

    log_step 2 "Setting up Docker"
    install_docker

    log_step 3 "Configuring"
    load_existing_config
    gather_config


    log_step 4 "Writing files"
    setup_directories
    setup_ssh
    write_files

    log_step 5 "Starting services"
    start_services

    log_step 6 "Installing management CLI"
    install_management_script

    show_complete
    send_telemetry "install_success" || true
}

main "$@"
