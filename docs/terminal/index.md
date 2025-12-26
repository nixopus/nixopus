# Terminal

Access your VPS directly from the Nixopus dashboard with a secure, browser based terminal. Execute commands, manage files, and troubleshoot deployments without switching to an external SSH client.

## Opening the Terminal

Press `Ctrl + J` to toggle the terminal, or click the terminal icon in the dashboard. Press `Ctrl + T` to switch between bottom and side positions.

## Sessions

Create up to **5 independent sessions**, each with its own shell connection. Click the **+** button to open a new tab, click any tab to switch, or click **×** to close a session.

The tab indicator shows session status: idle, loading, or active.

## Split Panes

Divide a session into up to **4 side by side panes**. Click the split icon in the header to add a pane, click any pane to focus it, or click **×** on the pane header to close it.

## Keyboard Shortcuts

```
Ctrl + J     Toggle terminal visibility
Ctrl + T     Switch terminal position
Ctrl + C     Interrupt running command
Ctrl + D     Close shell (if empty)
Ctrl + L     Clear screen
```

::: tip macOS
Replace `Ctrl` with `Cmd` for all shortcuts.
:::

::: warning
When focused, the terminal captures all keyboard input. `Ctrl + C` sends SIGINT instead of copying.
:::

## Resizing

Drag the top edge of the terminal panel to adjust height.

## Customization

Configure the terminal in **Settings > Advanced Settings**:

- **Text**: Font family, size, weight, letter spacing, line height
- **Cursor**: Style (bar/block/underline), blink, width
- **Behavior**: Scrollback lines, tab stop width

The terminal inherits your theme from **Settings > General**.

## How It Works

Commands travel from your browser through WebSocket to the Nixopus API, which establishes a secure SSH connection to your VPS. All connections are encrypted and authenticated.

```
Browser (XTerm.js) → WebSocket → Nixopus API → SSH → Your VPS
```
