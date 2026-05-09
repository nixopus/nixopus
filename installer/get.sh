#!/bin/bash
set -euo pipefail

NIXOPUS_VERSION="0.3.0"
NIXOPUS_HOME="${NIXOPUS_HOME:-/opt/nixopus}"
TELEMETRY_URL="${NIXOPUS_TELEMETRY_URL:-https://api.nixopus.com/api/v1/cli/telemetry}"
NIXOPUS_REPO="${NIXOPUS_REPO:-nixopus/nixopus}"
NIXOPUS_BRANCH="${NIXOPUS_BRANCH:-master}"
REPO_RAW="${NIXOPUS_REPO_RAW:-https://raw.githubusercontent.com/${NIXOPUS_REPO}/${NIXOPUS_BRANCH}/installer}"
INSTALL_START=$(date +%s)
# Set in main(); full installer transcript for support (tee). On failure, path is under /tmp.
INSTALL_LOG=""

start_install_session_log() {
    INSTALL_LOG="/tmp/nixopus-install-$(date +%Y%m%dT%H%M%S)-$$.log"
    : >"$INSTALL_LOG"
    chmod 600 "$INSTALL_LOG" || true
    # Duplicate stdout/stderr to a file for copy-paste debugging.
    exec > >(tee "$INSTALL_LOG") 2>&1
}

PREVIEW_PR=""
PREVIEW_FORK=""
PREVIEW_BRANCH=""

parse_args() {
    while [ $# -gt 0 ]; do
        case "$1" in
            --preview)
                shift
                PREVIEW_PR="${1:-}"
                [ -z "$PREVIEW_PR" ] && { echo "Usage: --preview <pr-number>" >&2; exit 1; }
                ;;
            --branch)
                shift
                PREVIEW_BRANCH="${1:-}"
                [ -z "$PREVIEW_BRANCH" ] && { echo "Usage: --branch <branch-name>" >&2; exit 1; }
                ;;
            --fork)
                shift
                PREVIEW_FORK="${1:-}"
                [ -z "$PREVIEW_FORK" ] && { echo "Usage: --fork <owner/repo>" >&2; exit 1; }
                ;;
            --help|-h)
                echo "Usage: curl -fsSL install.nixopus.com | sudo bash -s -- [options]"
                echo ""
                echo "Non-interactive / env-heavy installs: if variables are ignored or the"
                echo "completion summary is missing, sudo may be stripping the environment."
                echo "Use a nested shell so exports apply to the same bash that runs this script:"
                echo '  sudo bash -c '"'"'export ADMIN_EMAIL=… ADMIN_PASSWORD=…; curl -fsSL install.nixopus.com | bash'"'"
                echo ""
                echo "Options:"
                echo "  --preview <pr>        Install from a PR's preview images (pr number)"
                echo "  --branch <name>       Use installer files from a specific branch"
                echo "  --fork <owner/repo>   Use installer files and images from a fork"
                echo "  --help                Show this help"
                echo ""
                echo "Examples:"
                echo "  # Test PR #42"
                echo "  curl -fsSL install.nixopus.com | sudo bash -s -- --preview 42"
                echo ""
                echo "  # Test a branch from the main repo"
                echo "  curl -fsSL ... | sudo bash -s -- --branch feat/new-feature --preview 42"
                echo ""
                echo "  # Test a fork"
                echo "  curl -fsSL ... | sudo bash -s -- --fork someuser/nixopus --branch main"
                exit 0
                ;;
            *)
                echo "Unknown option: $1 (use --help for usage)" >&2
                exit 1
                ;;
        esac
        shift
    done
}

preview_image_exists() {
    local image="$1"
    docker manifest inspect "$image" >/dev/null 2>&1
}

