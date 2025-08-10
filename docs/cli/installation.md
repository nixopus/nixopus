# CLI Installation

Installation guide for the Nixopus CLI with multiple installation options.

## Prerequisites

- **Python 3.9 or higher** (supports up to Python 3.13)
- **Git** (for source installation)

Verify your Python version:
```bash
python3 --version
```

## Installation Options

### Option 1: Binary Installation (Recommended)

Download and install the pre-built binary for your platform:

```bash
sudo bash -c "$(curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install-cli.sh)"
```

**Install script options:**
- `--local`: Install to `~/.local/bin` (no sudo required)
- `--dir DIR`: Install to custom directory
- `--no-path`: Don't update PATH automatically

**Manual binary installation:**
```bash
# Download the appropriate binary for your platform
wget https://github.com/raghavyuva/nixopus/releases/latest/download/nixopus_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m)

# Make executable and install
chmod +x nixopus_*
sudo mv nixopus_* /usr/local/bin/nixopus

# Or install locally without sudo
mkdir -p ~/.local/bin
mv nixopus_* ~/.local/bin/nixopus
```

### Option 2: Poetry Installation (For Development)

Using Poetry for development work:

```bash
# Clone repository
git clone https://github.com/raghavyuva/nixopus.git
cd nixopus/cli

# Install with Poetry
poetry install

# Activate virtual environment
poetry shell

# Verify installation
nixopus --help
```

### Option 3: Python Package Installation

Install from source using pip:

```bash
# Clone repository
git clone https://github.com/raghavyuva/nixopus.git
cd nixopus/cli

# Install in development mode
pip install -e .

# Or install from wheel (if available)
pip install dist/nixopus-0.1.0-py3-none-any.whl
```

### Option 4: Build from Source

Build your own binary:

```bash
# Clone repository
git clone https://github.com/raghavyuva/nixopus.git
cd nixopus/cli

# Install Poetry dependencies
poetry install --with dev

# Build binary
./build.sh

# Install the built binary
./install.sh --local
```

## Verification

After installation, verify the CLI is working:

```bash
nixopus --help
nixopus version
```

Expected output:
```
┌───────────────── Version Info ─────────────────┐
│ Nixopus CLI v0.1.0                            │
└─────────────────────────────────────────────────┘
```

## Next Steps: Installing Nixopus

Once the CLI is installed, you can use it to install Nixopus on your VPS:

```bash
# Basic installation
nixopus install

# Installation with custom domains
nixopus install \
  --api-domain api.example.com \
  --view-domain app.example.com \
  --verbose

# Check system requirements first
nixopus preflight

# Install only dependencies
nixopus install deps
```

For detailed installation options, see the [Installation Guide](../install/index.md).

## Troubleshooting

### Command Not Found

If `nixopus` command is not found:

```bash
# Check if binary is in PATH
which nixopus

# For local installation, add to PATH
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Or for zsh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### Permission Errors

For permission issues during installation:

```bash
# Use local installation
curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/cli/install.sh | bash -s -- --local

# Or install to custom directory
curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/cli/install.sh | bash -s -- --dir ~/bin
```

### Python Version Issues

For Python version compatibility issues:

```bash
# Check Python version
python3 --version

# Install specific Python version if needed (example for Ubuntu)
sudo apt update
sudo apt install python3.9

# Or use pyenv for version management
curl https://pyenv.run | bash
pyenv install 3.9.0
pyenv local 3.9.0
```

## Development Installation

For contributing to the CLI:

```bash
# Clone and setup
git clone https://github.com/raghavyuva/nixopus.git
cd nixopus/cli

# Install with development dependencies
poetry install --with dev

# Activate environment
poetry shell

# Run tests to verify setup
make test
```

Available development commands:
```bash
make help          # Show available commands
make test          # Run test suite
make test-cov      # Run tests with coverage
make build         # Build binary
make format        # Format code
make lint          # Run linting
make clean         # Clean build artifacts
```

## Uninstallation

To uninstall the CLI:

```bash
# For binary installation
sudo rm /usr/local/bin/nixopus
# Or for local installation
rm ~/.local/bin/nixopus

# For Poetry installation
cd nixopus/cli
poetry env remove python

# For pip installation
pip uninstall nixopus
```