#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
MOCKS_TAR="/tmp/nixopus-mocks.tar.gz"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

DISTROS=(
    "ubuntu:22.04"
    "ubuntu:24.04"
    "debian:12"
    "rockylinux:9"
    "alpine:3.20"
)

PASSED=0
FAILED=0
SKIPPED=0

log()      { echo -e "$1"; }
log_head() { echo -e "\n${CYAN}${BOLD}━━━ $1 ━━━${NC}"; }
log_pass() { echo -e "  ${GREEN}PASS${NC}  $1"; PASSED=$((PASSED + 1)); }
log_fail() { echo -e "  ${RED}FAIL${NC}  $1"; FAILED=$((FAILED + 1)); }
log_skip() { echo -e "  ${YELLOW}SKIP${NC}  $1"; SKIPPED=$((SKIPPED + 1)); }

build_mocks() {
    log_head "Building mock service images"
    for svc in api auth view agent; do
        docker build -q -t "nixopus-mock-$svc" \
            -f "$ROOT_DIR/test/mocks/Dockerfile" \
            "$ROOT_DIR/test/mocks/$svc/" >/dev/null
        log_pass "nixopus-mock-$svc"
    done

    log "  ${DIM}Saving to $MOCKS_TAR...${NC}"
    docker save nixopus-mock-api nixopus-mock-auth nixopus-mock-view nixopus-mock-agent \
        | gzip > "$MOCKS_TAR"
    log_pass "Mocks saved ($(du -h "$MOCKS_TAR" | awk '{print $1}'))"
}

test_distro() {
    local distro="$1"
    local label
    label=$(echo "$distro" | tr ':/' '-')
    local container_name="nixopus-test-${label}"

    log_head "Testing: $distro"

    docker rm -f "$container_name" 2>/dev/null || true

    local start_time
    start_time=$(date +%s)

    if docker run --rm --privileged \
        --name "$container_name" \
        -v "$ROOT_DIR:/installer:ro" \
        -v "$SCRIPT_DIR/test-inside-container.sh:/test.sh:ro" \
        -v "$MOCKS_TAR:/mocks.tar.gz:ro" \
        -e NIXOPUS_INSTALLER_DIR=/installer \
        "$distro" \
        sh -c 'command -v bash >/dev/null 2>&1 || (apk add --no-cache bash >/dev/null 2>&1 || apt-get update -qq && apt-get install -y -qq bash >/dev/null 2>&1 || dnf install -y -q bash >/dev/null 2>&1); exec bash /test.sh' 2>&1 | tee "/tmp/nixopus-test-${label}.log"; then

        local duration=$(( $(date +%s) - start_time ))
        log_pass "$distro (${duration}s)"
    else
        local duration=$(( $(date +%s) - start_time ))
        log_fail "$distro (${duration}s) — log: /tmp/nixopus-test-${label}.log"
    fi
}

cleanup() {
    log_head "Cleanup"
    for distro in "${DISTROS[@]}"; do
        local label
        label=$(echo "$distro" | tr ':/' '-')
        docker rm -f "nixopus-test-${label}" 2>/dev/null || true
    done
    docker rm -f nixopus-db nixopus-redis nixopus-auth nixopus-api nixopus-view nixopus-caddy nixopus-agent 2>/dev/null || true
    log "  Done"
}

show_results() {
    log_head "Results"
    log "  ${GREEN}Passed: $PASSED${NC}"
    log "  ${RED}Failed: $FAILED${NC}"
    log "  ${YELLOW}Skipped: $SKIPPED${NC}"
    echo ""
    [ "$FAILED" -eq 0 ]
}

usage() {
    echo "Usage: $0 [options] [distro...]"
    echo ""
    echo "Options:"
    echo "  --build-mocks    Only build mock images"
    echo "  --all            Test all distros (default)"
    echo "  --cleanup        Remove test containers"
    echo "  --help           Show this help"
    echo ""
    echo "Distros: ${DISTROS[*]}"
    echo ""
    echo "Examples:"
    echo "  $0                       # Test all distros"
    echo "  $0 ubuntu:22.04          # Test single distro"
    echo "  $0 ubuntu:22.04 alpine:3.20  # Test specific distros"
}

main() {
    local targets=()
    local build_only=false

    for arg in "$@"; do
        case "$arg" in
            --build-mocks) build_only=true ;;
            --all)         targets=("${DISTROS[@]}") ;;
            --cleanup)     cleanup; exit 0 ;;
            --help|-h)     usage; exit 0 ;;
            *)             targets+=("$arg") ;;
        esac
    done

    [ ${#targets[@]} -eq 0 ] && targets=("${DISTROS[@]}")

    echo -e "${BOLD}Nixopus Installer Test Suite${NC}"
    echo ""

    build_mocks
    if [ "$build_only" = true ]; then exit 0; fi

    for distro in "${targets[@]}"; do
        test_distro "$distro"
    done

    show_results
}

main "$@"
