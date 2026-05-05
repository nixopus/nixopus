#!/usr/bin/env bats
# Unit tests for installer/nixopus.sh
# Coverage target: 100% of all commands and business-logic branches.
#
# Strategy: source nixopus.sh with NIXOPUS_HOME set to a temp dir that
# contains a real .env so every load_env / dc call is intercepted by stubs.

INSTALLER_DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")/../.." && pwd)"
NIXOPUS_SH="$INSTALLER_DIR/nixopus.sh"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

# Create a minimal .env in a temp NIXOPUS_HOME
# IMPORTANT: NIXOPUS_HOME in the .env must point to $1 so that load_env()
# does not revert it to /opt/nixopus after the test sets NIXOPUS_HOME.
make_env() {
    local home="$1"
    mkdir -p "$home/ssh"
    cat > "$home/.env" <<EOF
NIXOPUS_VERSION=0.3.0
NIXOPUS_HOME=${home}
DOMAIN=
SITE_ADDRESS=:80
HOST_IP=1.2.3.4
CADDY_HTTP_PORT=80
CADDY_HTTPS_PORT=443
SSH_HOST=1.2.3.4
SSH_PORT=22
SSH_USER=root
DATABASE_URL=postgres://nixopus:dbpass@nixopus-db:5432/nixopus?sslmode=disable
REDIS_URL=redis://default:redispass@nixopus-redis:6379
DB_PASSWORD=dbpass1234567890dbpass1234567890
REDIS_PASSWORD=redispass1234567890redispass1234
AUTH_SERVICE_SECRET=authsecret1234567890authsecret1234567890authsecret12345678
JWT_SECRET=jwtsecret1234567890jwtsecret1234567890jwtsecret123456789
ALLOWED_ORIGIN=http://1.2.3.4
BASE_URL=http://1.2.3.4
API_URL=http://1.2.3.4/api
AUTH_PUBLIC_URL=http://1.2.3.4
AUTH_COOKIE_DOMAIN=
AUTH_SECURE_COOKIES=false
USE_BUNDLED_DB=true
USE_BUNDLED_REDIS=true
ADMIN_EMAIL=admin@test.com
ADMIN_PASSWORD=Abc1!TestPass9
SELF_HOSTED=true
NIXOPUS_TELEMETRY=off
LOG_LEVEL=debug
USE_AGENT=true
LLM_PROVIDER=openrouter
OPENROUTER_API_KEY=sk-or-test
OPENAI_API_KEY=
ANTHROPIC_API_KEY=
GOOGLE_GENERATIVE_AI_API_KEY=
DEEPSEEK_API_KEY=
GROQ_API_KEY=
AGENT_MODEL=openrouter/anthropic/claude-sonnet-4
AGENT_LIGHT_MODEL=openrouter/openai/gpt-4o-mini
NIXOPUS_API_IMAGE=
NIXOPUS_VIEW_IMAGE=
NIXOPUS_AUTH_IMAGE=
NIXOPUS_AGENT_IMAGE=
EOF
    chmod 600 "$home/.env"
    touch "$home/docker-compose.yml"
    touch "$home/docker-compose.db.yml"
    touch "$home/docker-compose.redis.yml"
    touch "$home/docker-compose.agent.yml"
}

# Source nixopus.sh with the root check and case-dispatch disabled so we can call functions directly
load_nixopus() {
    local tmp
    tmp=$(mktemp)
    # Strip top-level guards and the trailing case dispatch so we can call functions directly:
    #   1) root check  (id -u != 0 → exit 1)
    #   2) .env check  ([ ! -f "$NIXOPUS_HOME/.env" ] → exit 1)
    #   3) case dispatch
    awk '
        index($0, "(id -u)") { in_root=1 }
        in_root && /^fi$/ { in_root=0; next }
        in_root { next }
        index($0, "[ ! -f \"$NIXOPUS_HOME") { in_env=1 }
        in_env && /^fi$/ { in_env=0; next }
        in_env { next }
        index($0, "case \"${1:-help}\"") { in_case=1 }
        !in_case { print }
    ' "$NIXOPUS_SH" > "$tmp"
    # shellcheck disable=SC1090
    source "$tmp"
    rm -f "$tmp"
}

# ---------------------------------------------------------------------------
# sedi
# ---------------------------------------------------------------------------

@test "sedi: replaces content in file" {
    load_nixopus
    local f
    f=$(mktemp)
    echo "KEY=old" > "$f"
    sedi "s|^KEY=.*|KEY=new|" "$f"
    grep -q "^KEY=new" "$f"
    rm -f "$f"
}

