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
        dc up -d --no-deps "$service"
    else
        dc up -d --remove-orphans
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
    echo "Host IP6:     ${HOST_IP6:-<none>}"
    echo "IP Family:    ${IP_FAMILY:-ipv4}"
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
        local provider="${LLM_PROVIDER:-openrouter}"
        local key=""
        case "$provider" in
            openrouter) key="${OPENROUTER_API_KEY:-}" ;;
            openai)     key="${OPENAI_API_KEY:-}" ;;
            anthropic)  key="${ANTHROPIC_API_KEY:-}" ;;
            google)     key="${GOOGLE_GENERATIVE_AI_API_KEY:-}" ;;
            deepseek)   key="${DEEPSEEK_API_KEY:-}" ;;
            groq)       key="${GROQ_API_KEY:-}" ;;
        esac
        if [ -n "$key" ]; then
            echo "LLM:          $provider ($(redact "$key"))"
        else
            echo "LLM:          $provider (key not set)"
        fi
        [ -n "${AGENT_MODEL:-}" ] && echo "Model:        ${AGENT_MODEL}"
    fi
}

cmd_domain() {
    local action="" domain="" ip_flag="" verify_mode="auto"

    while [ $# -gt 0 ]; do
        case "$1" in
            add|remove) action="$1" ;;
            --ipv4)     ip_flag="ipv4" ;;
            --ipv6)     ip_flag="ipv6" ;;
            --verify=*) verify_mode="${1#--verify=}" ;;
            -*)         echo "Unknown flag: $1" >&2; return 1 ;;
            *)          [ -z "$domain" ] && domain="$1" ;;
        esac
        shift
    done

    case "$verify_mode" in
        auto|strict|off) ;;
        *) echo "Unknown verify mode: $verify_mode (use auto, strict, or off)" >&2; return 1 ;;
    esac

    if [ "$action" != "add" ] && [ "$action" != "remove" ]; then
        echo "Usage: nixopus domain add <domain> [--ipv4|--ipv6] [--verify=auto|strict|off]"
        echo "       nixopus domain remove"
        return 1
    fi

    load_env

    if [ "$action" = "add" ]; then
        if [ -z "$domain" ]; then
            echo "Usage: nixopus domain add <domain> [--ipv4|--ipv6] [--verify=auto|strict|off]" >&2
            return 1
        fi

        local effective_family="${ip_flag:-${IP_FAMILY:-ipv4}}"

        if [ "$verify_mode" != "off" ]; then
            local mismatch=false

            if command -v getent >/dev/null 2>&1; then
                local resolved_v4="" resolved_v6=""
                resolved_v4=$(getent hosts "$domain" 2>/dev/null | awk '$1 !~ /:/ {print $1; exit}') || true
                resolved_v6=$(getent hosts "$domain" 2>/dev/null | awk '/:/{print $1; exit}') || true

                case "$effective_family" in
                    ipv4)
                        if [ -n "$resolved_v4" ] && [ "$resolved_v4" != "${HOST_IP:-}" ]; then
                            echo "WARNING: $domain resolves to $resolved_v4 but HOST_IP is ${HOST_IP:-<unset>}"
                            mismatch=true
                        fi
                        ;;
                    ipv6)
                        local effective_v6="${HOST_IP6:-${HOST_IP:-}}"
                        if [ -n "$resolved_v6" ] && [ "$resolved_v6" != "$effective_v6" ]; then
                            echo "WARNING: $domain resolves to $resolved_v6 but host IPv6 is ${effective_v6:-<unset>}"
                            mismatch=true
                        fi
                        ;;
                    dual)
                        if [ -n "$resolved_v4" ] && [ -n "${HOST_IP:-}" ] && [ "$resolved_v4" != "$HOST_IP" ]; then
                            echo "WARNING: $domain A record resolves to $resolved_v4 but HOST_IP is $HOST_IP"
                            mismatch=true
                        fi
                        if [ -n "$resolved_v6" ] && [ -n "${HOST_IP6:-}" ] && [ "$resolved_v6" != "$HOST_IP6" ]; then
                            echo "WARNING: $domain AAAA record resolves to $resolved_v6 but HOST_IP6 is $HOST_IP6"
                            mismatch=true
                        fi
                        ;;
                esac
            else
                echo "Note: 'getent' not available, skipping DNS verification"
            fi

            if [ "$mismatch" = true ] && [ "$verify_mode" = "strict" ]; then
                echo "Aborting (--verify=strict). Fix DNS or use --verify=off." >&2
                return 1
            fi
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
        case "$effective_family" in
            ipv4) echo "Ensure DNS A record for $domain points to ${HOST_IP:-your server IP}" ;;
            ipv6) echo "Ensure DNS AAAA record for $domain points to ${HOST_IP6:-${HOST_IP:-your server IP}}" ;;
            dual)
                [ -n "${HOST_IP:-}" ] && echo "Ensure DNS A record for $domain points to $HOST_IP"
                [ -n "${HOST_IP6:-}" ] && echo "Ensure DNS AAAA record for $domain points to $HOST_IP6"
                ;;
        esac
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
    local action="" new_ip="" ip_flag=""

    while [ $# -gt 0 ]; do
        case "$1" in
            set)    action="set" ;;
            --ipv4) ip_flag="ipv4" ;;
            --ipv6) ip_flag="ipv6" ;;
            -*)     echo "Unknown flag: $1" >&2; return 1 ;;
            *)      [ -z "$new_ip" ] && new_ip="$1" ;;
        esac
        shift
    done

    load_env

    if [ "$action" != "set" ] || [ -z "$new_ip" ]; then
        echo "── IP Configuration ──"
        echo "Host IP:    ${HOST_IP:-<unknown>}"
        echo "Host IP6:   ${HOST_IP6:-<none>}"
        echo "IP Family:  ${IP_FAMILY:-ipv4}"
        echo "Access:     ${ALLOWED_ORIGIN:-unknown}"
        echo ""
        echo "Usage: nixopus ip set [--ipv4|--ipv6] <ip>"
        return
    fi

    local target_var="HOST_IP"
    if [ -n "$ip_flag" ]; then
        [ "$ip_flag" = "ipv6" ] && target_var="HOST_IP6"
    elif [[ "$new_ip" == *:* ]]; then
        target_var="HOST_IP6"
    fi

    if grep -q "^${target_var}=" "$NIXOPUS_HOME/.env"; then
        sedi "s|^${target_var}=.*|${target_var}=${new_ip}|" "$NIXOPUS_HOME/.env"
    else
        echo "${target_var}=${new_ip}" >> "$NIXOPUS_HOME/.env"
    fi

    if ! grep -q "^IP_FAMILY=" "$NIXOPUS_HOME/.env"; then
        echo "IP_FAMILY=ipv4" >> "$NIXOPUS_HOME/.env"
    fi

    if [ "$target_var" = "HOST_IP" ] && [ -z "${DOMAIN:-}" ]; then
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
    echo "Set $target_var to $new_ip"
    if [ "$target_var" = "HOST_IP6" ] && [ "$(grep '^IP_FAMILY=' "$NIXOPUS_HOME/.env" | cut -d= -f2)" = "ipv4" ]; then
        echo "Note: IP_FAMILY is still 'ipv4'. Run 'nixopus config set IP_FAMILY=dual' to enable dual-stack."
    fi
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
    echo "  domain add <domain> [--ipv4|--ipv6] [--verify=auto|strict|off]"
    echo "                      Switch to domain-based HTTPS"
    echo "  domain remove       Switch back to IP-based HTTP"
    echo "  ip set [--ipv4|--ipv6] <ip>"
    echo "                      Change host IP"
    echo "  port                Show port configuration"
    echo "  port set <type> <n> Change port (http, https, ssh)"
    echo "  backup              Backup database and config"
    echo "  info                Show install info and status"
    echo "  admin-bootstrap     Create admin account via auth API (uses ADMIN_EMAIL/ADMIN_PASSWORD)"
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
