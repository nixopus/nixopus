#!/bin/bash

if [[ -f "/nixopus-cli/tests/test_config.sh" ]]; then
    source "/nixopus-cli/tests/test_config.sh"
fi
if [[ -f "/nixopus-cli/tests/test_framework.sh" ]]; then
    source "/nixopus-cli/tests/test_framework.sh"
fi

CLI_BINARY_DIR=$(find "${CLI_DIST_DIR}" -type d -name "nixopus_*" | head -1)
CLI_BINARY="${CLI_BINARY_DIR}/nixopus"

test_binary_exists() {
    [[ -f "$CLI_BINARY" ]]
}

test_binary_executable() {
    [[ -x "$CLI_BINARY" ]]
}

test_binary_version() {
    [[ -x "$CLI_WRAPPER" ]] && "$CLI_WRAPPER" --version &> /dev/null
}

test_binary_help() {
    [[ -x "$CLI_WRAPPER" ]] && "$CLI_WRAPPER" --help &> /dev/null
}

test_git_installed() {
    command -v git &> /dev/null
}

test_curl_installed() {
    command -v curl &> /dev/null
}

run_basic_tests() {
    test_init
    
    echo "Running basic tests..."
    
    test_run "CLI binary exists" test_binary_exists
    test_run "CLI binary is executable" test_binary_executable
    test_run "CLI binary runs (--version)" test_binary_version
    test_run "CLI binary shows help" test_binary_help
    test_run "Git installed" test_git_installed
    test_run "Curl installed" test_curl_installed
    
    test_summary
}
