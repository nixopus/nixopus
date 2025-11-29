# CLI Installation

Installation guide for the Nixopus CLI with multiple installation options.

## Prerequisites

Before installing the Nixopus CLI, ensure you have:

::: info Prerequisites Checklist
- **Python** 3.9 or higher (supports up to Python 3.13)
- **Git** (required for source installation methods)
:::

Verify your Python version:

```bash
python3 --version
```

::: tip Python Version Check
The CLI supports Python 3.9 through 3.13. If you have multiple Python versions installed, make sure `python3` points to a compatible version.
:::

## Installation Options

Choose the installation method that best fits your needs:

| Method | Best For | Requires Sudo | Speed |
|--------|----------|---------------|-------|
| Binary Installation | Most users | Optional | ⚡ Fastest |
| Poetry Installation | Development | No | 🐢 Slower |
| Python Package | Python users | No | 🐢 Slower |
| Build from Source | Custom builds | Optional | 🐌 Slowest |

### Option 1: Binary Installation (Recommended) ⭐

The fastest and easiest way to install the CLI. Pre-built binaries are available for all major platforms.

::: code-group

```bash [Quick Install (with sudo)]
sudo bash -c "$(curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install.sh)"
```

```bash [Local Install (no sudo)]
bash -c "$(curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install.sh)" -- --local
```

```bash [Custom Directory]
bash -c "$(curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install.sh)" -- --dir ~/bin
```

:::

::: tip Install Script Options
The install script supports several options:

- `--local`: Install to `~/.local/bin` (no sudo required)
- `--dir DIR`: Install to custom directory
- `--no-path`: Don't update PATH automatically (you'll need to add it manually)

Use `--local` if you don't have sudo access or prefer user-local installation.
:::

::: details Manual Binary Installation
If you prefer to download and install manually:

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

**Note**: Make sure to add `~/.local/bin` to your PATH if installing locally.
:::

### Option 2: Poetry Installation (For Development)

Best for contributors and developers who want to work on the CLI codebase.

::: warning Poetry Required
This method requires Poetry to be installed. Install it with:
```bash
curl -sSL https://install.python-poetry.org | python3 -
```
:::

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

::: tip Poetry Virtual Environment
Poetry automatically creates and manages a virtual environment. Use `poetry shell` to activate it, or run commands with `poetry run nixopus`.
:::

### Option 3: Python Package Installation

Install from source using pip. Good for Python developers familiar with pip workflows.

```bash
# Clone repository
git clone https://github.com/raghavyuva/nixopus.git
cd nixopus/cli

# Install in development mode
pip install -e .

# Or install from wheel (if available)
pip install dist/nixopus-0.1.0-py3-none-any.whl
```

::: tip Virtual Environment
It's recommended to use a virtual environment when installing with pip:
```bash
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -e .
```
:::

### Option 4: Build from Source

Build your own binary from source. Useful for custom builds or when pre-built binaries aren't available for your platform.

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

::: info Build Requirements
Building from source requires:
- Poetry (for dependency management)
- All development dependencies
- Build tools for your platform
:::

## Verification

After installation, verify the CLI is working correctly:

```bash
nixopus --help
nixopus version
```

::: info Expected Output
You should see something like:

```
┌───────────────── Version Info ─────────────────┐
│ Nixopus CLI v0.1.0                            │
└─────────────────────────────────────────────────┘
```

If you see "command not found", see the [Troubleshooting](#troubleshooting) section below.
:::

## Next Steps: Installing Nixopus

Once the CLI is installed, you can use it to install Nixopus on your VPS:

::: code-group

```bash [Basic Installation]
sudo nixopus install
```

```bash [With Custom Domains]
sudo nixopus install \
  --api-domain api.example.com \
  --view-domain app.example.com \
  --verbose
```

```bash [Check Requirements First]
nixopus preflight
```

```bash [Install Dependencies Only]
sudo nixopus install deps
```

:::

::: warning Root Privileges Required
The `nixopus install` command requires root privileges to install system dependencies like Docker. Always use `sudo` unless you're already running as root. If you encounter permission errors or "exit status 100", ensure you're using sudo.
:::

::: tip Preflight Check
Always run `nixopus preflight` before installation to verify your system meets all requirements. This can save time by catching issues early.
:::

For detailed installation options and configuration, see the [Installation Guide](../install/index.md).

## Troubleshooting

Common issues and their solutions:

### Command Not Found

If `nixopus` command is not found after installation:

```bash
# Check if binary is in PATH
which nixopus
```

::: warning PATH Configuration
If the command is not found, the binary might not be in your PATH. Add it based on your installation method:

**For local installation (`~/.local/bin`):**

::: code-group

```bash [Bash]
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

```bash [Zsh]
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

```bash [Fish]
echo 'set -gx PATH $HOME/.local/bin $PATH' >> ~/.config/fish/config.fish
source ~/.config/fish/config.fish
```

:::

**For Poetry installation:**
- Make sure you've activated the Poetry virtual environment with `poetry shell`
- Or use `poetry run nixopus` instead
:::

### Permission Errors

If you encounter permission issues during installation:

::: tip Solutions
**Option 1: Use local installation (no sudo required)**
```bash
curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install.sh | bash -s -- --local
```

**Option 2: Install to custom directory**
```bash
curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install.sh | bash -s -- --dir ~/bin
```

**Option 3: Fix permissions (if already installed)**
```bash
sudo chmod +x /usr/local/bin/nixopus
```
:::

### Python Version Issues

If you encounter Python version compatibility issues:

```bash
# Check Python version
python3 --version
```

::: details Install Specific Python Version

**For Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install python3.9
```

**For macOS (using Homebrew):**
```bash
brew install python@3.9
```

**Using pyenv (recommended for version management):**
```bash
curl https://pyenv.run | bash
pyenv install 3.9.0
pyenv local 3.9.0
```

**For Windows:**
- Download from [python.org](https://www.python.org/downloads/)
- Or use the Microsoft Store
:::

::: warning Python Version Range
The CLI requires Python 3.9 or higher (up to 3.13). Make sure your Python version is within this range.
:::

## Development Installation

For contributing to the CLI development:

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

::: info Development Commands
Available development commands:

| Command | Description |
|---------|------------|
| `make help` | Show all available commands |
| `make test` | Run the test suite |
| `make test-cov` | Run tests with coverage report |
| `make build` | Build the binary |
| `make format` | Format code with black |
| `make lint` | Run linting checks |
| `make clean` | Clean build artifacts |
:::

::: tip Development Workflow
1. Make your changes
2. Run `make format` to format code
3. Run `make lint` to check for issues
4. Run `make test` to verify everything works
5. Build with `make build` to test the binary
:::

## Uninstallation

To remove the CLI from your system:

::: code-group

```bash [Binary Installation]
# System-wide installation
sudo rm /usr/local/bin/nixopus

# Local installation
rm ~/.local/bin/nixopus
```

```bash [Poetry Installation]
cd nixopus/cli
poetry env remove python
```

```bash [Pip Installation]
pip uninstall nixopus
```

:::

::: tip Clean Up
After uninstallation, you may also want to:
- Remove configuration files: `rm -rf ~/.config/nixopus`
- Remove cache: `rm -rf ~/.cache/nixopus`
- Remove Poetry virtual environment: `poetry env remove python` (if using Poetry)
:::