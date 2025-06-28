#!/bin/bash

# Requirements:
# - LXD should be installed

# This script is used for testing the installation script across multiple distributions.
# It will create containers for each distribution, run the installation, and report results.

set -euo pipefail
OS="$(uname)"

CONTAINER_NAME=""
TEST_RESULTS=()

# DISTRO_MATRIX is the list of distributions to test these names are from the lxc image list command
# TODO: Uncomment the distributions to test once the first two tests seemed to be working
DISTRO_MATRIX=(
    "alpine/3.19"
    "fedora/41"
    # "archlinux"
    # "debian/11"
    # "centos/9-Stream"
    # "gentoo/openrc"
    # "ubuntu/22.04"
)

# Parse command line arguments
function parse_arguments() {
    local email=""
    local password=""
    local api_domain=""
    local app_domain=""
    local env="production"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --email=*)
                email="${1#*=}"
                shift
                ;;
            --password=*)
                password="${1#*=}"
                shift
                ;;
            --api-domain=*)
                api_domain="${1#*=}"
                shift
                ;;
            --app-domain=*)
                app_domain="${1#*=}"
                shift
                ;;
            --env=*)
                env="${1#*=}"
                shift
                ;;
            *)
                echo "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    echo "$email:$password:$api_domain:$app_domain:$env"
}

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

# Check if LXD is installed
function is_lxd_installed() {
    if ! command -v sudo lxc --version &>/dev/null; then
        echo "Error: LXD not found. Please install LXD first: https://canonical.com/lxd" >&2
        exit 1
    fi
}   

# Install the package
function install_package() {
    local pkg_manager
    pkg_manager=$(detect_package_manager)
    
    case $pkg_manager in
        "brew")
            brew install "$1"
            ;;
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

# Install the dependencies
function install_dependencies() {
    local dependencies
    dependencies=(
        "python3"
        "python3-pip"
        "git"
        "openssl"
    )
    for dependency in "${dependencies[@]}"; do
        install_package "$dependency"
    done
}

# Generate a random string and add it to the test-container- prefix
function get_lxd_container_name() {
    local distro="$1"
    echo "test-container-$(date +%s%N | md5sum | cut -c1-8)"
}


# Create a new LXD container with Docker privileges
create_lxd_container() {
    local distro="$1"
    CONTAINER_NAME=$(get_lxd_container_name "$distro")

    echo "Creating container: $CONTAINER_NAME with image: $distro"
    sudo lxc launch images:"$distro" "$CONTAINER_NAME"
    sudo lxc config set "$CONTAINER_NAME" security.privileged true
    sudo lxc config set "$CONTAINER_NAME" security.nesting true
    sudo lxc restart "$CONTAINER_NAME"
    echo "Waiting for container to be ready..."
    sleep 60
    sudo lxc exec "$CONTAINER_NAME" -- cloud-init status --wait || true
}

# Install dependencies in the container
function install_dependencies_in_container() {
    local container_name="$1"
    local distro="$2"
    echo "Installing dependencies in container: $container_name"

    # Extract base distribution name (e.g., "alpine" from "alpine/3.19")
    local base_distro=$(echo "$distro" | cut -d'/' -f1)
    echo "Base distribution: $base_distro"

    sudo lxc exec "$container_name" -- sh -c 'set -x'
    # Install the dependencies based on the distro
    case $base_distro in
        "alpine")
            echo "Installing dependencies for Alpine"
            sudo lxc exec "$container_name" -- apk add --no-cache python3 docker py3-pip git openssl curl
            ;;
        "fedora")
            echo "Installing dependencies for Fedora"
            sudo lxc exec "$container_name" -- dnf install -y python3 docker python3-pip git openssl curl
            ;;
        "debian")
            echo "Installing dependencies for Debian"
            sudo lxc exec "$container_name" -- apt-get update
            sudo lxc exec "$container_name" -- apt-get install -y python3 docker python3-pip git openssl curl
            ;;
        "archlinux")
            echo "Installing dependencies for Arch Linux"
            sudo lxc exec "$container_name" -- pacman -Sy --noconfirm python3 docker python3-pip git openssl curl
            ;;
        "centos")
            echo "Installing dependencies for CentOS"
            sudo lxc exec "$container_name" -- yum install -y python3 docker python3-pip git openssl curl
            ;;
        "gentoo")
            echo "Installing dependencies for Gentoo"
            sudo lxc exec "$container_name" -- emerge --ask nix python3 docker python3-pip git openssl curl
            ;;
        "ubuntu")
            echo "Installing dependencies for Ubuntu"
            sudo lxc exec "$container_name" -- apt-get update
            sudo lxc exec "$container_name" -- apt-get install -y python3 docker python3-pip git openssl curl
            ;;
        *)
            echo "Unknown distribution: $base_distro"
            return 1
            ;;
    esac
    echo "Finished installing dependencies in container: $container_name"
}