apply_preview_config() {
    local repo="${PREVIEW_FORK:-${NIXOPUS_REPO}}"
    local branch="${PREVIEW_BRANCH:-${NIXOPUS_BRANCH}}"

    if [ -n "$PREVIEW_PR" ]; then
        local api_tag="ghcr.io/${repo}-api:pr-${PREVIEW_PR}"
        local view_tag="ghcr.io/${repo}-view:pr-${PREVIEW_PR}"
        local api_ok=false view_ok=false

        if [ -z "${NIXOPUS_API_IMAGE:-}" ]; then
            if preview_image_exists "$api_tag"; then
                NIXOPUS_API_IMAGE="$api_tag"
                api_ok=true
            fi
        else
            api_ok=true
        fi

        if [ -z "${NIXOPUS_VIEW_IMAGE:-}" ]; then
            if preview_image_exists "$view_tag"; then
                NIXOPUS_VIEW_IMAGE="$view_tag"
                view_ok=true
            fi
        else
            view_ok=true
        fi

        if [ "$api_ok" = false ] && [ "$view_ok" = false ]; then
            echo -e "  \033[1;33mWARN\033[0m  No preview images found for PR #${PREVIEW_PR} — using latest for both API and View"
        elif [ "$api_ok" = false ]; then
            echo -e "  \033[0;34mINFO\033[0m  Preview image not available for API (pr-${PREVIEW_PR}) — using latest"
        elif [ "$view_ok" = false ]; then
            echo -e "  \033[0;34mINFO\033[0m  Preview image not available for View (pr-${PREVIEW_PR}) — using latest"
        fi

        export NIXOPUS_API_IMAGE NIXOPUS_VIEW_IMAGE
    fi

    if [ -n "$PREVIEW_FORK" ]; then
        branch="${PREVIEW_BRANCH:-main}"
    fi

    if [ -n "$PREVIEW_BRANCH" ] || [ -n "$PREVIEW_FORK" ]; then
        REPO_RAW="${NIXOPUS_REPO_RAW:-https://raw.githubusercontent.com/${repo}/${branch}/installer}"
    fi
}

parse_args "$@"

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
    if [ -n "${INSTALL_LOG:-}" ] && [ -f "$INSTALL_LOG" ]; then
        local log_path="$INSTALL_LOG"
        if [ -d "${NIXOPUS_HOME:-}" ]; then
            if cp -f "$INSTALL_LOG" "$NIXOPUS_HOME/install.log" 2>/dev/null; then
                chmod 600 "$NIXOPUS_HOME/install.log" 2>/dev/null || true
                log_path="$NIXOPUS_HOME/install.log"
                rm -f "$INSTALL_LOG" 2>/dev/null || true
            fi
        fi
        log_error "Installer log: $log_path"
        log_error "Attach when asking for help (may include secrets — redact before sharing)."
    fi
    send_telemetry "install_failure" "$1" || true
    exit 1
}

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

# gen_password generates a 16-char password that satisfies Better Auth's
# validation rules (>=8 chars, 1 upper, 1 lower, 1 digit, 1 special).
# Special set is restricted to characters that are JSON- and shell-safe so
# the resulting value can be embedded in .env and curl payloads without
# escaping headaches.  Exclude & ; | etc. — unquoted .env lines must remain
# valid when sourced by nixopus.sh under set -u (see load_env).
gen_password() {
    local upper lower digit special rest combined
    upper=$(LC_ALL=C tr -dc 'A-Z'             </dev/urandom 2>/dev/null | head -c 1)
    lower=$(LC_ALL=C tr -dc 'a-z'             </dev/urandom 2>/dev/null | head -c 1)
    digit=$(LC_ALL=C tr -dc '0-9'             </dev/urandom 2>/dev/null | head -c 1)
    special=$(LC_ALL=C tr -dc '!@#%^*'       </dev/urandom 2>/dev/null | head -c 1)
    rest=$(LC_ALL=C tr -dc 'A-Za-z0-9!@#%^*' </dev/urandom 2>/dev/null | head -c 12)
    combined="${upper}${lower}${digit}${special}${rest}"
    echo "$combined" | fold -w1 | shuf | tr -d '\n'
}

