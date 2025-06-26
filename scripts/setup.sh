#!/usr/bin/env bash

# Nixopus Development Environment Setup Script
# 
# This script sets up the development environment for the project
# Supported platforms: Linux (Ubuntu, CentOS, Fedora, Arch) and macOS

# Prerequisites:
# - Linux: Run with sudo privileges
# - macOS: Homebrew should be installed (https://brew.sh)
#         Docker Desktop for Mac should be installed and running

# Usage:
# - Linux: sudo ./setup.sh [OPTIONS]
# - macOS: ./setup.sh [OPTIONS] (no sudo required)
#
# Port Configuration:
# Use --help to see available port configuration options
# Example: ./setup.sh --api-port 8081 --view-port 3001

set -euo pipefail


BRANCH="feat/dev_environment"
OS="$(uname)"

# Default port configurations
DEFAULT_API_PORT=8080
DEFAULT_VIEW_PORT=7443
DEFAULT_DB_PORT=5432

function detect_package_manager() {
    if [[ "$OS" == "Darwin" ]]; then
        if command -v brew &>/dev/null; then
            echo "brew"
        else
            echo "Error: Homebrew not found. Please install Homebrew first: https://brew.sh" >&2
            exit 1
        fi
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

function install_package() {
    local pkg_manager
    pkg_manager=$(detect_package_manager)
    
    case $pkg_manager in
        "brew")
            brew install "$1"
            ;;
        "apt")
            apt-get update
            apt-get install -y "$1"
            ;;
        "dnf")
            dnf install -y "$1"
            ;;
        "yum")
            yum install -y "$1"
            ;;
        "pacman")
            pacman -Sy --noconfirm "$1"
            ;;
    esac
}

# check if the os is linux or macOS
function check_os() {
    if [[ "$OS" != "Linux" && "$OS" != "Darwin" ]]; then
        echo "Error: This script is only supported on Linux and macOS." >&2
        exit 1
    fi
}

# check for Docker availability globally
function check_docker() {
    if ! command -v docker &>/dev/null; then
        if [[ "$OS" == "Darwin" ]]; then
            echo "Error: Docker not found. Please ensure Docker Desktop for Mac is installed and running."
            echo "Download from: https://www.docker.com/products/docker-desktop"
        else
            echo "Error: Docker not found. Please install Docker first."
            echo "You can install it using your package manager or from: https://docs.docker.com/engine/install/"
        fi
        exit 1
    fi
    
    # Check if Docker daemon is running
    if ! docker info &>/dev/null; then
        echo "Error: Docker daemon is not running. Please start Docker service."
        exit 1
    fi
    
    echo "Docker check completed."
}

# check for prerequisites on macOS
function check_macos_prerequisites() {
    if [[ "$OS" == "Darwin" ]]; then
        echo "Checking macOS prerequisites"
        
        # Check for Homebrew
        if ! command -v brew &>/dev/null; then
            echo "Error: Homebrew is required on macOS but not found." >&2
            echo "Please install Homebrew first by running:" >&2
            echo '/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"' >&2
            exit 1
        fi
        
        echo "macOS prerequisites check completed."
    fi
}

# check if the script is running as root (only required for Linux)
function check_root() {
    if [[ "$OS" == "Linux" && "$EUID" -ne 0 ]]; then
        echo "Error: Please run as root (sudo) on Linux systems" >&2
        exit 1
    elif [[ "$OS" == "Darwin" && "$EUID" -eq 0 ]]; then
        echo "Warning: consider running without sudo on macos as it is not required." >&2
    fi
}

# check if the required commands are installed
function check_command() {
    local cmd="$1"
    if ! command -v "$cmd" &>/dev/null; then
        echo "Command '$cmd' not found. Attempting to install"
        case "$cmd" in
            "git")
                install_package "git"
                ;;
            "yarn")
                if [[ "$OS" == "Darwin" ]]; then
                    install_package "node"
                    install_package "yarn"
                else
                    install_package "nodejs"
                    install_package "npm"
                    npm install -g yarn
                fi
                ;;
            *)
                echo "Error: Automatic installation not supported for '$cmd'" >&2
                exit 1
                ;;
        esac
    fi
}