# Build installation command
function build_installation_command() {
    local email="$1"
    local password="$2"
    local api_domain="$3"
    local app_domain="$4"
    local env="$5"
    
    local cmd="sudo bash -c \"\$(curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install.sh)\""
    
    if [ -n "$email" ]; then
        cmd="$cmd --email=$email"
    fi
    
    if [ -n "$password" ]; then
        cmd="$cmd --password=$password"
    fi
    
    if [ -n "$api_domain" ]; then
        cmd="$cmd --api-domain=$api_domain"
    fi
    
    if [ -n "$app_domain" ]; then
        cmd="$cmd --app-domain=$app_domain"
    fi
    
    cmd="$cmd --env=$env"
    
    echo "$cmd"
}

# Run installation script
function run_installation_script() {
    local command="$1"
    echo "Running installation script in container: $CONTAINER_NAME"
    echo "This may take several minutes..."
    
    # Run with timeout of 10 minutes (600 seconds)
    if timeout 600 sudo lxc exec "$CONTAINER_NAME" -- bash -c "set -x; $command"; then
        echo "Installation script completed successfully"
        return 0
    else
        local exit_code=$?
        if [ $exit_code -eq 124 ]; then
            echo "Installation script timed out after 30 minutes"
        else
            echo "Installation script failed with exit code: $exit_code"
        fi
        return $exit_code
    fi
}

# Stop and delete container
function cleanup_container() {
    if [ -n "$CONTAINER_NAME" ]; then
        echo "Stopping container: $CONTAINER_NAME"
        sudo lxc stop "$CONTAINER_NAME" 2>/dev/null || true
        
        echo "Deleting container: $CONTAINER_NAME"
        sudo lxc delete "$CONTAINER_NAME" 2>/dev/null || true
    fi
}

# Setup trap for cleanup
function setup_cleanup_trap() {
    trap cleanup_container EXIT
}

# Test installation on a single distribution with specific parameters
function test_distribution_with_params() {
    local distro="$1"
    local test_name="$2"
    local env="$3"
    local email="$4"
    local password="$5"
    local api_domain="$6"
    local app_domain="$7"
    local test_result=""
    
    echo "=========================================="
    echo "Testing: $distro - $test_name"
    echo "Environment: $env"
    echo "Email: ${email:-<not set>}"
    echo "Password: ${password:+<set>}"
    echo "API Domain: ${api_domain:-<not set>}"
    echo "App Domain: ${app_domain:-<not set>}"
    echo "=========================================="
    
    # Reset container name for this test
    CONTAINER_NAME=""
    
    if ! create_lxd_container "$distro"; then
        TEST_RESULTS+=("$distro-$test_name: SKIPPED (not available)")
        return 0
    fi

    install_dependencies_in_container "$CONTAINER_NAME" "$distro"

    echo "Starting installation script..."
    if run_installation_script "$(build_installation_command "$email" "$password" "$api_domain" "$app_domain" "$env")"; then
        test_result="PASSED"
    else
        test_result="FAILED"
    fi
    echo "Installation script completed with result: $test_result"
    
    cleanup_container
    
    TEST_RESULTS+=("$distro-$test_name: $test_result")
    echo "Test result for $distro-$test_name: $test_result"
}

# Test installation on a single distribution
function test_distribution() {
    local distro="$1"
    local email="$2"
    local password="$3"
    local api_domain="$4"
    local app_domain="$5"
    local env="$6"
    
    local test_name="custom"
    test_distribution_with_params "$distro" "$test_name" "$env" "$email" "$password" "$api_domain" "$app_domain"
}

# Print test results summary
function print_test_summary() {
    echo ""
    echo "=========================================="
    echo "TEST RESULTS SUMMARY"
    echo "=========================================="
    
    local passed=0
    local failed=0
    local skipped=0
    
    for result in "${TEST_RESULTS[@]}"; do
        echo "$result"
        if [[ "$result" == *": PASSED" ]]; then
            ((passed++))
        elif [[ "$result" == *": FAILED" ]]; then
            ((failed++))
        elif [[ "$result" == *": SKIPPED" ]]; then
            ((skipped++))
        fi
    done
    
    echo ""
    echo "Summary:"
    echo "  Passed: $passed"
    echo "  Failed: $failed"
    echo "  Skipped: $skipped"
    echo "  Total: ${#TEST_RESULTS[@]}"
    
    if [ "$failed" -gt 0 ]; then
        echo ""
        echo "Some tests failed. Please check the logs above."
        exit 1
    else
        echo ""
        echo "All tests passed successfully!"
    fi
}

function main() {
    # Parse command line arguments
    local args
    args=$(parse_arguments "$@")
    IFS=':' read -r email password api_domain app_domain env <<< "$args"
    
    echo "Starting Nixopus installation matrix test..."
    echo "Environment: $env"
    echo "Email: ${email:-<not set>}"
    echo "Password: ${password:+<set>}"
    echo "API Domain: ${api_domain:-<not set>}"
    echo "App Domain: ${app_domain:-<not set>}"
    
    echo "Testing distributions: ${DISTRO_MATRIX[*]}"
    
    echo ""
    
    is_lxd_installed
    install_dependencies
    
    for distro in "${DISTRO_MATRIX[@]}"; do
        test_distribution "$distro" "$email" "$password" "$api_domain" "$app_domain" "$env"
    done
    
    print_test_summary
}

main "$@"
