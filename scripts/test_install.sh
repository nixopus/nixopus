#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/util.sh"

# This script is used for testing the installation script across multiple distributions.
# It will create containers for each distribution, run the installation, and report results.

set -euo pipefail
OS="$(uname)"

# Base URL for GitHub raw content
GITHUB_RAW_BASE="https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/feat/hosted_action_runner"

# DISTRO_MATRIX is the list of distributions to test these names are from the lxc image list command
# TODO: Uncomment the distributions to test once the first two tests seemed to be working
DISTRO_MATRIX=(
    # "alpine/3.19"
    # "fedora/41" = working distribution
    # "archlinux"
    # "debian/11" #  working distribution
    # "centos/9-Stream" # working distribution
    "gentoo/openrc"
    # "ubuntu:18.04" # working distribution
)

# Maximum timeout for the installation script to run in seconds
TIMEOUT=600

# Configuration for the test to run all the params
declare -A CONFIG

# Initialize  configuration with default values
function init_config() {
    local email="$1"
    local password="$2"
    local api_domain="$3"
    local app_domain="$4"
    local env="$5"
    local show_in_console="$6"
    
    CONFIG=(
        ["email"]="${email:-}"
        ["password"]="${password:-}"
        ["api_domain"]="${api_domain:-}"
        ["app_domain"]="${app_domain:-}"
        ["env"]="${env:-production}"
        ["distro"]=""
        ["base_distro"]=""
        ["container_name"]=""
        ["test_name"]=""
        ["show_in_console"]="${show_in_console:-false}"
    )
}

# Set current test parameters
function set_test_params() {
    local distro="$1"
    local test_name="${2:-custom}"
    
    CONFIG["distro"]="$distro"
    CONFIG["test_name"]="$test_name"
    CONFIG["container_name"]=$(generate_container_name)
    CONFIG["base_distro"]=$(echo "$distro" | cut -d'/' -f1)
}

# Parse command line arguments
function parse_arguments() {
    local email=""
    local password=""
    local api_domain=""
    local app_domain=""
    local env="production"
    local show_in_console="false"
    while [[ $# -gt 0 ]]; do
        case $1 in
            --email=*)
                email="${1#*=}"
                shift
                ;;
            --password=*)
                password="${1#*=}"
                shift
                ;;
            --api-domain=*)
                api_domain="${1#*=}"
                shift
                ;;
            --app-domain=*)
                app_domain="${1#*=}"
                shift
                ;;
            --show-in-console=*)
                show_in_console="${1#*=}"
                shift
                ;;
            *)
                echo "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    echo "$email:$password:$api_domain:$app_domain:$env:$show_in_console"
}

# Generate a random string and add it to the test-container- prefix
function generate_container_name() {
    echo "test-container-$(date +%s%N | md5sum | cut -c1-8)"
}

# Create a new LXD container with Docker privileges
function create_lxd_container() {
    log_message "INFO" "Creating LXD container: ${CONFIG[container_name]} with distro: ${CONFIG[distro]}" "${CONFIG[show_in_console]}"
    
    if sudo lxc launch images:"${CONFIG[distro]}" "${CONFIG[container_name]}"; then
        sudo lxc config set "${CONFIG[container_name]}" security.privileged true
        sudo lxc config set "${CONFIG[container_name]}" security.nesting true
        sudo lxc restart "${CONFIG[container_name]}"
        sleep 60
        sudo lxc exec "${CONFIG[container_name]}" -- cloud-init status --wait || true
        log_message "INFO" "Container ${CONFIG[container_name]} created successfully" "${CONFIG[show_in_console]}"
        return 0
    else
        log_message "ERROR" "Failed to create container ${CONFIG[container_name]}" "${CONFIG[show_in_console]}"
        return 1
    fi
}