# ---------------------------------------------------------------------------
# redact
# ---------------------------------------------------------------------------

@test "redact: masks middle of a string" {
    load_nixopus
    result=$(redact "abcdefghijklmn")
    # first 4 + **** + last 4
    [[ "$result" == "abcd****klmn" ]]
}

@test "redact: short secret still redacts" {
    load_nixopus
    result=$(redact "1234567890123456")
    [[ "$result" == "1234****3456" ]]
}

# ---------------------------------------------------------------------------
# format_ip_for_url (nixopus.sh copy)
# ---------------------------------------------------------------------------

@test "format_ip_for_url (nixopus): IPv6 wrapped in brackets" {
    load_nixopus
    result=$(format_ip_for_url "2001:db8::1")
    [ "$result" = "[2001:db8::1]" ]
}

@test "format_ip_for_url (nixopus): IPv4 unchanged" {
    load_nixopus
    result=$(format_ip_for_url "10.0.0.1")
    [ "$result" = "10.0.0.1" ]
}

# ---------------------------------------------------------------------------
# compose_files (nixopus.sh version)
# ---------------------------------------------------------------------------

@test "compose_files (nixopus.sh): includes base file" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    USE_BUNDLED_DB=true USE_BUNDLED_REDIS=true USE_AGENT=true
    result=$(compose_files)
    [[ "$result" == *"docker-compose.yml"* ]]
    rm -rf "$tmp_home"
}

@test "compose_files (nixopus.sh): omits db when USE_BUNDLED_DB=false" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    USE_BUNDLED_DB=false USE_BUNDLED_REDIS=true USE_AGENT=false
    result=$(compose_files)
    [[ "$result" != *"docker-compose.db.yml"* ]]
    rm -rf "$tmp_home"
}

@test "compose_files (nixopus.sh): omits redis when USE_BUNDLED_REDIS=false" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    USE_BUNDLED_DB=true USE_BUNDLED_REDIS=false USE_AGENT=false
    result=$(compose_files)
    [[ "$result" != *"docker-compose.redis.yml"* ]]
    rm -rf "$tmp_home"
}

@test "compose_files (nixopus.sh): omits agent when USE_AGENT=false" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    USE_BUNDLED_DB=false USE_BUNDLED_REDIS=false USE_AGENT=false
    result=$(compose_files)
    [[ "$result" != *"docker-compose.agent.yml"* ]]
    rm -rf "$tmp_home"
}

@test "compose_files (nixopus.sh): includes agent when USE_AGENT=true and file exists" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    USE_BUNDLED_DB=false USE_BUNDLED_REDIS=false USE_AGENT=true
    result=$(compose_files)
    [[ "$result" == *"docker-compose.agent.yml"* ]]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# sync_view_llm_env
# ---------------------------------------------------------------------------

@test "sync_view_llm_env: writes view-llm.env with model vars" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    AGENT_MODEL="openrouter/anthropic/claude-sonnet-4"
    AGENT_LIGHT_MODEL="openrouter/openai/gpt-4o-mini"
    LLM_PROVIDER="openrouter"

    sync_view_llm_env
    [ -f "$tmp_home/view-llm.env" ]
    grep -q "^AGENT_MODEL=" "$tmp_home/view-llm.env"
    grep -q "^LLM_PROVIDER=" "$tmp_home/view-llm.env"
    rm -rf "$tmp_home"
}

@test "sync_view_llm_env: view-llm.env has mode 600" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    AGENT_MODEL="" AGENT_LIGHT_MODEL="" LLM_PROVIDER="openrouter"

    sync_view_llm_env
    perms=$(stat -c "%a" "$tmp_home/view-llm.env" 2>/dev/null || stat -f "%Lp" "$tmp_home/view-llm.env")
    [ "$perms" = "600" ]
    rm -rf "$tmp_home"
}

@test "sync_view_llm_env: no-op when .env missing" {
    load_nixopus
    NIXOPUS_HOME="/nonexistent-$(date +%s)"
    run sync_view_llm_env
    [ "$status" -eq 0 ]
}

# ---------------------------------------------------------------------------
# cmd_status
# ---------------------------------------------------------------------------

@test "cmd_status: calls dc ps with table format" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    dc() { echo "dc $*"; }
    run cmd_status
    [[ "$output" == *"ps"* ]]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# cmd_logs
# ---------------------------------------------------------------------------

