#!/bin/bash

set -e

readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# Default GitHub repository info
DEFAULT_REPO_URL="https://github.com/raghavyuva/nixopus"
DEFAULT_BRANCH="master"

# Variables for custom repository and branch
REPO_URL="$DEFAULT_REPO_URL"
BRANCH="$DEFAULT_BRANCH"
readonly PACKAGE_JSON_URL_MASTER="https://raw.githubusercontent.com/raghavyuva/nixopus/master/package.json"

# Logging functions
log_error() { echo -e "${RED}[ERROR]${NC} $1" >&2; }
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }

# Validate repository URL format
validate_repo_url() {
    local repo_url="$1"
    if [[ ! "$repo_url" =~ ^https://github\.com/[^/]+/[^/]+/?$ ]]; then
        log_error "Invalid repository URL format. Expected: https://github.com/owner/repo"
        exit 1
    fi
}

# Validate branch name
validate_branch() {
    local branch="$1"
    if [[ ! "$branch" =~ ^[a-zA-Z0-9._/-]+$ ]]; then
        log_error "Invalid branch name. Branch names can only contain letters, numbers, dots, underscores, slashes, and hyphens."
        exit 1
    fi
}

# Show usage information
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS] [SUBCOMMAND] [SUBCOMMAND_OPTIONS]

This script installs the nixopus CLI and optionally runs 'nixopus install' with the provided options.

CLI Installation Options:
  --skip-nixopus-install    Skip running 'nixopus install' after CLI installation
  --repo REPOSITORY         GitHub repository URL for CLI installation and passed to 'nixopus install' (default: https://github.com/raghavyuva/nixopus). When custom repo/branch is used, docker-compose-staging.yml will be used instead of docker-compose.yml
  --branch BRANCH           Git branch to use for CLI installation and passed to 'nixopus install' (default: master). When custom repo/branch is used, docker-compose-staging.yml will be used instead of docker-compose.yml

nixopus install Options (passed through to 'nixopus install'):
  -v, --verbose             Show more details while installing
  -t, --timeout SECONDS     How long to wait for each step (default: 300)
  -f, --force               Replace files if they already exist
  -d, --dry-run             See what would happen, but don't make changes
  -c, --config-file PATH    Path to custom config file (defaults to built-in config)
  -ad, --api-domain DOMAIN  The domain where the nixopus api will be accessible
                           (e.g. api.nixopus.com), if not provided you can use
                           the ip address and port (e.g. 192.168.1.100:8443)
  -vd, --view-domain DOMAIN The domain where the nixopus view will be accessible
                           (e.g. nixopus.com), if not provided you can use
                           the ip address and port (e.g. 192.168.1.100:80)
  -h, --help               Show this help message

Subcommands (passed through to 'nixopus install SUBCOMMAND'):
  ssh                       Generate SSH key pair with proper permissions
  deps                      Install dependencies

For subcommand-specific options, use: $0 SUBCOMMAND --help (after CLI installation)

Quick Install (one-liner):
  curl -sSL https://install.nixopus.com | bash
  curl -sSL https://install.nixopus.com | sudo bash

Quick Install with Options:
  curl -sSL https://install.nixopus.com | bash -s -- --verbose
  curl -sSL https://install.nixopus.com | bash -s -- --dry-run
  curl -sSL https://install.nixopus.com | bash -s -- --api-domain api.example.com
  curl -sSL https://install.nixopus.com | bash -s -- ssh --verbose
  curl -sSL https://install.nixopus.com | bash -s -- --repo https://github.com/user/fork --branch develop

Local Examples:
  $0                                           # Install CLI and run 'nixopus install'
  $0 --skip-nixopus-install                    # Only install CLI, skip 'nixopus install'
  $0 --verbose                                 # Install CLI and run 'nixopus install' with verbose output
  $0 --force --timeout 600                     # Install CLI and run 'nixopus install' with force and custom timeout
  $0 --api-domain api.example.com --view-domain example.com  # Install CLI and run 'nixopus install' with custom domains
  $0 --dry-run --config-file /path/to/config   # Install CLI and run 'nixopus install' in dry-run mode with custom config
  $0 --repo https://github.com/user/fork --branch develop  # Install CLI from custom repository and branch
  $0 ssh --verbose                             # Install CLI and run 'nixopus install ssh --verbose'
  $0 deps --dry-run                            # Install CLI and run 'nixopus install deps --dry-run'

EOF
}

# Detect system architecture
detect_arch() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *) log_error "Unsupported architecture: $arch"; exit 1 ;;
    esac
}

