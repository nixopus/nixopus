#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/test_install.sh"
source "$SCRIPT_DIR/util.sh"

# Configuration for the virtualized installation
declare -A CONFIG

# Initialize configuration with default values and parameter processing
function init_virtualized_config() {
    local distro="${1:-}"
    local api_domain="${2:-}"
    local app_domain="${3:-}"
    local proxy_url="${4:-}"
    local app_port="${5:-}"
    local api_port="${6:-}"
    local email="${7:-}"
    local password="${8:-}"
    local env="${9:-}"
    local show_in_console="${10:-}"
    local internal_app_proxy_port="${11:-}"
    local internal_api_proxy_port="${12:-}"
    
    [[ -z "$internal_app_proxy_port" ]] && internal_app_proxy_port="7443"
    [[ -z "$internal_api_proxy_port" ]] && internal_api_proxy_port="8443"
    [[ -z "$distro" ]] && distro="debian/11"
    [[ -z "$proxy_url" ]] && proxy_url="127.0.0.1"
    [[ -z "$app_port" ]] && app_port=$(generate_random_available_port)
    [[ -z "$api_port" ]] && api_port=$(generate_random_available_port)
    [[ -z "$env" ]] && env="production"
    [[ -z "$show_in_console" ]] && show_in_console="true"
    if [[ -z "$app_domain" ]]; then
        app_domain=$(generate_random_string 10).nixopus.com
    fi
    if [[ -z "$api_domain" ]]; then
        api_domain=$(generate_random_string 10).nixopus.com
    fi
    
    local container_name=$(generate_container_name)
    local app_proxy_name="app-proxy-${container_name}"
    local api_proxy_name="api-proxy-${container_name}"
    
    CONFIG=(
        ["distro"]="$distro"
        ["api_domain"]="${api_domain:-}"
        ["app_domain"]="${app_domain:-}"
        ["proxy_url"]="$proxy_url"
        ["app_port"]="$app_port"
        ["api_port"]="$api_port"
        ["email"]="$email"
        ["password"]="$password"
        ["env"]="$env"
        ["container_name"]="$container_name"
        ["base_distro"]=""
        ["show_in_console"]="$show_in_console"
        ["internal_app_proxy_port"]="$internal_app_proxy_port"
        ["internal_api_proxy_port"]="$internal_api_proxy_port"
        ["app_proxy_name"]="$app_proxy_name"
        ["api_proxy_name"]="$api_proxy_name"
    )
    
    CONFIG["base_distro"]=$(echo "${CONFIG[distro]}" | cut -d'/' -f1)
}

# Set up proxy for a specific service (app or api)
function set_proxy_for_lxd() {
    local service_type="$1"  # "app" or "api"
    local container_name="${CONFIG[container_name]}"
    local proxy_url="${CONFIG[proxy_url]}"
    
    # Determine configuration based on service type
    if [[ "$service_type" == "app" ]]; then
        local proxy_internal_port="${CONFIG[internal_app_proxy_port]}"
        local proxy_external_port="${CONFIG[app_port]}"
        local proxy_name="${CONFIG[app_proxy_name]}"
    elif [[ "$service_type" == "api" ]]; then
        local proxy_internal_port="${CONFIG[internal_api_proxy_port]}"
        local proxy_external_port="${CONFIG[api_port]}"
        local proxy_name="${CONFIG[api_proxy_name]}"
    else
        log_message "ERROR" "Invalid service type: $service_type. Must be 'app' or 'api'"
        return 1
    fi
    
    log_message "INFO" "Setting up ${service_type^^} proxy for container: $container_name"
    log_message "INFO" "External port: $proxy_external_port -> Internal: $proxy_url:$proxy_internal_port"
    
    sudo lxc config device add "$container_name" "$proxy_name" proxy listen=tcp:0.0.0.0:"$proxy_external_port" connect=tcp:"$proxy_url":"$proxy_internal_port"
}

# Set up both app and API proxies
function setup_all_proxies() {
    set_proxy_for_lxd "app"
    set_proxy_for_lxd "api"
}

