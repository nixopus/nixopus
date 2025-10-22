import React, { useState } from 'react';
import { DialogWrapper } from '@/components/ui/dialog-wrapper';
import { FileData } from '@/redux/types/files';
import FileInfo from './FileInfo';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { FileContextMenu } from '../context-menu/FileContextMenu';
import { useTranslation } from '@/hooks/use-translation';
import { TrashIcon } from 'lucide-react';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { cn } from '@/lib/utils';
import { MobileActionSheet } from './MobileActionSheet';
import { MobileLayout } from './MobileLayout';
import { GridLayout } from './GridLayout';
import { ListLayout } from './ListLayout';

interface FileItemProps {
  file: FileData;
  onFolderClick: (filePath: string) => void;
  type: 'file' | 'folder';
  layout: 'grid' | 'list' | 'mobile';
  activePath: string;
  onFolderClickActive: (path: string) => void;
  refetch: () => void;
  setFileToCopy: React.Dispatch<React.SetStateAction<FileData | undefined>>;
  setFileToMove: React.Dispatch<React.SetStateAction<FileData | undefined>>;
  index: number;
  isEditing: boolean;
  editedFileName: string;
  setEditedFileName: React.Dispatch<React.SetStateAction<string>>;
  isDialogOpen: boolean;
  setIsDialogOpen: React.Dispatch<React.SetStateAction<boolean>>;
  isSizeLoading: boolean;
  fileSize: number | null;
  handleRename: (file: FileData) => void;
  handleKeyDown: (e: React.KeyboardEvent<HTMLInputElement>, file: FileData) => void;
  handleTextDoubleClick: () => void;
  handleDelete: (path: string) => void;
  handleCopy: (fromPath: string, toPath: string) => void;
  startRenaming: (file: FileData) => void;
}

export function FileItem({
  file,
  onFolderClick,
  type,
  layout,
  activePath,
  onFolderClickActive,
  refetch,
  setFileToCopy,
  setFileToMove,
  index,
  isEditing,
  editedFileName,
  setEditedFileName,
  isDialogOpen,
  setIsDialogOpen,
  isSizeLoading,
  fileSize,
  handleRename,
  handleKeyDown,
  handleTextDoubleClick,
  handleDelete,
  handleCopy,
  startRenaming
}: FileItemProps) {
  const { t } = useTranslation();
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [isMobileActionSheetOpen, setIsMobileActionSheetOpen] = useState(false);

  if (layout === 'mobile') {
    return (
      <ResourceGuard resource="file-manager" action="read" loadingFallback={null}>
        <div
          className={cn(
            'border-b border-border last:border-b-0',
            activePath === file.path && 'bg-muted/50 border-l-4 border-l-primary'
          )}
        >
          <MobileLayout
            file={file}
            type={type}
            activePath={activePath}
            isEditing={isEditing}
            editedFileName={editedFileName}
            setEditedFileName={setEditedFileName}
            handleRename={handleRename}
            handleKeyDown={handleKeyDown}
            handleTextDoubleClick={handleTextDoubleClick}
            onFolderClick={onFolderClick}
            onFolderClickActive={onFolderClickActive}
            onMoreClick={() => setIsMobileActionSheetOpen(true)}
          />
        </div>

        <DialogWrapper
          open={isDialogOpen}
          onOpenChange={setIsDialogOpen}
          size="lg"
        >
          <FileInfo file={file} isLoading={isSizeLoading} fileSize={fileSize || null} />
        </DialogWrapper>

        <MobileActionSheet
          file={file}
          type={type}
          isOpen={isMobileActionSheetOpen}
          onClose={() => setIsMobileActionSheetOpen(false)}
          onInfo={() => setIsDialogOpen(true)}
          onRename={() => startRenaming(file)}
          onCopy={() => {
            setFileToCopy(file);
            handleCopy(file.path, file.path);
          }}
          onMove={() => setFileToMove(file)}
          onDelete={() => setIsDeleteDialogOpen(true)}
        />

        <ResourceGuard resource="file-manager" action="delete" loadingFallback={null}>
          <DeleteDialog
            title={t('fileManager.deleteDialog.title')}
            description={
              type === 'folder'
                ? t('fileManager.deleteDialog.descriptionDirectory', { name: file.name })
                : t('fileManager.deleteDialog.descriptionFile', { name: file.name })
            }
            onConfirm={() => handleDelete(file.path)}
            open={isDeleteDialogOpen}
            onOpenChange={setIsDeleteDialogOpen}
            variant="destructive"
            confirmText={t('fileManager.deleteDialog.confirm')}
            cancelText={t('fileManager.deleteDialog.cancel')}
            icon={TrashIcon}
          />
        </ResourceGuard>
      </ResourceGuard>
    );
  }

  return (
    <ResourceGuard resource="file-manager" action="read" loadingFallback={null}>
      <FileContextMenu
        isItem
        onInfo={() => setIsDialogOpen(true)}
        onRename={() => startRenaming(file)}
        onCopy={() => {
          setFileToCopy(file);
          handleCopy(file.path, file.path);
        }}
        onMoveItem={() => setFileToMove(file)}
        onDelete={() => setIsDeleteDialogOpen(true)}
      >
        <div
          onClick={() => onFolderClickActive(file.path)}
          className="cursor-pointer hover:bg-muted/50 transition-colors rounded-sm"
        >
          {layout === 'grid' ? (
            <GridLayout
              file={file}
              type={type}
              activePath={activePath}
              isEditing={isEditing}
              editedFileName={editedFileName}
              setEditedFileName={setEditedFileName}
              handleRename={handleRename}
              handleKeyDown={handleKeyDown}
              handleTextDoubleClick={handleTextDoubleClick}
              onFolderClick={onFolderClick}
            />
          ) : (
            <ListLayout
              file={file}
              type={type}
              activePath={activePath}
              isEditing={isEditing}
              editedFileName={editedFileName}
              setEditedFileName={setEditedFileName}
              handleRename={handleRename}
              handleKeyDown={handleKeyDown}
              handleTextDoubleClick={handleTextDoubleClick}
              onFolderClick={onFolderClick}
            />
          )}
        </div>

        <DialogWrapper
          open={isDialogOpen}
          onOpenChange={setIsDialogOpen}
          size="lg"
        >
          <FileInfo file={file} isLoading={isSizeLoading} fileSize={fileSize || null} />
        </DialogWrapper>

        <ResourceGuard resource="file-manager" action="delete" loadingFallback={null}>
          <DeleteDialog
            title={t('fileManager.deleteDialog.title')}
            description={
              type === 'folder'
                ? t('fileManager.deleteDialog.descriptionDirectory', { name: file.name })
                : t('fileManager.deleteDialog.descriptionFile', { name: file.name })
            }
            onConfirm={() => handleDelete(file.path)}
            open={isDeleteDialogOpen}
            onOpenChange={setIsDeleteDialogOpen}
            variant="destructive"
            confirmText={t('fileManager.deleteDialog.confirm')}
            cancelText={t('fileManager.deleteDialog.cancel')}
            icon={TrashIcon}
          />
        </ResourceGuard>
      </FileContextMenu>
    </ResourceGuard>
  );
}
