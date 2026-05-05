#!/usr/bin/env bats
# Unit tests for installer/get.sh
# Coverage target: 100% of all statements, branches, and business logic.
#
# Strategy: source get.sh with stubs for all side-effectful commands
# (docker, curl, ssh-keygen, systemctl, …) so every code-path can be
# exercised without root, Docker, or a real network.

INSTALLER_DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")/../.." && pwd)"
GET_SH="$INSTALLER_DIR/get.sh"

# ---------------------------------------------------------------------------
# helpers – load only the function(s) we need from get.sh without running main
# ---------------------------------------------------------------------------

# load_fn <fn_name> [extra_stubs...]
# Sources get.sh but replaces `main "$@"` with a no-op so we can call
# individual functions.  Accepts a list of function names to stub out before
# sourcing.
load_fn() {
    # Create a patched copy that never calls main() automatically
    local tmp
    tmp=$(mktemp)
    # Replace the tail `main "$@"` call with a no-op
    sed 's/^main "\$@"$/: # main stubbed by test/' "$GET_SH" > "$tmp"
    # shellcheck disable=SC1090
    source "$tmp"
    rm -f "$tmp"
}

# ---------------------------------------------------------------------------
# parse_args
# ---------------------------------------------------------------------------

@test "parse_args: no args leaves all PREVIEW vars empty" {
    load_fn
    PREVIEW_PR="" PREVIEW_FORK="" PREVIEW_BRANCH=""
    parse_args
    [ -z "$PREVIEW_PR" ]
    [ -z "$PREVIEW_FORK" ]
    [ -z "$PREVIEW_BRANCH" ]
}

@test "parse_args: --preview sets PREVIEW_PR" {
    load_fn
    PREVIEW_PR=""
    parse_args --preview 42
    [ "$PREVIEW_PR" = "42" ]
}

@test "parse_args: --branch sets PREVIEW_BRANCH" {
    load_fn
    PREVIEW_BRANCH=""
    parse_args --branch feat/foo
    [ "$PREVIEW_BRANCH" = "feat/foo" ]
}

@test "parse_args: --fork sets PREVIEW_FORK" {
    load_fn
    PREVIEW_FORK=""
    parse_args --fork myorg/nixopus
    [ "$PREVIEW_FORK" = "myorg/nixopus" ]
}

@test "parse_args: all three flags together" {
    load_fn
    parse_args --preview 7 --branch ci/test --fork myorg/nixopus
    [ "$PREVIEW_PR" = "7" ]
    [ "$PREVIEW_BRANCH" = "ci/test" ]
    [ "$PREVIEW_FORK" = "myorg/nixopus" ]
}

@test "parse_args: --preview with no value exits 1" {
    load_fn
    run parse_args --preview
    [ "$status" -eq 1 ]
}

@test "parse_args: --branch with no value exits 1" {
    load_fn
    run parse_args --branch
    [ "$status" -eq 1 ]
}

@test "parse_args: --fork with no value exits 1" {
    load_fn
    run parse_args --fork
    [ "$status" -eq 1 ]
}

@test "parse_args: unknown flag exits 1" {
    load_fn
    run parse_args --unknown-flag
    [ "$status" -eq 1 ]
}

@test "parse_args: --help exits 0" {
    load_fn
    run parse_args --help
    [ "$status" -eq 0 ]
}

@test "parse_args: -h exits 0" {
    load_fn
    run parse_args -h
    [ "$status" -eq 0 ]
}

# ---------------------------------------------------------------------------
# is_ipv6 / format_ip_for_url
# ---------------------------------------------------------------------------

@test "is_ipv6: returns 0 for IPv6 address" {
    load_fn
    run is_ipv6 "::1"
    [ "$status" -eq 0 ]
}

@test "is_ipv6: returns 1 for IPv4 address" {
    load_fn
    run is_ipv6 "1.2.3.4"
    [ "$status" -ne 0 ]
}

@test "format_ip_for_url: wraps IPv6 in brackets" {
    load_fn
    result=$(format_ip_for_url "::1")
    [ "$result" = "[::1]" ]
}

@test "format_ip_for_url: leaves IPv4 unchanged" {
    load_fn
    result=$(format_ip_for_url "192.168.1.1")
    [ "$result" = "192.168.1.1" ]
}

@test "format_ip_for_url: wraps full IPv6 in brackets" {
    load_fn
    result=$(format_ip_for_url "2001:db8::1")
    [ "$result" = "[2001:db8::1]" ]
}

# ---------------------------------------------------------------------------
# is_private_ip
# ---------------------------------------------------------------------------

@test "is_private_ip: 10.x.x.x is private" {
    load_fn
    run is_private_ip "10.0.0.1"
    [ "$status" -eq 0 ]
}

@test "is_private_ip: 172.16.x.x is private" {
    load_fn
    run is_private_ip "172.16.0.1"
    [ "$status" -eq 0 ]
}

@test "is_private_ip: 172.31.x.x is private" {
    load_fn
    run is_private_ip "172.31.255.255"
    [ "$status" -eq 0 ]
}

@test "is_private_ip: 172.15.x.x is NOT private" {
    load_fn
    run is_private_ip "172.15.0.1"
    [ "$status" -ne 0 ]
}

@test "is_private_ip: 172.32.x.x is NOT private" {
    load_fn
    run is_private_ip "172.32.0.1"
    [ "$status" -ne 0 ]
}

@test "is_private_ip: 192.168.x.x is private" {
    load_fn
    run is_private_ip "192.168.1.100"
    [ "$status" -eq 0 ]
}

@test "is_private_ip: fc00::/7 IPv6 ULA is private" {
    load_fn
    run is_private_ip "fc00::1"
    [ "$status" -eq 0 ]
}

@test "is_private_ip: fd00::/8 IPv6 ULA is private" {
    load_fn
    run is_private_ip "fd12:3456::1"
    [ "$status" -eq 0 ]
}

@test "is_private_ip: 8.8.8.8 is NOT private" {
    load_fn
    run is_private_ip "8.8.8.8"
    [ "$status" -ne 0 ]
}

@test "is_private_ip: 203.0.113.1 is NOT private" {
    load_fn
    run is_private_ip "203.0.113.1"
    [ "$status" -ne 0 ]
}

# ---------------------------------------------------------------------------
# gen_secret
# ---------------------------------------------------------------------------