json_escape() {
    local s="$1"
    s="${s//\\/\\\\}"
    s="${s//\"/\\\"}"
    printf '%s' "$s"
}

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

    # Auto-generate so the installer can bootstrap non-interactively when an
    # email was provided. The generated value is printed at the end.
    if [ -n "${ADMIN_EMAIL:-}" ] && [ -z "${ADMIN_PASSWORD:-}" ]; then
        ADMIN_PASSWORD="$(gen_password)"
        ADMIN_PASSWORD_GENERATED=true
    else
        ADMIN_PASSWORD_GENERATED=false
    fi

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

    LLM_PROVIDER="${LLM_PROVIDER:-}"
    OPENROUTER_API_KEY="${OPENROUTER_API_KEY:-}"
    OPENAI_API_KEY="${OPENAI_API_KEY:-}"
    ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"
    GOOGLE_GENERATIVE_AI_API_KEY="${GOOGLE_GENERATIVE_AI_API_KEY:-}"
    DEEPSEEK_API_KEY="${DEEPSEEK_API_KEY:-}"
    GROQ_API_KEY="${GROQ_API_KEY:-}"

    has_any_llm_key() {
        [ -n "${OPENROUTER_API_KEY:-}" ] || [ -n "${OPENAI_API_KEY:-}" ] || \
        [ -n "${ANTHROPIC_API_KEY:-}" ] || [ -n "${GOOGLE_GENERATIVE_AI_API_KEY:-}" ] || \
        [ -n "${DEEPSEEK_API_KEY:-}" ] || [ -n "${GROQ_API_KEY:-}" ]
    }

    if [ -t 0 ] && ! has_any_llm_key; then
        echo ""
        echo -e "  ${BOLD}AI Agent — LLM Provider${NC}"
        echo -e "  ${DIM}The agent needs an LLM API key for deployments and diagnostics.${NC}"
        echo ""
        echo -e "  ${BOLD}1)${NC} OpenRouter  ${DIM}— access Claude, GPT-4, Gemini & more via one key${NC}"
        echo -e "  ${BOLD}2)${NC} OpenAI      ${DIM}— GPT-4o, GPT-4.1${NC}"
        echo -e "  ${BOLD}3)${NC} Anthropic   ${DIM}— Claude Sonnet, Haiku${NC}"
        echo -e "  ${BOLD}4)${NC} Google      ${DIM}— Gemini 2.5 Flash${NC}"
        echo -e "  ${BOLD}5)${NC} DeepSeek    ${DIM}— DeepSeek Chat${NC}"
        echo -e "  ${BOLD}6)${NC} Groq        ${DIM}— Llama 3.3 70B${NC}"
        echo ""
        prompt_if_tty LLM_PROVIDER "Provider [1-6]" "1"

        case "${LLM_PROVIDER}" in
            1|openrouter|"")
                LLM_PROVIDER="openrouter"
                prompt_if_tty OPENROUTER_API_KEY "OpenRouter API key (https://openrouter.ai)" ""
                AGENT_MODEL="${AGENT_MODEL:-openrouter/anthropic/claude-sonnet-4}"
                AGENT_LIGHT_MODEL="${AGENT_LIGHT_MODEL:-openrouter/openai/gpt-4o-mini}"
                ;;
            2|openai)
                LLM_PROVIDER="openai"
                prompt_if_tty OPENAI_API_KEY "OpenAI API key" ""
                AGENT_MODEL="${AGENT_MODEL:-openai/gpt-4o}"
                AGENT_LIGHT_MODEL="${AGENT_LIGHT_MODEL:-openai/gpt-4o-mini}"
                ;;
            3|anthropic)
                LLM_PROVIDER="anthropic"
                prompt_if_tty ANTHROPIC_API_KEY "Anthropic API key" ""
                AGENT_MODEL="${AGENT_MODEL:-anthropic/claude-sonnet-4}"
                AGENT_LIGHT_MODEL="${AGENT_LIGHT_MODEL:-anthropic/claude-haiku-3.5}"
                ;;
            4|google)
                LLM_PROVIDER="google"
                prompt_if_tty GOOGLE_GENERATIVE_AI_API_KEY "Google AI API key" ""
                AGENT_MODEL="${AGENT_MODEL:-google/gemini-2.5-flash}"
                AGENT_LIGHT_MODEL="${AGENT_LIGHT_MODEL:-google/gemini-2.0-flash}"
                ;;
            5|deepseek)
                LLM_PROVIDER="deepseek"
                prompt_if_tty DEEPSEEK_API_KEY "DeepSeek API key" ""
                AGENT_MODEL="${AGENT_MODEL:-deepseek/deepseek-chat}"
                AGENT_LIGHT_MODEL="${AGENT_LIGHT_MODEL:-deepseek/deepseek-chat}"
                ;;
            6|groq)
                LLM_PROVIDER="groq"
                prompt_if_tty GROQ_API_KEY "Groq API key" ""
                AGENT_MODEL="${AGENT_MODEL:-groq/llama-3.3-70b-versatile}"
                AGENT_LIGHT_MODEL="${AGENT_LIGHT_MODEL:-groq/llama-3.3-70b-versatile}"
                ;;
            *)
                log_warn "Unknown provider '$LLM_PROVIDER', defaulting to OpenRouter"
                LLM_PROVIDER="openrouter"
                prompt_if_tty OPENROUTER_API_KEY "OpenRouter API key (https://openrouter.ai)" ""
                AGENT_MODEL="${AGENT_MODEL:-openrouter/anthropic/claude-sonnet-4}"
                AGENT_LIGHT_MODEL="${AGENT_LIGHT_MODEL:-openrouter/openai/gpt-4o-mini}"
                ;;
        esac
    fi

    # Non-TTY / pre-seeded keys skip the menu above, so AGENT_MODEL may be empty
    # even when a provider key is set. OpenRouter was also missing defaults in the
    # interactive case until the previous block. Infer provider from keys.
    if has_any_llm_key; then
        if [ -z "${AGENT_MODEL:-}" ] || [ -z "${AGENT_LIGHT_MODEL:-}" ]; then
            _key_count=0
            [ -n "${OPENROUTER_API_KEY:-}" ] && { _key_count=$((_key_count+1)); _inferred=openrouter; }
            [ -n "${OPENAI_API_KEY:-}" ] && { _key_count=$((_key_count+1)); _inferred=openai; }
            [ -n "${ANTHROPIC_API_KEY:-}" ] && { _key_count=$((_key_count+1)); _inferred=anthropic; }
            [ -n "${GOOGLE_GENERATIVE_AI_API_KEY:-}" ] && { _key_count=$((_key_count+1)); _inferred=google; }
            [ -n "${DEEPSEEK_API_KEY:-}" ] && { _key_count=$((_key_count+1)); _inferred=deepseek; }
            [ -n "${GROQ_API_KEY:-}" ] && { _key_count=$((_key_count+1)); _inferred=groq; }
            if [ "$_key_count" -eq 1 ] && [ -n "${_inferred:-}" ]; then
                LLM_PROVIDER="${_inferred}"
            fi
            # Only set models if LLM_PROVIDER is set and recognized
            if [ -n "${LLM_PROVIDER:-}" ]; then
                case "${LLM_PROVIDER}" in
                    openrouter)
                        AGENT_MODEL="${AGENT_MODEL:-openrouter/anthropic/claude-sonnet-4}"
                        AGENT_LIGHT_MODEL="${AGENT_LIGHT_MODEL:-openrouter/openai/gpt-4o-mini}"
                        ;;
                    openai)
                        AGENT_MODEL="${AGENT_MODEL:-openai/gpt-4o}"
                        AGENT_LIGHT_MODEL="${AGENT_LIGHT_MODEL:-openai/gpt-4o-mini}"
                        ;;
                    anthropic)
                        AGENT_MODEL="${AGENT_MODEL:-anthropic/claude-sonnet-4}"
                        AGENT_LIGHT_MODEL="${AGENT_LIGHT_MODEL:-anthropic/claude-haiku-3.5}"
                        ;;
                    google)
                        AGENT_MODEL="${AGENT_MODEL:-google/gemini-2.5-flash}"
                        AGENT_LIGHT_MODEL="${AGENT_LIGHT_MODEL:-google/gemini-2.0-flash}"
                        ;;
                    deepseek)
                        AGENT_MODEL="${AGENT_MODEL:-deepseek/deepseek-chat}"
                        AGENT_LIGHT_MODEL="${AGENT_LIGHT_MODEL:-deepseek/deepseek-chat}"
                        ;;
                    groq)
                        AGENT_MODEL="${AGENT_MODEL:-groq/llama-3.3-70b-versatile}"
                        AGENT_LIGHT_MODEL="${AGENT_LIGHT_MODEL:-groq/llama-3.3-70b-versatile}"
                        ;;
                    *)
                        log_warn "Unsupported LLM_PROVIDER: ${LLM_PROVIDER} — skipping model defaults"
                        ;;
                esac
            fi
        fi
    fi

    if ! has_any_llm_key; then
        log_warn "No LLM API key provided — the AI agent will not function without one"
        log_warn "Set it later: nixopus config set <PROVIDER>_API_KEY=xxx && nixopus restart"
    fi

    check_firewall

    log_ok "Configuration ready (${DOMAIN:-IP: $HOST_IP})"
}

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
BASE_URL=${BASE_URL}
API_URL=${API_URL}
AUTH_PUBLIC_URL=${AUTH_PUBLIC_URL}
AUTH_COOKIE_DOMAIN=${AUTH_COOKIE_DOMAIN}
AUTH_SECURE_COOKIES=${AUTH_SECURE_COOKIES}