function check_required_commands() {
    local commands=("git" "yarn" "go")  
    for cmd in "${commands[@]}"; do
        check_command "$cmd"
    done
}

# check if the go version is installed 
function check_go_version() {
    local go_version=$(go version | awk '{print $3}' | sed 's/^go//')
    local required_version="1.23.4"
    
    local ver_num=$(echo "$go_version" | sed 's/\.//g')
    local req_num=$(echo "$required_version" | sed 's/\.//g')
    
    if [ "$ver_num" -lt "$req_num" ]; then
        echo "Error: Go version $required_version or higher is required. Current version: $go_version" >&2
        exit 1
    fi
}

# clone the nixopus repository
function clone_nixopus() {
    if [ -d "nixopus" ]; then
        echo "nixopus directory already exists, please remove it manually and run the script again"
        exit 1
    fi
    if ! git clone --branch "$BRANCH" https://github.com/raghavyuva/nixopus.git; then
        echo "Error: Failed to clone nixopus repository" >&2
        exit 1
    fi
}   

# checkout to the branch
function checkout_branch() {
    local branch="$1"
    if ! git checkout "$branch"; then
        echo "Error: Failed to checkout to $branch" >&2
        exit 1
    fi
}

# move to the folder
function move_to_folder() {
    local folder="$1"
    if ! cd "$folder"; then
        echo "Error: Failed to change directory to $folder" >&2
        exit 1
    fi
}

# get architecture name in format expected by most download URLs
function get_arch_name() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
    esac
}

# get OS name in lowercase format (for downloads, packages, etc.)
function get_os_name() {
    case "$OS" in
        "Linux") echo "linux" ;;
        "Darwin") echo "darwin" ;;
        *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
    esac
}

function install_go() {
    local version="1.23.4"
    local arch
    arch=$(get_arch_name)
    
    local os
    os=$(get_os_name)
    
    local temp_dir
    temp_dir=$(mktemp -d)
    
    echo "Downloading Go ${version}"
    if ! curl -L "https://go.dev/dl/go${version}.${os}-${arch}.tar.gz" -o "${temp_dir}/go.tar.gz"; then
        echo "Error: Failed to download Go" >&2
        rm -rf "$temp_dir"
        exit 1
    fi

    echo "Verifying checksum"
    local checksum_url="https://go.dev/dl/go${version}.${os}-${arch}.tar.gz.sha256"
    local expected_sum
    expected_sum=$(curl -sL "$checksum_url" | awk '{print $1}')
    local actual_sum
    if [[ "$OS" == "Darwin" ]]; then
        actual_sum=$(shasum -a 256 "${temp_dir}/go.tar.gz" | awk '{print $1}')
    else
        actual_sum=$(sha256sum "${temp_dir}/go.tar.gz" | awk '{print $1}')
    fi
    
    if [[ $expected_sum != $actual_sum ]]; then
        echo "Error: Checksum mismatch for Go archive" >&2
        rm -rf "$temp_dir"
        exit 1
    fi
    
    echo "Installing Go ${version}"
    local go_install_path
    if [[ "$OS" == "Darwin" ]]; then # incase of macOS
        go_install_path="/usr/local"
        if [[ "$EUID" -ne 0 ]]; then
            echo "Installing Go to user directory"
            go_install_path="$HOME"
        fi
    else
        go_install_path="/usr/local"
    fi
    
    if ! rm -rf "${go_install_path}/go" && tar -C "${go_install_path}" -xzf "${temp_dir}/go.tar.gz"; then
        echo "Error: Failed to install Go" >&2
        rm -rf "$temp_dir"
        exit 1
    fi
    
    rm -rf "$temp_dir"
    
    # Set up PATH
    if [[ "$OS" == "Darwin" ]]; then
        local shell_profile
        if [[ "$SHELL" == *"zsh"* ]]; then
            shell_profile="$HOME/.zshrc"
        else
            shell_profile="$HOME/.bash_profile"
        fi
        
        if ! grep -q "${go_install_path}/go/bin" "$shell_profile" 2>/dev/null; then
            echo "export PATH=\$PATH:${go_install_path}/go/bin" >> "$shell_profile"
            echo "Added Go to PATH in $shell_profile"
        fi
        export PATH="$PATH:${go_install_path}/go/bin"
    else
        if ! grep -q "/usr/local/go/bin" /etc/profile.d/go.sh 2>/dev/null; then
            echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile.d/go.sh
            chmod +x /etc/profile.d/go.sh
            source /etc/profile.d/go.sh
        fi
    fi
}