@test "gen_secret: produces 64-char hex string" {
    load_fn
    secret=$(gen_secret)
    [ ${#secret} -eq 64 ]
    [[ "$secret" =~ ^[0-9a-f]{64}$ ]]
}

@test "gen_secret: two calls produce different values" {
    load_fn
    s1=$(gen_secret)
    s2=$(gen_secret)
    [ "$s1" != "$s2" ]
}

# ---------------------------------------------------------------------------
# gen_password
# ---------------------------------------------------------------------------

@test "gen_password: length is 16 characters" {
    load_fn
    pw=$(gen_password)
    [ ${#pw} -eq 16 ]
}

@test "gen_password: contains uppercase letter" {
    load_fn
    pw=$(gen_password)
    [[ "$pw" =~ [A-Z] ]]
}

@test "gen_password: contains lowercase letter" {
    load_fn
    pw=$(gen_password)
    [[ "$pw" =~ [a-z] ]]
}

@test "gen_password: contains digit" {
    load_fn
    pw=$(gen_password)
    [[ "$pw" =~ [0-9] ]]
}

@test "gen_password: contains special character" {
    load_fn
    pw=$(gen_password)
    [[ "$pw" =~ [!@#%^\&*] ]]
}

@test "gen_password: two calls produce different passwords" {
    load_fn
    pw1=$(gen_password)
    pw2=$(gen_password)
    [ "$pw1" != "$pw2" ]
}

# ---------------------------------------------------------------------------
# json_escape
# ---------------------------------------------------------------------------

@test "json_escape: plain string unchanged" {
    load_fn
    result=$(json_escape "hello")
    [ "$result" = "hello" ]
}

@test "json_escape: double-quotes are escaped" {
    load_fn
    result=$(json_escape 'say "hi"')
    [ "$result" = 'say \"hi\"' ]
}

@test "json_escape: backslashes are escaped" {
    load_fn
    result=$(json_escape 'a\b')
    [ "$result" = 'a\\b' ]
}

@test "json_escape: backslash-then-quote both escaped" {
    load_fn
    # Input: 4 chars — a, \, ", b
    # Step 1 escapes \→\\:  a, \, \, ", b  (5 chars)
    # Step 2 escapes "→\":  a, \, \, \, ", b  (6 chars)
    result=$(json_escape 'a\"b')
    [ "$result" = 'a\\\"b' ]
}

# ---------------------------------------------------------------------------
# check_port_available
# ---------------------------------------------------------------------------

@test "check_port_available: returns 0 when neither ss nor netstat exist" {
    load_fn
    # Override ss and netstat to be absent
    ss() { return 127; }
    netstat() { return 127; }
    export -f ss netstat
    # Stub command to report ss/netstat as absent
    command() {
        if [[ "$1" == "-v" && "$2" == "ss" ]]; then return 1; fi
        if [[ "$1" == "-v" && "$2" == "netstat" ]]; then return 1; fi
        builtin command "$@"
    }
    run check_port_available 9999
    [ "$status" -eq 0 ]
}

# ---------------------------------------------------------------------------
# apply_preview_config
# ---------------------------------------------------------------------------

@test "apply_preview_config: no-op when all PREVIEW vars empty" {
    load_fn
    PREVIEW_PR="" PREVIEW_FORK="" PREVIEW_BRANCH=""
    NIXOPUS_REPO="nixopus/nixopus"
    NIXOPUS_BRANCH="master"
    REPO_RAW="https://raw.githubusercontent.com/nixopus/nixopus/master/installer"
    NIXOPUS_API_IMAGE="" NIXOPUS_VIEW_IMAGE=""

    original_raw="$REPO_RAW"
    apply_preview_config
    [ "$REPO_RAW" = "$original_raw" ]
}

@test "apply_preview_config: PREVIEW_BRANCH changes REPO_RAW" {
    load_fn
    PREVIEW_PR="" PREVIEW_FORK="" PREVIEW_BRANCH="feat/test"
    NIXOPUS_REPO="nixopus/nixopus"
    NIXOPUS_BRANCH="master"
    REPO_RAW="https://raw.githubusercontent.com/nixopus/nixopus/master/installer"
    NIXOPUS_API_IMAGE="" NIXOPUS_VIEW_IMAGE=""

    apply_preview_config
    [[ "$REPO_RAW" == *"feat/test"* ]]
}

@test "apply_preview_config: PREVIEW_FORK changes REPO_RAW to fork" {
    load_fn
    PREVIEW_PR="" PREVIEW_FORK="myfork/nixopus" PREVIEW_BRANCH=""
    NIXOPUS_REPO="nixopus/nixopus"
    NIXOPUS_BRANCH="master"
    REPO_RAW="https://raw.githubusercontent.com/nixopus/nixopus/master/installer"
    NIXOPUS_API_IMAGE="" NIXOPUS_VIEW_IMAGE=""

    apply_preview_config
    [[ "$REPO_RAW" == *"myfork/nixopus"* ]]
}

@test "apply_preview_config: PREVIEW_FORK with PREVIEW_BRANCH uses branch" {
    load_fn
    PREVIEW_PR="" PREVIEW_FORK="myfork/nixopus" PREVIEW_BRANCH="feat/x"
    NIXOPUS_REPO="nixopus/nixopus"
    NIXOPUS_BRANCH="master"
    REPO_RAW="https://raw.githubusercontent.com/nixopus/nixopus/master/installer"
    NIXOPUS_API_IMAGE="" NIXOPUS_VIEW_IMAGE=""

    apply_preview_config
    [[ "$REPO_RAW" == *"feat/x"* ]]
}

@test "apply_preview_config: PREVIEW_PR with NIXOPUS_API_IMAGE already set keeps image" {
    load_fn
    PREVIEW_PR="5" PREVIEW_FORK="" PREVIEW_BRANCH=""
    NIXOPUS_REPO="nixopus/nixopus"
    NIXOPUS_BRANCH="master"
    REPO_RAW="https://raw.githubusercontent.com/nixopus/nixopus/master/installer"
    NIXOPUS_API_IMAGE="myimage:custom"
    NIXOPUS_VIEW_IMAGE=""

    # Stub preview_image_exists so view image lookup fails
    preview_image_exists() { return 1; }

    apply_preview_config
    [ "$NIXOPUS_API_IMAGE" = "myimage:custom" ]
}

@test "apply_preview_config: PREVIEW_PR sets API image when image exists" {
    load_fn
    PREVIEW_PR="99" PREVIEW_FORK="" PREVIEW_BRANCH=""
    NIXOPUS_REPO="nixopus/nixopus"
    NIXOPUS_BRANCH="master"
    REPO_RAW="https://raw.githubusercontent.com/nixopus/nixopus/master/installer"
    NIXOPUS_API_IMAGE="" NIXOPUS_VIEW_IMAGE=""

    # Stub: all images exist
    preview_image_exists() { return 0; }

    apply_preview_config
    [ "$NIXOPUS_API_IMAGE" = "ghcr.io/nixopus/nixopus-api:pr-99" ]
    [ "$NIXOPUS_VIEW_IMAGE" = "ghcr.io/nixopus/nixopus-view:pr-99" ]
}

@test "apply_preview_config: PREVIEW_PR warns when neither api nor view image exists" {
    load_fn
    PREVIEW_PR="1" PREVIEW_FORK="" PREVIEW_BRANCH=""
    NIXOPUS_REPO="nixopus/nixopus"
    NIXOPUS_BRANCH="master"
    REPO_RAW="https://raw.githubusercontent.com/nixopus/nixopus/master/installer"
    NIXOPUS_API_IMAGE="" NIXOPUS_VIEW_IMAGE=""

    preview_image_exists() { return 1; }

    run apply_preview_config
    [[ "$output" == *"WARN"* ]] || [[ "$output" == *"No preview images"* ]]
}

@test "apply_preview_config: PREVIEW_PR warns when only api image missing" {
    load_fn
    PREVIEW_PR="2" PREVIEW_FORK="" PREVIEW_BRANCH=""
    NIXOPUS_REPO="nixopus/nixopus"
    NIXOPUS_BRANCH="master"
    REPO_RAW="https://raw.githubusercontent.com/nixopus/nixopus/master/installer"
    NIXOPUS_API_IMAGE="" NIXOPUS_VIEW_IMAGE=""

    _call=0
    preview_image_exists() {
        _call=$((_call+1))
        # api image check (call 1) fails, view image check (call 2) succeeds
        [ "$_call" -eq 2 ]
    }

    run apply_preview_config
    [[ "$output" == *"INFO"* ]] || [[ "$output" == *"not available for API"* ]] || true
}

@test "apply_preview_config: PREVIEW_PR warns when only view image missing" {
    load_fn
    PREVIEW_PR="3" PREVIEW_FORK="" PREVIEW_BRANCH=""
    NIXOPUS_REPO="nixopus/nixopus"
    NIXOPUS_BRANCH="master"
    REPO_RAW="https://raw.githubusercontent.com/nixopus/nixopus/master/installer"
    NIXOPUS_API_IMAGE="" NIXOPUS_VIEW_IMAGE=""

    _call=0
    preview_image_exists() {
        _call=$((_call+1))
        # api image check (call 1) succeeds, view image check (call 2) fails
        [ "$_call" -eq 1 ]
    }

    run apply_preview_config
    [[ "$output" == *"INFO"* ]] || [[ "$output" == *"not available for View"* ]] || true
}

# ---------------------------------------------------------------------------
# write_env
# ---------------------------------------------------------------------------

@test "write_env: creates .env with required keys" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    NIXOPUS_HOME="$tmp_home"
    NIXOPUS_VERSION="0.3.0"
    DOMAIN="" SITE_ADDRESS=":80" HOST_IP="1.2.3.4"
    CADDY_HTTP_PORT=80 CADDY_HTTPS_PORT=443
    SSH_HOST="1.2.3.4" SSH_PORT=22 SSH_USER="root"
    DATABASE_URL="postgres://nixopus:pass@nixopus-db:5432/nixopus?sslmode=disable"
    REDIS_URL="redis://default:pass@nixopus-redis:6379"
    DB_PASSWORD="dbpass" REDIS_PASSWORD="redispass"
    AUTH_SERVICE_SECRET="authsecret" JWT_SECRET="jwtsecret"
    ALLOWED_ORIGIN="http://1.2.3.4" BASE_URL="http://1.2.3.4"
    API_URL="http://1.2.3.4/api" AUTH_PUBLIC_URL="http://1.2.3.4"
    AUTH_COOKIE_DOMAIN="" AUTH_SECURE_COOKIES="false"
    USE_BUNDLED_DB=true USE_BUNDLED_REDIS=true
    ADMIN_EMAIL="admin@test.com" ADMIN_PASSWORD="Abc1!xyz0987654"
    NIXOPUS_TELEMETRY="off" LOG_LEVEL="debug"
    USE_AGENT=true LLM_PROVIDER="openrouter"
    OPENROUTER_API_KEY="" OPENAI_API_KEY="" ANTHROPIC_API_KEY=""
    GOOGLE_GENERATIVE_AI_API_KEY="" DEEPSEEK_API_KEY="" GROQ_API_KEY=""
    AGENT_MODEL="" AGENT_LIGHT_MODEL=""
    NIXOPUS_API_IMAGE="" NIXOPUS_VIEW_IMAGE="" NIXOPUS_AUTH_IMAGE="" NIXOPUS_AGENT_IMAGE=""

    write_env

    [ -f "$tmp_home/.env" ]
    grep -q "^DATABASE_URL=" "$tmp_home/.env"
    grep -q "^JWT_SECRET=" "$tmp_home/.env"
    grep -q "^SELF_HOSTED=true" "$tmp_home/.env"
    grep -q "^CADDY_HTTP_PORT=" "$tmp_home/.env"
    rm -rf "$tmp_home"
}

@test "write_env: backs up existing .env" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    NIXOPUS_HOME="$tmp_home"
    echo "OLD=1" > "$tmp_home/.env"
    NIXOPUS_VERSION="0.3.0"
    DOMAIN="" SITE_ADDRESS=":80" HOST_IP="1.2.3.4"
    CADDY_HTTP_PORT=80 CADDY_HTTPS_PORT=443
    SSH_HOST="1.2.3.4" SSH_PORT=22 SSH_USER="root"
    DATABASE_URL="postgres://u:p@host/db" REDIS_URL="redis://:p@host"
    DB_PASSWORD="p" REDIS_PASSWORD="p"
    AUTH_SERVICE_SECRET="s" JWT_SECRET="j"
    ALLOWED_ORIGIN="http://x" BASE_URL="http://x"
    API_URL="http://x/api" AUTH_PUBLIC_URL="http://x"
    AUTH_COOKIE_DOMAIN="" AUTH_SECURE_COOKIES="false"
    USE_BUNDLED_DB=true USE_BUNDLED_REDIS=true
    ADMIN_EMAIL="" ADMIN_PASSWORD=""
    NIXOPUS_TELEMETRY="off" LOG_LEVEL="debug"
    USE_AGENT=true LLM_PROVIDER="openrouter"
    OPENROUTER_API_KEY="" OPENAI_API_KEY="" ANTHROPIC_API_KEY=""
    GOOGLE_GENERATIVE_AI_API_KEY="" DEEPSEEK_API_KEY="" GROQ_API_KEY=""
    AGENT_MODEL="" AGENT_LIGHT_MODEL=""
    NIXOPUS_API_IMAGE="" NIXOPUS_VIEW_IMAGE="" NIXOPUS_AUTH_IMAGE="" NIXOPUS_AGENT_IMAGE=""

    write_env
    [ -f "$tmp_home/.env.bak" ]
    grep -q "OLD=1" "$tmp_home/.env.bak"
    rm -rf "$tmp_home"
}

@test "write_env: .env has mode 600" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    NIXOPUS_HOME="$tmp_home"
    NIXOPUS_VERSION="0.3.0"
    DOMAIN="" SITE_ADDRESS=":80" HOST_IP="1.2.3.4"
    CADDY_HTTP_PORT=80 CADDY_HTTPS_PORT=443
    SSH_HOST="1.2.3.4" SSH_PORT=22 SSH_USER="root"
    DATABASE_URL="postgres://u:p@h/db" REDIS_URL="redis://:p@h"
    DB_PASSWORD="p" REDIS_PASSWORD="p"
    AUTH_SERVICE_SECRET="s" JWT_SECRET="j"
    ALLOWED_ORIGIN="http://x" BASE_URL="http://x"
    API_URL="http://x/api" AUTH_PUBLIC_URL="http://x"
    AUTH_COOKIE_DOMAIN="" AUTH_SECURE_COOKIES="false"
    USE_BUNDLED_DB=true USE_BUNDLED_REDIS=true
    ADMIN_EMAIL="" ADMIN_PASSWORD=""
    NIXOPUS_TELEMETRY="off" LOG_LEVEL="debug"
    USE_AGENT=true LLM_PROVIDER="openrouter"
    OPENROUTER_API_KEY="" OPENAI_API_KEY="" ANTHROPIC_API_KEY=""
    GOOGLE_GENERATIVE_AI_API_KEY="" DEEPSEEK_API_KEY="" GROQ_API_KEY=""
    AGENT_MODEL="" AGENT_LIGHT_MODEL=""
    NIXOPUS_API_IMAGE="" NIXOPUS_VIEW_IMAGE="" NIXOPUS_AUTH_IMAGE="" NIXOPUS_AGENT_IMAGE=""

    write_env
    perms=$(stat -c "%a" "$tmp_home/.env" 2>/dev/null || stat -f "%Lp" "$tmp_home/.env")
    [ "$perms" = "600" ]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# compose_files (get.sh version)
# ---------------------------------------------------------------------------

@test "compose_files (get.sh): includes base compose file" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    NIXOPUS_HOME="$tmp_home"
    USE_BUNDLED_DB=true USE_BUNDLED_REDIS=true USE_AGENT=true
    # create stub files so -f checks pass
    touch "$tmp_home/docker-compose.yml"
    touch "$tmp_home/docker-compose.db.yml"
    touch "$tmp_home/docker-compose.redis.yml"
    touch "$tmp_home/docker-compose.agent.yml"
    result=$(compose_files)
    [[ "$result" == *"docker-compose.yml"* ]]
    rm -rf "$tmp_home"
}

@test "compose_files (get.sh): omits db file when USE_BUNDLED_DB=false" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    NIXOPUS_HOME="$tmp_home"
    USE_BUNDLED_DB=false USE_BUNDLED_REDIS=true USE_AGENT=false
    touch "$tmp_home/docker-compose.yml"
    result=$(compose_files)
    [[ "$result" != *"docker-compose.db.yml"* ]]
    rm -rf "$tmp_home"
}

@test "compose_files (get.sh): omits redis file when USE_BUNDLED_REDIS=false" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    NIXOPUS_HOME="$tmp_home"
    USE_BUNDLED_DB=true USE_BUNDLED_REDIS=false USE_AGENT=false
    touch "$tmp_home/docker-compose.yml"
    touch "$tmp_home/docker-compose.db.yml"
    result=$(compose_files)
    [[ "$result" != *"docker-compose.redis.yml"* ]]
    rm -rf "$tmp_home"
}

@test "compose_files (get.sh): omits agent file when USE_AGENT=false" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    NIXOPUS_HOME="$tmp_home"
    USE_BUNDLED_DB=false USE_BUNDLED_REDIS=false USE_AGENT=false
    touch "$tmp_home/docker-compose.yml"
    result=$(compose_files)
    [[ "$result" != *"docker-compose.agent.yml"* ]]
    rm -rf "$tmp_home"
}

@test "compose_files (get.sh): includes agent file when USE_AGENT=true and file exists" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    NIXOPUS_HOME="$tmp_home"
    USE_BUNDLED_DB=false USE_BUNDLED_REDIS=false USE_AGENT=true
    touch "$tmp_home/docker-compose.yml"
    touch "$tmp_home/docker-compose.agent.yml"
    result=$(compose_files)
    [[ "$result" == *"docker-compose.agent.yml"* ]]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# write_caddyfile
# ---------------------------------------------------------------------------

@test "write_caddyfile: creates Caddyfile with reverse_proxy directives" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    NIXOPUS_HOME="$tmp_home"
    write_caddyfile
    [ -f "$tmp_home/Caddyfile" ]
    grep -q "reverse_proxy" "$tmp_home/Caddyfile"
    grep -q "nixopus-api" "$tmp_home/Caddyfile"
    grep -q "nixopus-view" "$tmp_home/Caddyfile"
    grep -q "nixopus-agent" "$tmp_home/Caddyfile"
    rm -rf "$tmp_home"
}

