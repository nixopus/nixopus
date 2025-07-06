#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/test_install.sh"
source "$SCRIPT_DIR/util.sh"

# Configuration for the virtualized installation
declare -A CONFIG

# Initialize configuration with default values
function init_virtualized_config() {
    local distro="$1"
    local api_domain="$2"
    local app_domain="$3"
    local proxy_url="$4"
    local app_port="$5"
    local api_port="$6"
    local email="$7"
    local password="$8"
    local env="$9"
    
    CONFIG=(
        ["distro"]="${distro:-debian/11}"
        ["api_domain"]="${api_domain:-}"
        ["app_domain"]="${app_domain:-}"
        ["proxy_url"]="${proxy_url:-127.0.0.1}"
        ["app_port"]="${app_port:-9984}"
        ["api_port"]="${api_port:-9985}"
        ["email"]="${email:-admin@nixopus.com}"
        ["password"]="${password:-admin123}"
        ["env"]="${env:-production}"
        ["container_name"]=""
        ["base_distro"]=""
        ["show_in_console"]="true"
    )
    
    CONFIG["base_distro"]=$(echo "${CONFIG[distro]}" | cut -d'/' -f1)
    CONFIG["container_name"]=$(generate_container_name)
}

# sets the proxy for the container
function set_proxy_for_lxd() {
    local container_name="$1"
    echo "Setting proxy for container: $container_name"
    local proxy_url="$2"
    local proxy_internal_port="$3"
    local proxy_external_port="$4"
    local proxy_name="$5"
    sudo lxc config device add "$container_name" "$proxy_name" proxy listen=tcp:0.0.0.0:"$proxy_external_port" connect=tcp:"$proxy_url":"$proxy_internal_port"
}

