#!/bin/bash

source "$(dirname "$0")/messages.sh"

# DETECT THE PACKAGE MANAGER FOR THE OS
function detect_package_manager() {
    if [[ "$OS" == "Darwin" ]]; then
        log_error "$MACOS_UNSUPPORTED"
        exit 1
    elif command -v apt-get &>/dev/null; then
        echo "apt"
    elif command -v dnf &>/dev/null; then
        echo "dnf"
    elif command -v yum &>/dev/null; then
        echo "yum"
    elif command -v pacman &>/dev/null; then
        echo "pacman"
    else
        log_error "$UNSUPPORTED_PACKAGE_MANAGER"
        exit 1
    fi
}

# Install the package based on the package manager
function install_package() {
    local pkg_manager
    pkg_manager=$(detect_package_manager)
    
    case $pkg_manager in
        "apt")
            sudo apt-get update
            sudo apt-get install -y "$1"
            ;;
        "dnf")
            sudo dnf install -y "$1"
            ;;
        "yum")
            sudo yum install -y "$1"
            ;;
        "pacman")
            sudo pacman -Sy --noconfirm "$1"
            ;;
    esac
}

# ensure command is installed, install it if not present
function ensure_command_installed() {
    local cmd="$1"
    local package_name="$2"
    
    if ! command -v "$cmd" &>/dev/null; then
        log_info "$COMMAND_NOT_FOUND '$cmd' $ATTEMPTING_INSTALL"
        install_package "$package_name"
        log_info "$SUCCESSFULLY_INSTALLED $cmd"
    fi
}

# Install the dependencies
function install_dependencies() {
    ensure_command_installed "git" "git"
    ensure_command_installed "python3" "python3"
    ensure_command_installed "pip3" "python3-pip"
    ensure_command_installed "openssl" "openssl"
    ensure_command_installed "curl" "curl"
}

# Install Docker using the official installation script
# This script handles different distros and different package managers automatically
function install_docker(){    
    if command -v docker &>/dev/null; then
        log_info "$DOCKER_ALREADY_INSTALLED"
        return 0
    fi
    
    if curl -fsSL https://get.docker.com -o /tmp/get-docker.sh; then
        if sh /tmp/get-docker.sh; then            
            rm -f /tmp/get-docker.sh
            log_info "$DOCKER_INSTALLED_SUCCESSFULLY"
            return 0
        else
            log_error "$DOCKER_INSTALLATION_FAILED"
            rm -f /tmp/get-docker.sh
            return 1
        fi
    else
        log_error "$DOCKER_DOWNLOAD_FAILED"
        return 1
    fi
}

# Validate email format
function validate_email() {
    local email="$1"
    local strict="${2:-false}"
    
    if [ -n "$email" ]; then
        if [[ "$email" =~ ^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$ ]]; then
            log_info "$EMAIL_VALID"
            return 0
        else
            log_error "$INVALID_EMAIL_FORMAT $email"
            return 1
        fi
    else
        if [ "$strict" = "true" ]; then
            log_error "$EMAIL_REQUIRED"
            return 1
        else
            return 0
        fi
    fi
}

# Validate password
function validate_password() {
    local password="$1"
    local strict="${2:-false}"
    
    if [ -n "$password" ]; then
        if [ ${#password} -ge 6 ]; then
            log_info "$PASSWORD_PROVIDED ${#password})"
            return 0
        else
            log_error "$PASSWORD_TOO_SHORT"
            return 1
        fi
    else
        if [ "$strict" = "true" ]; then
            log_error "$PASSWORD_REQUIRED"
            return 1
        else
            return 0
        fi
    fi
}

# Validate domain format
function validate_domain() {
    local domain="$1"
    local domain_type="$2"
    local strict="${3:-false}"
    
    if [ -n "$domain" ]; then
        if [[ "$domain" =~ ^[A-Za-z0-9.-]+\.[A-Za-z]{2,}$ ]]; then
            log_info "$domain_type $DOMAIN_VALID $domain"   
            return 0
        else
            log_error "$INVALID_DOMAIN_FORMAT $domain_type domain format: $domain"
            return 1
        fi
    else
        if [ "$strict" = "true" ]; then
            log_error "$domain_type $DOMAIN_REQUIRED"
            return 1
        else
            return 0
        fi
    fi
}

# Validate environment
function validate_environment() {
    local env="$1"
    
    case "$env" in
        "development"|"staging"|"production"|"test")
            log_info "$ENVIRONMENT_VALID $env"
            return 0
            ;;
        *)
            log_error "$INVALID_ENVIRONMENT '$env'. $ENVIRONMENT_OPTIONS"
            return 1
            ;;
    esac
}

# Validate LXD installation and status
function validate_lxd() {
    if ! command -v sudo lxc &>/dev/null; then
        log_error "$LXD_NOT_FOUND"
        return 1
    fi
    log_info "$LXD_INSTALLED"
    
    if ! sudo lxc list &>/dev/null; then
        log_error "$LXD_NOT_RUNNING"
        return 1
    fi
    log_info "$LXD_RUNNING"
    return 0
}

# Validate sudo access
function validate_sudo() {
    if ! sudo -n true 2>/dev/null; then
        log_error "$SUDO_ACCESS_REQUIRED"
        return 1
    fi
    log_info "$SUDO_AVAILABLE"
    return 0
}

# Validate file exists
function validate_file() {
    local file_path="$1"
    local file_name="$2"
    
    if [ ! -f "$file_path" ]; then
        log_error "$file_name $FILE_NOT_FOUND $file_path"
        return 1
    fi
    log_info "$file_name: $FILE_FOUND"
    return 0
}