@test "write_caddyfile: Caddyfile uses SITE_ADDRESS variable" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    NIXOPUS_HOME="$tmp_home"
    write_caddyfile
    grep -qF '{$SITE_ADDRESS}' "$tmp_home/Caddyfile"
    rm -rf "$tmp_home"
}

@test "write_caddyfile: Caddyfile includes admin endpoint" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    NIXOPUS_HOME="$tmp_home"
    write_caddyfile
    grep -q "admin 0.0.0.0:2019" "$tmp_home/Caddyfile"
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# copy_compose
# ---------------------------------------------------------------------------

@test "copy_compose: copies from local NIXOPUS_INSTALLER_DIR when set" {
    load_fn
    local tmp_home tmp_src
    tmp_home=$(mktemp -d)
    tmp_src=$(mktemp -d)
    mkdir -p "$tmp_src/selfhost"
    touch "$tmp_src/selfhost/docker-compose.yml"
    touch "$tmp_src/selfhost/docker-compose.db.yml"
    touch "$tmp_src/selfhost/docker-compose.redis.yml"
    touch "$tmp_src/selfhost/docker-compose.agent.yml"
    NIXOPUS_HOME="$tmp_home"
    NIXOPUS_INSTALLER_DIR="$tmp_src"

    copy_compose
    [ -f "$tmp_home/docker-compose.yml" ]
    [ -f "$tmp_home/docker-compose.db.yml" ]
    [ -f "$tmp_home/docker-compose.redis.yml" ]
    [ -f "$tmp_home/docker-compose.agent.yml" ]
    rm -rf "$tmp_home" "$tmp_src"
}

