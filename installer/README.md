# Nixopus Self-Host Installer

One-line installer for self-hosting Nixopus.

```bash
curl -fsSL https://raw.githubusercontent.com/raghavyuva/nixopus/main/installer/get.sh | sudo bash
```

## Contents

- `get.sh` - Installer script
- `nixopus.sh` - Management CLI (installed to `/usr/local/bin/nixopus`)
- `selfhost/` - Docker Compose files (base, db, redis overlays)
- `test/` - Cross-distro test suite
