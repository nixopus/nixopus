'use client';
import React from 'react';
import { TrashIcon, CheckIcon } from 'lucide-react';
import Skeleton from './components/skeleton/Skeleton';
import Header from './components/layout/Header';
import { SearchBar } from '@/components/ui/search-bar';
import { Breadcrumbs } from './components/breadcrumbs/Breadcrumbs';
import SortMethods from './components/actions/SortMethods';
import LayoutSwitcher from './components/layout/LayoutSwitcher';
import Actions from './components/actions/Actions';
import { FileItem } from './components/file-list/FileItem';
import useFileManager from './hooks/ui/useFileManager';
import { useTranslation } from '@/hooks/use-translation';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { FileData } from '@/redux/types/files';
import { FileContextMenu } from './components/context-menu/FileContextMenu';
import { toast } from 'sonner';
import { useFeatureFlags } from '@/hooks/features_provider';
import DisabledFeature from '@/components/features/disabled-feature';
import { FeatureNames } from '@/types/feature-flags';
import { ResourceGuard, AnyPermissionGuard } from '@/components/rbac/PermissionGuard';

function FileManager() {
  const { t } = useTranslation();
  const [fileToDelete, setFileToDelete] = React.useState<FileData | null>(null);
  const { isFeatureEnabled, isLoading: isFeatureFlagsLoading } = useFeatureFlags();

  const {
    currentPath,
    layout,
    showHidden,
    selectedPath,
    fileToCopy,
    fileToMove,
    isCopyFileOrDirectoryLoading,
    showCopyFeedback,
    copyFeedbackMessage,
    handleFileSelect,
    handleCopy,
    handleMove,
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
    files,
    handleFileUpload,
    handleDelete,
    handleKeyDown,
    handleTextDoubleClick,
    handleRename,
    startRenaming,
    isEditing,
    editedFileName,
    setEditedFileName,
    isDialogOpen,
    setIsDialogOpen
  } = useFileManager();

  if (isFeatureFlagsLoading) {
    return <Skeleton />;
  }

  if (!isFeatureEnabled(FeatureNames.FeatureFileManager)) {
    return <DisabledFeature />;
  }

  if (isLoading || isCopyFileOrDirectoryLoading) {
    return <Skeleton />;
  }

  return (
    <ResourceGuard 
      resource="file-manager" 
      action="read"
      loadingFallback={<Skeleton />}
    >
      <FileContextMenu
        showHidden={showHidden}
        setShowHidden={setShowHidden}
        fileToMove={fileToMove}
        fileToCopy={fileToCopy}
        currentPath={currentPath}
        onCreateFolder={createNewFolder}
        onFileUpload={(e) => handleFileUpload(e, currentPath)}
        onMove={() => {
          if (fileToMove) {
            const destinationPath = `${currentPath}/${fileToMove.name}`;
            if (!files?.some((f) => f.path === destinationPath)) {
              handleMove(fileToMove.path, destinationPath);
              setFileToMove(undefined);
            } else {
              toast.error(t('fileManager.errors.fileExists'));
            }
          }
        }}
        onPaste={() => {
          if (fileToCopy) {
            handleCopy(fileToCopy.path, `${currentPath}/${fileToCopy.name}`);
            setFileToCopy(undefined);
          }
        }}
      >
        <div onDragOver={(e) => e.preventDefault()} className="min-h-[calc(100vh-100px)]">
          <div className="mx-auto max-w-7xl">
            {showCopyFeedback && (
              <div className="fixed bottom-4 right-4 flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-primary-foreground shadow-lg animate-in fade-in slide-in-from-bottom-4">
                <CheckIcon className="h-4 w-4" />
                <span>{copyFeedbackMessage}</span>
              </div>
            )}
            <div className="mb-6 flex flex-col gap-4 px-0 lg:px-6">
              <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <Header />
                <div className="w-full sm:w-auto">
                  <SearchBar
                    searchTerm={searchTerm}
                    handleSearchChange={handleSearchChange}
                    label={t('common.searchFiles')}
                  />
                </div>
              </div>
              <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <Breadcrumbs
                  breadcrumbs={currentPath.split('/').filter(Boolean)}
                  fileClicked={fileClicked}
                />
                <div className="flex flex-wrap items-center gap-2 sm:gap-4">
                  <SortMethods files={visibleFiles} onSortChange={handleSortChange} />
                  <LayoutSwitcher layout={layout} setLayout={setLayout} />
                  <AnyPermissionGuard 
                    permissions={['file-manager:create', 'file-manager:update', 'file-manager:delete']}
                  >
                    <Actions
                      refetch={refetch}
                      showHidden={showHidden}
                      setShowHidden={setShowHidden}
                      currentPath={currentPath}
                      setSelectedPath={setSelectedPath}
                      selectedPath={selectedPath}
                      files={files || []}
                    />
                  </AnyPermissionGuard>
                </div>
              </div>
            </div>

            <div
              className={`grid w-full ${
                layout === 'grid'
                  ? 'grid-cols-1 gap-2 xs:grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8'
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
                  isEditing={isEditing && selectedPath === file.path}
                  editedFileName={editedFileName}
                  setEditedFileName={setEditedFileName}
                  isDialogOpen={isDialogOpen}
                  setIsDialogOpen={setIsDialogOpen}
                  isSizeLoading={false}
                  fileSize={0}
                  handleRename={handleRename}
                  handleKeyDown={handleKeyDown}
                  handleTextDoubleClick={handleTextDoubleClick}
                  handleDelete={handleDelete}
                  handleCopy={handleCopy}
                  startRenaming={startRenaming}
                />
              ))}
              {visibleFiles.length === 0 && (
                <div className="col-span-full text-center text-5xl text-muted-foreground">
                  {t('fileManager.noFiles')}
                </div>
              )}
            </div>
            <ResourceGuard 
              resource="file-manager" 
              action="delete"
              loadingFallback={null}
            >
              <DeleteDialog
                title={t('fileManager.deleteDialog.title')}
                description={
                  fileToDelete?.file_type === 'Directory'
                    ? t('fileManager.deleteDialog.descriptionDirectory', {
                        name: fileToDelete.name || ''
                      })
                    : t('fileManager.deleteDialog.descriptionFile', { name: fileToDelete?.name || '' })
                }
                onConfirm={() => {
                  if (fileToDelete) {
                    handleDelete(fileToDelete.path);
                    setFileToDelete(null);
                  }
                }}
                open={!!fileToDelete}
                onOpenChange={(open) => !open && setFileToDelete(null)}
                variant="destructive"
                confirmText={t('fileManager.deleteDialog.confirm')}
                cancelText={t('fileManager.deleteDialog.cancel')}
                icon={TrashIcon}
              />
            </ResourceGuard>
          </div>
        </div>
      </FileContextMenu>
    </ResourceGuard>
  );
}

export default FileManager;