# check if the go version is installed else install it 1.23.4
function check_go_version() {
    if ! command -v go &>/dev/null; then
        echo "Go is not installed. Installing"
        install_go
    fi
}

# install air hot reload for golang
function install_air_hot_reload(){
    local user_home
    user_home=$(get_user_home)
    local air_path="$user_home/go/bin/air"
    
    # Check if air is already installed
    if command -v air &>/dev/null || [[ -f "$air_path" ]]; then
        echo "Air hot reload is already installed, skipping installation"
        export PATH="$PATH:$user_home/go/bin"
        return 0
    fi
    
    echo "Installing Air hot reload"
    sudo -u "${SUDO_USER:-$USER}" env GOPATH="$user_home/go" go install github.com/air-verse/air@latest
    export PATH="$PATH:$user_home/go/bin"
    
    # Verify installation
    if command -v air &>/dev/null || [[ -f "$air_path" ]]; then
        echo "Air hot reload installed successfully"
    else
        echo "Warning: Air installation may have failed"
    fi
}

# load the env variables from the api/.env.sample file
function load_api_env_variables(){
    move_to_folder "api"
    if [ -f .env.sample ]; then
        while IFS='=' read -r key value; do
            # Skip empty lines and comments
            [[ -z "$key" || "$key" =~ ^# ]] && continue
            # Remove any quotes from the value
            value=$(echo "$value" | tr -d '"'"'")
            # Export the variable
            export "$key=$value"
        done < .env.sample
    else
        echo "Error: .env.sample file not found in api directory" >&2
        exit 1
    fi
    move_to_folder ".."
}

# setup postgres with docker
function setup_postgres_with_docker(){
    load_api_env_variables

    # Check if container already exists
    if docker ps -a --format 'table {{.Names}}' | grep -q "^nixopus-db$"; then
        echo "Already nixopus-db container exists"
        return 0
    fi
    
    # Start PostgreSQL container with credentials matching .env.sample
    docker run -d --name nixopus-db \
        -e POSTGRES_USER="${USERNAME:-postgres}" \
        -e POSTGRES_PASSWORD="${PASSWORD:-12344}" \
        -e POSTGRES_DB="${DB_NAME:-postgres}" \
        -e POSTGRES_HOST_AUTH_METHOD=trust \
        -p "${DB_PORT:-5432}:5432" \
        --health-cmd="pg_isready -U ${USERNAME:-postgres} -d ${DB_NAME:-postgres}" \
        postgres:14-alpine
    
    echo "Waiting for PostgreSQL to be ready"
    sleep 5
    
    # Wait for PostgreSQL to be ready
    local max_attempts=30
    local attempt=1
    while [ $attempt -le $max_attempts ]; do
        if docker exec nixopus-db pg_isready -U "${USERNAME:-postgres}" -d "${DB_NAME:-postgres}" >/dev/null 2>&1; then
            echo "PostgreSQL is ready!"
            break
        fi
        echo "Waiting for PostgreSQL (attempt $attempt/$max_attempts)"
        sleep 2
        attempt=$((attempt + 1))
    done
    
    if [ $attempt -gt $max_attempts ]; then
        echo "Error: PostgreSQL failed to start within expected time" >&2
        exit 1
    fi
    
    echo "Postgres setup completed successfully"
}

# verify database connection
function verify_database_connection(){
    echo "Verifying database connection"
    load_api_env_variables
    
    # Test connection using docker exec
    if docker exec nixopus-db psql -U "${USERNAME:-postgres}" -d "${DB_NAME:-postgres}" -c "SELECT 1;" >/dev/null 2>&1; then
        echo "Database connection verified successfully"
    else
        echo "Error: Failed to connect to database" >&2
        echo "Details: user=${USERNAME:-postgres}, db=${DB_NAME:-postgres}, host=${HOST_NAME:-localhost}, port=${DB_PORT:-5432}" >&2
        exit 1
    fi
}

# setup ssh will create a ssh key and add it to the authorized_keys file
function setup_ssh(){
    local user_home
    user_home=$(get_user_home)
    local ssh_dir="$user_home/.ssh"
    local private_key="$ssh_dir/id_ed25519_nixopus"
    local public_key="$ssh_dir/id_ed25519_nixopus.pub"
    
    if [[ -f "$private_key" && -f "$public_key" ]]; then
        echo "SSH key for Nixopus already exists, skipping ssh setup"
        return 0
    fi
    
    echo "setting up SSH config"
    
    # Check if ssh-keygen is available and install if needed
    if ! command -v ssh-keygen &>/dev/null; then
        echo "Installing openssh"
        if [[ "$OS" == "Darwin" ]]; then
            install_package "openssh"
        else
            case $(detect_package_manager) in
                "apt") install_package "openssh-client" ;;
                "dnf"|"yum") install_package "openssh-clients" ;;
                "pacman") install_package "openssh" ;;
            esac
        fi
    fi
    
    # Check SSH daemon availability on macOS
    if [[ "$OS" == "Darwin" ]]; then
        echo "Checking SSH daemon (Remote Login) status on macOS "
        
        # Check if SSH daemon is running
        if ! sudo launchctl list | grep -q "com.openssh.sshd"; then
            echo ""
            echo "WARNING: SSH Remote Login is not enabled on this macOS system!"
            echo ""
            echo "To enable SSH Remote Login, please follow these steps:"
            echo "1. Open System Settings (or System Preferences on older macOS versions)"
            echo "2. Go to General → Sharing (or just Sharing on older versions)"
            echo "3. Turn on 'Remote Login'"
            echo "4. You can choose to allow access for:"
            echo "   - All users, or"
            echo "   - Only specific users (recommended for security)"
            echo ""
            echo "alternatively, you can enable it via command line by running:"
            echo "   sudo systemsetup -setremotelogin on"
            echo ""
            echo "after enabling Remote Login, please run this setup script again."
            echo ""
            read -p "press Enter to continue with SSH key generation (you'll still need to enable Remote Login) "
        else
            echo "SSH Remote Login is already enabled on macOS"
        fi
    fi
    
    local authorized_keys="$ssh_dir/authorized_keys"

    mkdir -p "$ssh_dir" && chmod 700 "$ssh_dir"
    
    echo "Generating Nixopus SSH key"
    ssh-keygen -t ed25519 -f "$private_key" -N "" -C "nixopus-$(whoami)@$(hostname)-$(date +%Y%m%d)"
    chmod 600 "$private_key" && chmod 644 "$public_key"
    
    if [[ ! -f "$authorized_keys" || ! $(grep -Fq "$(cat "$public_key")" "$authorized_keys" 2>/dev/null) ]]; then
        cat "$public_key" >> "$authorized_keys" && chmod 600 "$authorized_keys"
        echo "Nixopus public key added to authorized_keys"
    fi
    
    echo "Nixopus SSH setup is done"
}