@test "cmd_logs: without service tails all logs" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    dc() { echo "dc $*"; }
    run cmd_logs
    [[ "$output" == *"logs"* ]]
    [[ "$output" != *"nixopus-api"* ]]
    rm -rf "$tmp_home"
}

@test "cmd_logs: with service name passes it to dc logs" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    dc() { echo "dc $*"; }
    run cmd_logs "nixopus-api"
    [[ "$output" == *"nixopus-api"* ]]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# cmd_update
# ---------------------------------------------------------------------------

@test "cmd_update: pulls and restarts services" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    _pull=false _up=false _ps=false
    dc() {
        case "$1" in
            pull) _pull=true ;;
            up)   _up=true ;;
            ps)   _ps=true ;;
        esac
    }
    cmd_update
    [ "$_pull" = "true" ]
    [ "$_up" = "true" ]
    [ "$_ps" = "true" ]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# cmd_restart
# ---------------------------------------------------------------------------

@test "cmd_restart: without service restarts all with remove-orphans" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    dc() { echo "dc $*"; }
    run cmd_restart
    [[ "$output" == *"remove-orphans"* ]]
    rm -rf "$tmp_home"
}

@test "cmd_restart: with service passes --no-deps" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    dc() { echo "dc $*"; }
    run cmd_restart "nixopus-api"
    [[ "$output" == *"nixopus-api"* ]]
    [[ "$output" == *"no-deps"* ]]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# cmd_stop
# ---------------------------------------------------------------------------

@test "cmd_stop: stops all services" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    dc() { echo "dc $*"; }
    run cmd_stop
    [[ "$output" == *"stop"* ]]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# cmd_uninstall
# ---------------------------------------------------------------------------

@test "cmd_uninstall: aborts without confirmation" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    dc() { echo "dc $*"; }
    # Use <<< to feed "n" as stdin without spawning a subshell (preserves var scope)
    run cmd_uninstall <<< "n"
    [[ "$output" == *"Cancelled"* ]]
    rm -rf "$tmp_home"
}

@test "cmd_uninstall: proceeds with --purge and confirmation" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    _dc_down=false
    dc() {
        case "$1" in
            down) _dc_down=true ;;
        esac
    }
    # Override rm to avoid writing to /usr/local/bin
    rm() { return 0; }

    cmd_uninstall "--purge" <<< "y"
    [ "$_dc_down" = "true" ]
    rm -rf "$tmp_home"
}

@test "cmd_uninstall: proceeds without --purge on confirmation" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    _dc_down=false
    dc() {
        case "$1" in
            down) _dc_down=true ;;
        esac
    }
    rm() { return 0; }

    cmd_uninstall <<< "y"
    [ "$_dc_down" = "true" ]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# cmd_config
# ---------------------------------------------------------------------------

@test "cmd_config: shows config values without args" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    run cmd_config
    [[ "$output" == *"Nixopus Configuration"* ]]
    rm -rf "$tmp_home"
}

@test "cmd_config set: updates existing key in .env" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    sync_view_llm_env() { return 0; }

    cmd_config set "LOG_LEVEL=info"
    grep -q "^LOG_LEVEL=info" "$tmp_home/.env"
    rm -rf "$tmp_home"
}

@test "cmd_config set: appends new key to .env" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    sync_view_llm_env() { return 0; }

    cmd_config set "NEW_KEY=newval"
    grep -q "^NEW_KEY=newval" "$tmp_home/.env"
    rm -rf "$tmp_home"
}

@test "cmd_config: shows LLM key info when key set" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    run cmd_config
    [[ "$output" == *"openrouter"* ]]
    rm -rf "$tmp_home"
}

@test "cmd_config: shows agent model when AGENT_MODEL set" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    # Add AGENT_MODEL to env
    echo "AGENT_MODEL=openrouter/anthropic/claude-sonnet-4" >> "$tmp_home/.env"
    NIXOPUS_HOME="$tmp_home"

    run cmd_config
    [[ "$output" == *"Model"* ]]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# cmd_domain add
# ---------------------------------------------------------------------------

@test "cmd_domain add: updates DOMAIN and related keys" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    dc() { return 0; }

    cmd_domain add "nixopus.example.com"
    grep -q "^DOMAIN=nixopus.example.com" "$tmp_home/.env"
    grep -q "^AUTH_SECURE_COOKIES=true" "$tmp_home/.env"
    grep -q "^ALLOWED_ORIGIN=https://nixopus.example.com" "$tmp_home/.env"
    rm -rf "$tmp_home"
}