# overrides the env variables present inside the /etc/nixopus/source/api/.env and /etc/nixopus/source/view/.env files
# overrriding is carried out because the nixopus will be installed in the container, so we need to update the ports and set proxy for the services run inside the container
function override_env_variables() {
    local container_name="$1"
    local app_domain="$2"
    local api_domain="$3"

    local app_env_file="/etc/nixopus/source/view/.env"
    local api_env_file="/etc/nixopus/source/api/.env"

    log_message "INFO" "Overriding environment variables in container: $container_name" "${CONFIG[show_in_console]}"
    
    sudo lxc exec "$container_name" -- bash -c "
        if [[ ! -f \"$app_env_file\" ]]; then
            echo \"File not found: $app_env_file\"
            exit 1
        fi
        if [[ ! -f \"$api_env_file\" ]]; then
            echo \"File not found: $api_env_file\"
            exit 1
        fi

        sed -i \"s|^WEBSOCKET_URL=.*|WEBSOCKET_URL=wss://$app_domain/ws|\" \"$app_env_file\"
        sed -i \"s|^WEBHOOK_URL=.*|WEBHOOK_URL=https://$app_domain/api/v1/webhook|\" \"$app_env_file\"
        sed -i \"s|^API_URL=.*|API_URL=https://$api_domain/api|\" \"$app_env_file\"
        
        sed -i \"s|^ALLOWED_ORIGIN=.*|ALLOWED_ORIGIN=https://$app_domain|\" \"$api_env_file\"
        
        echo \"Environment variables updated successfully\"
    "
}

# Validate  installation parameters
function validate_virtualized_params() {
    log_message "INFO" "Validating virtualized installation parameters..." "${CONFIG[show_in_console]}"
    local validation_failed=false
    
    validate_environment "${CONFIG[env]}" || validation_failed=true
    validate_lxd || validation_failed=true
    validate_sudo || validation_failed=true
    validate_file "${SCRIPT_DIR}/util.sh" "util.sh" || validation_failed=true
    
    # Check if ports are available
    log_message "INFO" "Checking if ports are available..." "${CONFIG[show_in_console]}"
    is_port_available "${CONFIG[app_port]}" || validation_failed=true
    is_port_available "${CONFIG[api_port]}" || validation_failed=true
    
    # Validate domains if provided
    if [ -n "${CONFIG[api_domain]}" ]; then
        validate_domain "${CONFIG[api_domain]}" "API" "false" || validation_failed=true
    fi
    if [ -n "${CONFIG[app_domain]}" ]; then
        validate_domain "${CONFIG[app_domain]}" "App" "false" || validation_failed=true
    fi
    
    if [ "$validation_failed" = true ]; then
        log_message "ERROR" "Validation failed!" "${CONFIG[show_in_console]}"
        return 1
    fi
    
    log_message "INFO" "All validations passed!" "${CONFIG[show_in_console]}"
    return 0
}

function main() {
    echo "Installing Nixopus through virtualizer"
    local distro="$1"
    local api_domain="$2"
    local app_domain="$3"
    local proxy_url="$4"
    local app_port="$5" 
    local api_port="$6" 
    
    local email="admin@nixopus.com"
    local password="admin123"
    local env="production"
    
    if [[ ! "$distro" ]]; then
        distro="debian/11"
    fi
    if [[ ! "$proxy_url" ]]; then
        proxy_url="127.0.0.1"
    fi
    if [[ ! "$app_port" ]]; then
        app_port="9984"
    fi
    if [[ ! "$api_port" ]]; then
        api_port="9985"
    fi

    init_virtualized_config "$distro" "$api_domain" "$app_domain" "$proxy_url" "$app_port" "$api_port" "$email" "$password" "$env"

    log_message "INFO" "Starting Nixopus virtualized installation" "${CONFIG[show_in_console]}"
    log_message "INFO" "Distribution: ${CONFIG[distro]}" "${CONFIG[show_in_console]}"
    log_message "INFO" "Container: ${CONFIG[container_name]}" "${CONFIG[show_in_console]}"
    log_message "INFO" "Environment: ${CONFIG[env]}" "${CONFIG[show_in_console]}"
    log_message "INFO" "App Port: ${CONFIG[app_port]}" "${CONFIG[show_in_console]}"
    log_message "INFO" "API Port: ${CONFIG[api_port]}" "${CONFIG[show_in_console]}"

    if ! validate_virtualized_params; then
        log_message "ERROR" "Parameter validation failed, exiting" "${CONFIG[show_in_console]}"
        exit 1
    fi

    log_message "INFO" "Installing dependencies..." "${CONFIG[show_in_console]}"
    install_dependencies
    
    log_message "INFO" "Creating LXD container..." "${CONFIG[show_in_console]}"
    if ! create_lxd_container; then
        log_message "ERROR" "Failed to create container, exiting..." "${CONFIG[show_in_console]}"
        exit 1
    fi
    
    log_message "INFO" "Created Container name: '${CONFIG[container_name]}'" "${CONFIG[show_in_console]}"
    
    log_message "INFO" "Installing dependencies in container..." "${CONFIG[show_in_console]}"
    install_dependencies_in_container
    
    log_message "INFO" "Starting installation script..." "${CONFIG[show_in_console]}"
    if run_installation_script "$(build_installation_command)" "${CONFIG[container_name]}"; then
        test_result="PASSED"
    else
        test_result="FAILED"
    fi
    
    log_message "INFO" "Installation script completed with result: $test_result" "${CONFIG[show_in_console]}"
    if [[ "$test_result" != "PASSED" ]]; then
        log_message "ERROR" "Installation failed, exiting..." "${CONFIG[show_in_console]}"
        exit 1
    fi
    
    override_env_variables "${CONFIG[container_name]}" "${CONFIG[app_domain]}" "${CONFIG[api_domain]}"
    set_proxy_for_lxd "${CONFIG[container_name]}" "${CONFIG[proxy_url]}" "7443" "${CONFIG[app_port]}" "app"
    set_proxy_for_lxd "${CONFIG[container_name]}" "${CONFIG[proxy_url]}" "8443" "${CONFIG[api_port]}" "api"
    
    log_message "INFO" "Proxy set for container: ${CONFIG[container_name]}" "${CONFIG[show_in_console]}"
    log_message "INFO" "Virtualized installation completed successfully!" "${CONFIG[show_in_console]}"
}

main "$@"
