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
import type { translationKey } from '@/hooks/use-translation';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { FileData } from '@/redux/types/files';
import { FileContextMenu } from './components/context-menu/FileContextMenu';
import { toast } from 'sonner';
import { useFeatureFlags } from '@/hooks/features_provider';
import DisabledFeature from '@/components/features/disabled-feature';
import { FeatureNames } from '@/types/feature-flags';
import { ResourceGuard, AnyPermissionGuard } from '@/components/rbac/PermissionGuard';
import PageLayout from '@/components/layout/page-layout';
import { useIsMobile } from '@/hooks/use-mobile';

const CopyFeedback = ({ show, message }: { show: boolean; message: string }) => {
  if (!show) return null;

  return (
    <div className="fixed bottom-4 right-4 flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-primary-foreground shadow-lg animate-in fade-in slide-in-from-bottom-4">
      <CheckIcon className="h-4 w-4" />
      <span>{message}</span>
    </div>
  );
};

interface MobileNavigationProps {
  currentPath: string;
  fileClicked: (path: string) => void;
  visibleFiles: FileData[];
  handleSortChange: (method: any) => void;
  refetch: () => void;
  showHidden: boolean;
  setShowHidden: React.Dispatch<React.SetStateAction<boolean>>;
  setSelectedPath: React.Dispatch<React.SetStateAction<string>>;
  selectedPath: string;
  files: FileData[];
}

const MobileNavigation = ({
  currentPath,
  fileClicked,
  visibleFiles,
  handleSortChange,
  refetch,
  showHidden,
  setShowHidden,
  setSelectedPath,
  selectedPath,
  files
}: MobileNavigationProps) => (
  <div className="flex items-center justify-between">
    <button
      onClick={() => {
        const parentPath = currentPath.split('/').slice(0, -1).join('/') || '/';
        fileClicked(parentPath);
      }}
      className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors"
      disabled={currentPath === '/'}
    >
      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
      </svg>
      <span className="text-sm">Back</span>
    </button>
    <div className="flex items-center gap-2">
      <SortMethods files={visibleFiles} onSortChange={handleSortChange} />
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
          files={files}
        />
      </AnyPermissionGuard>
    </div>
  </div>
);

interface DesktopNavigationProps {
  currentPath: string;
  fileClicked: (path: string) => void;
  visibleFiles: FileData[];
  handleSortChange: (method: any) => void;
  layout: 'grid' | 'list';
  setLayout: React.Dispatch<React.SetStateAction<'grid' | 'list'>>;
  refetch: () => void;
  showHidden: boolean;
  setShowHidden: React.Dispatch<React.SetStateAction<boolean>>;
  setSelectedPath: React.Dispatch<React.SetStateAction<string>>;
  selectedPath: string;
  files: FileData[];
}

const DesktopNavigation = ({
  currentPath,
  fileClicked,
  visibleFiles,
  handleSortChange,
  layout,
  setLayout,
  refetch,
  showHidden,
  setShowHidden,
  setSelectedPath,
  selectedPath,
  files
}: DesktopNavigationProps) => (
  <>
    <Breadcrumbs breadcrumbs={currentPath.split('/').filter(Boolean)} fileClicked={fileClicked} />
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
          files={files}
        />
      </AnyPermissionGuard>
    </div>
  </>
);

interface FileListProps {
  layout: 'grid' | 'list' | 'mobile';
  isMobile: boolean;
  visibleFiles: FileData[];
  fileClicked: (path: string) => void;
  handleFileSelect: (path: string) => void;
  refetch: () => void;
  setFileToCopy: React.Dispatch<React.SetStateAction<FileData | undefined>>;
  setFileToMove: React.Dispatch<React.SetStateAction<FileData | undefined>>;
  isEditing: boolean;
  selectedPath: string;
  editedFileName: string;
  setEditedFileName: React.Dispatch<React.SetStateAction<string>>;
  isDialogOpen: boolean;
  setIsDialogOpen: React.Dispatch<React.SetStateAction<boolean>>;
  handleRename: (file: FileData) => void;
  handleKeyDown: (e: React.KeyboardEvent<HTMLInputElement>, file: FileData) => void;
  handleTextDoubleClick: () => void;
  handleDelete: (path: string) => void;
  handleCopy: (fromPath: string, toPath: string) => void;
  startRenaming: (file: FileData) => void;
  t: (key: translationKey, params?: Record<string, string>) => string;
}

