#!/bin/bash
GOBIN=/Users/shaale/.gvm/gos/go1.25/bin/go

cd /Users/shaale/nixopus/api

echo "=== Running integration tests ==="
echo ""

run_suite() {
  local pkg=$1
  local label=$2
  echo "--- $label ---"
  $GOBIN test -p 1 -v -count=1 -timeout 180s "$pkg" 2>&1
  echo ""
}

run_suite "./internal/tests/notification/" "NOTIFICATION"
run_suite "./internal/tests/domain/" "DOMAIN"
run_suite "./internal/tests/mcp/" "MCP"
run_suite "./internal/tests/healthcheck/" "HEALTHCHECK"
run_suite "./internal/tests/machine/" "MACHINE (billing + existing)"
run_suite "./internal/tests/container/" "CONTAINER"

echo "=== All suites done ==="