# Validate URL accessibility
function validate_url() {
    local url="$1"
    local url_name="$2"
    local strict="${3:-false}"
    
    if [ -n "$url" ]; then
        if ! curl -fsSL --head "$url" &>/dev/null; then
            if [ "$strict" = "true" ]; then
                log_error "$url_name $URL_NOT_ACCESSIBLE $url"
                return 1
            else
                log_warn "$url_name $URL_NOT_ACCESSIBLE_CONTINUING $url $CONTINUING_ANYWAY"
                return 0
            fi
        fi
        log_info "$url_name: $URL_ACCESSIBLE"
        return 0
    else
        if [ "$strict" = "true" ]; then
            log_error "$url_name $URL_REQUIRED"
            return 1
        else
            return 0
        fi
    fi
}

# Validate array is not empty
function validate_array() {
    local array_name="$1"
    local array_ref="$2"
    
    if [ ${#array_ref[@]} -eq 0 ]; then
        log_error "$array_name $ARRAY_EMPTY"
        return 1
    fi
    log_info "$array_name: ${#array_ref[@]} $ARRAY_CONFIGURED"
    return 0
}

# Log message to console if show_in_console is true
# If no third parameter is provided and CONFIG[show_in_console] exists, use that value
function log_message() {
    local level="$1"
    local message="$2"
    local show_in_console="${3:-}"
    local timestamp="$(date -Iseconds)"

    # If no third parameter provided, try to use CONFIG[show_in_console] if it exists
    if [ -z "$show_in_console" ] && [ -n "${CONFIG[show_in_console]:-}" ]; then
        show_in_console="${CONFIG[show_in_console]}"
    fi
    
    # Default to false if still not set
    show_in_console="${show_in_console:-false}"

    if [ "$show_in_console" = "false" ]; then
        return 0
    fi
    
    local color_reset="\033[0m"
    local color_info="\033[32m"
    local color_warn="\033[33m"
    local color_error="\033[31m"
    local color_debug="\033[36m"
    
    case "$level" in
        "INFO")
            echo -e "[$timestamp] [${color_info}${level}${color_reset}] $message"
            ;;
        "WARN")
            echo -e "[$timestamp] [${color_warn}${level}${color_reset}] $message"
            ;;
        "ERROR")
            echo -e "[$timestamp] [${color_error}${level}${color_reset}] $message"
            ;;
        "DEBUG")
            echo -e "[$timestamp] [${color_debug}${level}${color_reset}] $message"
            ;;
        *)
            echo -e "[$timestamp] [${level}] $message"
            ;;
    esac
}

# Log info message to console if show_in_console is true
# If no second parameter is provided and CONFIG[show_in_console] exists, use that value
function log_info() {
    log_message "INFO" "$1"
}

# Log error message to console if show_in_console is true
# If no second parameter is provided and CONFIG[show_in_console] exists, use that value
function log_error() {
    log_message "ERROR" "$1"
}

# Log warning message to console if show_in_console is true
# If no second parameter is provided and CONFIG[show_in_console] exists, use that value
function log_warn() {
    log_message "WARN" "$1"
}

# Log debug message to console if show_in_console is true
# If no second parameter is provided and CONFIG[show_in_console] exists, use that value
function log_debug() {
    log_message "DEBUG" "$1"
}

# checks if the port is available
function is_port_available() {
    local port="$1"
    if sudo lsof -i :"$port" > /dev/null 2>&1; then
        log_error "$PORT_IN_USE $port $PORT_ALREADY_IN_USE"
        exit 1
    fi
}

# Generate a random available port between 1024 and 65535
function generate_random_available_port() {
    local min_port="${1:-1024}"
    local max_port="${2:-65535}"
    local max_attempts="${3:-100}"
    local attempt=0
    
    if ! [[ "$min_port" =~ ^[0-9]+$ ]] || ! [[ "$max_port" =~ ^[0-9]+$ ]]; then
        log_error "$INVALID_PORT_RANGE$min_port, max=$max_port"
        return 1
    fi
    
    if [ "$min_port" -gt "$max_port" ]; then
        log_error "$MIN_PORT_GREATER ($min_port) $CANNOT_BE_GREATER ($max_port)"
        return 1
    fi
    
    if [ "$min_port" -lt 1 ] || [ "$max_port" -gt 65535 ]; then
        log_error "$PORT_RANGE_LIMITS"
        return 1
    fi
    
    local used_ports
    used_ports=$(sudo lsof -i -P -n | grep LISTEN | awk '{print $9}' | sed 's/.*://' | sort -u 2>/dev/null)
    
    while [ $attempt -lt $max_attempts ]; do
        local random_port=$((RANDOM % (max_port - min_port + 1) + min_port))
        
        if ! echo "$used_ports" | grep -q "^$random_port$"; then
            if ! sudo lsof -i :"$random_port" >/dev/null 2>&1; then
                log_info "$GENERATED_RANDOM_PORT $random_port"
                echo "$random_port"
                return 0
            fi
        fi
        
        attempt=$((attempt + 1))
    done
    
    log_error "$FAILED_FIND_PORT $max_attempts"
    return 1
}

# Generate a random string of a given length
function generate_random_string() {
    local length="${1:-10}"
    local charset="abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    local result=""
    
    for ((i=0; i<length; i++)); do
        result+="${charset:$((RANDOM % ${#charset})):1}"
    done
    echo "$result"
}