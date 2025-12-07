#!/bin/bash

# Container paths (inside Docker container)
export CLI_ROOT="/nixopus-cli"
export CLI_DIST_DIR="${CLI_ROOT}/dist"
export CLI_BINARY_DIR="${CLI_DIST_DIR}/nixopus_linux_amd64"
export CLI_BINARY="${CLI_BINARY_DIR}/nixopus"
export CLI_WRAPPER="${CLI_DIST_DIR}/nixopus"
export INSTALL_SCRIPT="${CLI_ROOT}/install.sh"

# Test framework paths
export TEST_ROOT="${CLI_ROOT}/tests"
export TEST_FRAMEWORK="${TEST_ROOT}/test_framework.sh"
export TEST_CASES_DIR="${TEST_ROOT}/cases"

# Temporary test directories
export TEMP_TEST_DIR="/tmp/nixopus-install-test"
export LOCAL_BIN_DIR="${HOME}/local/bin"

# Container settings
export CONTAINER_NAME="nixopus-test-debian12-clean"
export COMPOSE_FILE="${TEST_ROOT}/docker-compose.test.yml"

# Host paths (for main.sh)
export HOST_DIST_DIR="dist"
export HOST_BINARY_PATTERN="nixopus_*"

