'use client';
import React from 'react';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger
} from '@/components/ui/context-menu';
import {
  Copy,
  EyeIcon,
  EyeOffIcon,
  FolderPlusIcon,
  Move,
  TrashIcon,
  UploadCloudIcon
} from 'lucide-react';
import FileManagerSkeleton from './components/FileManagerSkeleton';
import FileManagerHeader from './components/FileManagerHeader';
import { SearchBar } from '@/components/search-bar';
import { FileBreadCrumbs } from './components/FileBreadCrumbs';
import FileSortMethods from './components/FileSortMethods';
import FileManagerLayouts from './components/FileManagerLayoutSwitcher';
import FileManagerActions from './components/FileManagerActions';
import { FileItem } from './components/FileItem';
import use_file_manager from './hooks/use_file_manager';

function FileManager() {
  const {
    currentPath,
    layout,
    showHidden,
    selectedPath,
    fileToCopy,
    fileToMove,
    isCopyFileOrDirectoryLoading,
    isMoveOrRenameDirectoryLoading,
    handleFileSelect,
    handleFilePaste,
    handleFileMove,
    handleSearchChange,
    handleSortChange,
    visibleFiles,
    createNewFolder,
    fileClicked,
    isLoading,
    refetch,
    searchTerm,
    setLayout,
    setShowHidden,
    setFileToCopy,
    setFileToMove,
    setSelectedPath,
    files
  } = use_file_manager();

  if (isLoading || isCopyFileOrDirectoryLoading || isMoveOrRenameDirectoryLoading) {
    return <FileManagerSkeleton />;
  }

  return (
    <div>
      <div className="mx-auto max-w-7xl">
        <div className="mb-6 flex items-center justify-between px-0 lg:px-6">
          <FileManagerHeader />
          <div className="relative">
            <SearchBar
              searchTerm={searchTerm}
              handleSearchChange={handleSearchChange}
              label="Search files..."
            />
          </div>
        </div>
        <div className="mb-6 flex items-center justify-between px-0 lg:px-6">
          <FileBreadCrumbs
            breadcrumbs={currentPath.split('/').filter(Boolean)}
            fileClicked={fileClicked}
          />
          <div className="flex items-center gap-4">
            <FileSortMethods files={visibleFiles} onSortChange={handleSortChange} />
            <FileManagerLayouts layout={layout} setLayout={setLayout} />
            <FileManagerActions
              refetch={refetch}
              showHidden={showHidden}
              setShowHidden={setShowHidden}
              currentPath={currentPath}
              setSelectedPath={setSelectedPath}
              selectedPath={selectedPath}
              files={files || []}
            />
          </div>
        </div>

        <ContextMenu>
          <ContextMenuTrigger>
            <div
              className={`grid w-full ${
                layout === 'grid'
                  ? 'grid-cols-2 gap-2 sm:grid-cols-3 sm:gap-3 md:grid-cols-4 md:gap-4 lg:grid-cols-6 xl:grid-cols-8'
                  : 'grid-cols-1 gap-2'
              }`}
            >
              {visibleFiles.map((file, index) => (
                <FileItem
                  key={file.path}
                  file={file}
                  onFolderClick={fileClicked}
                  type={file.file_type === 'Directory' ? 'folder' : 'file'}
                  layout={layout}
                  activePath={selectedPath}
                  onFolderClickActive={handleFileSelect}
                  refetch={refetch}
                  setFileToCopy={setFileToCopy}
                  setFileToMove={setFileToMove}
                  index={index}
                />
              ))}
              {visibleFiles.length === 0 && (
                <div className="col-span-full text-center text-5xl text-muted-foreground">
                  No files found
                </div>
              )}
            </div>
          </ContextMenuTrigger>
          <ContextMenuContent>
            <ContextMenuItem>
              <UploadCloudIcon className="mr-2 h-5 w-5" />
              <span>New File</span>
            </ContextMenuItem>
            {fileToMove && (
              <ContextMenuItem
                onSelect={() => {
                  handleFileMove(fileToMove.path, currentPath + '/' + fileToMove.name);
                  refetch();
                  setFileToMove(undefined);
                }}
              >
                <Move className="mr-2 h-5 w-5" />
                <span>Move here</span>
              </ContextMenuItem>
            )}
            {fileToCopy && (
              <ContextMenuItem
                onSelect={() => {
                  handleFilePaste(fileToCopy.path, currentPath + '/' + fileToCopy.name);
                  refetch();
                  setFileToCopy(undefined);
                }}
              >
                <Copy className="mr-2 h-5 w-5" />
                <span>Paste here</span>
              </ContextMenuItem>
            )}
            <ContextMenuItem onSelect={createNewFolder}>
              <FolderPlusIcon className="mr-2 h-5 w-5" />
              <span>New Folder</span>
            </ContextMenuItem>
            <ContextMenuItem onSelect={() => setShowHidden(!showHidden)}>
              {showHidden ? (
                <EyeOffIcon className="mr-2 h-5 w-5" />
              ) : (
                <EyeIcon className="mr-2 h-5 w-5" />
              )}
              <span>{showHidden ? 'Hide Hidden Files' : 'Show Hidden Files'}</span>
            </ContextMenuItem>
            <ContextMenuItem>
              <TrashIcon className="mr-2 h-5 w-5" />
              <span>Trash</span>
            </ContextMenuItem>
          </ContextMenuContent>
        </ContextMenu>
      </div>
    </div>
  );
}

export default FileManager;
