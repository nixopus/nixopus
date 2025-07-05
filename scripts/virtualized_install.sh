#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/test_install.sh"

# checks if the port is available
function is_port_available() {
    local port="$1"
    if sudo lsof -i :"$port" > /dev/null; then
        echo "Port $port is already in use"
        exit 1
    fi
}

# sets the proxy for the container
function set_proxy_for_lxd() {
    local container_name="$1"
    echo "Setting proxy for container: $container_name"
    local proxy_url="$2"
    local proxy_port="$3"
    local proxy_name="$4"
    sudo lxc config device add "$container_name" "$proxy_name" proxy listen=tcp:0.0.0.0:"$proxy_port" connect=tcp:"$proxy_url":"$proxy_port"
}

# checks if the file is present
function is_file_present() {
    local file_path="$1"
    if [[ ! -f "$file_path" ]]; then
        echo "File not found: $file_path" 
        exit 1
    fi
}

# replaces the env variable in the file with the new value
# the env variable is replaced only if it is present in the file does not handle appending the variable if it is not present in the file
function replace_env_variable() {
    local file_path="$1"
    local variable_name="$2"
    local variable_value="$3"
    sed -i "s|^$variable_name=.*|$variable_name=$variable_value|" "$file_path"
}

function get_env_variable() {
    local file_path="$1"
    local variable_name="$2"
    local variable_value=$(grep "^$variable_name=" "$file_path" | cut -d'=' -f2)
    echo "$variable_value"
}

# replaces the env variable in the view file
function replace_env_variable_in_view(){
    local app_env_file="$1"
    local app_domain="$2"
    local api_domain="$3"
    replace_env_variable "$app_env_file" "WEBSOCKET_URL" "wss://$app_domain/ws"
    replace_env_variable "$app_env_file" "WEBHOOK_URL" "https://$app_domain/api/v1/webhook"
    replace_env_variable "$app_env_file" "API_URL" "https://$api_domain/api"
}

# replaces the env variable in the api file
function replace_env_variable_in_api(){
    local api_env_file="$1"
    local app_domain="$2"
    replace_env_variable "$api_env_file" "ALLOWED_ORIGIN" "https://$app_domain"
}

# overrides the env variables present inside the /etc/nixopus/source/api/.env and /etc/nixopus/source/view/.env files
# overrriding is carried out because the nixopus will be installed in the container, so we need to update the ports and set proxy for the services run inside the container
function override_env_variables() {
    local container_name="$1"
    local app_domain="$2"
    local api_domain="$3"

    local app_env_file="/etc/nixopus/source/view/.env"
    local api_env_file="/etc/nixopus/source/api/.env"

    echo "Overriding environment variables in container: $container_name"
    
    # Execute the file operations inside the container
    sudo lxc exec "$container_name" -- bash -c "
        if [[ ! -f \"$app_env_file\" ]]; then
            echo \"File not found: $app_env_file\"
            exit 1
        fi
        if [[ ! -f \"$api_env_file\" ]]; then
            echo \"File not found: $api_env_file\"
            exit 1
        fi
        
        # Replace environment variables in view file
        sed -i \"s|^WEBSOCKET_URL=.*|WEBSOCKET_URL=wss://$app_domain/ws|\" \"$app_env_file\"
        sed -i \"s|^WEBHOOK_URL=.*|WEBHOOK_URL=https://$app_domain/api/v1/webhook|\" \"$app_env_file\"
        sed -i \"s|^API_URL=.*|API_URL=https://$api_domain/api|\" \"$app_env_file\"
        
        # Replace environment variables in api file
        sed -i \"s|^ALLOWED_ORIGIN=.*|ALLOWED_ORIGIN=https://$app_domain|\" \"$api_env_file\"
        
        echo \"Environment variables updated successfully\"
    "
}


function main() {
    echo "Installing Nixopus through virtualizer"
    local distro="$1"
    local api_domain="$2"
    local app_domain="$3"
    local container_name="$4"
    local proxy_url="$5"
 
    # these are used to set the proxy for the app and api from the host machine 
    # e.g. irrespective of what port is used inside the container for app and api, we will use these ports from the host machine to container as a reverse proxy
    local app_port="$6" 
    local api_port="$7" 
    
    # Set default values for missing parameters
    local email="admin@nixopus.com"
    local password="admin123"
    local env="production"
    
    if [[ ! "$container_name" ]]; then
        container_name="nixopus-dev"
    fi
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

    # check if the ports are available before starting the installation
    echo "Checking if the ports are available..."
    is_port_available "$app_port"
    is_port_available "$api_port"

    echo "Checking if lxd is installed..."
    is_lxd_installed

    echo "Installing dependencies..."
    install_dependencies
    echo "Creating LXD container..."
    create_lxd_container "$distro"
    echo "Installing dependencies in container..."
    install_dependencies_in_container "$container_name" "$distro"
    echo "Starting installation script..."
    local install_cmd
    install_cmd=$(build_installation_command "$email" "$password" "$api_domain" "$app_domain" "$env")
    echo "Installation command: $install_cmd"
    if run_installation_script "$install_cmd"; then
        test_result="PASSED"
    else
        test_result="FAILED"
    fi
    echo "Installation script completed with result: $test_result"
    override_env_variables "$container_name" "$app_domain" "$api_domain"
    set_proxy_for_lxd "$container_name" "$proxy_url" "$app_port" "app"
    set_proxy_for_lxd "$container_name" "$proxy_url" "$api_port" "api"
    echo "Proxy set for container: $container_name"

    # we need to reverse proxy to the domain from host machine for the app and api to be accessible from the host machine using caddy here
}

main "$@"
