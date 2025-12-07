#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

test_counter=0
passed_tests=0
failed_tests=0

test_init() {
    test_counter=0
    passed_tests=0
    failed_tests=0
}

test_run() {
    local test_name="$1"
    shift
    
    test_counter=$((test_counter + 1))
    echo -e "${BLUE}[TEST]${NC} #${test_counter}: ${test_name}"
    
    if "$@"; then
        echo -e "${GREEN}[PASS]${NC} ${test_name}"
        passed_tests=$((passed_tests + 1))
        return 0
    else
        echo -e "${RED}[FAIL]${NC} ${test_name}"
        failed_tests=$((failed_tests + 1))
        return 1
    fi
}

test_summary() {
    echo ""
    echo "Test Summary:"
    echo "Total: ${test_counter}"
    echo -e "${GREEN}Passed: ${passed_tests}${NC}"
    echo -e "${RED}Failed: ${failed_tests}${NC}"
    
    if [[ $failed_tests -eq 0 ]]; then
        echo -e "${GREEN}All tests passed!${NC}"
        return 0
    else
        echo -e "${RED}Some tests failed!${NC}"
        return 1
    fi
}