# Function to update SSH configuration in environment files
function update_ssh_env_config(){
    local user_home
    user_home=$(get_user_home)
    local ssh_dir="$user_home/.ssh"
    local private_key="$ssh_dir/id_ed25519_nixopus"
    local current_user=$(whoami)
    
    echo "Updating SSH configuration in environment files "
    
    # Update API environment file
    if [[ -f "api/.env" ]]; then
        # Update SSH settings in API .env file
        sed -i.bak "s|SSH_HOST=.*|SSH_HOST=localhost|g" api/.env
        sed -i.bak "s|SSH_PORT=.*|SSH_PORT=22|g" api/.env
        sed -i.bak "s|SSH_USER=.*|SSH_USER=$current_user|g" api/.env
        
        # For development environment, set up SSH private key (recommended)
        # Users can optionally use SSH_PASSWORD for development if preferred
        if grep -q "SSH_PRIVATE_KEY=" api/.env; then
            sed -i.bak "s|SSH_PRIVATE_KEY=.*|SSH_PRIVATE_KEY=$private_key|g" api/.env
        else
            # Add SSH_PRIVATE_KEY if it doesn't exist
            echo "SSH_PRIVATE_KEY=$private_key" >> api/.env
        fi
        
        # Ensure SSH_PASSWORD exists as commented option for development
        if ! grep -q "SSH_PASSWORD=" api/.env; then
            echo "# SSH_PASSWORD=<YOUR_SSH_PASSWORD_HERE>" >> api/.env
        fi
        
        # Remove backup file
        rm -f api/.env.bak
        
        echo "SSH configuration updated in api/.env"
        echo "  - SSH_HOST: localhost"
        echo "  - SSH_PORT: 22"
        echo "  - SSH_USER: $current_user"
        echo "  - SSH_PRIVATE_KEY: $private_key (recommended for production)"
        echo "  - SSH_PASSWORD: Available as commented option for development"
        echo ""
        echo "Note: For development, you can uncomment SSH_PASSWORD and comment out SSH_PRIVATE_KEY if preferred"
    else
        echo "Warning: api/.env file not found, SSH configuration not updated"
    fi
}