# overrides the env variables present inside the /etc/nixopus/source/api/.env and /etc/nixopus/source/view/.env files
# overriding is carried out because the nixopus will be installed in the container, so we need to update the ports and set proxy for the services run inside the container
function override_env_variables() {
    local container_name="${CONFIG[container_name]}"
    local app_domain="${CONFIG[app_domain]}"
    local api_domain="${CONFIG[api_domain]}"

    local app_env_file="/etc/nixopus/source/view/.env"
    local api_env_file="/etc/nixopus/source/api/.env"

    log_message "INFO" "Overriding environment variables in container: $container_name"
    
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

# Validate installation parameters
function validate_virtualized_params() {
    log_message "INFO" "Validating virtualized installation parameters..."
    local validation_failed=false
    
    validate_environment "${CONFIG[env]}" || validation_failed=true
    validate_lxd || validation_failed=true
    validate_sudo || validation_failed=true
    validate_file "${SCRIPT_DIR}/util.sh" "util.sh" || validation_failed=true
    
    log_message "INFO" "Checking if ports are available..."
    is_port_available "${CONFIG[app_port]}" || validation_failed=true
    is_port_available "${CONFIG[api_port]}" || validation_failed=true
    
    if [ -n "${CONFIG[api_domain]}" ]; then
        validate_domain "${CONFIG[api_domain]}" "API" "false" || validation_failed=true
    fi
    if [ -n "${CONFIG[app_domain]}" ]; then
        validate_domain "${CONFIG[app_domain]}" "App" "false" || validation_failed=true
    fi
    
    if [ "$validation_failed" = true ]; then
        log_message "ERROR" "Validation failed!"
        return 1
    fi
    
    log_message "INFO" "All validations passed!"
    return 0
}

function display_config_summary() {
    log_message "INFO" "Configuration Summary:"
    log_message "INFO" "  Distribution: ${CONFIG[distro]}"
    log_message "INFO" "  Container: ${CONFIG[container_name]}"
    log_message "INFO" "  Environment: ${CONFIG[env]}"
    log_message "INFO" "  App Port: ${CONFIG[app_port]}"
    log_message "INFO" "  API Port: ${CONFIG[api_port]}"
    log_message "INFO" "  Proxy URL: ${CONFIG[proxy_url]}"
    log_message "INFO" "  Internal App Proxy Port: ${CONFIG[internal_app_proxy_port]}"
    log_message "INFO" "  Internal API Proxy Port: ${CONFIG[internal_api_proxy_port]}"
    log_message "INFO" "  App Proxy Name: ${CONFIG[app_proxy_name]}"
    log_message "INFO" "  API Proxy Name: ${CONFIG[api_proxy_name]}"
    if [ -n "${CONFIG[api_domain]}" ]; then
        log_message "INFO" "  API Domain: ${CONFIG[api_domain]}"
    fi
    if [ -n "${CONFIG[app_domain]}" ]; then
        log_message "INFO" "  App Domain: ${CONFIG[app_domain]}"
    fi
}

# Create Caddy configuration directory and copy files
function setup_caddy_config_directory() {
    local caddy_config_dir="/etc/nixopus/caddy"
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local project_root="$(dirname "$script_dir")"
    
    log_message "INFO" "Setting up Caddy configuration directory: $caddy_config_dir"
    
    sudo mkdir -p "$caddy_config_dir"
    
    local caddyfile_source="$project_root/helpers/Caddyfile"
    local caddy_json_source="$project_root/helpers/caddy.json"
    
    validate_file "$caddyfile_source" "Caddyfile" || return 1
    validate_file "$caddy_json_source" "caddy.json" || return 1
    
    sudo cp "$caddyfile_source" "$caddy_config_dir/"
    sudo cp "$caddy_json_source" "$caddy_config_dir/caddy.json"
    
    validate_file "$caddy_config_dir/Caddyfile" "Caddyfile" || return 1
    validate_file "$caddy_config_dir/caddy.json" "caddy.json" || return 1
    
    echo "$caddy_config_dir"
}

# Update Caddy configuration with domain and proxy settings
function update_caddy_configuration() {
    local caddy_config_dir="$1"
    local app_domain="${CONFIG[app_domain]}"
    local api_domain="${CONFIG[api_domain]}"
    local app_port="${CONFIG[app_port]}"
    local api_port="${CONFIG[api_port]}"
    local proxy_url="${CONFIG[proxy_url]}"
    
    local caddy_config_file="$caddy_config_dir/caddy.json"
    local app_proxy_url="$proxy_url:$app_port"
    local api_proxy_url="$proxy_url:$api_port"
    
    validate_file "$caddy_config_file" "Caddy configuration" || return 1
    
    log_message "INFO" "Updating Caddy configuration:"
    log_message "INFO" "  App Domain: $app_domain -> $app_proxy_url"
    log_message "INFO" "  API Domain: $api_domain -> $api_proxy_url"
    
    sudo sed -i "s/{env.APP_DOMAIN}/$app_domain/g" "$caddy_config_file"
    sudo sed -i "s/{env.API_DOMAIN}/$api_domain/g" "$caddy_config_file"
    sudo sed -i "s/{env.APP_REVERSE_PROXY_URL}/$app_proxy_url/g" "$caddy_config_file"
    sudo sed -i "s/{env.API_REVERSE_PROXY_URL}/$api_proxy_url/g" "$caddy_config_file"
}

# Create docker-compose file for Caddy
function create_caddy_compose_file() {
    local caddy_config_dir="$1"
    local compose_file="$caddy_config_dir/docker-compose.yml"
    
    log_message "INFO" "Creating Caddy docker-compose file: $compose_file"
    
    cat << EOF | sudo tee "$compose_file" > /dev/null
version: '3.8'

services:
  nixopus-caddy:
    image: caddy:latest
    container_name: nixopus-caddy-virtualized
    ports:
      - "2019:2019"
      - "80:80"
      - "443:443"
    volumes:
      - $caddy_config_dir/Caddyfile:/etc/caddy/Caddyfile
      - $caddy_config_dir:/data
      - $caddy_config_dir:/config
    command: ["caddy", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]
    restart: unless-stopped
    networks:
      - nixopus-network

networks:
  nixopus-network:
    driver: bridge
EOF
    
    log_message "INFO" "Caddy docker-compose file created successfully"
}

# Start Caddy container
function start_caddy_container() {
    local caddy_config_dir="$1"
    
    log_message "INFO" "Starting Caddy container..."
    
    cd "$caddy_config_dir"
    if sudo docker-compose up -d; then
        log_message "INFO" "Caddy container started successfully"
        return 0
    else
        log_message "ERROR" "Failed to start Caddy container"
        return 1
    fi
}

# Wait for Caddy to be ready
function wait_for_caddy_ready() {
    local max_attempts=30
    local attempt=1
    
    log_message "INFO" "Waiting for Caddy to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "http://localhost:2019/config" > /dev/null 2>&1; then
            log_message "INFO" "Caddy is ready"
            return 0
        fi
        log_message "INFO" "Waiting for Caddy... (attempt $attempt/$max_attempts)"
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log_message "ERROR" "Caddy failed to start within expected time"
    return 1
}

# Check if Caddy container is already running
function is_caddy_running() {
    sudo docker ps --format "table {{.Names}}" | grep -q "nixopus-caddy-virtualized"
}

# Get existing Caddy configuration
function get_existing_caddy_config() {
    curl -s "http://localhost:2019/config"
}

# Check if domain already exists in configuration
function domain_exists_in_config() {
    local config_json="$1"
    local domain="$2"
    
    echo "$config_json" | jq -e --arg domain "$domain" '
        .apps.http.servers.nixopus.routes[] | 
        select(.match[0].host[] == $domain)
    ' > /dev/null 2>&1
}

# Merge new domains with existing configuration
function merge_caddy_configuration() {
    local caddy_config_dir="$1"
    local caddy_config_file="$caddy_config_dir/caddy.json"
    local app_domain="${CONFIG[app_domain]}"
    local api_domain="${CONFIG[api_domain]}"
    local app_port="${CONFIG[app_port]}"
    local api_port="${CONFIG[api_port]}"
    local proxy_url="${CONFIG[proxy_url]}"
    
    validate_file "$caddy_config_file" "Caddy configuration" || return 1
    
    local app_proxy_url="$proxy_url:$app_port"
    local api_proxy_url="$proxy_url:$api_port"
    
    log_message "INFO" "Merging new domains with existing Caddy configuration..."
    
    local existing_config
    existing_config=$(get_existing_caddy_config)
    
    if domain_exists_in_config "$existing_config" "$app_domain"; then
        log_message "WARNING" "App domain '$app_domain' already exists in Caddy configuration"
    fi
    
    if domain_exists_in_config "$existing_config" "$api_domain"; then
        log_message "WARNING" "API domain '$api_domain' already exists in Caddy configuration"
    fi
    
    local temp_config_file="$caddy_config_dir/temp_caddy.json"
    sudo cp "$caddy_config_file" "$temp_config_file"
    
    sudo sed -i "s/{env.APP_DOMAIN}/$app_domain/g" "$temp_config_file"
    sudo sed -i "s/{env.API_DOMAIN}/$api_domain/g" "$temp_config_file"
    sudo sed -i "s/{env.APP_REVERSE_PROXY_URL}/$app_proxy_url/g" "$temp_config_file"
    sudo sed -i "s/{env.API_REVERSE_PROXY_URL}/$api_proxy_url/g" "$temp_config_file"
    
    local existing_json=$(echo "$existing_config" | jq -c '.')
    local new_json=$(sudo cat "$temp_config_file" | jq -c '.')
    
    local merged_config
    merged_config=$(echo "$existing_json" | jq --argjson new "$new_json" '
        .apps.http.servers.nixopus.routes += $new.apps.http.servers.nixopus.routes
    ')
    
    echo "$merged_config" | curl -s -X POST "http://localhost:2019/load" \
        -H "Content-Type: application/json" \
        -d @- > /dev/null 2>&1
    
    sudo rm -f "$temp_config_file"
    
    log_message "INFO" "Configuration merged successfully"
}

# Load configuration into Caddy
function load_caddy_configuration() {
    local caddy_config_dir="$1"
    local caddy_config_file="$caddy_config_dir/caddy.json"
    
    log_message "INFO" "Loading Caddy configuration..."
    
    curl -s -X POST "http://localhost:2019/load" \
        -H "Content-Type: application/json" \
        -d @"$caddy_config_file" > /dev/null 2>&1
}

# setup caddy reverse proxy for the domains
function setup_caddy_reverse_proxy() {
    local app_domain="${CONFIG[app_domain]}"
    local api_domain="${CONFIG[api_domain]}"

    log_message "INFO" "Setting up Caddy reverse proxy for domains: $app_domain and $api_domain"

    if is_caddy_running; then
        log_message "INFO" "Caddy container is already running, appending new domains..."
        local caddy_config_dir
        caddy_config_dir=$(setup_caddy_config_directory) || return 1
        merge_caddy_configuration "$caddy_config_dir" || return 1
        load_caddy_configuration "$caddy_config_dir" || return 1
    else
        log_message "INFO" "Setting up Caddy for first time..."
        local caddy_config_dir
        caddy_config_dir=$(setup_caddy_config_directory) || return 1
        update_caddy_configuration "$caddy_config_dir" || return 1
        create_caddy_compose_file "$caddy_config_dir" || return 1
        start_caddy_container "$caddy_config_dir" || return 1
        wait_for_caddy_ready || return 1
        load_caddy_configuration "$caddy_config_dir" || return 1
    fi
    
    log_message "INFO" "Caddy reverse proxy setup completed successfully"
    log_message "INFO" "App available at: https://$app_domain"
    log_message "INFO" "API available at: https://$api_domain"
}

function main() {
    log_message "INFO" "Installing Nixopus through virtualizer"
    init_virtualized_config "$@"
    display_config_summary

    if ! validate_virtualized_params; then
        log_message "ERROR" "Parameter validation failed, exiting"
        exit 1
    fi

    log_message "INFO" "Installing dependencies..."
    install_dependencies
    
    log_message "INFO" "Creating LXD container..."
    if ! create_lxd_container; then
        log_message "ERROR" "Failed to create container, exiting..."
        exit 1
    fi
    
    log_message "INFO" "Created Container name: '${CONFIG[container_name]}'"
    
    log_message "INFO" "Installing dependencies in container..."
    install_dependencies_in_container
    
    log_message "INFO" "Starting installation script..."
    if run_installation_script "$(build_installation_command)" "${CONFIG[container_name]}"; then
        test_result="PASSED"
    else
        test_result="FAILED"
    fi
    
    log_message "INFO" "Installation script completed with result: $test_result"
    if [[ "$test_result" != "PASSED" ]]; then
        log_message "ERROR" "Installation failed, exiting..."
        exit 1
    fi 
    
    override_env_variables
    setup_all_proxies
    
    if [[ -n "${CONFIG[app_domain]}" && -n "${CONFIG[api_domain]}" ]]; then
        log_message "INFO" "Setting up Caddy reverse proxy..."
        if ! setup_caddy_reverse_proxy; then
            log_message "ERROR" "Failed to setup Caddy reverse proxy"
            exit 1
        fi
    else
        log_message "INFO" "Skipping Caddy setup - no domains provided"
    fi
    
    log_message "INFO" "Proxy set for container: ${CONFIG[container_name]}"
    log_message "INFO" "Virtualized installation completed successfully!"
}

main "$@"
