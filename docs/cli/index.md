# Nixopus CLI

Command line interface for installing and managing Nixopus on your server.

## Quick Start

::: code-group

```bash [Install CLI]
sudo bash -c "$(curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install.sh)"
```

```bash [Install Nixopus]
sudo nixopus install
```

:::

## Commands

| Command | Description |
|---------|-------------|
| `preflight` | Check system requirements |
| `install` | Install Nixopus |
| `uninstall` | Remove Nixopus |
| `update` | Update Nixopus or CLI |
| `version` | Show CLI version |

::: tip Getting Help
Run `nixopus --help` or `nixopus <command> --help` for detailed usage information.
:::

## Installation Options

| Method | Best For | Requires Sudo |
|--------|----------|---------------|
| Binary Installation | Most users | Optional |
| Poetry Installation | Development | No |
| Build from Source | Custom builds | Optional |

### Binary Installation (Recommended)

::: code-group

```bash [System Install]
sudo bash -c "$(curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install.sh)"
```

```bash [Local Install]
bash -c "$(curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install.sh)" -- --local
```

```bash [Custom Directory]
bash -c "$(curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install.sh)" -- --dir ~/bin
```

:::

**Script Options:**
- `--local`: Install to `~/.local/bin` (no sudo required)
- `--dir DIR`: Install to custom directory
- `--no-path`: Don't update PATH automatically

::: details Manual Binary Installation
```bash
# Download binary for your platform
wget https://github.com/raghavyuva/nixopus/releases/latest/download/nixopus_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m)

# Make executable and install
chmod +x nixopus_*
sudo mv nixopus_* /usr/local/bin/nixopus
```
:::

::: details Poetry Installation (Development)
```bash
git clone https://github.com/raghavyuva/nixopus.git
cd nixopus/cli
poetry install
poetry shell
nixopus --help
```
:::

::: details Build from Source
```bash
git clone https://github.com/raghavyuva/nixopus.git
cd nixopus/cli
poetry install --with dev
./build.sh
./install.sh --local
```
:::

## Installing Nixopus

Once the CLI is installed, use it to install Nixopus:

::: code-group

```bash [Basic]
sudo nixopus install
```

```bash [With Domains]
sudo nixopus install \
  --api-domain api.example.com \
  --view-domain app.example.com
```

```bash [Preflight Check]
nixopus preflight
```

:::

::: warning Root Privileges Required
The `nixopus install` command requires root privileges to install system dependencies like Docker.
:::

## Verification

```bash
nixopus version
```

Expected output:
```
┌───────────────── Version Info ─────────────────┐
│ Nixopus CLI v0.1.0                            │
└─────────────────────────────────────────────────┘
```

## Troubleshooting

::: details Command Not Found
If `nixopus` is not found, add it to your PATH:

```bash
# For ~/.local/bin installation
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

For Zsh, use `~/.zshrc` instead.
:::

::: details Permission Errors
Use local installation if you don't have sudo access:
```bash
bash -c "$(curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/scripts/install.sh)" -- --local
```
:::

## Uninstallation

::: code-group

```bash [Binary]
sudo rm /usr/local/bin/nixopus
# Or for local: rm ~/.local/bin/nixopus
```

```bash [Poetry]
cd nixopus/cli && poetry env remove python
```

:::

## Reference

For complete command documentation with all options, see the [CLI Reference](./cli-reference.md).