# Update environment files with custom port configurations
function update_port_configurations(){
    echo "Updating port configurations "
    
    # Update API environment file
    if [[ -f "api/.env" ]]; then
        sed -i.bak "s|PORT=.*|PORT=$API_PORT|g" api/.env
        sed -i.bak "s|DB_PORT=.*|DB_PORT=$DB_PORT|g" api/.env
        sed -i.bak "s|ALLOWED_ORIGIN=.*|ALLOWED_ORIGIN=http://localhost:$VIEW_PORT|g" api/.env
        
        rm -f api/.env.bak
        echo "Updated API environment with custom ports"
    fi
    
    # Update view environment file
    if [[ -f "view/.env" ]]; then
        sed -i.bak "s|PORT=.*|PORT=$VIEW_PORT|g" view/.env
        sed -i.bak "s|NEXT_PUBLIC_PORT=.*|NEXT_PUBLIC_PORT=$VIEW_PORT|g" view/.env
        
        rm -f view/.env.bak
        echo "Updated view environment with custom ports"
    fi
    
    echo "Port configurations updated successfully"
    echo "  - API Port: $API_PORT"
    echo "  - Frontend Port: $VIEW_PORT"
    echo "  - Database Port: $DB_PORT"
}

# setup environment variables
function setup_environment_variables(){
    move_to_folder "api"
    if [ -f .env.sample ]; then
        cp .env.sample .env || { echo "Error: Failed to copy api/.env.sample to .env" >&2; exit 1; }
    else
        echo "Error: api/.env.sample file not found" >&2
        exit 1
    fi
    move_to_folder ".."
    
    move_to_folder "view"
    if [ -f .env.sample ]; then
        cp .env.sample .env || { echo "Error: Failed to copy view/.env.sample to .env" >&2; exit 1; }
    else
        echo "Error: view/.env.sample file not found" >&2
        exit 1
    fi
    move_to_folder ".."
    echo "Environment variables setup completed successfully"
}

