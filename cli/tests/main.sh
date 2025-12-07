#!/bin/bash

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

cd "$SCRIPT_DIR"

source test_config.sh
source test_framework.sh

echo -e "${BLUE}Nixopus CLI Test Suite${NC}"

cd ..

BINARY_DIR=$(find "${HOST_DIST_DIR}" -type d -name "${HOST_BINARY_PATTERN}" 2>/dev/null | head -1)

if [[ -z "$BINARY_DIR" ]] || [[ ! -f "$BINARY_DIR/nixopus" ]] || [[ ! -x "$BINARY_DIR/nixopus" ]]; then
    echo -e "${RED}Error: CLI binary not found${NC}"
    echo "Please build it first: cd $(pwd) && bash build.sh"
    exit 1
fi

echo -e "${GREEN}Using binary at $BINARY_DIR/nixopus${NC}"

cd "$SCRIPT_DIR"

echo "Starting container..."
docker compose -f docker-compose.test.yml up -d debian12-clean 2>&1 > /dev/null

sleep 2

run_test_suite() {
    local test_file="$1"
    local test_function="$2"
    
    docker exec "${CONTAINER_NAME}" /bin/bash -c "
        source ${TEST_ROOT}/test_config.sh
        source ${TEST_FRAMEWORK}
        source ${TEST_CASES_DIR}/$test_file
        $test_function
    "
}

echo "Running basic tests..."
BASIC_EXIT=0
run_test_suite "basic.sh" "run_basic_tests" || BASIC_EXIT=$?

echo ""
echo "Running scenario tests..."
SCENARIO_EXIT=0
run_test_suite "scenarios.sh" "run_scenario_tests" || SCENARIO_EXIT=$?

echo ""

if [[ $BASIC_EXIT -eq 0 ]] && [[ $SCENARIO_EXIT -eq 0 ]]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