# Detect OS and package manager
detect_os() {
    case "$(uname -s)" in
        Darwin*)
            echo "tar"  # macOS uses tar fallback
            ;;
        Linux*)
            if command -v apt &>/dev/null; then
                echo "deb"
            elif command -v yum &>/dev/null || command -v dnf &>/dev/null; then
                echo "rpm"
            elif command -v apk &>/dev/null; then
                echo "apk"
            else
                echo "tar"
            fi
            ;;
        *)
            echo "tar"  # Default fallback
            ;;
    esac
}

# Get CLI version and package list
get_package_info() {
    local package_json
    package_json=$(curl -fsSL "$PACKAGE_JSON_URL_MASTER" 2>/dev/null || true)
    if [[ -z "$package_json" || "$package_json" != \{* ]]; then
        log_error "Failed to fetch package.json from master branch"
        exit 1
    fi

    # Extract version and packages
    CLI_VERSION=$(echo "$package_json" | grep -o '"cli-version":[[:space:]]*"[^"]*"' | cut -d'"' -f4)
    CLI_PACKAGES=$(echo "$package_json" | grep -A 200 '"cli-packages"' | sed -n '/\[/,/\]/p' | grep -o '"[^"]*\..*"' | tr -d '"')

    if [[ -z "$CLI_VERSION" ]]; then
        log_error "Could not find cli-version in package.json"
        exit 1
    fi
}

# Build package name based on system
build_package_name() {
    local arch="$1"
    local pkg_type="$2"
    
    case "$pkg_type" in
        deb) echo "nixopus_${CLI_VERSION}_${arch}.deb" ;;
        rpm) echo "nixopus-${CLI_VERSION}-1.$([ "$arch" = "amd64" ] && echo "x86_64" || echo "aarch64").rpm" ;;
        apk) echo "nixopus_${CLI_VERSION}_${arch}.apk" ;;
        tar) echo "nixopus-${CLI_VERSION}.tar" ;;
        *) log_error "Unknown package type: $pkg_type"; exit 1 ;;
    esac
}

# Check if package exists in CLI packages list
package_exists() {
    local package_name="$1"
    echo "$CLI_PACKAGES" | grep -q "^$package_name$"
}

# Download and install package
install_package() {
    local arch="$1"
    local pkg_type="$2"
    local package_name
    local download_url
    local temp_file
    
    package_name=$(build_package_name "$arch" "$pkg_type")
    
    if ! package_exists "$package_name"; then
        log_error "Package $package_name not found in available packages"
        echo "$CLI_PACKAGES"
        exit 1
    fi
    
    download_url="$REPO_URL/releases/download/nixopus-$CLI_VERSION/$package_name"
    temp_file="/tmp/$package_name"
    
    curl -L -o "$temp_file" "$download_url" || {
        log_error "Failed to download package"
        exit 1
    }
    
    case "$pkg_type" in
        deb)
            sudo dpkg -i "$temp_file" || sudo apt-get install -f -y
            ;;
        rpm)
            if command -v dnf &>/dev/null; then
                sudo dnf install -y "$temp_file"
            else
                sudo yum install -y "$temp_file"
            fi
            ;;
        apk)
            sudo apk add --allow-untrusted "$temp_file"
            ;;
        tar)
            tar -xf "$temp_file" -C /tmp
            
            # Try to install without sudo first (for macOS with writable /usr/local/bin)
            if [[ -w /usr/local/bin ]] || mkdir -p /usr/local/bin 2>/dev/null; then
                cp /tmp/usr/local/bin/nixopus /usr/local/bin/
                chmod +x /usr/local/bin/nixopus
            else
                # Fall back to sudo
                sudo mkdir -p /usr/local/bin
                sudo cp /tmp/usr/local/bin/nixopus /usr/local/bin/
                sudo chmod +x /usr/local/bin/nixopus
            fi
            
            # On macOS, ensure /usr/local/bin is in PATH
            if [[ "$(uname -s)" == "Darwin" ]]; then
                if [[ ":$PATH:" != *":/usr/local/bin:"* ]]; then
                    export PATH="/usr/local/bin:$PATH"
                fi
            fi
            ;;
    esac
    
    # Cleanup
    rm -f "$temp_file"
}

