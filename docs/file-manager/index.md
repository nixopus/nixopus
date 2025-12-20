# File Manager

The file manager in Nixopus is unique and is inspired by many file managers on the internet, with added customizability. Whether you're coming from Windows, macOS, or Linux, you'll find our file manager familiar yet refreshingly different.

## Working with Files

Managing your files is a breeze with our intuitive interface. You can copy files using `CTRL + C` (or `CMD + C` on Mac), cut them with `CTRL + X` (or `CMD + X`), and paste them using `CTRL + V` (or `CMD + V`). Need to move files? Just drag and drop them into the folder you want - it's that simple!

Want to rename a file? Just double-click on its name and type away. When you're done, click anywhere else to save your changes. Looking for a specific file? You can sort your files by clicking on any column header - we support sorting by size, name, type, and date.

## View Options

We know everyone has their preferred way of viewing files. That's why we give you the flexibility to switch between grid and list views with just a click of the layout button in the top right corner. Need to see hidden files? Just click the three dots menu and select "Show Hidden Files" - it's all there when you need it.

## File Information

Curious about a file's details? Right-click on any file and select "Get Info" from the menu to see everything you need to know.
Need to organize your files? Creating a new folder is as easy as clicking the three dots menu and selecting "New Folder". It's these little touches that make working with files feel natural and effortless.

## File Operations

### Downloading Files and Folders

Downloading your files is straightforward:

* **Single File Download**: Right-click on any file and select "Download" to save it to your computer
* **Folder Download**: Right-click on any folder and select "Download as ZIP" to download the entire folder and its contents as a compressed ZIP file

### Uploading Files and Folders

You can upload both individual files and entire folders:

* **File Upload**: Click the upload button or drag and drop files into the file manager
* **Folder Upload**: Click the folder upload button to select and upload entire directory structures while preserving the folder hierarchy

### Advanced Upload Features

* **Batch Upload**: Upload multiple files at once
* **Path Preservation**: When uploading folders, the directory structure is maintained on the server
* **Progress Tracking**: Monitor upload progress for large files and folders

## API Reference

### Download Endpoints

* `GET /v1/file-manager/download?path=/path/to/file` - Download a single file
* `GET /v1/file-manager/download-folder?path=/path/to/folder` - Download a folder as ZIP

### Upload Endpoint

* `POST /v1/file-manager/upload` - Upload files
  * Form fields:
    * `path`: Base path where files should be uploaded
    * `files`: Array of files to upload
    * `relativePaths`: Array of relative paths for each file (optional, defaults to filename)

## What's Coming Next

We're always working to make the file manager even better. Here's what we're planning to add:

* A tree view for easier navigation through your files
* Enhanced privacy features to keep your files secure
* Persistent file positioning, similar to Mac Finder
* Custom icon sets to personalize your experience
* Built-in support for zip and tarball files
* File synchronization options for both local and remote servers
* Easy file sharing and downloading capabilities ✅ **Implemented**
* Folder upload with structure preservation ✅ **Implemented**