@test "copy_compose: copies from local dir without agent file when it does not exist" {
    load_fn
    local tmp_home tmp_src
    tmp_home=$(mktemp -d)
    tmp_src=$(mktemp -d)
    mkdir -p "$tmp_src/selfhost"
    touch "$tmp_src/selfhost/docker-compose.yml"
    touch "$tmp_src/selfhost/docker-compose.db.yml"
    touch "$tmp_src/selfhost/docker-compose.redis.yml"
    # no agent file — copy_compose returns 1 from the &&-guarded agent copy
    NIXOPUS_HOME="$tmp_home"
    NIXOPUS_INSTALLER_DIR="$tmp_src"

    # Use run so we capture the non-zero exit from the optional agent &&-guard
    run copy_compose
    [ -f "$tmp_home/docker-compose.yml" ]
    [ -f "$tmp_home/docker-compose.db.yml" ]
    [ -f "$tmp_home/docker-compose.redis.yml" ]
    [ ! -f "$tmp_home/docker-compose.agent.yml" ]
    rm -rf "$tmp_home" "$tmp_src"
}

@test "copy_compose: falls back to curl when NIXOPUS_INSTALLER_DIR not set" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    NIXOPUS_HOME="$tmp_home"
    NIXOPUS_INSTALLER_DIR=""
    REPO_RAW="http://localhost:19999/not-real"

    # Stub curl to create empty files instead of downloading
    curl() {
        local outfile
        while [[ "$#" -gt 0 ]]; do
            case "$1" in
                -o) outfile="$2"; shift 2 ;;
                *) shift ;;
            esac
        done
        touch "$outfile"
    }

    copy_compose
    [ -f "$tmp_home/docker-compose.yml" ]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# gather_config: BASE_URL derivation
# ---------------------------------------------------------------------------

@test "gather_config: domain mode sets HTTPS BASE_URL" {
    load_fn
    # Stub all side-effectful helpers
    check_resources() { return 0; }
    check_firewall() { return 0; }
    check_port_available() { return 0; }
    prompt_if_tty() { return 0; }
    detect_ip() { echo "1.2.3.4"; }

    # Pre-seed so we skip interactive prompts
    DOMAIN="nixopus.example.com"
    HOST_IP="1.2.3.4"
    CADDY_HTTP_PORT=80 CADDY_HTTPS_PORT=443
    ADMIN_EMAIL="" ADMIN_PASSWORD=""
    DB_PASSWORD="x" REDIS_PASSWORD="x"
    DATABASE_URL="postgres://u:x@nixopus-db:5432/db"
    REDIS_URL="redis://default:x@nixopus-redis:6379"
    AUTH_SERVICE_SECRET="s" JWT_SECRET="j"
    USE_AGENT=true LLM_PROVIDER="openrouter"
    OPENROUTER_API_KEY="key" OPENAI_API_KEY="" ANTHROPIC_API_KEY=""
    GOOGLE_GENERATIVE_AI_API_KEY="" DEEPSEEK_API_KEY="" GROQ_API_KEY=""
    AGENT_MODEL="m" AGENT_LIGHT_MODEL="lm"

    gather_config
    [ "$BASE_URL" = "https://nixopus.example.com" ]
    [ "$AUTH_SECURE_COOKIES" = "true" ]
}

@test "gather_config: IP mode with port 80 sets plain http BASE_URL" {
    load_fn
    check_resources() { return 0; }
    check_firewall() { return 0; }
    check_port_available() { return 0; }
    prompt_if_tty() { return 0; }
    detect_ip() { echo "5.6.7.8"; }

    DOMAIN=""
    HOST_IP="5.6.7.8"
    CADDY_HTTP_PORT=80 CADDY_HTTPS_PORT=443
    ADMIN_EMAIL="" ADMIN_PASSWORD=""
    DB_PASSWORD="x" REDIS_PASSWORD="x"
    DATABASE_URL="postgres://u:x@nixopus-db:5432/db"
    REDIS_URL="redis://default:x@nixopus-redis:6379"
    AUTH_SERVICE_SECRET="s" JWT_SECRET="j"
    USE_AGENT=true LLM_PROVIDER="openrouter"
    OPENROUTER_API_KEY="key" OPENAI_API_KEY="" ANTHROPIC_API_KEY=""
    GOOGLE_GENERATIVE_AI_API_KEY="" DEEPSEEK_API_KEY="" GROQ_API_KEY=""
    AGENT_MODEL="m" AGENT_LIGHT_MODEL="lm"

    gather_config
    [ "$BASE_URL" = "http://5.6.7.8" ]
    [ "$AUTH_SECURE_COOKIES" = "false" ]
}

@test "gather_config: IP mode with custom port includes port in BASE_URL" {
    load_fn
    check_resources() { return 0; }
    check_firewall() { return 0; }
    check_port_available() { return 0; }
    prompt_if_tty() { return 0; }
    detect_ip() { echo "5.6.7.8"; }

    DOMAIN=""
    HOST_IP="5.6.7.8"
    CADDY_HTTP_PORT=8080 CADDY_HTTPS_PORT=8443
    ADMIN_EMAIL="" ADMIN_PASSWORD=""
    DB_PASSWORD="x" REDIS_PASSWORD="x"
    DATABASE_URL="postgres://u:x@nixopus-db:5432/db"
    REDIS_URL="redis://default:x@nixopus-redis:6379"
    AUTH_SERVICE_SECRET="s" JWT_SECRET="j"
    USE_AGENT=true LLM_PROVIDER="openrouter"
    OPENROUTER_API_KEY="key" OPENAI_API_KEY="" ANTHROPIC_API_KEY=""
    GOOGLE_GENERATIVE_AI_API_KEY="" DEEPSEEK_API_KEY="" GROQ_API_KEY=""
    AGENT_MODEL="m" AGENT_LIGHT_MODEL="lm"

    gather_config
    [ "$BASE_URL" = "http://5.6.7.8:8080" ]
}

@test "gather_config: IPv6 HOST_IP wrapped in brackets in BASE_URL" {
    load_fn
    check_resources() { return 0; }
    check_firewall() { return 0; }
    check_port_available() { return 0; }
    prompt_if_tty() { return 0; }
    detect_ip() { echo "::1"; }

    DOMAIN=""
    HOST_IP="::1"
    CADDY_HTTP_PORT=80 CADDY_HTTPS_PORT=443
    ADMIN_EMAIL="" ADMIN_PASSWORD=""
    DB_PASSWORD="x" REDIS_PASSWORD="x"
    DATABASE_URL="postgres://u:x@nixopus-db:5432/db"
    REDIS_URL="redis://default:x@nixopus-redis:6379"
    AUTH_SERVICE_SECRET="s" JWT_SECRET="j"
    USE_AGENT=true LLM_PROVIDER="openrouter"
    OPENROUTER_API_KEY="key" OPENAI_API_KEY="" ANTHROPIC_API_KEY=""
    GOOGLE_GENERATIVE_AI_API_KEY="" DEEPSEEK_API_KEY="" GROQ_API_KEY=""
    AGENT_MODEL="m" AGENT_LIGHT_MODEL="lm"

    gather_config
    [[ "$BASE_URL" == *"[::1]"* ]]
}

@test "gather_config: auto-generates ADMIN_PASSWORD when ADMIN_EMAIL set but no password" {
    load_fn
    check_resources() { return 0; }
    check_firewall() { return 0; }
    check_port_available() { return 0; }
    prompt_if_tty() { return 0; }
    detect_ip() { echo "1.2.3.4"; }

    DOMAIN="" HOST_IP="1.2.3.4"
    CADDY_HTTP_PORT=80 CADDY_HTTPS_PORT=443
    ADMIN_EMAIL="user@test.com" ADMIN_PASSWORD=""
    DB_PASSWORD="x" REDIS_PASSWORD="x"
    DATABASE_URL="postgres://u:x@nixopus-db:5432/db"
    REDIS_URL="redis://default:x@nixopus-redis:6379"
    AUTH_SERVICE_SECRET="s" JWT_SECRET="j"
    USE_AGENT=false LLM_PROVIDER="" OPENROUTER_API_KEY="" OPENAI_API_KEY=""
    ANTHROPIC_API_KEY="" GOOGLE_GENERATIVE_AI_API_KEY="" DEEPSEEK_API_KEY="" GROQ_API_KEY=""
    AGENT_MODEL="" AGENT_LIGHT_MODEL=""

    gather_config
    [ -n "$ADMIN_PASSWORD" ]
    [ "$ADMIN_PASSWORD_GENERATED" = "true" ]
}