# Verify installation
verify_installation() {
    if ! command -v nixopus &>/dev/null; then
        log_error "Installation verification failed. nixopus command not found."
        exit 1
    fi
}

# Main installation flow for CLI
install_cli() {
    # Detect system
    local arch pkg_type
    arch=$(detect_arch)
    pkg_type=$(detect_os)
    
    # Get package information
    get_package_info
    
    # Install package
    install_package "$arch" "$pkg_type"
    
    # Verify installation
    verify_installation
}

# Check if running as root for package managers that need it
check_permissions() {
    local pkg_type="$1"
    case "$pkg_type" in
        deb|rpm|apk)
            if [[ $EUID -ne 0 ]] && ! sudo -n true 2>/dev/null; then
                echo "This script requires sudo privileges for package installation."
            fi
            ;;
        tar)
            if [[ "$(uname -s)" != "Darwin" ]] || [[ ! -w /usr/local/bin ]]; then
                if [[ $EUID -ne 0 ]] && ! sudo -n true 2>/dev/null; then
                    echo "This script requires sudo privileges to install to /usr/local/bin."
                fi
            fi
            ;;
    esac
}

main() {
    # Default behavior
    SKIP_NIXOPUS_INSTALL=false
    NIXOPUS_INSTALL_ARGS=()
    SUBCOMMAND=""

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-nixopus-install)
                SKIP_NIXOPUS_INSTALL=true
                shift
                ;;
            --repo)
                REPO_URL="$2"
                validate_repo_url "$REPO_URL"
                shift 2
                ;;
            --branch)
                BRANCH="$2"
                validate_branch "$BRANCH"
                shift 2
                ;;
            --verbose|-v)
                NIXOPUS_INSTALL_ARGS+=("$1")
                shift
                ;;
            --timeout|-t)
                NIXOPUS_INSTALL_ARGS+=("$1" "$2")
                shift 2
                ;;
            --force|-f)
                NIXOPUS_INSTALL_ARGS+=("$1")
                shift
                ;;
            --dry-run|-d)
                NIXOPUS_INSTALL_ARGS+=("$1")
                shift
                ;;
            --config-file|-c)
                NIXOPUS_INSTALL_ARGS+=("$1" "$2")
                shift 2
                ;;
            --api-domain|-ad)
                NIXOPUS_INSTALL_ARGS+=("$1" "$2")
                shift 2
                ;;
            --view-domain|-vd)
                NIXOPUS_INSTALL_ARGS+=("$1" "$2")
                shift 2
                ;;
            --help|-h)
                show_usage
                exit 0
                ;;
            ssh|deps)
                # encounter a subcommand, capture it and all remaining args
                SUBCOMMAND="$1"
                shift
                # add all remaining args to the nixopus install command
                NIXOPUS_INSTALL_ARGS+=("$SUBCOMMAND" "$@")
                break
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done

    # Show repository and branch info
    log_info "Using repository: $REPO_URL"
    log_info "Using branch: $BRANCH"
    
    # Run main function with permission check
    pkg_type=$(detect_os)
    check_permissions "$pkg_type"
    install_cli

    if [ "$SKIP_NIXOPUS_INSTALL" = false ]; then
        # Add repository and branch info to nixopus install command
        local cli_args=()
        if [ "$REPO_URL" != "$DEFAULT_REPO_URL" ]; then
            cli_args+=("--repo" "$REPO_URL")
        fi
        if [ "$BRANCH" != "$DEFAULT_BRANCH" ]; then
            cli_args+=("--branch" "$BRANCH")
        fi
        cli_args+=("${NIXOPUS_INSTALL_ARGS[@]}")
        
        if [ -n "$SUBCOMMAND" ]; then
            log_info "Running 'nixopus install' with subcommand and options: ${cli_args[*]}"
        else
            log_info "Running 'nixopus install' with options: ${cli_args[*]}"
        fi
        nixopus install "${cli_args[@]}"
        log_success "nixopus install completed successfully!"
    else
        log_info "Skipping 'nixopus install' as requested..."
        log_success "CLI installation completed! You can now run 'nixopus install' manually with your preferred options."
    fi
}

main "$@"
