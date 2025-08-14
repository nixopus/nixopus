#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

APP_NAME="nixopus"
BUILD_DIR="dist"
BINARY_DIR="binaries"
SPEC_FILE="nixopus.spec"

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

check_requirements() {
    log_info "Checking requirements..."
    
    if ! command -v poetry &> /dev/null; then
        log_error "Poetry is not installed. Please install Poetry first."
        exit 1
    fi
    
    if ! command -v python3 &> /dev/null; then
        log_error "Python3 is not installed."
        exit 1
    fi
    
    log_success "All requirements met"
}

setup_environment() {
    log_info "Setting up build environment..."
    
    if ! poetry check; then
        log_info "Updating poetry lock file..."
        poetry lock
    fi
    
    poetry install
    
    if ! poetry run python -c "import PyInstaller" &> /dev/null; then
        log_info "Installing PyInstaller..."
        poetry add --group dev pyinstaller
    fi
    
    log_success "Environment setup complete"
}

create_spec_file() {
    log_info "Creating PyInstaller spec file..."
    
    cat > $SPEC_FILE << 'EOF'
# -*- mode: python ; coding: utf-8 -*-

block_cipher = None

a = Analysis(
    ['app/main.py'],
    pathex=[],
    binaries=[],
    datas=[
        ('../helpers/config.prod.yaml', 'helpers/'),
    ],
    hiddenimports=[
        'app.commands.clone.command',
        'app.commands.conf.command',
        'app.commands.install.command',
        'app.commands.preflight.command',
        'app.commands.proxy.command',
        'app.commands.service.command',
        'app.commands.test.command',
        'app.commands.uninstall.command',
        'app.commands.version.command',
    ],
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=[],
    win_no_prefer_redirects=False,
    win_private_assemblies=False,
    cipher=block_cipher,
    noarchive=False,
)

pyz = PYZ(a.pure, a.zipped_data, cipher=block_cipher)

exe = EXE(
    pyz,
    a.scripts,
    [],
    exclude_binaries=True,
    name='nixopus',
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=True,
    upx_exclude=[],
    runtime_tmpdir=None,
    console=True,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
)

coll = COLLECT(
    exe,
    a.binaries,
    a.zipfiles,
    a.datas,
    strip=False,
    upx=True,
    upx_exclude=[],
    name='nixopus'
)
EOF
    
    log_success "Spec file created: $SPEC_FILE"
}

run_pyinstaller_build() {
    # Ensure spec file exists even if manually deleted
    if [[ ! -f "$SPEC_FILE" ]]; then
        log_warning "Spec file missing; regenerating..."
        create_spec_file
    fi

    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    if [[ "$OS" == "linux" ]] && command -v docker &> /dev/null && [[ -z "$NIXOPUS_DISABLE_DOCKER" ]]; then
        case $ARCH in
            x86_64)
                MANYLINUX_IMAGE="quay.io/pypa/manylinux2014_x86_64"
                PYTAG="cp311-cp311"
                ;;
            aarch64|arm64)
                MANYLINUX_IMAGE="quay.io/pypa/manylinux2014_aarch64"
                PYTAG="cp311-cp311"
                ;;
            *)
                MANYLINUX_IMAGE=""
                ;;
        esac

        if [[ -n "$MANYLINUX_IMAGE" ]]; then
            log_info "Building with PyInstaller inside $MANYLINUX_IMAGE for wide glibc compatibility..."
            docker run --rm -v "$(cd .. && pwd)":/work -w /work/cli "$MANYLINUX_IMAGE" bash -lc \
"export PATH=/opt/python/$PYTAG/bin:\$PATH && \
python3 -m pip install -U pip && \
python3 -m pip install 'poetry==1.8.3' && \
poetry install --with dev && \
poetry run pyinstaller --clean --noconfirm $SPEC_FILE" || {
                log_error "Dockerized build failed"
                exit 1
            }
            return
        fi

        log_warning "Unsupported arch $ARCH for manylinux; building on host (may require newer glibc)"
    fi

    log_info "Building with PyInstaller on host..."
    poetry run pyinstaller --clean --noconfirm $SPEC_FILE
}

build_wheel() {
    log_info "Building wheel package..."
    
    poetry build
    
    log_success "Wheel package built in $BUILD_DIR/"
}