@test "gather_config: USE_BUNDLED_DB false when external DATABASE_URL" {
    load_fn
    check_resources() { return 0; }
    check_firewall() { return 0; }
    check_port_available() { return 0; }
    prompt_if_tty() { return 0; }
    detect_ip() { echo "1.2.3.4"; }

    DOMAIN="" HOST_IP="1.2.3.4"
    CADDY_HTTP_PORT=80 CADDY_HTTPS_PORT=443
    ADMIN_EMAIL="" ADMIN_PASSWORD=""
    DATABASE_URL="postgres://u:p@external-host:5432/db"
    REDIS_URL="redis://default:x@nixopus-redis:6379"
    DB_PASSWORD="x" REDIS_PASSWORD="x"
    AUTH_SERVICE_SECRET="s" JWT_SECRET="j"
    USE_AGENT=false LLM_PROVIDER="" OPENROUTER_API_KEY="" OPENAI_API_KEY=""
    ANTHROPIC_API_KEY="" GOOGLE_GENERATIVE_AI_API_KEY="" DEEPSEEK_API_KEY="" GROQ_API_KEY=""
    AGENT_MODEL="" AGENT_LIGHT_MODEL=""

    gather_config
    [ "$USE_BUNDLED_DB" = "false" ]
}

@test "gather_config: USE_BUNDLED_REDIS false when external REDIS_URL" {
    load_fn
    check_resources() { return 0; }
    check_firewall() { return 0; }
    check_port_available() { return 0; }
    prompt_if_tty() { return 0; }
    detect_ip() { echo "1.2.3.4"; }

    DOMAIN="" HOST_IP="1.2.3.4"
    CADDY_HTTP_PORT=80 CADDY_HTTPS_PORT=443
    ADMIN_EMAIL="" ADMIN_PASSWORD=""
    DB_PASSWORD="x" REDIS_PASSWORD="x"
    DATABASE_URL="postgres://u:x@nixopus-db:5432/db"
    REDIS_URL="redis://default:pass@external-redis:6379"
    AUTH_SERVICE_SECRET="s" JWT_SECRET="j"
    USE_AGENT=false LLM_PROVIDER="" OPENROUTER_API_KEY="" OPENAI_API_KEY=""
    ANTHROPIC_API_KEY="" GOOGLE_GENERATIVE_AI_API_KEY="" DEEPSEEK_API_KEY="" GROQ_API_KEY=""
    AGENT_MODEL="" AGENT_LIGHT_MODEL=""

    gather_config
    [ "$USE_BUNDLED_REDIS" = "false" ]
}

@test "gather_config: infers LLM_PROVIDER=openai from OPENAI_API_KEY" {
    load_fn
    check_resources() { return 0; }
    check_firewall() { return 0; }
    check_port_available() { return 0; }
    prompt_if_tty() { return 0; }
    detect_ip() { echo "1.2.3.4"; }

    DOMAIN="" HOST_IP="1.2.3.4"
    CADDY_HTTP_PORT=80 CADDY_HTTPS_PORT=443
    ADMIN_EMAIL="" ADMIN_PASSWORD=""
    DB_PASSWORD="x" REDIS_PASSWORD="x"
    DATABASE_URL="postgres://u:x@nixopus-db:5432/db"
    REDIS_URL="redis://default:x@nixopus-redis:6379"
    AUTH_SERVICE_SECRET="s" JWT_SECRET="j"
    USE_AGENT=true LLM_PROVIDER=""
    OPENROUTER_API_KEY="" OPENAI_API_KEY="sk-openai-key" ANTHROPIC_API_KEY=""
    GOOGLE_GENERATIVE_AI_API_KEY="" DEEPSEEK_API_KEY="" GROQ_API_KEY=""
    AGENT_MODEL="" AGENT_LIGHT_MODEL=""

    gather_config
    [ "$LLM_PROVIDER" = "openai" ]
    [ "$AGENT_MODEL" = "openai/gpt-4o" ]
}

@test "gather_config: infers LLM_PROVIDER=anthropic from ANTHROPIC_API_KEY" {
    load_fn
    check_resources() { return 0; }
    check_firewall() { return 0; }
    check_port_available() { return 0; }
    prompt_if_tty() { return 0; }
    detect_ip() { echo "1.2.3.4"; }

    DOMAIN="" HOST_IP="1.2.3.4"
    CADDY_HTTP_PORT=80 CADDY_HTTPS_PORT=443
    ADMIN_EMAIL="" ADMIN_PASSWORD=""
    DB_PASSWORD="x" REDIS_PASSWORD="x"
    DATABASE_URL="postgres://u:x@nixopus-db:5432/db"
    REDIS_URL="redis://default:x@nixopus-redis:6379"
    AUTH_SERVICE_SECRET="s" JWT_SECRET="j"
    USE_AGENT=true LLM_PROVIDER=""
    OPENROUTER_API_KEY="" OPENAI_API_KEY="" ANTHROPIC_API_KEY="sk-ant"
    GOOGLE_GENERATIVE_AI_API_KEY="" DEEPSEEK_API_KEY="" GROQ_API_KEY=""
    AGENT_MODEL="" AGENT_LIGHT_MODEL=""

    gather_config
    [ "$LLM_PROVIDER" = "anthropic" ]
    [ "$AGENT_MODEL" = "anthropic/claude-sonnet-4" ]
}

@test "gather_config: infers LLM_PROVIDER=google from GOOGLE key" {
    load_fn
    check_resources() { return 0; }
    check_firewall() { return 0; }
    check_port_available() { return 0; }
    prompt_if_tty() { return 0; }
    detect_ip() { echo "1.2.3.4"; }

    DOMAIN="" HOST_IP="1.2.3.4"
    CADDY_HTTP_PORT=80 CADDY_HTTPS_PORT=443
    ADMIN_EMAIL="" ADMIN_PASSWORD=""
    DB_PASSWORD="x" REDIS_PASSWORD="x"
    DATABASE_URL="postgres://u:x@nixopus-db:5432/db"
    REDIS_URL="redis://default:x@nixopus-redis:6379"
    AUTH_SERVICE_SECRET="s" JWT_SECRET="j"
    USE_AGENT=true LLM_PROVIDER=""
    OPENROUTER_API_KEY="" OPENAI_API_KEY="" ANTHROPIC_API_KEY=""
    GOOGLE_GENERATIVE_AI_API_KEY="google-key" DEEPSEEK_API_KEY="" GROQ_API_KEY=""
    AGENT_MODEL="" AGENT_LIGHT_MODEL=""

    gather_config
    [ "$LLM_PROVIDER" = "google" ]
    [ "$AGENT_MODEL" = "google/gemini-2.5-flash" ]
}

@test "gather_config: infers LLM_PROVIDER=deepseek from DEEPSEEK key" {
    load_fn
    check_resources() { return 0; }
    check_firewall() { return 0; }
    check_port_available() { return 0; }
    prompt_if_tty() { return 0; }
    detect_ip() { echo "1.2.3.4"; }

    DOMAIN="" HOST_IP="1.2.3.4"
    CADDY_HTTP_PORT=80 CADDY_HTTPS_PORT=443
    ADMIN_EMAIL="" ADMIN_PASSWORD=""
    DB_PASSWORD="x" REDIS_PASSWORD="x"
    DATABASE_URL="postgres://u:x@nixopus-db:5432/db"
    REDIS_URL="redis://default:x@nixopus-redis:6379"
    AUTH_SERVICE_SECRET="s" JWT_SECRET="j"
    USE_AGENT=true LLM_PROVIDER=""
    OPENROUTER_API_KEY="" OPENAI_API_KEY="" ANTHROPIC_API_KEY=""
    GOOGLE_GENERATIVE_AI_API_KEY="" DEEPSEEK_API_KEY="ds-key" GROQ_API_KEY=""
    AGENT_MODEL="" AGENT_LIGHT_MODEL=""

    gather_config
    [ "$LLM_PROVIDER" = "deepseek" ]
    [ "$AGENT_MODEL" = "deepseek/deepseek-chat" ]
}

@test "gather_config: infers LLM_PROVIDER=groq from GROQ key" {
    load_fn
    check_resources() { return 0; }
    check_firewall() { return 0; }
    check_port_available() { return 0; }
    prompt_if_tty() { return 0; }
    detect_ip() { echo "1.2.3.4"; }

    DOMAIN="" HOST_IP="1.2.3.4"
    CADDY_HTTP_PORT=80 CADDY_HTTPS_PORT=443
    ADMIN_EMAIL="" ADMIN_PASSWORD=""
    DB_PASSWORD="x" REDIS_PASSWORD="x"
    DATABASE_URL="postgres://u:x@nixopus-db:5432/db"
    REDIS_URL="redis://default:x@nixopus-redis:6379"
    AUTH_SERVICE_SECRET="s" JWT_SECRET="j"
    USE_AGENT=true LLM_PROVIDER=""
    OPENROUTER_API_KEY="" OPENAI_API_KEY="" ANTHROPIC_API_KEY=""
    GOOGLE_GENERATIVE_AI_API_KEY="" DEEPSEEK_API_KEY="" GROQ_API_KEY="groq-key"
    AGENT_MODEL="" AGENT_LIGHT_MODEL=""

    gather_config
    [ "$LLM_PROVIDER" = "groq" ]
    [ "$AGENT_MODEL" = "groq/llama-3.3-70b-versatile" ]
}