# Install dependencies in the container
function install_dependencies_in_container() {
    log_message "INFO" "Installing dependencies in container: ${CONFIG[container_name]}" "${CONFIG[show_in_console]}"
    log_message "DEBUG" "Base distribution: ${CONFIG[base_distro]}" "${CONFIG[show_in_console]}"

    # Push the util.sh script to the container
    log_message "DEBUG" "Pushing util.sh to container..." "${CONFIG[show_in_console]}"
    sudo lxc file push "$SCRIPT_DIR/util.sh" "${CONFIG[container_name]}/tmp/util.sh"
    
    # Set the package manager for the container
    sudo lxc exec "${CONFIG[container_name]}" -- env OS=Linux bash -c "
        source /tmp/util.sh
        ensure_command_installed 'python3' 'python3'
        ensure_command_installed 'pip3' 'python3-pip'
        ensure_command_installed 'git' 'git'
        ensure_command_installed 'openssl' 'openssl'
        ensure_command_installed 'curl' 'curl'
    "
    
    # Install additional distro-specific packages
    case ${CONFIG[base_distro]} in
        "debian"|"ubuntu")
            log_message "DEBUG" "Installing additional packages for ${CONFIG[base_distro]}" "${CONFIG[show_in_console]}"
            sudo lxc exec "${CONFIG[container_name]}" -- apt-get install -y python3-venv
            ;;
        "gentoo")
            log_message "DEBUG" "Installing additional packages for ${CONFIG[base_distro]}" "${CONFIG[show_in_console]}"
            sudo lxc exec "${CONFIG[container_name]}" -- emerge --noreplace nix
            ;;
    esac
    
    # Install Docker using the utility function
    log_message "INFO" "Installing Docker..." "${CONFIG[show_in_console]}"
    sudo lxc exec "${CONFIG[container_name]}" -- bash -c "
        source /tmp/util.sh
        install_docker
    "
    
    # Clean up the temporary file
    sudo lxc exec "${CONFIG[container_name]}" -- rm -f /tmp/util.sh
    
    log_message "INFO" "Finished installing dependencies in container: ${CONFIG[container_name]}" "${CONFIG[show_in_console]}"
}

# Build installation command using global config
function build_installation_command() {
    local cmd="sudo bash -c \"\$(curl -sSL $GITHUB_RAW_BASE/scripts/install.sh)\""
    
    if [ -n "${CONFIG[email]}" ]; then
        cmd="$cmd --email=${CONFIG[email]}"
    fi
    
    if [ -n "${CONFIG[password]}" ]; then
        cmd="$cmd --password=${CONFIG[password]}"
    fi
    
    if [ -n "${CONFIG[api_domain]}" ]; then
        cmd="$cmd --api-domain=${CONFIG[api_domain]}"
    fi
    
    if [ -n "${CONFIG[app_domain]}" ]; then
        cmd="$cmd --app-domain=${CONFIG[app_domain]}"
    fi
    
    cmd="$cmd --env=${CONFIG[env]}"
    cmd="$cmd --debug"
    
    echo "$cmd"
}

# Run installation script
function run_installation_script() {
    local command="$1"
    local container_name="$2"
    local start_time=$(date +%s)
    local error_message=""
    
    log_message "INFO" "Running installation script in container: $container_name" "${CONFIG[show_in_console]}"
    log_message "INFO" "This may take several minutes..." "${CONFIG[show_in_console]}"
    
    if timeout "$TIMEOUT" sudo lxc exec "$container_name" -- bash -c "set -x; $command" 2>&1; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        log_message "INFO" "Installation script completed successfully in ${duration}s" "${CONFIG[show_in_console]}"
        return 0
    else
        local exit_code=$?
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        if [ $exit_code -eq 124 ]; then
            error_message="Installation script timed out after $TIMEOUT seconds"
            log_message "ERROR" "$error_message" "${CONFIG[show_in_console]}"
        else
            error_message="Installation script failed with exit code: $exit_code"
            log_message "ERROR" "$error_message" "${CONFIG[show_in_console]}"
        fi
        
        CONFIG["error_message"]="$error_message"
        CONFIG["exit_code"]="$exit_code"
        CONFIG["duration"]="$duration"
        
        return $exit_code
    fi
}

# Stop and delete container
function cleanup_container() {
    if [ -n "${CONFIG[container_name]}" ]; then
        log_message "INFO" "Stopping container: ${CONFIG[container_name]}" "${CONFIG[show_in_console]}"
        sudo lxc stop "${CONFIG[container_name]}" 2>/dev/null || true
        
        log_message "INFO" "Deleting container: ${CONFIG[container_name]}" "${CONFIG[show_in_console]}"
        sudo lxc delete "${CONFIG[container_name]}" 2>/dev/null || true
    fi
}