@test "cmd_domain add: cookie domain uses eTLD+1" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    dc() { return 0; }

    cmd_domain add "app.nixopus.example.com"
    grep -q "^AUTH_COOKIE_DOMAIN=.example.com" "$tmp_home/.env"
    rm -rf "$tmp_home"
}

@test "cmd_domain add: fails when domain argument missing" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    run cmd_domain add ""
    [ "$status" -ne 0 ]
    rm -rf "$tmp_home"
}

@test "cmd_domain remove: switches back to IP mode" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    dc() { return 0; }

    cmd_domain remove
    grep -q "^DOMAIN=$" "$tmp_home/.env"
    grep -q "^AUTH_SECURE_COOKIES=false" "$tmp_home/.env"
    grep -q "^AUTH_COOKIE_DOMAIN=$" "$tmp_home/.env"
    rm -rf "$tmp_home"
}

@test "cmd_domain remove: URL uses port when HTTP port not 80" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    # Set non-default port
    sed -i.bak 's/^CADDY_HTTP_PORT=80/CADDY_HTTP_PORT=8080/' "$tmp_home/.env"
    NIXOPUS_HOME="$tmp_home"
    dc() { return 0; }

    cmd_domain remove
    grep -q "^ALLOWED_ORIGIN=http://1.2.3.4:8080" "$tmp_home/.env"
    rm -rf "$tmp_home"
}

@test "cmd_domain: returns 1 for unknown action" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    run cmd_domain "invalid"
    [ "$status" -ne 0 ]
    rm -rf "$tmp_home"
}

@test "cmd_domain: returns 1 with no action" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    run cmd_domain
    [ "$status" -ne 0 ]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# cmd_ip
# ---------------------------------------------------------------------------

@test "cmd_ip set: updates HOST_IP in .env" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    dc() { return 0; }

    cmd_ip set "9.9.9.9"
    grep -q "^HOST_IP=9.9.9.9" "$tmp_home/.env"
    rm -rf "$tmp_home"
}

@test "cmd_ip set: updates ALLOWED_ORIGIN when no domain" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    dc() { return 0; }

    cmd_ip set "9.9.9.9"
    grep -q "^ALLOWED_ORIGIN=http://9.9.9.9" "$tmp_home/.env"
    rm -rf "$tmp_home"
}

@test "cmd_ip set: does not update ALLOWED_ORIGIN when domain is set" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    # Set a domain
    echo "DOMAIN=nixopus.example.com" >> "$tmp_home/.env"
    sed -i.bak 's/^DOMAIN=$/DOMAIN=nixopus.example.com/' "$tmp_home/.env"
    NIXOPUS_HOME="$tmp_home"
    dc() { return 0; }

    cmd_ip set "9.9.9.9"
    # ALLOWED_ORIGIN should still be https://nixopus.example.com
    grep -qv "^ALLOWED_ORIGIN=http://9.9.9.9" "$tmp_home/.env" || true
    rm -rf "$tmp_home"
}

@test "cmd_ip: shows current IP when called without set" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    run cmd_ip
    [[ "$output" == *"IP Configuration"* ]]
    rm -rf "$tmp_home"
}

@test "cmd_ip: shows current IP when called with set but no IP" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    run cmd_ip set
    [[ "$output" == *"IP Configuration"* ]]
    rm -rf "$tmp_home"
}

@test "cmd_ip set: includes port in URL when http port not 80" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    sed -i.bak 's/^CADDY_HTTP_PORT=80/CADDY_HTTP_PORT=8080/' "$tmp_home/.env"
    NIXOPUS_HOME="$tmp_home"
    dc() { return 0; }

    cmd_ip set "5.5.5.5"
    grep -q "^ALLOWED_ORIGIN=http://5.5.5.5:8080" "$tmp_home/.env"
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# cmd_port
# ---------------------------------------------------------------------------

@test "cmd_port set http: updates CADDY_HTTP_PORT" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    dc() { return 0; }

    cmd_port set http 8080
    grep -q "^CADDY_HTTP_PORT=8080" "$tmp_home/.env"
    rm -rf "$tmp_home"
}

@test "cmd_port set https: updates CADDY_HTTPS_PORT" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    dc() { return 0; }

    cmd_port set https 9443
    grep -q "^CADDY_HTTPS_PORT=9443" "$tmp_home/.env"
    rm -rf "$tmp_home"
}

