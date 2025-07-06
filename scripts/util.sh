#!/bin/bash

# DETECT THE PACKAGE MANAGER FOR THE OS
function detect_package_manager() {
    if [[ "$OS" == "Darwin" ]]; then
        log_message "ERROR" "This script is not supported on macOS"
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
        log_message "ERROR" "Unsupported package manager"
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
        log_message "INFO" "Command '$cmd' not found. Attempting to install..."
        install_package "$package_name"
        log_message "INFO" "Successfully installed $cmd"
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
        log_message "INFO" "Docker is already installed."
        return 0
    fi
    
    if curl -fsSL https://get.docker.com -o /tmp/get-docker.sh; then
        if sh /tmp/get-docker.sh; then            
            rm -f /tmp/get-docker.sh
            log_message "INFO" "Docker installed successfully!"
            return 0
        else
            log_message "ERROR" "Docker installation failed!"
            rm -f /tmp/get-docker.sh
            return 1
        fi
    else
        log_message "ERROR" "Failed to download Docker installation script!"
        return 1
    fi
}

# Validate email format
function validate_email() {
    local email="$1"
    local strict="${2:-false}"
    
    if [ -n "$email" ]; then
        if [[ "$email" =~ ^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$ ]]; then
            log_message "INFO" "Email: valid"
            return 0
        else
            log_message "ERROR" "Invalid email format: $email"
            return 1
        fi
    else
        if [ "$strict" = "true" ]; then
            log_message "ERROR" "Email is required"
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
            log_message "INFO" "Password: provided (length: ${#password})"
            return 0
        else
            log_message "ERROR" "Password too short (minimum 6 characters)"
            return 1
        fi
    else
        if [ "$strict" = "true" ]; then
            log_message "ERROR" "Password is required"
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
            log_message "INFO" "$domain_type Domain: $domain"   
            return 0
        else
            log_message "ERROR" "Invalid $domain_type domain format: $domain"
            return 1
        fi
    else
        if [ "$strict" = "true" ]; then
            log_message "ERROR" "$domain_type domain is required"
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
            log_message "INFO" "Environment: $env"
            return 0
            ;;
        *)
            log_message "ERROR" "Invalid environment '$env'. Must be one of: development, staging, production, test"
            return 1
            ;;
    esac
}

# Validate LXD installation and status
function validate_lxd() {
    if ! command -v sudo lxc &>/dev/null; then
        log_message "ERROR" "LXD not found. Please install LXD first: https://canonical.com/lxd"
        return 1
    fi
    log_message "INFO" "LXD: installed"
    
    if ! sudo lxc list &>/dev/null; then
        log_message "ERROR" "LXD is not running. Please start LXD first: lxc start"
        return 1
    fi
    log_message "INFO" "LXD: running"
    return 0
}

# Validate sudo access
function validate_sudo() {
    if ! sudo -n true 2>/dev/null; then
        log_message "ERROR" "Sudo access required for LXD operations"
        return 1
    fi
    log_message "INFO" "Sudo: available"
    return 0
}

# Validate file exists
function validate_file() {
    local file_path="$1"
    local file_name="$2"
    
    if [ ! -f "$file_path" ]; then
        log_message "ERROR" "$file_name not found at $file_path"
        return 1
    fi
    log_message "INFO" "$file_name: found"
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
                log_message "ERROR" "$url_name not accessible at $url"
                return 1
            else
                log_message "WARN" "$url_name not accessible at $url (continuing anyway)"
                return 0
            fi
        fi
        log_message "INFO" "$url_name: accessible"
        return 0
    else
        if [ "$strict" = "true" ]; then
            log_message "ERROR" "$url_name is required"
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
        log_message "ERROR" "$array_name is empty"
        return 1
    fi
    log_message "INFO" "$array_name: ${#array_ref[@]} items configured"
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

# checks if the port is available
function is_port_available() {
    local port="$1"
    if sudo lsof -i :"$port" > /dev/null 2>&1; then
        log_message "ERROR" "Port $port is already in use"
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
        log_message "ERROR" "Invalid port range: min=$min_port, max=$max_port"
        return 1
    fi
    
    if [ "$min_port" -gt "$max_port" ]; then
        log_message "ERROR" "Min port ($min_port) cannot be greater than max port ($max_port)"
        return 1
    fi
    
    if [ "$min_port" -lt 1 ] || [ "$max_port" -gt 65535 ]; then
        log_message "ERROR" "Port range must be between 1 and 65535"
        return 1
    fi
    
    local used_ports
    used_ports=$(sudo lsof -i -P -n | grep LISTEN | awk '{print $9}' | sed 's/.*://' | sort -u 2>/dev/null)
    
    while [ $attempt -lt $max_attempts ]; do
        local random_port=$((RANDOM % (max_port - min_port + 1) + min_port))
        
        if ! echo "$used_ports" | grep -q "^$random_port$"; then
            if ! sudo lsof -i :"$random_port" >/dev/null 2>&1; then
                log_message "INFO" "Generated random available port: $random_port"
                echo "$random_port"
                return 0
            fi
        fi
        
        attempt=$((attempt + 1))
    done
    
    log_message "ERROR" "Failed to find available port after $max_attempts attempts"
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