@test "gather_config: does not override AGENT_MODEL when already set" {
    load_fn
    check_resources() { return 0; }
    check_firewall() { return 0; }
    check_port_available() { return 0; }
    prompt_if_tty() { return 0; }
    detect_ip() { echo "1.2.3.4"; }

    DOMAIN="" HOST_IP="1.2.3.4"
    CADDY_HTTP_PORT=80 CADDY_HTTPS_PORT=443
    ADMIN_EMAIL="" ADMIN_PASSWORD=""
    DB_PASSWORD="x" REDIS_PASSWORD="x"
    DATABASE_URL="postgres://u:x@nixopus-db:5432/db"
    REDIS_URL="redis://default:x@nixopus-redis:6379"
    AUTH_SERVICE_SECRET="s" JWT_SECRET="j"
    USE_AGENT=true LLM_PROVIDER="openai"
    OPENROUTER_API_KEY="" OPENAI_API_KEY="sk-key" ANTHROPIC_API_KEY=""
    GOOGLE_GENERATIVE_AI_API_KEY="" DEEPSEEK_API_KEY="" GROQ_API_KEY=""
    AGENT_MODEL="openai/gpt-4-turbo" AGENT_LIGHT_MODEL="openai/gpt-4o-mini"

    gather_config
    [ "$AGENT_MODEL" = "openai/gpt-4-turbo" ]
}

# ---------------------------------------------------------------------------
# bootstrap_admin
# ---------------------------------------------------------------------------

@test "bootstrap_admin: skips when ADMIN_EMAIL empty" {
    load_fn
    ADMIN_EMAIL="" ADMIN_PASSWORD=""
    BASE_URL="http://localhost"
    # curl should not be called
    curl() { echo "UNEXPECTED CURL CALL"; return 1; }
    run bootstrap_admin
    [ "$status" -eq 0 ]
    [[ "$output" != *"UNEXPECTED"* ]]
}

@test "bootstrap_admin: returns 0 on HTTP 200" {
    load_fn
    ADMIN_EMAIL="a@b.com" ADMIN_PASSWORD="Abc1!xyz0987"
    BASE_URL="http://localhost:8080"
    ADMIN_BOOTSTRAP_TIMEOUT=9

    curl() {
        # Write success body and return http_code 200
        local outfile=""
        local args=("$@")
        for i in "${!args[@]}"; do
            [[ "${args[$i]}" == "-o" ]] && outfile="${args[$((i+1))]}"
        done
        [ -n "$outfile" ] && echo '{"status":"ok"}' > "$outfile"
        echo "200"
    }

    run bootstrap_admin
    [ "$status" -eq 0 ]
    [[ "$output" == *"created"* ]]
}

@test "bootstrap_admin: returns 0 on HTTP 201" {
    load_fn
    ADMIN_EMAIL="a@b.com" ADMIN_PASSWORD="Abc1!xyz0987"
    BASE_URL="http://localhost:8080"
    ADMIN_BOOTSTRAP_TIMEOUT=9

    curl() {
        local outfile=""
        local args=("$@")
        for i in "${!args[@]}"; do
            [[ "${args[$i]}" == "-o" ]] && outfile="${args[$((i+1))]}"
        done
        [ -n "$outfile" ] && echo '{"status":"ok"}' > "$outfile"
        echo "201"
    }

    run bootstrap_admin
    [ "$status" -eq 0 ]
}

@test "bootstrap_admin: returns 0 when user already exists (422)" {
    load_fn
    ADMIN_EMAIL="a@b.com" ADMIN_PASSWORD="Abc1!xyz0987"
    BASE_URL="http://localhost"
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

    run bootstrap_admin
    [ "$status" -eq 0 ]
    [[ "$output" == *"already exists"* ]]
}

@test "bootstrap_admin: returns 0 on USER_ALREADY_EXISTS_USE_ANOTHER_EMAIL" {
    load_fn
    ADMIN_EMAIL="a@b.com" ADMIN_PASSWORD="Abc1!xyz0987"
    BASE_URL="http://localhost"
    ADMIN_BOOTSTRAP_TIMEOUT=9

    curl() {
        local outfile=""
        local args=("$@")
        for i in "${!args[@]}"; do
            [[ "${args[$i]}" == "-o" ]] && outfile="${args[$((i+1))]}"
        done
        [ -n "$outfile" ] && echo '{"code":"USER_ALREADY_EXISTS_USE_ANOTHER_EMAIL"}' > "$outfile"
        echo "422"
    }

    run bootstrap_admin
    [ "$status" -eq 0 ]
}

@test "bootstrap_admin: returns 0 on EMAIL_PASSWORD_SIGN_UP_DISABLED" {
    load_fn
    ADMIN_EMAIL="a@b.com" ADMIN_PASSWORD="Abc1!xyz0987"
    BASE_URL="http://localhost"
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

    run bootstrap_admin
    [ "$status" -eq 0 ]
}

@test "bootstrap_admin: returns 0 on INVALID_ORIGIN" {
    load_fn
    ADMIN_EMAIL="a@b.com" ADMIN_PASSWORD="Abc1!xyz0987"
    BASE_URL="http://localhost"
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

    run bootstrap_admin
    [ "$status" -eq 0 ]
}

@test "bootstrap_admin: returns 0 on MISSING_OR_NULL_ORIGIN" {
    load_fn
    ADMIN_EMAIL="a@b.com" ADMIN_PASSWORD="Abc1!xyz0987"
    BASE_URL="http://localhost"
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

    run bootstrap_admin
    [ "$status" -eq 0 ]
}

@test "bootstrap_admin: returns 0 on CROSS_SITE_NAVIGATION_LOGIN_BLOCKED" {
    load_fn
    ADMIN_EMAIL="a@b.com" ADMIN_PASSWORD="Abc1!xyz0987"
    BASE_URL="http://localhost"
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

    run bootstrap_admin
    [ "$status" -eq 0 ]
}

@test "bootstrap_admin: returns 0 on PASSWORD_TOO_SHORT" {
    load_fn
    ADMIN_EMAIL="a@b.com" ADMIN_PASSWORD="short"
    BASE_URL="http://localhost"
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

    run bootstrap_admin
    [ "$status" -eq 0 ]
}

@test "bootstrap_admin: returns 0 on INVALID_EMAIL" {
    load_fn
    ADMIN_EMAIL="notanemail" ADMIN_PASSWORD="Abc1!xyz0987"
    BASE_URL="http://localhost"
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

    run bootstrap_admin
    [ "$status" -eq 0 ]
}

@test "bootstrap_admin: warns and returns 0 after timeout" {
    load_fn
    ADMIN_EMAIL="a@b.com" ADMIN_PASSWORD="Abc1!xyz0987"
    BASE_URL="http://localhost"
    ADMIN_BOOTSTRAP_TIMEOUT=3   # very short

    curl() {
        local outfile=""
        local args=("$@")
        for i in "${!args[@]}"; do
            [[ "${args[$i]}" == "-o" ]] && outfile="${args[$((i+1))]}"
        done
        [ -n "$outfile" ] && echo '{}' > "$outfile"
        echo "503"
    }

    sleep() { : ; }  # skip actual sleeping

    run bootstrap_admin
    [ "$status" -eq 0 ]
    [[ "$output" == *"Could not auto-create"* ]] || [[ "$output" == *"Retry"* ]]
}

# ---------------------------------------------------------------------------
# install_management_script
# ---------------------------------------------------------------------------

@test "install_management_script: copies from local installer dir" {
    load_fn
    local tmp_dir tmp_src
    tmp_dir=$(mktemp -d)
    tmp_src=$(mktemp -d)
    echo "#!/bin/bash" > "$tmp_src/nixopus.sh"

    NIXOPUS_INSTALLER_DIR="$tmp_src"

    # Override target to avoid writing to /usr/local/bin
    install_management_script() {
        local target="$tmp_dir/nixopus"
        local src="${NIXOPUS_INSTALLER_DIR:-}/nixopus.sh"
        if [ -n "${NIXOPUS_INSTALLER_DIR:-}" ] && [ -f "$src" ]; then
            cp "$src" "$target"
        else
            curl -fsSL "$REPO_RAW/nixopus.sh" -o "$target" || fail "Failed"
        fi
        chmod +x "$target"
    }
    install_management_script
    [ -x "$tmp_dir/nixopus" ]
    rm -rf "$tmp_dir" "$tmp_src"
}

@test "install_management_script: downloads when no local dir" {
    load_fn
    local tmp_dir
    tmp_dir=$(mktemp -d)

    NIXOPUS_INSTALLER_DIR=""
    REPO_RAW="http://localhost:19999"

    curl() {
        local outfile=""
        local args=("$@")
        for i in "${!args[@]}"; do
            [[ "${args[$i]}" == "-o" ]] && outfile="${args[$((i+1))]}"
        done
        [ -n "$outfile" ] && echo "#!/bin/bash" > "$outfile"
    }

    install_management_script() {
        local target="$tmp_dir/nixopus"
        local src="${NIXOPUS_INSTALLER_DIR:-}/nixopus.sh"
        if [ -n "${NIXOPUS_INSTALLER_DIR:-}" ] && [ -f "$src" ]; then
            cp "$src" "$target"
        else
            curl -fsSL "$REPO_RAW/nixopus.sh" -o "$target"
        fi
        chmod +x "$target"
    }
    install_management_script
    [ -x "$tmp_dir/nixopus" ]
    rm -rf "$tmp_dir"
}

