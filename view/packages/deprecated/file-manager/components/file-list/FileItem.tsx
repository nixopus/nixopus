import React, { useState } from 'react';
import { DialogWrapper } from '@/components/ui/dialog-wrapper';
import { FileData } from '@/redux/types/files';
import FileInfo from './FileInfo';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { FileContextMenu } from '../context-menu/FileContextMenu';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { TrashIcon } from 'lucide-react';
import { ResourceGuard } from '@/packages/components/rbac';
import { GridLayout } from './GridLayout';

interface FileItemProps {
  file: FileData;
  onFolderClick: (filePath: string) => void;
  type: 'file' | 'folder';
  activePath: string;
  onFolderClickActive: (path: string) => void;
  setFileToCopy: React.Dispatch<React.SetStateAction<FileData | undefined>>;
  setFileToMove: React.Dispatch<React.SetStateAction<FileData | undefined>>;
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
  activePath,
  onFolderClickActive,
  setFileToCopy,
  setFileToMove,
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
        <div onClick={() => onFolderClickActive(file.path)}>
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
        </div>

        <DialogWrapper open={isDialogOpen} onOpenChange={setIsDialogOpen} size="lg">
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
