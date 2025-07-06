#!/bin/bash

# DETECT THE PACKAGE MANAGER FOR THE OS
function detect_package_manager() {
    if [[ "$OS" == "Darwin" ]]; then
        echo "This script is not supported on macOS" >&2
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
        echo "Error: Unsupported package manager" >&2
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
        echo "Command '$cmd' not found. Attempting to install..."
        install_package "$package_name"
        echo "Successfully installed $cmd"
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
        echo "Docker is already installed."
        return 0
    fi
    
    if curl -fsSL https://get.docker.com -o /tmp/get-docker.sh; then
        if sh /tmp/get-docker.sh; then            
            rm -f /tmp/get-docker.sh
            echo "Docker installed successfully!"
            return 0
        else
            echo "Error: Docker installation failed!" >&2
            rm -f /tmp/get-docker.sh
            return 1
        fi
    else
        echo "Error: Failed to download Docker installation script!" >&2
        return 1
    fi
}

# Validate email format
function validate_email() {
    local email="$1"
    local strict="${2:-false}"
    
    if [ -n "$email" ]; then
        if [[ "$email" =~ ^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$ ]]; then
            echo "Email: valid"
            return 0
        else
            echo "Error: Invalid email format: $email" >&2
            return 1
        fi
    else
        if [ "$strict" = "true" ]; then
            echo "Error: Email is required" >&2
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
            echo "Password: provided (length: ${#password})"
            return 0
        else
            echo "Error: Password too short (minimum 6 characters)" >&2
            return 1
        fi
    else
        if [ "$strict" = "true" ]; then
            echo "Error: Password is required" >&2
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
            echo "$domain_type Domain: $domain"
            return 0
        else
            echo "Error: Invalid $domain_type domain format: $domain" >&2
            return 1
        fi
    else
        if [ "$strict" = "true" ]; then
            echo "Error: $domain_type domain is required" >&2
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
            echo "Environment: $env"
            return 0
            ;;
        *)
            echo "Error: Invalid environment '$env'. Must be one of: development, staging, production, test" >&2
            return 1
            ;;
    esac
}

# Validate LXD installation and status
function validate_lxd() {
    if ! command -v sudo lxc &>/dev/null; then
        echo "Error: LXD not found. Please install LXD first: https://canonical.com/lxd" >&2
        return 1
    fi
    echo "LXD: installed"
    
    if ! sudo lxc list &>/dev/null; then
        echo "Error: LXD is not running. Please start LXD first: lxc start" >&2
        return 1
    fi
    echo "LXD: running"
    return 0
}

# Validate sudo access
function validate_sudo() {
    if ! sudo -n true 2>/dev/null; then
        echo "Error: Sudo access required for LXD operations" >&2
        return 1
    fi
    echo "Sudo: available"
    return 0
}

# Validate file exists
function validate_file() {
    local file_path="$1"
    local file_name="$2"
    
    if [ ! -f "$file_path" ]; then
        echo "Error: $file_name not found at $file_path" >&2
        return 1
    fi
    echo "$file_name: found"
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
                echo "Error: $url_name not accessible at $url" >&2
                return 1
            else
                echo "Warning: $url_name not accessible at $url (continuing anyway)"
                return 0
            fi
        fi
        echo "$url_name: accessible"
        return 0
    else
        if [ "$strict" = "true" ]; then
            echo "Error: $url_name is required" >&2
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
        echo "Error: $array_name is empty" >&2
        return 1
    fi
    echo "$array_name: ${#array_ref[@]} items configured"
    return 0
}

# Log message to console if show_in_console is true
# default to false
function log_message() {
    local level="$1"
    local message="$2"
    local show_in_console="${3:-false}"
    local timestamp="$(date -Iseconds)"

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
    if sudo lsof -i :"$port" > /dev/null; then
        echo "Port $port is already in use"
        exit 1
    fi
}