USE_BUNDLED_DB=${USE_BUNDLED_DB}
USE_BUNDLED_REDIS=${USE_BUNDLED_REDIS}

ADMIN_EMAIL=${ADMIN_EMAIL:-}
ADMIN_PASSWORD=${ADMIN_PASSWORD:-}
SELF_HOSTED=true
NIXOPUS_TELEMETRY=${NIXOPUS_TELEMETRY:-on}
LOG_LEVEL=${LOG_LEVEL:-debug}

LLM_PROVIDER=${LLM_PROVIDER:-openrouter}
OPENROUTER_API_KEY=${OPENROUTER_API_KEY:-}
OPENAI_API_KEY=${OPENAI_API_KEY:-}
ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY:-}
GOOGLE_GENERATIVE_AI_API_KEY=${GOOGLE_GENERATIVE_AI_API_KEY:-}
DEEPSEEK_API_KEY=${DEEPSEEK_API_KEY:-}
GROQ_API_KEY=${GROQ_API_KEY:-}
AGENT_MODEL=${AGENT_MODEL:-}
AGENT_LIGHT_MODEL=${AGENT_LIGHT_MODEL:-}

NIXOPUS_API_IMAGE=${NIXOPUS_API_IMAGE:-}
NIXOPUS_VIEW_IMAGE=${NIXOPUS_VIEW_IMAGE:-}
NIXOPUS_AUTH_IMAGE=${NIXOPUS_AUTH_IMAGE:-}
EOF
    chmod 600 "$NIXOPUS_HOME/.env"
}

