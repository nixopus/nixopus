#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/test_install.sh"
source "$SCRIPT_DIR/util.sh"
source "$SCRIPT_DIR/messages.sh"
source "$SCRIPT_DIR/caddy_setup_util.sh"

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
    local force_rebuild="${9:-}"
    local base_image_name="${10:-}"
    
    [[ -z "$internal_app_proxy_port" ]] && internal_app_proxy_port="7443"
    [[ -z "$internal_api_proxy_port" ]] && internal_api_proxy_port="8443"
    [[ -z "$distro" ]] && distro="debian/11"
    [[ -z "$proxy_url" ]] && proxy_url="127.0.0.1"
    [[ -z "$app_port" ]] && app_port=$(generate_random_available_port)
    [[ -z "$api_port" ]] && api_port=$(generate_random_available_port)
    [[ -z "$env" ]] && env="production"
    [[ -z "$show_in_console" ]] && show_in_console="true"
    [[ -z "$force_rebuild" ]] && force_rebuild="false"
    
    # Generate base image name if not provided
    if [[ -z "$base_image_name" ]]; then
        local base_distro=$(echo "$distro" | cut -d'/' -f1)
        base_image_name="nixopus-base-${base_distro}"
    fi
    
    echo "$internal_app_proxy_port $internal_api_proxy_port $distro $proxy_url $app_port $api_port $env $show_in_console $force_rebuild $base_image_name"
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
    local force_rebuild="${13:-false}"
    local base_image_name="${14:-}"
    
    local defaults
    defaults=$(set_default_values "$internal_app_proxy_port" "$internal_api_proxy_port" "$distro" "$proxy_url" "$app_port" "$api_port" "$env" "$show_in_console" "$force_rebuild" "$base_image_name")
    internal_app_proxy_port=$(echo "$defaults" | cut -d' ' -f1)
    internal_api_proxy_port=$(echo "$defaults" | cut -d' ' -f2)
    distro=$(echo "$defaults" | cut -d' ' -f3)
    proxy_url=$(echo "$defaults" | cut -d' ' -f4)
    app_port=$(echo "$defaults" | cut -d' ' -f5)
    api_port=$(echo "$defaults" | cut -d' ' -f6)
    env=$(echo "$defaults" | cut -d' ' -f7)
    show_in_console=$(echo "$defaults" | cut -d' ' -f8)
    force_rebuild=$(echo "$defaults" | cut -d' ' -f9)
    base_image_name=$(echo "$defaults" | cut -d' ' -f10)
    
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
        ["force_rebuild"]="$force_rebuild"
        ["base_image_name"]="$base_image_name"
        ["base_image_ready"]="false"
    )
    
    CONFIG["base_distro"]=$(echo "${CONFIG[distro]}" | cut -d'/' -f1)
}

# Check if base image exists so we can skip building the base image and installing nixopus in the container, thus by 
# improving the installation time, 
function check_base_image_exists() {
    local image_name="${CONFIG[base_image_name]}"
    if sudo lxc image list | grep -q "$image_name"; then
        log_info "$BASE_IMAGE_EXISTS"
        return 0
    fi
    log_info "$BASE_IMAGE_DOES_NOT_EXIST"
    return 1
}

# Build base image with Nixopus pre-installed, this is only done if the base image does not exist
function build_base_image() {
    local image_name="${CONFIG[base_image_name]}"
    local temp_container="temp-build-$(date +%s%N | md5sum | cut -c1-8)"
    local distro="${CONFIG[distro]}"
    
    log_info "$BUILDING_BASE_IMAGE $image_name $FROM_DISTRO $distro"
    
    local original_container_name="${CONFIG[container_name]}"
    CONFIG["container_name"]="$temp_container"
    
    if ! setup_container_with_nixopus "minimal"; then
        log_error "$FAILED_TO_SETUP_CONTAINER_FOR_BASE_IMAGE"
        CONFIG["container_name"]="$original_container_name"
        return 1
    fi
    
    log_info "$CREATING_BASE_IMAGE_FROM_CONTAINER"
    log_info "$STOPPING_CONTAINER_FOR_IMAGE_CREATION"
    sudo lxc stop "$temp_container"
    
    if sudo lxc publish "$temp_container" --alias "$image_name"; then
        log_info "$BASE_IMAGE_CREATED_SUCCESSFULLY $image_name"
        CONFIG["base_image_ready"]="true"
    else
        log_error "$FAILED_TO_CREATE_BASE_IMAGE"
        cleanup_container
        CONFIG["container_name"]="$original_container_name"
        return 1
    fi
    
    log_info "$CLEANING_UP_TEMPORARY_BUILD_CONTAINER"
    cleanup_container
    CONFIG["container_name"]="$original_container_name"
    return 0
}

