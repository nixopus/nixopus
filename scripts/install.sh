#!/bin/bash

set -euo pipefail

# check if the script is running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: Please run as root (sudo)" >&2
    exit 1
fi

# check if the required commands are installed
function check_command() {
    local cmd="$1"
    if ! command -v "$cmd" &>/dev/null; then
        echo "Error: '$cmd' is not installed. Please install '$cmd' before running this script." >&2
        exit 1
    fi
}

# check if the required python version is installed
function check_python_version() {
    if ! python3 --version | grep -q "Python 3.10"; then
        echo "Error: Python 3.10 is not installed. Please install Python 3.10 before running this script." >&2
        exit 1
    fi
}

# check if the required dependencies are installed
function check_dependencies() {
    echo "Checking dependencies..."
    check_command "python3"
    check_command "pip3"
    check_command "git"
    # check_python_version
}

function parse_command_line_arguments() {
    local env="production"
    for arg in "$@"; do
        case "$arg" in
            --env=*)
                env="${arg#*=}"
                ;;
        esac
    done
    echo "$env"
}

function setup_config_based_on_environment() {
    local env="$1"
    echo "Setting up environment: $env"
    if [ "$env" == "staging" ]; then
        NIXOPUS_DIR="/etc/nixopus-staging"
        SOURCE_DIR="$NIXOPUS_DIR/source"
        BRANCH="feat/develop"
    else
        NIXOPUS_DIR="/etc/nixopus"
        SOURCE_DIR="$NIXOPUS_DIR/source"
        BRANCH="feat/develop" # TODO: change to master
    fi
}

function create_nixopus_directories() {
    echo "Creating Nixopus directories..."
    mkdir -p "${NIXOPUS_DIR:?}"
    mkdir -p "${SOURCE_DIR:?}"
}

function clone_nixopus_repository() {
    echo "Cloning Nixopus repository..."
    if [ -d "${SOURCE_DIR:?}/.git" ]; then
        cd "${SOURCE_DIR:?}" || exit 1
        git fetch --all
        git reset --hard "origin/${BRANCH:?}"
        git checkout "${BRANCH:?}"
        git pull
    else
        rm -rf "${SOURCE_DIR:?}"/* "${SOURCE_DIR:?}"/.[!.]*
        git clone https://github.com/raghavyuva/nixopus.git "${SOURCE_DIR:?}"
        cd "${SOURCE_DIR:?}" || exit 1
        git checkout "${BRANCH:?}"
    fi
}

function setup_caddy_configuration() {
    echo "Setting up Caddy configuration..."
    rm -rf "${NIXOPUS_DIR:?}/caddy"
    mkdir -p "${NIXOPUS_DIR:?}/caddy"
    echo '{
        admin 0.0.0.0:2019
        log {
            format json
            level INFO
        }
    }' > "${NIXOPUS_DIR:?}/caddy/Caddyfile"
}

function setup_nixopus_installation_environment() {
    echo "Setting up Python virtual environment..."
    cd "${SOURCE_DIR:?}/installer" || exit 1
    python3 -m venv venv
    source venv/bin/activate
    echo "Upgrading pip..."
    pip install --upgrade pip
    echo "Installing requirements..."
    pip install -r requirements.txt
}

function run_installer() {
    echo "Running installer..."
    PYTHONPATH="${SOURCE_DIR:?}/installer" python3 install.py "$@"
}

function deactivate_virtual_environment() {
    echo "Deactivating virtual environment..."
    deactivate
}

function main() {
    echo "Starting Nixopus installation..."
    check_dependencies
    ENV=$(parse_command_line_arguments "$@")
    setup_config_based_on_environment "$ENV"
    create_nixopus_directories
    clone_nixopus_repository
    setup_caddy_configuration
    setup_nixopus_installation_environment
    run_installer "$@"
    deactivate_virtual_environment
    echo "Installation completed successfully!"
}

main "$@"