const FileList = ({
  layout,
  isMobile,
  visibleFiles,
  fileClicked,
  handleFileSelect,
  refetch,
  setFileToCopy,
  setFileToMove,
  isEditing,
  selectedPath,
  editedFileName,
  setEditedFileName,
  isDialogOpen,
  setIsDialogOpen,
  handleRename,
  handleKeyDown,
  handleTextDoubleClick,
  handleDelete,
  handleCopy,
  startRenaming,
  t
}: FileListProps) => (
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
        layout={isMobile ? 'mobile' : layout}
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
);

interface FileDeleteDialogProps {
  fileToDelete: FileData | null;
  setFileToDelete: (file: FileData | null) => void;
  handleDelete: (path: string) => void;
  t: (key: translationKey, params?: Record<string, string>) => string;
}

const FileDeleteDialog = ({
  fileToDelete,
  setFileToDelete,
  handleDelete,
  t
}: FileDeleteDialogProps) => (
  <ResourceGuard resource="file-manager" action="delete" loadingFallback={null}>
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
);

function FileManager() {
  const isMobile = useIsMobile();
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

  const handleMoveFile = () => {
    if (fileToMove) {
      const destinationPath = `${currentPath}/${fileToMove.name}`;
      if (!files?.some((f) => f.path === destinationPath)) {
        handleMove(fileToMove.path, destinationPath);
        setFileToMove(undefined);
      } else {
        toast.error(t('fileManager.errors.fileExists'));
      }
    }
  };

  const handlePasteFile = () => {
    if (fileToCopy) {
      handleCopy(fileToCopy.path, `${currentPath}/${fileToCopy.name}`);
      setFileToCopy(undefined);
    }
  };

  return (
    <ResourceGuard resource="file-manager" action="read" loadingFallback={<Skeleton />}>
      <FileContextMenu
        showHidden={showHidden}
        setShowHidden={setShowHidden}
        fileToMove={fileToMove}
        fileToCopy={fileToCopy}
        currentPath={currentPath}
        onCreateFolder={createNewFolder}
        onFileUpload={(e) => handleFileUpload(e, currentPath)}
        onMove={handleMoveFile}
        onPaste={handlePasteFile}
      >
        <PageLayout maxWidth="6xl" padding="md" spacing="lg" className="min-h-[calc(100vh-100px)]">
          <CopyFeedback show={showCopyFeedback} message={copyFeedbackMessage} />

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
              {isMobile ? (
                <MobileNavigation
                  currentPath={currentPath}
                  fileClicked={fileClicked}
                  visibleFiles={visibleFiles}
                  handleSortChange={handleSortChange}
                  refetch={refetch}
                  showHidden={showHidden}
                  setShowHidden={setShowHidden}
                  setSelectedPath={setSelectedPath}
                  selectedPath={selectedPath}
                  files={files || []}
                />
              ) : (
                <DesktopNavigation
                  currentPath={currentPath}
                  fileClicked={fileClicked}
                  visibleFiles={visibleFiles}
                  handleSortChange={handleSortChange}
                  layout={layout}
                  setLayout={setLayout}
                  refetch={refetch}
                  showHidden={showHidden}
                  setShowHidden={setShowHidden}
                  setSelectedPath={setSelectedPath}
                  selectedPath={selectedPath}
                  files={files || []}
                />
              )}
            </div>
          </div>

          <FileList
            layout={layout}
            isMobile={isMobile}
            visibleFiles={visibleFiles}
            fileClicked={fileClicked}
            handleFileSelect={handleFileSelect}
            refetch={refetch}
            setFileToCopy={setFileToCopy}
            setFileToMove={setFileToMove}
            isEditing={isEditing}
            selectedPath={selectedPath}
            editedFileName={editedFileName}
            setEditedFileName={setEditedFileName}
            isDialogOpen={isDialogOpen}
            setIsDialogOpen={setIsDialogOpen}
            handleRename={handleRename}
            handleKeyDown={handleKeyDown}
            handleTextDoubleClick={handleTextDoubleClick}
            handleDelete={handleDelete}
            handleCopy={handleCopy}
            startRenaming={startRenaming}
            t={t}
          />
          <FileDeleteDialog
            fileToDelete={fileToDelete}
            setFileToDelete={setFileToDelete}
            handleDelete={handleDelete}
            t={t}
          />
        </PageLayout>
      </FileContextMenu>
    </ResourceGuard>
  );
}

export default FileManager;