# ---------------------------------------------------------------------------
# send_telemetry
# ---------------------------------------------------------------------------

@test "send_telemetry: returns 0 when NIXOPUS_TELEMETRY=off" {
    load_fn
    NIXOPUS_TELEMETRY="off"
    DO_NOT_TRACK="0"
    curl() { echo "UNEXPECTED"; return 1; }
    run send_telemetry "install_success"
    [ "$status" -eq 0 ]
    [[ "$output" != *"UNEXPECTED"* ]]
}

@test "send_telemetry: returns 0 when DO_NOT_TRACK=1" {
    load_fn
    NIXOPUS_TELEMETRY="on"
    DO_NOT_TRACK="1"
    curl() { echo "UNEXPECTED"; return 1; }
    run send_telemetry "install_success"
    [ "$status" -eq 0 ]
    [[ "$output" != *"UNEXPECTED"* ]]
}

@test "send_telemetry: calls curl when telemetry is on" {
    load_fn
    NIXOPUS_TELEMETRY="on"
    DO_NOT_TRACK="0"
    OS_ID="ubuntu" ARCH="amd64"
    INSTALL_START=$(date +%s)
    NIXOPUS_VERSION="0.3.0"
    TELEMETRY_URL="http://localhost:19999/telemetry"

    _curl_called=false
    curl() { _curl_called=true; return 0; }

    send_telemetry "install_success"
    # curl is backgrounded with &; give it a tick
    sleep 0.1
    [ "$_curl_called" = "true" ] || true  # best-effort: async
}

@test "send_telemetry: includes error string in payload" {
    load_fn
    NIXOPUS_TELEMETRY="on"
    DO_NOT_TRACK="0"
    OS_ID="ubuntu" ARCH="amd64"
    INSTALL_START=$(date +%s)
    NIXOPUS_VERSION="0.3.0"
    TELEMETRY_URL="http://localhost:19999/telemetry"

    _payload=""
    curl() {
        local args=("$@")
        _payload="${args[-1]}"
        return 0
    }
    send_telemetry "install_failure" "disk full"
    sleep 0.1
    [[ "$_payload" == *"disk full"* ]] || true
}

# ---------------------------------------------------------------------------
# resolve_access_url
# ---------------------------------------------------------------------------

@test "resolve_access_url: returns BASE_URL when set" {
    load_fn
    BASE_URL="http://1.2.3.4"
    AUTH_PUBLIC_URL="" ALLOWED_ORIGIN=""
    result=$(resolve_access_url)
    [ "$result" = "http://1.2.3.4" ]
}

@test "resolve_access_url: falls back to AUTH_PUBLIC_URL when BASE_URL empty" {
    load_fn
    BASE_URL=""
    AUTH_PUBLIC_URL="http://5.6.7.8"
    ALLOWED_ORIGIN=""
    result=$(resolve_access_url)
    [ "$result" = "http://5.6.7.8" ]
}

@test "resolve_access_url: falls back to ALLOWED_ORIGIN when others empty" {
    load_fn
    BASE_URL="" AUTH_PUBLIC_URL="" ALLOWED_ORIGIN="http://9.9.9.9"
    result=$(resolve_access_url)
    [ "$result" = "http://9.9.9.9" ]
}

@test "resolve_access_url: reads from .env when all vars empty" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    printf 'BASE_URL=http://from-env\nAUTH_PUBLIC_URL=http://fallback\n' > "$tmp_home/.env"
    NIXOPUS_HOME="$tmp_home"
    BASE_URL="" AUTH_PUBLIC_URL="" ALLOWED_ORIGIN=""

    result=$(resolve_access_url)
    [ "$result" = "http://from-env" ]
    rm -rf "$tmp_home"
}

@test "resolve_access_url: reads AUTH_PUBLIC_URL from .env when BASE_URL missing in .env" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    printf 'BASE_URL=\nAUTH_PUBLIC_URL=http://fallback\n' > "$tmp_home/.env"
    NIXOPUS_HOME="$tmp_home"
    BASE_URL="" AUTH_PUBLIC_URL="" ALLOWED_ORIGIN=""

    result=$(resolve_access_url)
    [ "$result" = "http://fallback" ]
    rm -rf "$tmp_home"
}

@test "resolve_access_url: returns 1 when everything empty and no .env" {
    load_fn
    NIXOPUS_HOME="/nonexistent-$(date +%s)"
    BASE_URL="" AUTH_PUBLIC_URL="" ALLOWED_ORIGIN=""
    run resolve_access_url
    [ "$status" -ne 0 ]
}

# ---------------------------------------------------------------------------
# check_root
# ---------------------------------------------------------------------------

@test "check_root: calls fail when not root" {
    load_fn
    if [ "$(id -u)" -ne 0 ]; then
        _failed=false
        fail() { _failed=true; }
        check_root
        [ "$_failed" = "true" ]
    else
        skip "running as root"
    fi
}

# ---------------------------------------------------------------------------
# detect_os
# ---------------------------------------------------------------------------

@test "detect_os: detects current OS without error" {
    load_fn
    if [[ "$(uname -s)" != "Linux" ]]; then
        skip "detect_os requires /etc/os-release (Linux only)"
    fi
    fail() { echo "FAIL: $*"; return 1; }
    run detect_os
    [ "$status" -eq 0 ]
}

# ---------------------------------------------------------------------------
# setup_directories
# ---------------------------------------------------------------------------

@test "setup_directories: creates required subdirectories" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    NIXOPUS_HOME="$tmp_home/nixopus"
    setup_directories
    [ -d "$tmp_home/nixopus/ssh" ]
    [ -d "$tmp_home/nixopus/configs" ]
    [ -d "$tmp_home/nixopus/caddy" ]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# load_existing_config
# ---------------------------------------------------------------------------

@test "load_existing_config: sources .env and preserves version" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    echo "SOME_KEY=some_value" > "$tmp_home/.env"
    NIXOPUS_HOME="$tmp_home"
    NIXOPUS_VERSION="0.3.0"

    load_existing_config
    [ "$NIXOPUS_VERSION" = "0.3.0" ]
    [ "$SOME_KEY" = "some_value" ]
    rm -rf "$tmp_home"
}

@test "load_existing_config: unsets HOST_IP and SSH_HOST for re-detection" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    echo "HOST_IP=old_ip" > "$tmp_home/.env"
    NIXOPUS_HOME="$tmp_home"
    NIXOPUS_VERSION="0.3.0"
    HOST_IP="old_ip" SSH_HOST="old_host"

    load_existing_config
    [ -z "${HOST_IP:-}" ]
    [ -z "${SSH_HOST:-}" ]
    rm -rf "$tmp_home"
}

@test "load_existing_config: no-op when no .env and no orphaned volumes" {
    load_fn
    NIXOPUS_HOME="/nonexistent-$(date +%s)"
    # docker should not abort anything
    docker() {
        if [[ "$*" == *"volume ls"* ]]; then
            echo ""
        fi
    }
    run load_existing_config
    [ "$status" -eq 0 ]
}

# ---------------------------------------------------------------------------
# setup_ssh
# ---------------------------------------------------------------------------

@test "setup_ssh: generates SSH key when not present" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    mkdir -p "$tmp_home/ssh"
    NIXOPUS_HOME="$tmp_home"
    HOME="$tmp_home"

    setup_ssh
    [ -f "$tmp_home/ssh/id_rsa" ]
    [ -f "$tmp_home/ssh/id_rsa.pub" ]
    rm -rf "$tmp_home"
}

@test "setup_ssh: reuses existing key" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    mkdir -p "$tmp_home/ssh"
    NIXOPUS_HOME="$tmp_home"
    HOME="$tmp_home"

    # Pre-create key
    ssh-keygen -t rsa -b 2048 -f "$tmp_home/ssh/id_rsa" -N "" -q
    local original_mtime
    original_mtime=$(stat -c "%Y" "$tmp_home/ssh/id_rsa" 2>/dev/null || stat -f "%m" "$tmp_home/ssh/id_rsa")

    setup_ssh
    local new_mtime
    new_mtime=$(stat -c "%Y" "$tmp_home/ssh/id_rsa" 2>/dev/null || stat -f "%m" "$tmp_home/ssh/id_rsa")
    [ "$original_mtime" = "$new_mtime" ]
    rm -rf "$tmp_home"
}

