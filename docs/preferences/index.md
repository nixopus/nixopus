# Preferences

Nixopus provides extensive customization through its settings panel. Access settings from the user menu in the sidebar or navigate to any settings page directly.

## General Settings

### Account

Update your profile information:
- **Username**: Change your display name
- **Email**: View your email address (verification status shown)
- **Avatar**: Upload a custom profile picture

::: details Appearance
**Font Family**: Outfit (default), Geist, Inter, Roboto, Poppins, Montserrat, Space Grotesk, Jakarta, System

**Auto Update**: Toggle automatic updates to keep Nixopus current with the latest features and security patches.
:::

::: details Language
Supported languages: English (default), Spanish, French, Kannada, Malayalam

Want to help translate Nixopus? We welcome contributions from native speakers to make the experience more authentic.
:::

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl + J` | Toggle terminal |
| `Ctrl + T` | Change terminal position |
| `Ctrl + B` | Toggle sidebar |
| `Ctrl + C` | Copy file |
| `Ctrl + X` | Cut file |
| `Ctrl + V` | Paste file |
| `Ctrl + H` | Toggle hidden files |
| `Ctrl + L` | Toggle layout (grid/list) |
| `Ctrl + Shift + N` | Create new folder |
| `F2` | Rename file |

::: tip macOS Users
Replace `Ctrl` with `Cmd` for all shortcuts.
:::

## Terminal Settings

Customize the web terminal appearance and behavior:

| Setting | Range | Default |
|---------|-------|---------|
| Font Size | 8 to 24px | 13px |
| Font Weight | Normal, Bold | Normal |
| Line Height | 1.0 to 2.5 | 1.4 |
| Letter Spacing | 0 to 2px | 0 |
| Cursor Style | Bar, Block, Underline | Bar |
| Cursor Blink | On/Off | On |
| Cursor Width | 1 to 5px | 2 |
| Scrollback | 1,000 to 50,000 lines | 5,000 |
| Tab Stop Width | 2 to 8 spaces | 4 |

**Font Families**: JetBrains Mono (default), Fira Code, Cascadia Code, SF Mono, Menlo, Monaco, Courier New

## Network Settings

Configure connection behavior for real time features:

**WebSocket**
- **Reconnect Attempts**: Number of retry attempts (1 to 20, default: 5)
- **Reconnect Interval**: Time between retries in milliseconds (1,000 to 30,000, default: 3,000)

**API**
- **Retry Attempts**: Number of API retry attempts (0 to 5, default: 1)
- **Disable Cache**: Turn off API response caching for debugging

## Container Settings

Configure default behavior for container operations:

| Setting | Range | Default |
|---------|-------|---------|
| Log Tail Lines | 50 to 10,000 | 100 |
| Stop Timeout | 1 to 300 seconds | 10 |
| Auto Prune Dangling Images | On/Off | Off |
| Auto Prune Build Cache | On/Off | Off |

**Restart Policy Options**

| Policy | Behavior |
|--------|----------|
| `no` | Never restart |
| `always` | Always restart regardless of exit status |
| `on-failure` | Restart only on non-zero exit status |
| `unless-stopped` | Restart unless explicitly stopped (default) |

## Troubleshooting Settings

Enable diagnostic features when debugging issues:

- **Debug Mode**: Enable verbose logging in the browser console
- **Show API Error Details**: Display detailed error information from API responses

::: warning Development Use Only
These options are intended for troubleshooting. Disable them during normal use for better performance.
:::

## Organization Settings

These settings are scoped to your current organization and affect all team members.

::: details Teams
Manage your organization and team members:
- Edit organization name and description
- Invite new members via email
- Update member roles (Admin, Developer, Viewer)
- Remove members from the organization
- View team statistics and recent activity
:::

::: details Domains
Register and manage custom domains for your deployments:
- Add new domains
- Configure domain type (primary, alias)
- View domain verification status
- Remove domains
:::

::: details Feature Flags
Control which features are enabled for your organization. Toggle individual features on or off based on your needs.

Feature flag management requires administrator privileges.
:::

## Resetting to Defaults

Each settings section displays a **Reset to Defaults** button when you have made changes. Click this to restore all options in that section to their original values.
