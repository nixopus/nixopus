#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/test_install.sh"
source "$SCRIPT_DIR/util.sh"
source "$SCRIPT_DIR/messages.sh"

# Configuration for the virtualized installation
declare -A CONFIG

# Set default values for configuration parameters
function set_default_values() {
    local internal_app_proxy_port="${1:-}"
    local internal_api_proxy_port="${2:-}"
    local distro="${3:-}"
    local proxy_url="${4:-}"
    local app_port="${5:-}"
    local api_port="${6:-}"
    local env="${7:-}"
    local show_in_console="${8:-}"
    
    [[ -z "$internal_app_proxy_port" ]] && internal_app_proxy_port="7443"
    [[ -z "$internal_api_proxy_port" ]] && internal_api_proxy_port="8443"
    [[ -z "$distro" ]] && distro="debian/11"
    [[ -z "$proxy_url" ]] && proxy_url="127.0.0.1"
    [[ -z "$app_port" ]] && app_port=$(generate_random_available_port)
    [[ -z "$api_port" ]] && api_port=$(generate_random_available_port)
    [[ -z "$env" ]] && env="production"
    [[ -z "$show_in_console" ]] && show_in_console="true"
    
    echo "$internal_app_proxy_port $internal_api_proxy_port $distro $proxy_url $app_port $api_port $env $show_in_console"
}