# start the api server
function start_api(){
    move_to_folder "api"
    go mod tidy
    go mod download
    
    local user_home
    user_home=$(get_user_home)
    
    echo "API server started with air hot reload"
    echo "Logs can be found in api.log"
    echo "You can stop the server using 'pkill -f air' command"
    nohup "$user_home/go/bin/air" > api.log 2>&1 &
    
}

open_discord_gh_link() {

  local url="https://discord.com/invite/skdcq39Wpv"
  local gh_url="https://github.com/raghavyuva/nixopus/"

  case "$OS" in
    Darwin)
      open "$url" 2>/dev/null || echo "Could not open browser on macOS"
      open "$gh_url" 2>/dev/null || echo "Could not open browser on macOS"
      ;;
    Linux)
      if command -v xdg-open &>/dev/null; then
        xdg-open "$url" 2>/dev/null || echo "Could not open Discord link"
        xdg-open "$gh_url" 2>/dev/null || echo "Could not open GitHub link"
      else
        echo "Warning: Could not auto-open browser." >&2
      fi
      ;;
    *)
      echo "Warning: Unsupported OS for browser launch." >&2
      ;;
  esac
}


# start the view server
function start_view(){
    move_to_folder "view"
    yarn install --frozen-lockfile
    
    # Read PORT from .env file
    local view_port=7443  # default fallback
    if [[ -f ".env" ]]; then
        view_port=$(grep "^PORT=" .env | cut -d'=' -f2 | tr -d ' ')
        view_port=${view_port:-7443}  # fallback if empty
    fi
    
    echo "View server started on port $view_port"
    echo "Logs can be found in view.log"
    echo "You can stop the server using 'pkill -f yarn' command"
    nohup yarn run dev -- -p "$view_port" > view.log 2>&1 &

}

# Check if a port is available
function is_port_available() {
    local port=$1
    local host=${2:-localhost}
    
    # Try to connect to the port, if it fails the port is available
    ! nc -z "$host" "$port" 2>/dev/null
}

# Validate that a port number is valid
function validate_port() {
    local port=$1
    local port_name=$2
    
    if ! [[ "$port" =~ ^[0-9]+$ ]] || [ "$port" -lt 1 ] || [ "$port" -gt 65535 ]; then
        echo "Error: Invalid $port_name port '$port'. Port must be a number between 1 and 65535."
        return 1
    fi
    
    if [ "$port" -lt 1024 ] && [ "$EUID" -ne 0 ]; then
        echo "Warning: $port_name port $port is below 1024 and may require root privileges."
    fi
    
    return 0
}

# Parse command line arguments for custom ports
function parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --api-port)
                API_PORT="$2"
                shift 2
                ;;
            --view-port)
                VIEW_PORT="$2"
                shift 2
                ;;
            --db-port)
                DB_PORT="$2"
                shift 2
                ;;
            --help|-h)
                show_usage
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # Set defaults if not provided
    API_PORT=${API_PORT:-$DEFAULT_API_PORT}
    VIEW_PORT=${VIEW_PORT:-$DEFAULT_VIEW_PORT}
    DB_PORT=${DB_PORT:-$DEFAULT_DB_PORT}
    
    # Validate all ports
    validate_port "$API_PORT" "API" || exit 1
    validate_port "$VIEW_PORT" "View" || exit 1
    validate_port "$DB_PORT" "Database" || exit 1
}

# Show usage information
function show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --api-port PORT      Set API server port (default: $DEFAULT_API_PORT)"
    echo "  --view-port PORT     Set frontend server port (default: $DEFAULT_VIEW_PORT)"
    echo "  --db-port PORT       Set database port (default: $DEFAULT_DB_PORT)"
    echo "  --help, -h           Show this help message"
    echo ""
    echo "After setup completion, default admin credentials will be created:"
    echo "  Email: \$USER@example.com (where \$USER is your system username)"
    echo "  Password: Nixopus123!"
    echo ""
}