@test "cmd_port set ssh: updates SSH_PORT" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    dc() { return 0; }

    cmd_port set ssh 2222
    grep -q "^SSH_PORT=2222" "$tmp_home/.env"
    rm -rf "$tmp_home"
}

@test "cmd_port set: unknown type returns 1" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    run cmd_port set ftp 21
    [ "$status" -ne 0 ]
    rm -rf "$tmp_home"
}

@test "cmd_port: shows current ports when no args" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    run cmd_port
    [[ "$output" == *"Port Configuration"* ]]
    rm -rf "$tmp_home"
}

@test "cmd_port: shows usage when set called with missing args" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    run cmd_port set
    [[ "$output" == *"Port Configuration"* ]]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# cmd_backup
# ---------------------------------------------------------------------------

@test "cmd_backup: creates backup directory and backs up .env" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    docker() {
        if [[ "$*" == *"exec"* ]]; then return 1; fi
    }

    cmd_backup
    # env.bak is inside the timestamped subdirectory
    find "$tmp_home/backups" -name "env.bak" | grep -q "env.bak"
    rm -rf "$tmp_home"
}

@test "cmd_backup: creates backup dir with timestamp prefix" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    docker() { return 1; }

    cmd_backup
    [ -d "$tmp_home/backups" ]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# cmd_info
# ---------------------------------------------------------------------------

@test "cmd_info: shows version and home" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    dc() { echo "dc $*"; }
    run cmd_info
    [[ "$output" == *"Nixopus"* ]]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# cmd_admin_bootstrap
# ---------------------------------------------------------------------------

@test "cmd_admin_bootstrap: fails when ADMIN_EMAIL missing" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    # Remove ADMIN_EMAIL from env
    sed -i.bak 's/^ADMIN_EMAIL=.*/ADMIN_EMAIL=/' "$tmp_home/.env"
    NIXOPUS_HOME="$tmp_home"

    run cmd_admin_bootstrap
    [ "$status" -ne 0 ]
    rm -rf "$tmp_home"
}

@test "cmd_admin_bootstrap: fails when ADMIN_PASSWORD missing" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    sed -i.bak 's/^ADMIN_PASSWORD=.*/ADMIN_PASSWORD=/' "$tmp_home/.env"
    NIXOPUS_HOME="$tmp_home"

    run cmd_admin_bootstrap
    [ "$status" -ne 0 ]
    rm -rf "$tmp_home"
}

@test "cmd_admin_bootstrap: fails when ALLOWED_ORIGIN missing" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    sed -i.bak 's/^ALLOWED_ORIGIN=.*/ALLOWED_ORIGIN=/' "$tmp_home/.env"
    sed -i.bak 's/^AUTH_PUBLIC_URL=.*/AUTH_PUBLIC_URL=/' "$tmp_home/.env"
    NIXOPUS_HOME="$tmp_home"

    run cmd_admin_bootstrap
    [ "$status" -ne 0 ]
    rm -rf "$tmp_home"
}

@test "cmd_admin_bootstrap: succeeds on HTTP 200" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    ADMIN_BOOTSTRAP_TIMEOUT=9

    curl() {
        local outfile=""
        local args=("$@")
        for i in "${!args[@]}"; do
            [[ "${args[$i]}" == "-o" ]] && outfile="${args[$((i+1))]}"
        done
        [ -n "$outfile" ] && echo '{"status":"ok"}' > "$outfile"
        echo "200"
    }

    run cmd_admin_bootstrap
    [ "$status" -eq 0 ]
    [[ "$output" == *"created"* ]]
    rm -rf "$tmp_home"
}

@test "cmd_admin_bootstrap: succeeds when user already exists" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    ADMIN_BOOTSTRAP_TIMEOUT=9

    curl() {
        local outfile=""
        local args=("$@")
        for i in "${!args[@]}"; do
            [[ "${args[$i]}" == "-o" ]] && outfile="${args[$((i+1))]}"
        done
        [ -n "$outfile" ] && echo '{"code":"USER_ALREADY_EXISTS"}' > "$outfile"
        echo "422"
    }

    run cmd_admin_bootstrap
    [ "$status" -eq 0 ]
    [[ "$output" == *"already exists"* ]]
    rm -rf "$tmp_home"
}