copy_compose() {
    local src="${NIXOPUS_INSTALLER_DIR:-}/selfhost"
    if [ -n "${NIXOPUS_INSTALLER_DIR:-}" ] && [ -f "$src/docker-compose.yml" ]; then
        cp "$src/docker-compose.yml" "$NIXOPUS_HOME/"
        cp "$src/docker-compose.db.yml" "$NIXOPUS_HOME/"
        cp "$src/docker-compose.redis.yml" "$NIXOPUS_HOME/"
    else
        curl -fsSL "$REPO_RAW/selfhost/docker-compose.yml" -o "$NIXOPUS_HOME/docker-compose.yml"
        curl -fsSL "$REPO_RAW/selfhost/docker-compose.db.yml" -o "$NIXOPUS_HOME/docker-compose.db.yml"
        curl -fsSL "$REPO_RAW/selfhost/docker-compose.redis.yml" -o "$NIXOPUS_HOME/docker-compose.redis.yml"
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

    handle /api/auth/* {
        reverse_proxy nixopus-auth:9090
    }

    handle /ws {
        reverse_proxy nixopus-api:8443
    }

    handle /ws/* {
        reverse_proxy nixopus-api:8443
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
    umask 077
    {
        echo "AGENT_MODEL=${AGENT_MODEL:-}"
        echo "AGENT_LIGHT_MODEL=${AGENT_LIGHT_MODEL:-}"
        echo "LLM_PROVIDER=${LLM_PROVIDER:-openrouter}"
    } >"$NIXOPUS_HOME/view-llm.env"
    chmod 600 "$NIXOPUS_HOME/view-llm.env" 2>/dev/null || true
    copy_compose
    write_caddyfile
    log_ok "Config, Compose, and Caddyfile written to $NIXOPUS_HOME"
}

compose_files() {
    local args="-f $NIXOPUS_HOME/docker-compose.yml"
    [ "$USE_BUNDLED_DB" = true ] && args="$args -f $NIXOPUS_HOME/docker-compose.db.yml"
    [ "$USE_BUNDLED_REDIS" = true ] && args="$args -f $NIXOPUS_HOME/docker-compose.redis.yml"
    echo "$args"
}

dc() { docker compose $(compose_files) --env-file "$NIXOPUS_HOME/.env" "$@"; }

start_services() {
    cd "$NIXOPUS_HOME"

    local expected=4
    [ "$USE_BUNDLED_DB" = true ] && expected=$((expected + 1))
    [ "$USE_BUNDLED_REDIS" = true ] && expected=$((expected + 1))

    if [ "$USE_BUNDLED_DB" = false ]; then
        log_info "Using external database"
    fi
    if [ "$USE_BUNDLED_REDIS" = false ]; then
        log_info "Using external Redis"
    fi
    if has_any_llm_key; then
        log_info "AI agent enabled (LLM provider: ${LLM_PROVIDER:-openrouter})"
    fi

    log_info "Pulling container images (this may take several minutes on first install)..."
    # --parallel pulls all images concurrently (default in Compose v2, explicit for v1 compat)
    dc pull --parallel 2>&1 | while IFS= read -r line; do
        case "$line" in
            *Pulling*|*Pull*complete*|*Downloaded*|*"Already exists"*|*Waiting*|*Extracting*|*Verifying*)
                printf "\r  ${DIM}%s${NC}  " "$line"
                ;;
        esac
    done || true
    echo ""
    log_ok "Images pulled"

    log_info "Starting containers..."
    dc up -d --remove-orphans

    log_info "Waiting for services to become healthy..."
    local timeout=180 elapsed=0 interval=5

    while [ $elapsed -lt $timeout ]; do
        local healthy
        healthy=$(dc ps 2>/dev/null | grep -c "(healthy)" || echo "0")

        if [ "$healthy" -ge "$expected" ]; then
            log_ok "All services healthy"
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

# bootstrap_admin creates the initial admin user via Better Auth's sign-up
# endpoint so the operator does not have to register through the browser.
# Skipped when ADMIN_EMAIL is empty. Treats "user already exists" as success
# so re-running the installer is idempotent.
ADMIN_BOOTSTRAPPED=false
bootstrap_admin() {
    if [ -z "${ADMIN_EMAIL:-}" ] || [ -z "${ADMIN_PASSWORD:-}" ]; then
        return 0
    fi

    local url="$BASE_URL/api/auth/sign-up/email"
    local name="${ADMIN_EMAIL%@*}"
    local payload
    payload=$(printf '{"email":"%s","password":"%s","name":"%s"}' \
        "$(json_escape "$ADMIN_EMAIL")" \
        "$(json_escape "$ADMIN_PASSWORD")" \
        "$(json_escape "$name")")

    log_info "Bootstrapping admin account ($ADMIN_EMAIL)"

    local timeout="${ADMIN_BOOTSTRAP_TIMEOUT:-60}" elapsed=0 interval=3
    local http_code="000" body code tmp
    tmp=$(mktemp)
    while [ "$elapsed" -lt "$timeout" ]; do
        http_code=$(curl -sS -k -o "$tmp" -w "%{http_code}" \
            --connect-timeout 5 --max-time 15 \
            -X POST "$url" \
            -H "Content-Type: application/json" \
            -H "Origin: $BASE_URL" \
            -d "$payload" 2>/dev/null || echo "000")

        body=$(cat "$tmp" 2>/dev/null || true)
        # Avoid pipefail errexit when the body has no "code" (e.g. HTML from wrong route).
        code=$(printf '%s' "$body" | grep -oE '"code"[[:space:]]*:[[:space:]]*"[^"]+"' | head -n1 | sed -E 's/.*"code"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/' || true)

        case "$http_code" in
            200|201)
                rm -f "$tmp"
                log_ok "Admin account created"
                ADMIN_BOOTSTRAPPED=true
                return 0
                ;;
            422)
                if [ "$code" = "USER_ALREADY_EXISTS_USE_ANOTHER_EMAIL" ] || [ "$code" = "USER_ALREADY_EXISTS" ]; then
                    rm -f "$tmp"
                    log_info "Admin account already exists — skipping bootstrap"
                    return 0
                fi
                ;;
            400|403)
                # 403 with "Registration is disabled" (no code field) means the
                # self-hosted guard blocked sign-up because the auth service
                # already seeded the admin account on startup. Treat as success.
                if printf '%s' "$body" | grep -q '"Registration is disabled"'; then
                    rm -f "$tmp"
                    log_info "Admin account seeded by auth service — skipping HTTP bootstrap"
                    ADMIN_BOOTSTRAPPED=true
                    return 0
                fi
                case "$code" in
                    EMAIL_PASSWORD_SIGN_UP_DISABLED)
                        rm -f "$tmp"
                        log_warn "Email/password sign-up is disabled in the auth service — cannot bootstrap admin"
                        return 0
                        ;;
                    INVALID_ORIGIN|MISSING_OR_NULL_ORIGIN|CROSS_SITE_NAVIGATION_LOGIN_BLOCKED)
                        rm -f "$tmp"
                        log_warn "Auth service rejected origin '$BASE_URL' ($code). Check ALLOWED_ORIGIN in $NIXOPUS_HOME/.env"
                        return 0
                        ;;
                    PASSWORD_TOO_SHORT|PASSWORD_TOO_LONG|INVALID_PASSWORD|INVALID_EMAIL)
                        rm -f "$tmp"
                        log_warn "Auth service rejected credentials ($code). Adjust ADMIN_EMAIL / ADMIN_PASSWORD and re-run: nixopus admin-bootstrap"
                        return 0
                        ;;
                esac
                ;;
        esac

        sleep "$interval"
        elapsed=$((elapsed + interval))
    done

    rm -f "$tmp"
    if [ -n "$code" ]; then
        log_warn "Could not auto-create admin (HTTP $http_code, code: $code). Retry: nixopus admin-bootstrap"
    else
        log_warn "Could not auto-create admin (HTTP $http_code). Retry: nixopus admin-bootstrap"
    fi
    [ -n "$body" ] && log_warn "Response: $(printf '%s' "$body" | head -c 200)"
    return 0
}

# install_management_script installs the standalone nixopus.sh CLI from the
# local installer dir (when running tests) or downloads it from the repo.
install_management_script() {
    local target="/usr/local/bin/nixopus"
    local src="${NIXOPUS_INSTALLER_DIR:-}/nixopus.sh"

    if [ -n "${NIXOPUS_INSTALLER_DIR:-}" ] && [ -f "$src" ]; then
        cp "$src" "$target"
    else
        curl -fsSL "$REPO_RAW/nixopus.sh" -o "$target" \
            || fail "Failed to download management CLI from $REPO_RAW/nixopus.sh"
    fi

    chmod +x "$target"
    log_ok "Management CLI installed: nixopus"
}

send_telemetry() {
    local event="${1:-install_success}" error="${2:-}"
    [ "${NIXOPUS_TELEMETRY:-on}" = "off" ] && return 0
    [ "${DO_NOT_TRACK:-0}" = "1" ] && return 0

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

# Public web URL for humans (duplicates AUTH_PUBLIC_URL; used by completion output).
resolve_access_url() {
    local u="${BASE_URL:-${AUTH_PUBLIC_URL:-${ALLOWED_ORIGIN:-}}}"
    if [ -n "$u" ]; then
        printf '%s' "$u"
        return 0
    fi
    [ -f "${NIXOPUS_HOME:-}/.env" ] || return 1
    local fallback=""
    while IFS= read -r line || [ -n "$line" ]; do
        case "$line" in
            BASE_URL=*)
                u="${line#BASE_URL=}"
                [ -n "$u" ] && { printf '%s' "$u"; return 0; }
                ;;
            AUTH_PUBLIC_URL=*)
                [ -z "$fallback" ] && fallback="${line#AUTH_PUBLIC_URL=}"
                ;;
        esac
    done < "$NIXOPUS_HOME/.env"
    [ -n "$fallback" ] && { printf '%s' "$fallback"; return 0; }
    return 1
}

show_banner() {
    echo ""
    echo -e "${BOLD}  Nixopus Self-Host Installer v${NIXOPUS_VERSION}${NC}"
    if [ -n "$PREVIEW_PR" ]; then
        echo -e "  ${YELLOW}PREVIEW MODE — PR #${PREVIEW_PR}${NC}"
    elif [ -n "$PREVIEW_BRANCH" ] || [ -n "$PREVIEW_FORK" ]; then
        echo -e "  ${YELLOW}PREVIEW MODE — ${PREVIEW_FORK:-${NIXOPUS_REPO}}@${PREVIEW_BRANCH:-${NIXOPUS_BRANCH}}${NC}"
    fi
    echo -e "  ${DIM}https://nixopus.com${NC}"
    echo ""
}

finalize_install_session_log() {
    [ -n "${INSTALL_LOG:-}" ] && [ -f "$INSTALL_LOG" ] || return 0
    [ -d "${NIXOPUS_HOME:-}" ] || return 0
    if cp -f "$INSTALL_LOG" "$NIXOPUS_HOME/install.log" 2>/dev/null; then
        chmod 600 "$NIXOPUS_HOME/install.log" 2>/dev/null || true
        rm -f "$INSTALL_LOG" 2>/dev/null || true
    fi
}

show_complete() {
    local duration=$(( $(date +%s) - INSTALL_START ))
    local access_url
    access_url="$(resolve_access_url || true)"
    if [ -z "$access_url" ]; then
        log_warn "Could not determine access URL. Run: nixopus info"
        access_url="(run: nixopus info)"
    fi
    echo ""
    echo -e "  ${GREEN}${BOLD}Nixopus is running!${NC}"
    echo ""
    echo -e "  ${BOLD}Access:${NC}    $access_url"
    echo -e "  ${BOLD}Config:${NC}    $NIXOPUS_HOME/.env"
    echo -e "  ${BOLD}Installed:${NC} ${duration}s"
    if [ -n "${ADMIN_EMAIL:-}" ] && [ "${ADMIN_BOOTSTRAPPED:-false}" = true ]; then
        echo ""
        echo -e "  ${BOLD}Admin credentials:${NC}"
        echo -e "    ${BOLD}Email:${NC}    $ADMIN_EMAIL"
        echo -ne "    ${BOLD}Password:${NC} "
        printf '%s\n' "$ADMIN_PASSWORD"
        if [ "${ADMIN_PASSWORD_GENERATED:-false}" = true ]; then
            echo -e "    ${YELLOW}Auto-generated. Save it now and change it after first login.${NC}"
        fi
        echo -e "    ${DIM}Also stored in $NIXOPUS_HOME/.env (ADMIN_PASSWORD)${NC}"
    elif [ -n "${ADMIN_EMAIL:-}" ]; then
        echo ""
        echo -e "  ${BOLD}Admin:${NC}     register at ${access_url}/register"
    fi
    if [ -n "$PREVIEW_PR" ] || [ -n "$PREVIEW_BRANCH" ] || [ -n "$PREVIEW_FORK" ]; then
        echo ""
        echo -e "  ${YELLOW}${BOLD}Preview images:${NC}"
        if [ -n "${NIXOPUS_API_IMAGE:-}" ]; then
            echo -e "    ${DIM}API:  ${NIXOPUS_API_IMAGE}${NC}"
        elif [ -n "$PREVIEW_PR" ]; then
            echo -e "    ${DIM}API:  latest (pr-${PREVIEW_PR} not available)${NC}"
        fi
        if [ -n "${NIXOPUS_VIEW_IMAGE:-}" ]; then
            echo -e "    ${DIM}View: ${NIXOPUS_VIEW_IMAGE}${NC}"
        elif [ -n "$PREVIEW_PR" ]; then
            echo -e "    ${DIM}View: latest (pr-${PREVIEW_PR} not available)${NC}"
        fi
        echo -e "  ${YELLOW}To revert to stable: re-run the installer without --preview${NC}"
    fi
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

    if [ "${NIXOPUS_TELEMETRY:-on}" != "off" ] && [ "${DO_NOT_TRACK:-0}" != "1" ]; then
        echo -e "  ${DIM}Anonymous telemetry enabled. Disable: NIXOPUS_TELEMETRY=off or DO_NOT_TRACK=1${NC}"
    fi
    echo -e "  ${DIM}Problems? Attach $NIXOPUS_HOME/install.log or run: sudo nixopus report${NC}"
    echo ""
    finalize_install_session_log
}

main() {
    start_install_session_log
    show_banner
    send_telemetry "install_started" || true

    log_step 1 "Checking requirements"
    check_root
    detect_os
    install_prereqs

    log_step 2 "Setting up Docker"
    install_docker
    apply_preview_config

    log_step 3 "Configuring"
    load_existing_config
    gather_config

    log_step 4 "Writing files"
    setup_directories
    setup_ssh
    write_files

    log_step 5 "Starting services"
    start_services
    bootstrap_admin

    log_step 6 "Installing management CLI"
    install_management_script

    show_complete
    send_telemetry "install_success" || true
}

main "$@"