# Check availability of all required ports
function check_port_availability() {
    echo "Checking port availability "
    
    # Check API port
    if ! is_port_available "$API_PORT"; then
        echo "Error: API port $API_PORT is already in use."
        echo "Please use a different port: ./setup.sh --api-port <PORT>"
        exit 1
    fi
    
    # Check Frontend port
    if ! is_port_available "$VIEW_PORT"; then
        echo "Error: Frontend port $VIEW_PORT is already in use."
        echo "Please use a different port: ./setup.sh --view-port <PORT>"
        exit 1
    fi
    
    # Check Database port
    if ! is_port_available "$DB_PORT"; then
        echo "Error: Database port $DB_PORT is already in use."
        echo "Please use a different port: ./setup.sh --db-port <PORT>"
        exit 1
    fi
    
    echo "All required ports are available."
    return 0
}

# Function to perform comprehensive SSH health checks
function ssh_health_check(){
    local user_home
    user_home=$(get_user_home)
    local ssh_dir="$user_home/.ssh"
    local private_key="$ssh_dir/id_ed25519_nixopus"
    local current_user=$(whoami)
    
    echo "Running SSH health check "
    
    # 1. Check SSH keys exist
    if [[ ! -f "$private_key" || ! -f "$private_key.pub" ]]; then
        echo "Error: SSH keys missing"
        return 1
    fi
    
    # 2. Quick SSH connection test
    if timeout 10 ssh -o ConnectTimeout=5 -o BatchMode=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i "$private_key" "$current_user@localhost" "exit" &>/dev/null; then
        echo "SSH connection successful"
        return 0
    else
        echo "SSH connection failed - please enable Remote Login in System Settings → Sharing (macOS) or start SSH service (Linux)"
        return 1
    fi
}

# Get the correct user home directory (handles sudo scenarios)
function get_user_home(){
    local user_home
    if [[ "$OS" == "Darwin" ]]; then
        user_home="$HOME"
        if [[ "$EUID" -eq 0 && -n "${SUDO_USER:-}" ]]; then
            user_home=$(eval echo "~$SUDO_USER")
        fi
    else
        user_home=$(eval echo ~${SUDO_USER:-$USER})
    fi
    echo "$user_home"
}

function create_admin_credentials() {
    echo "Creating default admin credentials "
    
    # Simple wait for API to be ready
    local retries=3
    while [ $retries -gt 0 ]; do
        if curl -s "http://localhost:$API_PORT/api/v1/health" >/dev/null 2>&1; then
            break
        fi
        echo "Waiting for API server  ($retries attempts left)"
        sleep 2
        retries=$((retries - 1))
    done
    
    if [ $retries -eq 0 ]; then
        echo "Warning: API server not responding, skipping admin creation"
        return 1
    fi
    
    # Check if admin already exists
    if curl -s "http://localhost:$API_PORT/api/v1/auth/is-admin-registered" | grep -q '"admin_registered":true'; then
        echo "Admin already registered, skipping"
        return 0
    fi
    
    # Create admin user
    local username=${USER:-"admin"}
    local email="${username}@example.com"
    local password="Nixopus123!"
    
    echo "Creating admin: $email"
    
    if curl -s "http://localhost:$API_PORT/api/v1/auth/register" \
        -H 'content-type: application/json' \
        -d "{\"email\":\"$email\",\"password\":\"$password\",\"username\":\"$username\",\"type\":\"admin\"}" \
        | grep -q '"status":"success"'; then
        
        echo "Admin credentials created successfully!"
        echo "Email: $email | Password: $password"
        echo "Login at: http://localhost:$VIEW_PORT"
    else
        echo "Failed to create admin credentials"
        echo "Manual command:"
        echo "curl -v 'http://localhost:$API_PORT/api/v1/auth/register' -H 'content-type: application/json' -d '{\"email\":\"$email\",\"password\":\"$password\",\"username\":\"$username\",\"type\":\"admin\"}'"
    fi
}

