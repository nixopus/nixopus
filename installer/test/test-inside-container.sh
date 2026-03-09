#!/bin/bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive

install_prereqs() {
    if command -v apk &>/dev/null; then
        apk add --no-cache bash curl openssl openssh-keygen iptables device-mapper e2fsprogs-extra >/dev/null 2>&1
    elif command -v apt-get &>/dev/null; then
        apt-get update -qq >/dev/null 2>&1
        apt-get install -y -qq curl openssl iptables >/dev/null 2>&1
    elif command -v dnf &>/dev/null; then
        dnf install -y -q --allowerasing curl openssl iptables >/dev/null 2>&1
    elif command -v yum &>/dev/null; then
        yum install -y -q curl openssl iptables >/dev/null 2>&1
    fi
}

start_dockerd() {
    if command -v dockerd &>/dev/null; then
        echo "dockerd already available"
    else
        echo "Installing Docker..."
        local os_id=""
        [ -f /etc/os-release ] && . /etc/os-release && os_id="${ID:-}"

        case "$os_id" in
            rocky|alma|centos|rhel)
                dnf install -y dnf-plugins-core >/dev/null 2>&1
                dnf config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo >/dev/null 2>&1
                dnf install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin >/dev/null 2>&1
                ;;
            alpine)
                apk add --no-cache docker docker-cli-compose >/dev/null 2>&1
                ;;
            *)
                curl -fsSL https://get.docker.com | sh >/dev/null 2>&1 || {
                    echo "Docker install via get.docker.com failed"
                    exit 1
                }
                ;;
        esac
    fi

    echo "Starting dockerd..."
    dockerd --storage-driver=vfs &>/var/log/dockerd.log &

    local wait=0
    while ! docker info &>/dev/null; do
        sleep 2
        wait=$((wait + 2))
        if [ $wait -ge 60 ]; then
            echo "FATAL: dockerd failed to start"
            tail -20 /var/log/dockerd.log
            exit 1
        fi
    done
    echo "dockerd ready (${wait}s)"
}

load_mocks() {
    echo "Loading mock images..."
    gunzip -c /mocks.tar.gz | docker load >/dev/null 2>&1
    echo "Mock images loaded"
}

run_installer() {
    echo "Running get.sh..."
    export CADDY_HTTP_PORT="${CADDY_HTTP_PORT:-8080}"
    export CADDY_HTTPS_PORT="${CADDY_HTTPS_PORT:-8443}"
    local script="${NIXOPUS_INSTALLER_DIR:-/installer}/get.sh"
    NIXOPUS_API_IMAGE=nixopus-mock-api \
    NIXOPUS_AUTH_IMAGE=nixopus-mock-auth \
    NIXOPUS_VIEW_IMAGE=nixopus-mock-view \
    HOST_IP=127.0.0.1 \
    ADMIN_EMAIL=test@nixopus.dev \
    NIXOPUS_TELEMETRY=off \
    CADDY_HTTP_PORT="$CADDY_HTTP_PORT" \
    CADDY_HTTPS_PORT="$CADDY_HTTPS_PORT" \
    bash "$script"
}

verify() {
    local pass=0 fail=0
    check() {
        if eval "$2" &>/dev/null; then
            echo "  PASS  $1"
            pass=$((pass + 1))
        else
            echo "  FAIL  $1"
            fail=$((fail + 1))
        fi
    }

    echo ""
    echo "── Verification ──"

    check ".env exists"                   "[ -f /opt/nixopus/.env ]"
    check "docker-compose.yml exists"      "[ -f /opt/nixopus/docker-compose.yml ]"
    check "docker-compose.db.yml exists"   "[ -f /opt/nixopus/docker-compose.db.yml ]"
    check "docker-compose.redis.yml exists" "[ -f /opt/nixopus/docker-compose.redis.yml ]"
    check "Caddyfile exists"               "[ -f /opt/nixopus/Caddyfile ]"
    check "SSH key exists"            "[ -f /opt/nixopus/ssh/id_rsa ]"
    check "SSH pubkey exists"         "[ -f /opt/nixopus/ssh/id_rsa.pub ]"
    check "nixopus CLI exists"        "[ -x /usr/local/bin/nixopus ]"

    check "DATABASE_URL set"           "grep -q '^DATABASE_URL=postgres://' /opt/nixopus/.env"
    check "DB_PASSWORD is random"     "grep -q '^DB_PASSWORD=.\{32,\}' /opt/nixopus/.env"
    check "REDIS_PASSWORD is random"  "grep -q '^REDIS_PASSWORD=.\{32,\}' /opt/nixopus/.env"
    check "AUTH_SERVICE_SECRET set"   "grep -q '^AUTH_SERVICE_SECRET=.\{32,\}' /opt/nixopus/.env"
    check "JWT_SECRET set"            "grep -q '^JWT_SECRET=.\{32,\}' /opt/nixopus/.env"
    check "SELF_HOSTED=true"          "grep -q '^SELF_HOSTED=true' /opt/nixopus/.env"
    check "SSH_HOST set"              "grep -q '^SSH_HOST=' /opt/nixopus/.env"
    check "SSH_PORT set"              "grep -q '^SSH_PORT=' /opt/nixopus/.env"
    check "SSH_USER set"              "grep -q '^SSH_USER=' /opt/nixopus/.env"
    check "CADDY_HTTP_PORT set"       "grep -q '^CADDY_HTTP_PORT=' /opt/nixopus/.env"

    check "nixopus-db running"       "docker ps --format '{{.Names}}' | grep -q nixopus-db"
    check "nixopus-redis running"    "docker ps --format '{{.Names}}' | grep -q nixopus-redis"
    check "nixopus-auth running"     "docker ps --format '{{.Names}}' | grep -q nixopus-auth"
    check "nixopus-api running"      "docker ps --format '{{.Names}}' | grep -q nixopus-api"
    check "nixopus-view running"     "docker ps --format '{{.Names}}' | grep -q nixopus-view"
    check "nixopus-caddy running"    "docker ps --format '{{.Names}}' | grep -q nixopus-caddy"

    sleep 10

    local http_port="${CADDY_HTTP_PORT:-80}"
    check "API health"    "curl -sf http://127.0.0.1:${http_port}/api/v1/health | grep -q success"
    check "View responds" "curl -sf http://127.0.0.1:${http_port}/ | grep -q Nixopus"

    check "nixopus status runs"  "nixopus status"
    check "nixopus info runs"    "nixopus info"
    check "nixopus config runs"  "nixopus config"

    echo ""
    echo "Results: $pass passed, $fail failed"
    [ "$fail" -eq 0 ] && return 0 || return 1
}

echo "=== Nixopus Installer Test ==="
echo "OS: $(cat /etc/os-release 2>/dev/null | grep PRETTY_NAME | cut -d= -f2 | tr -d '"' || uname -s)"
echo "Arch: $(uname -m)"
echo ""

install_prereqs
start_dockerd
load_mocks
run_installer
verify