# Unified function to setup container with Nixopus installation
function setup_container_with_nixopus() {
    local config_type="${1:-full}" 
    
    log_info "$SETTING_UP_CONTAINER_WITH_NIXOPUS ($config_type $CONFIG_TYPE)"
    
    if ! create_lxd_container; then
        log_error "$FAILED_TO_CREATE_CONTAINER"
        return 1
    fi
    
    log_info "$INSTALLING_DEPENDENCIES_IN_CONTAINER"
    install_dependencies_in_container
    
    log_info "$INSTALLING_NIXOPUS_IN_CONTAINER"
    local install_cmd
    if [ "$config_type" = "minimal" ]; then
        install_cmd="sudo bash -c \"\$(curl -sSL $GITHUB_RAW_BASE/scripts/install.sh)\" --env=production --debug"
    else
        install_cmd="$(build_installation_command)"
    fi
    
    if ! run_installation_script "$install_cmd" "${CONFIG[container_name]}"; then
        log_error "$FAILED_TO_INSTALL_NIXOPUS"
        return 1
    fi
    
    return 0
}

# Handle base image management - check existence, force rebuild, and build when needed
function manage_base_image() {
    local force_rebuild="${CONFIG[force_rebuild]}"
    local image_name="${CONFIG[base_image_name]}"
    
    log_info "$MANAGING_BASE_IMAGE $image_name"
    
    log_info "About to call check_base_image_exists"
    check_base_image_exists
    log_info "check_base_image_exists returned: $?"
    
    if check_base_image_exists; then
        if [ "$force_rebuild" = "true" ]; then
            log_info "$FORCE_REBUILD_REQUESTED $image_name"
            sudo lxc image delete "$image_name" 2>/dev/null || true
            if ! build_base_image; then
                log_error "$FAILED_TO_REBUILD_BASE_IMAGE"
                return 1
            fi
        else
            log_info "$USING_EXISTING_BASE_IMAGE $image_name"
            CONFIG["base_image_ready"]="true"
        fi
    else
        log_info "$BASE_IMAGE_NOT_FOUND $image_name"
        if ! build_base_image; then
            log_error "$FAILED_TO_BUILD_BASE_IMAGE"
            return 1
        fi
    fi
    
    return 0
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
    log_info "  $BASE_IMAGE_LABEL ${CONFIG[base_image_name]}"
    log_info "  $FORCE_REBUILD_LABEL ${CONFIG[force_rebuild]}"
    if [ -n "${CONFIG[api_domain]}" ]; then
        log_info "  $API_DOMAIN ${CONFIG[api_domain]}"
    fi
    if [ -n "${CONFIG[app_domain]}" ]; then
        log_info "  $APP_DOMAIN ${CONFIG[app_domain]}"
    fi
}

# Launch container from base image
function launch_from_base_image() {
    local container_name="${CONFIG[container_name]}"
    local image_name="${CONFIG[base_image_name]}"
    
    log_info "$LAUNCHING_CONTAINER_FROM_BASE_IMAGE $image_name"
    
    if sudo lxc launch "$image_name" "$container_name"; then
        log_info "$CONTAINER_LAUNCHED_SUCCESSFULLY $container_name"
        log_info "$CONFIGURING_CONTAINER_WITH_DOCKER_PRIVILEGES"
        sudo lxc config set "$container_name" security.privileged true
        sudo lxc config set "$container_name" security.nesting true
        sudo lxc restart "$container_name"
        sleep 60
        
        log_info "$CONTAINER_CONFIGURED_SUCCESSFULLY"
        return 0
    else
        log_error "$FAILED_TO_LAUNCH_CONTAINER_FROM_BASE_IMAGE"
        return 1
    fi
}

# Handle container creation - either use base image or traditional approach
function create_or_launch_container() {
    local force_rebuild="${CONFIG[force_rebuild]}"
    local image_name="${CONFIG[base_image_name]}"
    
    log_info "$HANDLING_CONTAINER_CREATION ${CONFIG[container_name]}"
    
    if ! manage_base_image; then
        log_error "$FAILED_TO_MANAGE_BASE_IMAGE"
        return 1
    fi
    
    if [ "${CONFIG[base_image_ready]}" = "true" ]; then
        log_info "$LAUNCHING_FROM_PRE_BUILT_IMAGE"
        if ! launch_from_base_image; then
            log_error "$FAILED_TO_LAUNCH_CONTAINER_FROM_BASE_IMAGE_AGAIN"
            return 1
        fi
    else
        log_info "$CREATING_NEW_CONTAINER_AND_INSTALLING"
        if ! setup_container_with_nixopus "full"; then
            log_error "$FAILED_TO_SETUP_CONTAINER_WITH_NIXOPUS"
            return 1
        fi
    fi
    
    return 0
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
    if ! create_or_launch_container; then
        log_error "$FAILED_CREATE_CONTAINER ${CONFIG[container_name]}, exiting..."
        exit 1
    fi
    
    log_info "$CREATED_CONTAINER_NAME '${CONFIG[container_name]}'"
    
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