# main function
function main() {
    # Parse command line arguments first
    parse_arguments "$@"
    
    echo "Starting Nixopus development environment setup"
    check_os
    check_macos_prerequisites
    check_root
    check_docker
    check_required_commands
    
    # Check if ports are available before proceeding
    if ! check_port_availability; then
        echo "Setup cannot continue due to port conflicts. Please resolve them and try again."
        exit 1
    fi
    
    check_go_version
    clone_nixopus
    move_to_folder "nixopus"
    checkout_branch "$BRANCH"
    install_air_hot_reload
    echo "Nixopus repository cloned and configured successfully"
    
    setup_postgres_with_docker
    verify_database_connection
    setup_environment_variables
    setup_ssh
    update_ssh_env_config
    update_port_configurations
    
    # Perform SSH health checks
    echo "Running SSH health checks "
    if ! ssh_health_check; then
        if [[ "$OS" == "Darwin" ]]; then
            echo "Enabling SSH Remote Login..."
            sudo systemsetup -setremotelogin on &>/dev/null
        else
            echo "Starting SSH service..."
            sudo systemctl start ssh &>/dev/null || sudo systemctl start sshd &>/dev/null
        fi
        
        sleep 2
        if ssh_health_check; then
            echo "SSH working!"
        else
            echo "SSH still not working, continuing setup..."
        fi
    fi
    
    echo "SSH setup completed successfully"
    
    start_api
    echo "API server started successfully"
    move_to_folder ".."
    start_view
    echo "View server started successfully"
    
    echo "Waiting for applications to fully initialize"
    sleep 3

    RED='\e[31m'
    GREEN='\e[32m'
    YELLOW='\e[33m'
    BLUE='\e[34m'
    MAGENTA='\e[35m'
    CYAN='\e[36m'
    RESET='\e[0m'

    printf "${CYAN}    _   __    ____   _  __   ____     ____    __  __   _____${RESET}\n"
    printf "${GREEN}   / | / /   /  _/  | |/ /  / __ \\\\   / __ \\\\  / / / /  / ___/${RESET}\n"
    printf "${YELLOW}  /  |/ /    / /    |   /  / / / /  / /_/ / / / / /   \\\\__ \\\\ ${RESET}\n"
    printf "${BLUE} / /|  /   _/ /    /   |  / /_/ /  / ____/ / /_/ /   ___/ / ${RESET}\n"
    printf "${MAGENTA}/_/ |_/   /___/   /_/|_|  \\\\____/  /_/      \\\\____/   /____/  ${RESET}\n"
    printf "\n"


    # Create default admin credentials
    create_admin_credentials
    echo "Nixopus development environment setup completed successfully"
    echo "-------------------------------------------------------------"    
    echo ""
    echo "=== Application Access ==="
    echo "Frontend: http://localhost:$VIEW_PORT"
    echo "API: http://localhost:$API_PORT"
    echo "Database: localhost:$DB_PORT"
    echo ""
    echo "=== Default Login Credentials ==="
    echo "Email: ${USER:-admin}@example.com"
    echo "Password: Nixopus123!"
    echo "Change these credentials after first login!"
    echo ""
    echo "=== Troubleshooting ==="
    echo "If you encounter database connection issues:"
    echo "1. Check if Docker container is running: docker ps | grep nixopus-db"
    echo "2. Check database logs: docker logs nixopus-db"
    echo "3. Verify connection: docker exec nixopus-db psql -U postgres -d postgres -c 'SELECT 1;'"
    echo "4. Restart the database: docker restart nixopus-db"
    echo ""
    echo "To manually create admin credentials later:"
    echo "  Use the curl command shown above"
    echo ""
    echo "Log files:"
    echo "- API logs: nixopus/api/api.log"
    echo "- View logs: nixopus/view/view.log"
    echo "----------------------------------------------------------------------------"
    
    echo ""
    echo "Need help or have questions?"
    echo ">>>> Join our Discord :: https://discord.com/invite/skdcq39Wpv"
    echo ">>>> Star us on GitHub: https://github.com/raghavyuva/nixopus/"
    echo ">>>> Raise issues on GitHub Issues: https://github.com/raghavyuva/nixopus/issues"
    open_discord_gh_link

}

main "$@"