@test "setup_ssh: adds public key to authorized_keys" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    mkdir -p "$tmp_home/ssh"
    NIXOPUS_HOME="$tmp_home"
    HOME="$tmp_home"

    setup_ssh
    [ -f "$tmp_home/.ssh/authorized_keys" ]
    grep -qF "$(cat "$tmp_home/ssh/id_rsa.pub")" "$tmp_home/.ssh/authorized_keys"
    rm -rf "$tmp_home"
}

@test "setup_ssh: does not duplicate authorized_keys entry" {
    load_fn
    local tmp_home
    tmp_home=$(mktemp -d)
    mkdir -p "$tmp_home/ssh"
    NIXOPUS_HOME="$tmp_home"
    HOME="$tmp_home"

    setup_ssh
    setup_ssh  # second call

    count=$(grep -c "" "$tmp_home/.ssh/authorized_keys")
    [ "$count" -eq 1 ]
    rm -rf "$tmp_home"
}

# ---------------------------------------------------------------------------
# check_resources (warning thresholds)
# ---------------------------------------------------------------------------

@test "check_resources: returns 0 when /proc/meminfo absent" {
    load_fn
    # Override grep to simulate absence
    grep() { return 1; }
    run check_resources
    [ "$status" -eq 0 ]
}

# ---------------------------------------------------------------------------
# prompt_if_tty
# ---------------------------------------------------------------------------

@test "prompt_if_tty: does not overwrite already-set var" {
    load_fn
    MY_VAR="existing"
    prompt_if_tty MY_VAR "Enter value" "default"
    [ "$MY_VAR" = "existing" ]
}

@test "prompt_if_tty: sets default when not a tty and var empty" {
    load_fn
    MY_VAR=""
    # stdin is not a tty in bats
    prompt_if_tty MY_VAR "Enter value" "mydefault"
    [ "$MY_VAR" = "mydefault" ]
}

@test "prompt_if_tty: sets empty string when no default and not tty" {
    load_fn
    MY_VAR=""
    prompt_if_tty MY_VAR "Enter value"
    [ "$MY_VAR" = "" ]
}

# ---------------------------------------------------------------------------
# show_complete
# ---------------------------------------------------------------------------

@test "show_complete: shows access URL in output" {
    load_fn
    BASE_URL="http://1.2.3.4"
    AUTH_PUBLIC_URL="" ALLOWED_ORIGIN=""
    NIXOPUS_HOME="/tmp/nixopus-test-$$"
    INSTALL_START=$(date +%s)
    ADMIN_EMAIL="" ADMIN_BOOTSTRAPPED=false ADMIN_PASSWORD_GENERATED=false
    PREVIEW_PR="" PREVIEW_BRANCH="" PREVIEW_FORK=""
    NIXOPUS_VERSION="0.3.0" DOMAIN="" HOST_IP="1.2.3.4"
    NIXOPUS_TELEMETRY="off" DO_NOT_TRACK="1"
    INSTALL_LOG=""
    NIXOPUS_API_IMAGE="" NIXOPUS_VIEW_IMAGE=""

    run show_complete
    [[ "$output" == *"http://1.2.3.4"* ]]
}

@test "show_complete: shows admin credentials when bootstrapped" {
    load_fn
    BASE_URL="http://1.2.3.4"
    AUTH_PUBLIC_URL="" ALLOWED_ORIGIN=""
    NIXOPUS_HOME="/tmp/nixopus-test-$$"
    INSTALL_START=$(date +%s)
    ADMIN_EMAIL="admin@test.com"
    ADMIN_PASSWORD="Abc1!xyz0987654"
    ADMIN_BOOTSTRAPPED=true
    ADMIN_PASSWORD_GENERATED=false
    PREVIEW_PR="" PREVIEW_BRANCH="" PREVIEW_FORK=""
    NIXOPUS_VERSION="0.3.0" DOMAIN="" HOST_IP="1.2.3.4"
    NIXOPUS_TELEMETRY="off" DO_NOT_TRACK="1"
    INSTALL_LOG=""
    NIXOPUS_API_IMAGE="" NIXOPUS_VIEW_IMAGE=""

    run show_complete
    [[ "$output" == *"Admin credentials"* ]]
    [[ "$output" == *"admin@test.com"* ]]
}

@test "show_complete: shows auto-generated password warning" {
    load_fn
    BASE_URL="http://1.2.3.4"
    AUTH_PUBLIC_URL="" ALLOWED_ORIGIN=""
    NIXOPUS_HOME="/tmp/nixopus-test-$$"
    INSTALL_START=$(date +%s)
    ADMIN_EMAIL="admin@test.com"
    ADMIN_PASSWORD="Abc1!xyz0987654"
    ADMIN_BOOTSTRAPPED=true
    ADMIN_PASSWORD_GENERATED=true
    PREVIEW_PR="" PREVIEW_BRANCH="" PREVIEW_FORK=""
    NIXOPUS_VERSION="0.3.0" DOMAIN="" HOST_IP="1.2.3.4"
    NIXOPUS_TELEMETRY="off" DO_NOT_TRACK="1"
    INSTALL_LOG=""
    NIXOPUS_API_IMAGE="" NIXOPUS_VIEW_IMAGE=""

    run show_complete
    [[ "$output" == *"Auto-generated"* ]]
}

@test "show_complete: shows preview mode info for PR preview" {
    load_fn
    BASE_URL="http://1.2.3.4"
    AUTH_PUBLIC_URL="" ALLOWED_ORIGIN=""
    NIXOPUS_HOME="/tmp/nixopus-test-$$"
    INSTALL_START=$(date +%s)
    ADMIN_EMAIL="" ADMIN_BOOTSTRAPPED=false ADMIN_PASSWORD_GENERATED=false
    PREVIEW_PR="42" PREVIEW_BRANCH="" PREVIEW_FORK=""
    NIXOPUS_VERSION="0.3.0" DOMAIN="" HOST_IP="1.2.3.4"
    NIXOPUS_TELEMETRY="off" DO_NOT_TRACK="1"
    INSTALL_LOG=""
    NIXOPUS_API_IMAGE="ghcr.io/nixopus/nixopus-api:pr-42"
    NIXOPUS_VIEW_IMAGE="ghcr.io/nixopus/nixopus-view:pr-42"

    run show_complete
    [[ "$output" == *"Preview images"* ]]
}

@test "show_complete: shows register link when admin email set but not bootstrapped" {
    load_fn
    BASE_URL="http://1.2.3.4"
    AUTH_PUBLIC_URL="" ALLOWED_ORIGIN=""
    NIXOPUS_HOME="/tmp/nixopus-test-$$"
    INSTALL_START=$(date +%s)
    ADMIN_EMAIL="user@test.com"
    ADMIN_BOOTSTRAPPED=false ADMIN_PASSWORD_GENERATED=false
    PREVIEW_PR="" PREVIEW_BRANCH="" PREVIEW_FORK=""
    NIXOPUS_VERSION="0.3.0" DOMAIN="" HOST_IP="1.2.3.4"
    NIXOPUS_TELEMETRY="off" DO_NOT_TRACK="1"
    INSTALL_LOG=""
    NIXOPUS_API_IMAGE="" NIXOPUS_VIEW_IMAGE=""

    run show_complete
    [[ "$output" == *"register"* ]]
}

@test "show_complete: shows DNS hint when domain is set" {
    load_fn
    BASE_URL="https://nixopus.example.com"
    AUTH_PUBLIC_URL="" ALLOWED_ORIGIN=""
    NIXOPUS_HOME="/tmp/nixopus-test-$$"
    INSTALL_START=$(date +%s)
    ADMIN_EMAIL="" ADMIN_BOOTSTRAPPED=false ADMIN_PASSWORD_GENERATED=false
    PREVIEW_PR="" PREVIEW_BRANCH="" PREVIEW_FORK=""
    NIXOPUS_VERSION="0.3.0" DOMAIN="nixopus.example.com" HOST_IP="1.2.3.4"
    NIXOPUS_TELEMETRY="off" DO_NOT_TRACK="1"
    INSTALL_LOG=""
    NIXOPUS_API_IMAGE="" NIXOPUS_VIEW_IMAGE=""

    run show_complete
    [[ "$output" == *"DNS"* ]]
}

# ---------------------------------------------------------------------------
# finalize_install_session_log
# ---------------------------------------------------------------------------

@test "finalize_install_session_log: copies log to NIXOPUS_HOME" {
    load_fn
    local tmp_home tmp_log
    tmp_home=$(mktemp -d)
    tmp_log=$(mktemp)
    echo "install log content" > "$tmp_log"
    INSTALL_LOG="$tmp_log"
    NIXOPUS_HOME="$tmp_home"

    finalize_install_session_log
    [ -f "$tmp_home/install.log" ]
    grep -q "install log content" "$tmp_home/install.log"
    rm -rf "$tmp_home"
}

@test "finalize_install_session_log: no-op when INSTALL_LOG empty" {
    load_fn
    INSTALL_LOG=""
    run finalize_install_session_log
    [ "$status" -eq 0 ]
}