# Generate domains and container names
function generate_names() {
    local app_domain="${1:-}"
    local api_domain="${2:-}"
    local default_domain="${3:-.nixopus.com}"
    
    if [[ -z "$app_domain" ]]; then
        app_domain=$(generate_random_string 10)$default_domain
    fi
    if [[ -z "$api_domain" ]]; then
        api_domain=$(generate_random_string 10)$default_domain
    fi
    
    local container_name=$(generate_container_name)
    local app_proxy_name="app-proxy-${container_name}"
    local api_proxy_name="api-proxy-${container_name}"
    
    echo "$app_domain $api_domain $container_name $app_proxy_name $api_proxy_name"
}

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
    
    local defaults
    defaults=$(set_default_values "$internal_app_proxy_port" "$internal_api_proxy_port" "$distro" "$proxy_url" "$app_port" "$api_port" "$env" "$show_in_console")
    internal_app_proxy_port=$(echo "$defaults" | cut -d' ' -f1)
    internal_api_proxy_port=$(echo "$defaults" | cut -d' ' -f2)
    distro=$(echo "$defaults" | cut -d' ' -f3)
    proxy_url=$(echo "$defaults" | cut -d' ' -f4)
    app_port=$(echo "$defaults" | cut -d' ' -f5)
    api_port=$(echo "$defaults" | cut -d' ' -f6)
    env=$(echo "$defaults" | cut -d' ' -f7)
    show_in_console=$(echo "$defaults" | cut -d' ' -f8)
    
    local names
    names=$(generate_names "$app_domain" "$api_domain" ".nixopus.com")
    app_domain=$(echo "$names" | cut -d' ' -f1)
    api_domain=$(echo "$names" | cut -d' ' -f2)
    local container_name=$(echo "$names" | cut -d' ' -f3)
    local app_proxy_name=$(echo "$names" | cut -d' ' -f4)
    local api_proxy_name=$(echo "$names" | cut -d' ' -f5)
    
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local project_root="$(dirname "$script_dir")"
    
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
        ["project_root"]="$project_root"
        ["caddy_config_dir"]="/etc/nixopus/caddy"
        ["caddy_compose_template"]="$project_root/helpers/docker-compose.caddy.yml"
        ["caddyfile_template"]="$project_root/helpers/Caddyfile"
        ["caddy_json_template"]="$project_root/helpers/caddy.json"
        ["caddy_admin_url"]="http://localhost:2019"
        ["default_domain"]=".nixopus.com"
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
        log_error "$INVALID_SERVICE_TYPE $service_type. $SERVICE_TYPE_MUST_BE"
        return 1
    fi
    
    log_info "$SETTING_UP_PROXY ${service_type^^} $PROXY_FOR_CONTAINER $container_name"
    log_info "$EXTERNAL_PORT $proxy_external_port -> $INTERNAL_PORT $proxy_url:$proxy_internal_port"
    
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

    log_info "$OVERRIDING_ENV_VARIABLES $container_name"
    
    sudo lxc exec "$container_name" -- bash -c "
        if [[ ! -f \"$app_env_file\" ]]; then
            echo \"$FILE_NOT_FOUND_ERROR $app_env_file\"
            exit 1
        fi
        if [[ ! -f \"$api_env_file\" ]]; then
            echo \"$FILE_NOT_FOUND_ERROR $api_env_file\"
            exit 1
        fi

        sed -i \"s|^WEBSOCKET_URL=.*|WEBSOCKET_URL=wss://$app_domain/ws|\" \"$app_env_file\"
        sed -i \"s|^WEBHOOK_URL=.*|WEBHOOK_URL=https://$app_domain/api/v1/webhook|\" \"$app_env_file\"
        sed -i \"s|^API_URL=.*|API_URL=https://$api_domain/api|\" \"$app_env_file\"
        
        sed -i \"s|^ALLOWED_ORIGIN=.*|ALLOWED_ORIGIN=https://$app_domain|\" \"$api_env_file\"
        
        echo \"$ENVIRONMENT_VARIABLES_UPDATED\"
    "
}

# Validate installation parameters
function validate_virtualized_params() {
    log_info "$VALIDATING_VIRTUALIZED_PARAMS"
    local validation_failed=false
    
    validate_environment "${CONFIG[env]}" || validation_failed=true
    validate_lxd || validation_failed=true
    validate_sudo || validation_failed=true
    validate_file "${SCRIPT_DIR}/util.sh" "util.sh" || validation_failed=true
    
    log_info "$CHECKING_PORTS_AVAILABLE"
    is_port_available "${CONFIG[app_port]}" || validation_failed=true
    is_port_available "${CONFIG[api_port]}" || validation_failed=true
    
    if [ -n "${CONFIG[api_domain]}" ]; then
        validate_domain "${CONFIG[api_domain]}" "API" "false" || validation_failed=true
    fi
    if [ -n "${CONFIG[app_domain]}" ]; then
        validate_domain "${CONFIG[app_domain]}" "App" "false" || validation_failed=true
    fi
    
    if [ "$validation_failed" = true ]; then
        log_error "$VALIDATION_FAILED"
        return 1
    fi
    
    log_info "$ALL_VALIDATIONS_PASSED"
    return 0
}

function display_config_summary() {
    log_info "$CONFIGURATION_SUMMARY"
    log_info "  $DISTRIBUTION ${CONFIG[distro]}"
    log_info "  $CONTAINER ${CONFIG[container_name]}"
    log_info "  $ENVIRONMENT ${CONFIG[env]}"
    log_info "  $APP_PORT ${CONFIG[app_port]}"
    log_info "  $API_PORT ${CONFIG[api_port]}"
    log_info "  $PROXY_URL ${CONFIG[proxy_url]}"
    log_info "  $INTERNAL_APP_PROXY_PORT ${CONFIG[internal_app_proxy_port]}"
    log_info "  $INTERNAL_API_PROXY_PORT ${CONFIG[internal_api_proxy_port]}"
    log_info "  $APP_PROXY_NAME ${CONFIG[app_proxy_name]}"
    log_info "  $API_PROXY_NAME ${CONFIG[api_proxy_name]}"
    if [ -n "${CONFIG[api_domain]}" ]; then
        log_info "  $API_DOMAIN ${CONFIG[api_domain]}"
    fi
    if [ -n "${CONFIG[app_domain]}" ]; then
        log_info "  $APP_DOMAIN ${CONFIG[app_domain]}"
    fi
}

# Create Caddy configuration directory and copy files
function setup_caddy_config_directory() {
    local caddy_config_dir="${CONFIG[caddy_config_dir]}"
    
    log_info "$SETTING_UP_CADDY_CONFIG $caddy_config_dir" >&2
    
    sudo mkdir -p "$caddy_config_dir"
    
    local caddyfile_source="${CONFIG[caddyfile_template]}"
    local caddy_json_source="${CONFIG[caddy_json_template]}"
    
    if ! validate_file "$caddyfile_source" "Caddyfile" >&2; then
        return 1
    fi
    
    if ! validate_file "$caddy_json_source" "caddy.json" >&2; then
        return 1
    fi
    
    sudo cp "$caddyfile_source" "$caddy_config_dir/"
    sudo cp "$caddy_json_source" "$caddy_config_dir/caddy.json"
    
    if ! validate_file "$caddy_config_dir/Caddyfile" "Caddyfile" >&2; then
        return 1
    fi
    
    if ! validate_file "$caddy_config_dir/caddy.json" "caddy.json" >&2; then
        return 1
    fi
    
    echo "$caddy_config_dir"
}

# Update Caddy configuration with domain and proxy settings
function update_caddy_configuration() {
    local caddy_config_dir="${CONFIG[caddy_config_dir]}"
    local app_domain="${CONFIG[app_domain]}"
    local api_domain="${CONFIG[api_domain]}"
    local app_port="${CONFIG[app_port]}"
    local api_port="${CONFIG[api_port]}"
    local proxy_url="${CONFIG[proxy_url]}"
    
    local caddy_config_file="$caddy_config_dir/caddy.json"
    local app_proxy_url="$proxy_url:$app_port"
    local api_proxy_url="$proxy_url:$api_port"
    
    validate_file "$caddy_config_file" "Caddy configuration" || return 1
    
    log_info "$UPDATING_CADDY_CONFIG"
    log_info "  $APP_DOMAIN_PROXY $app_domain -> $app_proxy_url"
    log_info "  $API_DOMAIN_PROXY $api_domain -> $api_proxy_url"
    
    sudo sed -i "s/{env.APP_DOMAIN}/$app_domain/g" "$caddy_config_file"
    sudo sed -i "s/{env.API_DOMAIN}/$api_domain/g" "$caddy_config_file"
    sudo sed -i "s/{env.APP_REVERSE_PROXY_URL}/$app_proxy_url/g" "$caddy_config_file"
    sudo sed -i "s/{env.API_REVERSE_PROXY_URL}/$api_proxy_url/g" "$caddy_config_file"
}

# Create docker-compose file for Caddy
function create_caddy_compose_file() {
    local caddy_config_dir="${CONFIG[caddy_config_dir]}"
    local compose_file="$caddy_config_dir/docker-compose.yml"
    local template_file="${CONFIG[caddy_compose_template]}"
    
    log_info "$CREATING_CADDY_COMPOSE $compose_file"
    
    if ! validate_file "$template_file" "Caddy docker-compose template" >&2; then
        return 1
    fi
    
    CADDY_CONFIG_DIR="$caddy_config_dir" envsubst < "$template_file" | sudo tee "$compose_file" > /dev/null
    
    if ! validate_file "$compose_file" "Generated docker-compose file" >&2; then
        return 1
    fi
    
    log_info "$CADDY_COMPOSE_CREATED"
}

# Start Caddy container
function start_caddy_container() {
    local caddy_config_dir="${CONFIG[caddy_config_dir]}"
    
    log_info "$STARTING_CADDY_CONTAINER"
    
    cd "$caddy_config_dir"
    if sudo docker-compose up -d; then
        log_info "$CADDY_CONTAINER_STARTED"
        return 0
    else
        log_error "$FAILED_START_CADDY"
        return 1
    fi
}

# Wait for Caddy to be ready
function wait_for_caddy_ready() {
    local max_attempts=30
    local attempt=1
    
    log_info "$WAITING_FOR_CADDY"
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "${CONFIG[caddy_admin_url]}/config" > /dev/null 2>&1; then
            log_info "$CADDY_IS_READY"
            return 0
        fi
        log_info "$WAITING_FOR_CADDY_ATTEMPT $attempt/$max_attempts)"
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log_error "$CADDY_FAILED_TO_START"
    return 1
}

# Check if Caddy container is already running
function is_caddy_running() {
    sudo docker ps --format "table {{.Names}}" | grep -q "nixopus-caddy-virtualized"
}

# Get existing Caddy configuration
function get_existing_caddy_config() {
    curl -s "${CONFIG[caddy_admin_url]}/config"
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
    local caddy_config_dir="${CONFIG[caddy_config_dir]}"
    local caddy_config_file="$caddy_config_dir/caddy.json"
    local app_domain="${CONFIG[app_domain]}"
    local api_domain="${CONFIG[api_domain]}"
    local app_port="${CONFIG[app_port]}"
    local api_port="${CONFIG[api_port]}"
    local proxy_url="${CONFIG[proxy_url]}"
    
    validate_file "$caddy_config_file" "Caddy configuration" || return 1
    
    local app_proxy_url="$proxy_url:$app_port"
    local api_proxy_url="$proxy_url:$api_port"
    
    log_info "$MERGING_CADDY_CONFIG"
    
    local existing_config
    existing_config=$(get_existing_caddy_config)
    
    if domain_exists_in_config "$existing_config" "$app_domain"; then
        log_warn "$APP_DOMAIN_ALREADY_EXISTS '$app_domain' $ALREADY_EXISTS_IN_CONFIG"
    fi
    
    if domain_exists_in_config "$existing_config" "$api_domain"; then
        log_warn "$API_DOMAIN_ALREADY_EXISTS '$api_domain' $ALREADY_EXISTS_IN_CONFIG"
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
    
    echo "$merged_config" | curl -s -X POST "${CONFIG[caddy_admin_url]}/load" \
        -H "Content-Type: application/json" \
        -d @- > /dev/null 2>&1
    
    sudo rm -f "$temp_config_file"
    
    log_info "$CONFIGURATION_MERGED"
}

# Load configuration into Caddy
function load_caddy_configuration() {
    local caddy_config_dir="${CONFIG[caddy_config_dir]}"
    local caddy_config_file="$caddy_config_dir/caddy.json"
    
    log_info "$LOADING_CADDY_CONFIG"
    
    curl -s -X POST "${CONFIG[caddy_admin_url]}/load" \
        -H "Content-Type: application/json" \
        -d @"$caddy_config_file" > /dev/null 2>&1
}

# setup caddy reverse proxy for the domains
function setup_caddy_reverse_proxy() {
    local app_domain="${CONFIG[app_domain]}"
    local api_domain="${CONFIG[api_domain]}"

    log_info "$SETTING_UP_CADDY_REVERSE_PROXY $app_domain and $api_domain"

    if is_caddy_running; then
        log_info "$CADDY_ALREADY_RUNNING"
        setup_caddy_config_directory || return 1
        merge_caddy_configuration || return 1
    else
        log_info "$SETTING_UP_CADDY_FIRST_TIME"
        setup_caddy_config_directory || return 1
        update_caddy_configuration || return 1
        create_caddy_compose_file || return 1
        start_caddy_container || return 1
        wait_for_caddy_ready || return 1
        load_caddy_configuration || return 1
    fi
    
    log_info "$CADDY_REVERSE_PROXY_SETUP_COMPLETED"
    log_info "$APP_AVAILABLE_AT https://$app_domain"
    log_info "$API_AVAILABLE_AT https://$api_domain"
}

function main() {
    log_info "$INSTALLING_NIXOPUS_VIRTUALIZER"
    init_virtualized_config "$@"
    display_config_summary

    if ! validate_virtualized_params; then
        log_error "$PARAMETER_VALIDATION_FAILED"
        exit 1
    fi

    log_info "$INSTALLING_DEPENDENCIES"
    install_dependencies
    
    log_info "$CREATING_CONTAINER"
    if ! create_lxd_container; then
        log_error "$FAILED_CREATE_CONTAINER ${CONFIG[container_name]}, exiting..."
        exit 1
    fi
    
    log_info "$CREATED_CONTAINER_NAME '${CONFIG[container_name]}'"
    
    log_info "$INSTALLING_DEPENDENCIES"
    install_dependencies_in_container
    
    log_info "$STARTING_INSTALLATION_SCRIPT"
    if run_installation_script "$(build_installation_command)" "${CONFIG[container_name]}"; then
        test_result="PASSED"
    else
        test_result="FAILED"
    fi
    
    log_info "$INSTALLATION_SCRIPT_COMPLETED_RESULT $test_result"
    if [[ "$test_result" != "PASSED" ]]; then
        log_error "$INSTALLATION_FAILED_EXITING"
        exit 1
    fi 
    
    override_env_variables
    setup_all_proxies
    
    if [[ -n "${CONFIG[app_domain]}" && -n "${CONFIG[api_domain]}" ]]; then
        log_info "$SETTING_UP_CADDY_REVERSE_PROXY_MSG"
        if ! setup_caddy_reverse_proxy; then
            log_error "$FAILED_SETUP_CADDY"
            exit 1
        fi
    else
        log_info "$SKIPPING_CADDY_SETUP"
    fi
    
    log_info "$PROXY_SET_FOR_CONTAINER ${CONFIG[container_name]}"
    log_info "$VIRTUALIZED_INSTALLATION_COMPLETED"
}

main "$@"
