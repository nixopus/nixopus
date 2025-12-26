# File Manager

Browse and manage files on your VPS through a visual interface. Upload files, create directories, and organize your server without using command line tools.

## Navigation

Click any folder to open it. The breadcrumb trail at the top shows your current location and lets you jump back to any parent directory.

Switch between **Grid View** (icons) and **List View** (table with sortable columns) using the layout buttons in the header. Click any column header in list view to sort.

Right click anywhere and select **Show Hidden Files** to reveal files starting with a dot.

## File Operations

All operations are available through the right click context menu:

- **Upload**: Opens a dialog where you can drag files or click to browse. Multiple files supported.
- **New Folder**: Creates a folder named "New Folder". Double click to rename it.
- **Copy / Move**: Select an item, then right click in the destination and choose **Paste Here** or **Move Here**.
- **Rename**: Double click any file or folder name to edit. Press Enter to save, Escape to cancel.
- **Delete**: Removes the selected item. This cannot be undone.
- **Get Info**: Shows file details including name, path, type, size, and last modified date.

::: tip
You can also drag files directly from your computer into the file manager window to upload.
:::

## Keyboard Shortcuts

```
Ctrl + C     Copy selected file
Ctrl + X     Cut selected file
Ctrl + V     Paste file
Enter        Confirm rename
Escape       Cancel rename
```

::: tip macOS
Replace `Ctrl` with `Cmd` for all shortcuts.
:::

## Permissions

File manager actions respect your role permissions. If an action is not available in the context menu, you may not have the required permission:

- **Read**: View files and directories
- **Create**: Upload files and create folders
- **Update**: Rename, copy, and move items
- **Delete**: Remove files and directories
