#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/messages.sh"

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