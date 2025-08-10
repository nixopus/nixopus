#!/bin/bash

readonly RED='\033[0;31m'
readonly NC='\033[0m'

# GitHub repository info
readonly REPO_URL="https://github.com/raghavyuva/nixopus"
readonly PACKAGE_JSON_URL="$REPO_URL/raw/master/package.json"

# Logging functions
log_error() { echo -e "${RED}[ERROR]${NC} $1" >&2; }

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
    package_json=$(curl -s "$PACKAGE_JSON_URL" || {
        log_error "Failed to fetch package.json from repository"
        exit 1
    })
    
    # Extract version and packages
    CLI_VERSION=$(echo "$package_json" | grep -o '"cli-version":[[:space:]]*"[^"]*"' | cut -d'"' -f4)
    CLI_PACKAGES=$(echo "$package_json" | grep -A 100 '"cli-packages"' | sed -n '/\[/,/\]/p' | grep -o '"[^"]*\..*"' | tr -d '"')
    
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

# Main installation flow
main() {
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

# Run main function with permission check
pkg_type=$(detect_os)
check_permissions "$pkg_type"
main "$@"