# Test installation on a single distribution with specific parameters
function test_distribution_with_params() {
    local test_start_time=$(date +%s)
    local test_result=""
    local exit_code=0
    local error_message=""
    local test_name="${CONFIG[distro]}-${CONFIG[test_name]}"
    local metadata="container:${CONFIG[container_name]},base_distro:${CONFIG[base_distro]}"
    
    log_message "INFO" "==========================================" "${CONFIG[show_in_console]}"
    log_message "INFO" "Testing: ${CONFIG[distro]} - ${CONFIG[test_name]}" "${CONFIG[show_in_console]}"
    log_message "INFO" "Base Distro: ${CONFIG[base_distro]}" "${CONFIG[show_in_console]}"
    log_message "INFO" "Container: ${CONFIG[container_name]}" "${CONFIG[show_in_console]}"
    log_message "INFO" "Environment: ${CONFIG[env]}" "${CONFIG[show_in_console]}"
    log_message "INFO" "Email: ${CONFIG[email]:-<not set>}" "${CONFIG[show_in_console]}"
    log_message "INFO" "Password: ${CONFIG[password]:+<set>}" "${CONFIG[show_in_console]}"
    log_message "INFO" "API Domain: ${CONFIG[api_domain]:-<not set>}" "${CONFIG[show_in_console]}"
    log_message "INFO" "App Domain: ${CONFIG[app_domain]:-<not set>}" "${CONFIG[show_in_console]}"
    log_message "INFO" "==========================================" "${CONFIG[show_in_console]}"
    
    if ! create_lxd_container; then
        test_result="SKIPPED"
        error_message="Failed to create container"
        exit_code=1
    else
        install_dependencies_in_container

        log_message "INFO" "Starting installation script..." "${CONFIG[show_in_console]}"
        if run_installation_script "$(build_installation_command)" "${CONFIG[container_name]}"; then
            test_result="PASSED"
            exit_code=0
            error_message=""
        else
            test_result="FAILED"
            exit_code="${CONFIG[exit_code]:-1}"
            error_message="${CONFIG[error_message]:-Unknown error}"
        fi
    fi
    
    local test_end_time=$(date +%s)
    local test_duration=$((test_end_time - test_start_time))
    
    log_message "INFO" "Installation script completed with result: $test_result" "${CONFIG[show_in_console]}"
    
    cleanup_container
    
    return $exit_code
}

# Validate all the parameters for the test
function validate_test_params() {
    log_message "INFO" "Validating environment and arguments..." "${CONFIG[show_in_console]}"
    local validation_failed=false
    
    validate_environment "${CONFIG[env]}" || validation_failed=true
    validate_lxd || validation_failed=true
    validate_sudo || validation_failed=true
    validate_file "${SCRIPT_DIR}/util.sh" "util.sh" || validation_failed=true
    validate_array "Distro matrix" "${DISTRO_MATRIX[@]}" || validation_failed=true

    local email_strict="false"      
    local password_strict="false"   
    local api_domain_strict="false" 
    local app_domain_strict="false" 
    local url_strict="false"        
    
    validate_email "${CONFIG[email]}" "$email_strict" || validation_failed=true
    validate_password "${CONFIG[password]}" "$password_strict" || validation_failed=true
    validate_domain "${CONFIG[api_domain]}" "API" "$api_domain_strict" || validation_failed=true
    validate_domain "${CONFIG[app_domain]}" "App" "$app_domain_strict" || validation_failed=true
    validate_url "${GITHUB_RAW_BASE}/scripts/install.sh" "Installation script" "$url_strict" || validation_failed=true
    
    if [ "$validation_failed" = true ]; then
        log_message "ERROR" "Validation failed!" "${CONFIG[show_in_console]}"
        return 1
    fi
    
    log_message "INFO" "All validations passed!" "${CONFIG[show_in_console]}"
    return 0
}

function main() {
    local args
    args=$(parse_arguments "$@")
    IFS=':' read -r email password api_domain app_domain env show_in_console <<< "$args"
    init_config "$email" "$password" "$api_domain" "$app_domain" "$env" "$show_in_console"

    log_message "INFO" "Starting Nixopus installation test suite" "${CONFIG[show_in_console]}"
    log_message "INFO" "Testing ${#DISTRO_MATRIX[@]} distributions with timeout: ${TIMEOUT}s" "${CONFIG[show_in_console]}"
    log_message "INFO" "Show in console: ${CONFIG[show_in_console]}" "${CONFIG[show_in_console]}"
    
    if ! validate_test_params; then
        log_message "ERROR" "Parameter validation failed, exiting" "${CONFIG[show_in_console]}"
        exit 1
    fi
    
    install_dependencies

    # TODO: Enable parallel testing
    for distro in "${DISTRO_MATRIX[@]}"; do
        set_test_params "$distro" "custom"
        test_distribution_with_params
    done
    
    exit 0
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
