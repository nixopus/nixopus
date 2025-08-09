#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

APP_NAME="nixopus"
INSTALL_DIR="/usr/local/bin"
BUILD_DIR="dist"

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --local       Install to ~/.local/bin instead of /usr/local/bin"
    echo "  --dir DIR     Install to custom directory"
    echo "  --no-path     Don't automatically update PATH in shell profile"
    echo "  --help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                    # Install to /usr/local/bin (requires sudo)"
    echo "  $0 --local           # Install to ~/.local/bin (no sudo required)"
    echo "  $0 --dir ~/bin       # Install to custom directory"
    echo "  $0 --local --no-path # Install locally but don't update PATH"
}

detect_binary() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
    esac
    
    BINARY_NAME="${APP_NAME}_${OS}_${ARCH}"
    
    if [[ "$OS" == "mingw"* || "$OS" == "cygwin"* || "$OS" == "msys"* ]]; then
        BINARY_NAME="${BINARY_NAME}.exe"
    fi
    
    BINARY_PATH="$BUILD_DIR/$BINARY_NAME"
    
    if [[ ! -f "$BINARY_PATH" ]]; then
        log_error "Binary not found: $BINARY_PATH"
        log_info "Please run './build.sh' first to build the binary"
        exit 1
    fi
    
    log_info "Found binary: $BINARY_PATH"
}

install_binary() {
    log_info "Installing $APP_NAME to $INSTALL_DIR..."
    
    if [[ ! -d "$INSTALL_DIR" ]]; then
        log_info "Creating directory: $INSTALL_DIR"
        mkdir -p "$INSTALL_DIR"
    fi
    
    if [[ "$INSTALL_DIR" == "/usr/local/bin" ]] && [[ $EUID -ne 0 ]]; then
        log_info "Installing to system directory requires sudo..."
        sudo cp "$BINARY_PATH" "$INSTALL_DIR/$APP_NAME"
        sudo chmod +x "$INSTALL_DIR/$APP_NAME"
    else
        cp "$BINARY_PATH" "$INSTALL_DIR/$APP_NAME"
        chmod +x "$INSTALL_DIR/$APP_NAME"
    fi
    
    log_success "$APP_NAME installed to $INSTALL_DIR/$APP_NAME"
}

update_shell_profile() {
    shell_profile=""
    local current_shell=$(basename "$SHELL")
    
    case $current_shell in
        bash)
            if [[ -f "$HOME/.bash_profile" ]]; then
                shell_profile="$HOME/.bash_profile"
            elif [[ -f "$HOME/.bashrc" ]]; then
                shell_profile="$HOME/.bashrc"
            else
                shell_profile="$HOME/.bash_profile"
            fi
            ;;
        zsh)
            shell_profile="$HOME/.zshrc"
            ;;
        fish)
            shell_profile="$HOME/.config/fish/config.fish"
            ;;
        *)
            shell_profile="$HOME/.profile"
            ;;
    esac
    
    log_info "Detected shell: $current_shell"
    log_info "Using profile: $shell_profile"
    
    return 0
}

add_to_path() {
    if [[ ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
        log_success "$INSTALL_DIR is already in your PATH"
        return 0
    fi
    
    update_shell_profile
    local shell_profile_used=$shell_profile
    
    mkdir -p "$(dirname "$shell_profile_used")"
    
    if [[ -f "$shell_profile_used" ]] && grep -q "export PATH.*$INSTALL_DIR" "$shell_profile_used"; then
        log_info "PATH entry already exists in $shell_profile_used"
        return 0
    fi
    
    log_info "Adding $INSTALL_DIR to PATH in $shell_profile_used..."
    
    {
        echo ""
        echo "# Added by nixopus installer"
        echo "export PATH=\"$INSTALL_DIR:\$PATH\""
    } >> "$shell_profile_used"
    
    log_success "Added $INSTALL_DIR to PATH in $shell_profile_used"
    
    log_info "Updating PATH for current session..."
    export PATH="$INSTALL_DIR:$PATH"
    log_success "PATH updated for current session"
        
    if [[ -f "$shell_profile_used" ]]; then
        log_info "Sourcing $shell_profile_used for future sessions..."
        source "$shell_profile_used" 2>/dev/null || true
    fi
    
    return 0
}

test_installation() {
    log_info "Testing installation..."
    
    if command -v "$APP_NAME" &> /dev/null; then
        if "$APP_NAME" --version; then
            log_success "Installation test passed!"
            echo ""
            log_info "You can now use: $APP_NAME --help"
            log_info "The command is available in new shell sessions or by running:"
            log_info "  export PATH=\"$INSTALL_DIR:\$PATH\" && $APP_NAME --help"
        else
            log_error "Installation test failed - binary exists but doesn't work"
            exit 1
        fi
    else
        log_warning "Command '$APP_NAME' not found in PATH"
        log_info "You may need to restart your shell or update your PATH"
        log_info "You can run directly: $INSTALL_DIR/$APP_NAME --help"
    fi
}

main() {
    local use_local=false
    local custom_dir=""
    local skip_path=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --local)
                use_local=true
                shift
                ;;
            --dir)
                custom_dir="$2"
                shift 2
                ;;
            --no-path)
                skip_path=true
                shift
                ;;
            --help)
                show_usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    if [[ -n "$custom_dir" ]]; then
        INSTALL_DIR="$custom_dir"
    elif [[ "$use_local" == true ]]; then
        INSTALL_DIR="$HOME/.local/bin"
    fi
    
    log_info "Starting $APP_NAME installation..."
    log_info "Target directory: $INSTALL_DIR"
    
    detect_binary
    install_binary
    
    if [[ "$skip_path" == false ]]; then
        add_to_path
    else
        log_info "Skipping PATH update (--no-path specified)"
        if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
            log_warning "$INSTALL_DIR is not in your PATH"
            log_info "You can run: $INSTALL_DIR/$APP_NAME --help"
        fi
    fi
    
    test_installation
    
    log_success "Installation completed!"
    echo ""
    log_info "To use nixopus immediately in this session:"
    echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
    echo "  nixopus --help"
    echo ""
    log_info "Or open a new shell session and run: nixopus --help"
}

main "$@"