@test "cmd_admin_bootstrap: accepts email and password from CLI args" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    # Clear env values; pass via args
    sed -i.bak 's/^ADMIN_EMAIL=.*/ADMIN_EMAIL=/' "$tmp_home/.env"
    sed -i.bak 's/^ADMIN_PASSWORD=.*/ADMIN_PASSWORD=/' "$tmp_home/.env"
    NIXOPUS_HOME="$tmp_home"
    ADMIN_BOOTSTRAP_TIMEOUT=9

    curl() {
        local outfile=""
        local args=("$@")
        for i in "${!args[@]}"; do
            [[ "${args[$i]}" == "-o" ]] && outfile="${args[$((i+1))]}"
        done
        [ -n "$outfile" ] && echo '{"status":"ok"}' > "$outfile"
        echo "200"
    }

    run cmd_admin_bootstrap "cli@test.com" "Abc1!Pass1234"
    [ "$status" -eq 0 ]
    rm -rf "$tmp_home"
}

@test "cmd_admin_bootstrap: handles EMAIL_PASSWORD_SIGN_UP_DISABLED gracefully" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    ADMIN_BOOTSTRAP_TIMEOUT=9

    curl() {
        local outfile=""
        local args=("$@")
        for i in "${!args[@]}"; do
            [[ "${args[$i]}" == "-o" ]] && outfile="${args[$((i+1))]}"
        done
        [ -n "$outfile" ] && echo '{"code":"EMAIL_PASSWORD_SIGN_UP_DISABLED"}' > "$outfile"
        echo "400"
    }

    run cmd_admin_bootstrap
    [ "$status" -ne 0 ]
    rm -rf "$tmp_home"
}

@test "cmd_admin_bootstrap: handles INVALID_ORIGIN gracefully" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    ADMIN_BOOTSTRAP_TIMEOUT=9

    curl() {
        local outfile=""
        local args=("$@")
        for i in "${!args[@]}"; do
            [[ "${args[$i]}" == "-o" ]] && outfile="${args[$((i+1))]}"
        done
        [ -n "$outfile" ] && echo '{"code":"INVALID_ORIGIN"}' > "$outfile"
        echo "403"
    }

    run cmd_admin_bootstrap
    [ "$status" -ne 0 ]
    rm -rf "$tmp_home"
}

@test "cmd_admin_bootstrap: handles MISSING_OR_NULL_ORIGIN gracefully" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    ADMIN_BOOTSTRAP_TIMEOUT=9

    curl() {
        local outfile=""
        local args=("$@")
        for i in "${!args[@]}"; do
            [[ "${args[$i]}" == "-o" ]] && outfile="${args[$((i+1))]}"
        done
        [ -n "$outfile" ] && echo '{"code":"MISSING_OR_NULL_ORIGIN"}' > "$outfile"
        echo "403"
    }

    run cmd_admin_bootstrap
    [ "$status" -ne 0 ]
    rm -rf "$tmp_home"
}

@test "cmd_admin_bootstrap: handles CROSS_SITE_NAVIGATION_LOGIN_BLOCKED gracefully" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    ADMIN_BOOTSTRAP_TIMEOUT=9

    curl() {
        local outfile=""
        local args=("$@")
        for i in "${!args[@]}"; do
            [[ "${args[$i]}" == "-o" ]] && outfile="${args[$((i+1))]}"
        done
        [ -n "$outfile" ] && echo '{"code":"CROSS_SITE_NAVIGATION_LOGIN_BLOCKED"}' > "$outfile"
        echo "403"
    }

    run cmd_admin_bootstrap
    [ "$status" -ne 0 ]
    rm -rf "$tmp_home"
}

@test "cmd_admin_bootstrap: handles PASSWORD_TOO_SHORT gracefully" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    ADMIN_BOOTSTRAP_TIMEOUT=9

    curl() {
        local outfile=""
        local args=("$@")
        for i in "${!args[@]}"; do
            [[ "${args[$i]}" == "-o" ]] && outfile="${args[$((i+1))]}"
        done
        [ -n "$outfile" ] && echo '{"code":"PASSWORD_TOO_SHORT"}' > "$outfile"
        echo "400"
    }

    run cmd_admin_bootstrap
    [ "$status" -ne 0 ]
    rm -rf "$tmp_home"
}

@test "cmd_admin_bootstrap: handles PASSWORD_TOO_LONG gracefully" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    ADMIN_BOOTSTRAP_TIMEOUT=9

    curl() {
        local outfile=""
        local args=("$@")
        for i in "${!args[@]}"; do
            [[ "${args[$i]}" == "-o" ]] && outfile="${args[$((i+1))]}"
        done
        [ -n "$outfile" ] && echo '{"code":"PASSWORD_TOO_LONG"}' > "$outfile"
        echo "400"
    }

    run cmd_admin_bootstrap
    [ "$status" -ne 0 ]
    rm -rf "$tmp_home"
}

