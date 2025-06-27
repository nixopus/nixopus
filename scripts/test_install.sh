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
    if ! command -v lxc --version &>/dev/null; then
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

# Get first standard container alias for a distribution
function get_first_standard_container_alias() {
    local distro="$1"
    
    # List all images for the given distro in CSV format
    local all_images
    all_images=$(lxc image list images:"$distro" --format=csv)
    
    # Filter only CONTAINER entries
    local containers
    containers=$(echo "$all_images" | grep ',CONTAINER,')
    
    # Exclude cloud images
    local non_cloud
    non_cloud=$(echo "$containers" | grep -v 'cloud')
    
    # Extract the alias column (first field) and remove ' (n more)'
    local aliases
    aliases=$(echo "$non_cloud" | awk -F',' '{print $1}' | sed 's/ .*//')
    
    # Pick the first available alias
    local first_alias
    first_alias=$(echo "$aliases" | head -n1)
    
    echo "$first_alias"
}

# Generate a random string
function generate_random_string() {
    local random_string
    random_string=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)
    echo "$random_string"
}

# Generate a random string and add it to the test-container- prefix
function get_lxd_container_name() {
    local random_string
    random_string=$(generate_random_string)
    echo "test-container-$random_string"
}


# Create a new lxd container for a specific distribution
function create_lxd_container() {
    local distro="$1"
    local distro_alias
    distro_alias=$(get_first_standard_container_alias "$distro")
    CONTAINER_NAME=$(get_lxd_container_name)
    
    # if empty alias then skip
    if [ -z "$distro_alias" ]; then
        echo "Distribution $distro is not available, skipping..."
        return 1
    fi

    echo "Creating container: $CONTAINER_NAME with image: $distro_alias"
    lxc launch images:"$distro_alias" "$CONTAINER_NAME"
    
    # Wait for container to be ready
    echo "Waiting for container to be ready..."
    lxc exec "$CONTAINER_NAME" -- cloud-init status --wait 2>/dev/null || true
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
    lxc exec "$CONTAINER_NAME" -- bash -c "$command"
}

# Stop and delete container
function cleanup_container() {
    if [ -n "$CONTAINER_NAME" ]; then
        echo "Stopping container: $CONTAINER_NAME"
        lxc stop "$CONTAINER_NAME" 2>/dev/null || true
        
        echo "Deleting container: $CONTAINER_NAME"
        lxc delete "$CONTAINER_NAME" 2>/dev/null || true
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
    
    # Setup cleanup trap for this test
    trap cleanup_container EXIT
    
    if ! create_lxd_container "$distro"; then
        TEST_RESULTS+=("$distro-$test_name: SKIPPED (not available)")
        return 0
    fi

    if run_installation_script "$(build_installation_command "$email" "$password" "$api_domain" "$app_domain" "$env")"; then
        test_result="PASSED"
    else
        test_result="FAILED"
    fi
    
    cleanup_container
    
    # Remove the trap for this test
    trap - EXIT
    
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