build_binary() {
    log_info "Building binary"
    
    run_pyinstaller_build
    
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
    esac
    
    BINARY_DIR_NAME="${APP_NAME}_${OS}_${ARCH}"
    
    
    if [[ -d "$BUILD_DIR/$APP_NAME" ]]; then        
        mv "$BUILD_DIR/$APP_NAME" "$BUILD_DIR/$BINARY_DIR_NAME"

        
        cat > "$BUILD_DIR/$APP_NAME" << EOF
#!/bin/bash
# Nixopus CLI wrapper
SCRIPT_DIR="\$(cd "\$(dirname "\${BASH_SOURCE[0]}")" && pwd)"
exec "\$SCRIPT_DIR/$BINARY_DIR_NAME/$APP_NAME" "\$@"
EOF
        chmod +x "$BUILD_DIR/$APP_NAME"

        log_success "Binary directory built: $BUILD_DIR/$BINARY_DIR_NAME/"
        log_success "Wrapper script created: $BUILD_DIR/$APP_NAME"
    else
        log_error "Build failed - directory $BUILD_DIR/$APP_NAME not found"
        exit 1
    fi
}

test_binary() {
    
    log_info "Testing binary..."

    WRAPPER_PATH="$BUILD_DIR/$APP_NAME"
    
    if [[ -f "$WRAPPER_PATH" ]]; then
        chmod +x "$WRAPPER_PATH"
        
        if "$WRAPPER_PATH" --version; then
            log_success "Binary test passed"
        else
            log_error "Binary test failed"
            exit 1
        fi
    else
        log_error "Wrapper script not found for testing: $WRAPPER_PATH"
        exit 1
    fi
}

create_release_archive() {
    log_info "Creating release archive..."
    
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
    esac
    
    ARCHIVE_NAME="${APP_NAME}_${OS}_${ARCH}"
    BINARY_DIR_NAME="${APP_NAME}_${OS}_${ARCH}"
    
    cd $BUILD_DIR
    

    if [[ "$OS" == "darwin" || "$OS" == "linux" ]]; then
        tar -czf "${ARCHIVE_NAME}.tar.gz" "$BINARY_DIR_NAME" "$APP_NAME"
        log_success "Archive created: $BUILD_DIR/${ARCHIVE_NAME}.tar.gz"
    elif [[ "$OS" == "mingw"* || "$OS" == "cygwin"* || "$OS" == "msys"* ]]; then
        zip -r "${ARCHIVE_NAME}.zip" "$BINARY_DIR_NAME" "$APP_NAME"
        log_success "Archive created: $BUILD_DIR/${ARCHIVE_NAME}.zip"
    fi
    
    cd ..
}

cleanup() {
    log_info "Cleaning up temporary files..."
    rm -rf build/
    rm -f $SPEC_FILE
    log_success "Cleanup complete"
}

show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --no-test     Skip binary testing"
    echo "  --no-archive  Skip creating release archive"
    echo "  --no-cleanup  Skip cleanup of temporary files"
    echo "  --help        Show this help message"
    echo ""
    echo "Example:"
    echo "  $0                    # Full build with all steps"
    echo "  $0 --no-test         # Build without testing"
    echo "  $0 --no-archive      # Build without creating archive"
}

main() {
    local skip_test=false
    local skip_archive=false
    local skip_cleanup=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --no-test)
                skip_test=true
                shift
                ;;
            --no-archive)
                skip_archive=true
                shift
                ;;
            --no-cleanup)
                skip_cleanup=true
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
    
    log_info "Starting Nixopus CLI binary build process..."
    
    check_requirements
    setup_environment
    create_spec_file
    build_wheel
    build_binary
    
    if [[ $skip_test == false ]]; then
        test_binary
    fi
    
    if [[ $skip_archive == false ]]; then
        create_release_archive
    fi
    
    if [[ $skip_cleanup == false ]]; then
        cleanup
    fi
    
    log_success "Build process completed!"
    log_info "Binary location: $BUILD_DIR/"
    
    if [[ -d "$BUILD_DIR" ]]; then
        echo ""
        log_info "Built binaries:"
        ls -la $BUILD_DIR/
    fi
}

main "$@"