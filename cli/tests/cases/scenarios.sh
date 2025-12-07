#!/bin/bash

if [[ -f "/nixopus-cli/tests/test_config.sh" ]]; then
    source "/nixopus-cli/tests/test_config.sh"
fi
if [[ -f "/nixopus-cli/tests/test_framework.sh" ]]; then
    source "/nixopus-cli/tests/test_framework.sh"
fi

test_cli_help() {
    [[ -x "$CLI_WRAPPER" ]] && "$CLI_WRAPPER" --help &> /dev/null
}

test_cli_version() {
    [[ -x "$CLI_WRAPPER" ]] && "$CLI_WRAPPER" --version &> /dev/null
}

test_local_installation() {
    local temp_dir="${TEMP_TEST_DIR}"
    mkdir -p "$temp_dir"
    cp -r "${CLI_DIST_DIR}" "$temp_dir/"
    cp "${INSTALL_SCRIPT}" "$temp_dir/"
    
    cd "$temp_dir"
    bash "$temp_dir/install.sh" --local &> /dev/null
    
    [[ -f "${HOME}/.local/bin/nixopus" ]] && "${HOME}/.local/bin/nixopus" --version &> /dev/null
}

test_install_dry_run() {
    [[ -x "$CLI_WRAPPER" ]] && "$CLI_WRAPPER" install --dry-run &> /dev/null
}

test_install_deps_dry_run() {
    [[ -x "$CLI_WRAPPER" ]] && "$CLI_WRAPPER" install deps --dry-run &> /dev/null
}

test_install_ssh_dry_run() {
    [[ -x "$CLI_WRAPPER" ]] && "$CLI_WRAPPER" install ssh --dry-run &> /dev/null
}

run_scenario_tests() {
    test_init
    
    echo "Running scenario tests..."
    
    test_run "CLI help command" test_cli_help
    test_run "CLI version command" test_cli_version
    test_run "Local installation (--local)" test_local_installation
    test_run "Install command (--dry-run)" test_install_dry_run
    test_run "Install deps (--dry-run)" test_install_deps_dry_run
    test_run "Install ssh (--dry-run)" test_install_ssh_dry_run
    
    test_summary
}