@test "cmd_admin_bootstrap: handles INVALID_PASSWORD gracefully" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    ADMIN_BOOTSTRAP_TIMEOUT=9

    curl() {
        local outfile=""
        local args=("$@")
        for i in "${!args[@]}"; do
            [[ "${args[$i]}" == "-o" ]] && outfile="${args[$((i+1))]}"
        done
        [ -n "$outfile" ] && echo '{"code":"INVALID_PASSWORD"}' > "$outfile"
        echo "400"
    }

    run cmd_admin_bootstrap
    [ "$status" -ne 0 ]
    rm -rf "$tmp_home"
}

@test "cmd_admin_bootstrap: handles INVALID_EMAIL gracefully" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    ADMIN_BOOTSTRAP_TIMEOUT=9

    curl() {
        local outfile=""
        local args=("$@")
        for i in "${!args[@]}"; do
            [[ "${args[$i]}" == "-o" ]] && outfile="${args[$((i+1))]}"
        done
        [ -n "$outfile" ] && echo '{"code":"INVALID_EMAIL"}' > "$outfile"
        echo "400"
    }

    run cmd_admin_bootstrap
    [ "$status" -ne 0 ]
    rm -rf "$tmp_home"
}

@test "cmd_admin_bootstrap: warns after timeout on repeated failure" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"
    ADMIN_BOOTSTRAP_TIMEOUT=3

    sleep() { : ; }
    curl() {
        local outfile=""
        local args=("$@")
        for i in "${!args[@]}"; do
            [[ "${args[$i]}" == "-o" ]] && outfile="${args[$((i+1))]}"
        done
        [ -n "$outfile" ] && echo '{}' > "$outfile"
        echo "503"
    }

    run cmd_admin_bootstrap
    [ "$status" -ne 0 ]
    [[ "$output" == *"Failed"* ]] || [[ "$output" == *"failed"* ]]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# redact_install_log_lines (nixopus.sh copy)
# ---------------------------------------------------------------------------

@test "redact_install_log_lines (nixopus.sh): redacts secrets" {
    load_nixopus
    result=$(echo "JWT_SECRET=mysecret" | redact_install_log_lines)
    [[ "$result" == *"<redacted>"* ]]
    [[ "$result" != *"mysecret"* ]]
}

@test "redact_install_log_lines (nixopus.sh): strips ANSI codes" {
    load_nixopus
    result=$(printf '\033[1;32mhello\033[0m' | redact_install_log_lines)
    [ "$result" = "hello" ]
}

@test "redact_install_log_lines (nixopus.sh): preserves non-secret lines" {
    load_nixopus
    result=$(echo "NIXOPUS_VERSION=0.3.0" | redact_install_log_lines)
    [ "$result" = "NIXOPUS_VERSION=0.3.0" ]
}

# ---------------------------------------------------------------------------
# cmd_report
# ---------------------------------------------------------------------------

@test "cmd_report: includes version and home in output" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    dc() { return 0; }
    docker() { return 0; }

    run cmd_report
    [[ "$output" == *"support report"* ]]
    [[ "$output" == *"Version"* ]]
    rm -rf "$tmp_home"
}

@test "cmd_report: includes install.log content when present" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    echo "install log line" > "$tmp_home/install.log"
    NIXOPUS_HOME="$tmp_home"

    dc() { return 0; }
    docker() { return 0; }

    run cmd_report
    [[ "$output" == *"Installer log"* ]]
    rm -rf "$tmp_home"
}

@test "cmd_report: handles missing install.log gracefully" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    # no install.log
    NIXOPUS_HOME="$tmp_home"

    dc() { return 0; }
    docker() { return 0; }

    run cmd_report
    [[ "$output" != *"install log line"* ]]
    rm -rf "$tmp_home"
}

@test "cmd_report: redacts secrets in .env output" {
    load_nixopus
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    NIXOPUS_HOME="$tmp_home"

    dc() { return 0; }
    docker() { return 0; }

    run cmd_report
    [[ "$output" != *"dbpass1234567890"* ]]
    [[ "$output" == *"<redacted>"* ]]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# cmd_help
# ---------------------------------------------------------------------------

@test "cmd_help: shows all available commands" {
    load_nixopus
    run cmd_help
    [[ "$output" == *"status"* ]]
    [[ "$output" == *"update"* ]]
    [[ "$output" == *"domain"* ]]
    [[ "$output" == *"admin-bootstrap"* ]]
    [[ "$output" == *"report"* ]]
}

# ---------------------------------------------------------------------------
# Main dispatch (case block)
# ---------------------------------------------------------------------------

run_dispatch() {
    local tmp_home="$1"; shift
    # Write patched script (guards removed) to a temp file so we can source it cleanly
    local tmp
    tmp=$(mktemp)
    awk '
        index($0, "(id -u)") { in_root=1 }
        in_root && /^fi$/ { in_root=0; next }
        in_root { next }
        index($0, "[ ! -f \"$NIXOPUS_HOME") { in_env=1 }
        in_env && /^fi$/ { in_env=0; next }
        in_env { next }
        index($0, "case \"${1:-help}\"") { in_case=1 }
        !in_case { print }
    ' "$NIXOPUS_SH" > "$tmp"
    # Pass -- as $0 so $1 is the first real arg (the command name)
    env NIXOPUS_HOME="$tmp_home" bash -c "
        source \"$tmp\"
        NIXOPUS_HOME=\"$tmp_home\"
        dc() { echo \"DC: \$*\"; }
        docker() { return 1; }
        sync_view_llm_env() { return 0; }
        case \"\${1:-help}\" in
            status)    cmd_status ;;
            logs)      shift; cmd_logs \"\${1:-}\" ;;
            update)    cmd_update ;;
            restart)   shift; cmd_restart \"\${1:-}\" ;;
            stop)      cmd_stop ;;
            uninstall) shift; cmd_uninstall \"\${1:-}\" ;;
            config)    shift; cmd_config \"\$@\" ;;
            domain)    shift; cmd_domain \"\$@\" ;;
            ip)        shift; cmd_ip \"\$@\" ;;
            port)      shift; cmd_port \"\$@\" ;;
            backup)    cmd_backup ;;
            info)      cmd_info ;;
            admin-bootstrap) shift; cmd_admin_bootstrap \"\$@\" ;;
            report)    cmd_report ;;
            help|--help|-h) cmd_help ;;
            *)         echo \"Unknown command: \$1\"; cmd_help; exit 1 ;;
        esac
    " -- "$@"
    local rc=$?
    rm -f "$tmp"
    return $rc
}

@test "dispatch: status" {
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    run run_dispatch "$tmp_home" status
    [[ "$output" == *"DC:"* ]]
    rm -rf "$tmp_home"
}

@test "dispatch: logs" {
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    run run_dispatch "$tmp_home" logs
    [[ "$output" == *"DC:"* ]]
    rm -rf "$tmp_home"
}

@test "dispatch: update" {
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    run run_dispatch "$tmp_home" update
    [[ "$output" == *"DC:"* ]]
    rm -rf "$tmp_home"
}

@test "dispatch: restart" {
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    run run_dispatch "$tmp_home" restart
    [[ "$output" == *"DC:"* ]]
    rm -rf "$tmp_home"
}

@test "dispatch: stop" {
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    run run_dispatch "$tmp_home" stop
    [[ "$output" == *"DC:"* ]]
    rm -rf "$tmp_home"
}

@test "dispatch: help" {
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    run run_dispatch "$tmp_home" help
    [[ "$output" == *"Commands"* ]]
    rm -rf "$tmp_home"
}

@test "dispatch: --help" {
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    run run_dispatch "$tmp_home" --help
    [[ "$output" == *"Commands"* ]]
    rm -rf "$tmp_home"
}

@test "dispatch: -h" {
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    run run_dispatch "$tmp_home" -h
    [[ "$output" == *"Commands"* ]]
    rm -rf "$tmp_home"
}

@test "dispatch: no args defaults to help" {
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    run run_dispatch "$tmp_home"
    [[ "$output" == *"Commands"* ]]
    rm -rf "$tmp_home"
}

@test "dispatch: unknown command exits 1" {
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    run run_dispatch "$tmp_home" notacommand
    [ "$status" -ne 0 ]
    [[ "$output" == *"Unknown command"* ]]
    rm -rf "$tmp_home"
}

@test "dispatch: info" {
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    run run_dispatch "$tmp_home" info
    [[ "$output" == *"Nixopus"* ]]
    rm -rf "$tmp_home"
}

@test "dispatch: report" {
    local tmp_home
    tmp_home=$(mktemp -d)
    make_env "$tmp_home"
    run run_dispatch "$tmp_home" report
    [[ "$output" == *"report"* ]]
    rm -rf "$tmp_home"
}
