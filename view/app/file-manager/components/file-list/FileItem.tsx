import React from 'react';
import { Dialog } from '@/components/ui/dialog';
import { FileData, FileType } from '@/redux/types/files';
import FileInfo from './FileInfo';
import { getFileIcons } from '@/app/self-host/utils/getFileIcons';
import { formatFileSize } from '@/app/self-host/utils/formatFileSize';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import { FileContextMenu } from '../context-menu/FileContextMenu';
import { useTranslation } from '@/hooks/use-translation';
import { TrashIcon } from 'lucide-react';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';

interface FileItemProps {
  file: FileData;
  onFolderClick: (filePath: string) => void;
  type: 'file' | 'folder';
  layout: 'grid' | 'list';
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

  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = React.useState(false);

  const renderFileName = () =>
    isEditing ? (
      <input
        type="text"
        value={editedFileName}
        onChange={(e) => setEditedFileName(e.target.value)}
        onBlur={() => handleRename(file)}
        onKeyDown={(e) => handleKeyDown(e, file)}
        autoFocus
        className={`w-full px-2 py-1 ${activePath === file.path ? 'bg-secondary text-white' : 'bg-transparent'} rounded-md`}
        onClick={(e) => e.stopPropagation()}
      />
    ) : (
      <span
        className={`px-2 py-1 ${layout === 'grid' ? 'text-center' : 'text-left'} 
                ${activePath === file.path ? 'bg-secondary text-white' : ''} break-words rounded-md leading-normal`}
        title={file.name}
        onDoubleClick={handleTextDoubleClick}
      >
        {file.name}
      </span>
    );

  const gridLayout = (
    <div
      className={`flex flex-col items-center w-full min-h-[120px] sm:min-h-[130px] md:min-h-[140px] p-2 sm:p-3`}
    >
      <div
        className="mb-3 flex items-center justify-center flex-1"
        onDoubleClick={() => {
          if (type === 'folder') onFolderClick(file.path);
        }}
      >
        <div className="h-12 w-12 sm:h-16 sm:w-16 md:h-20 md:w-20 flex items-center justify-center">
          {getFileIcons(type, file.name.split('.').pop() as string, layout)}
        </div>
      </div>
      <div className="w-full px-1 sm:px-2 text-center text-xs sm:text-sm truncate">
        {renderFileName()}
      </div>
    </div>
  );

  const listLayout = (
    <div className={`flex items-center p-2`}>
      <div
        className="flex flex-1 items-center"
        onDoubleClick={() => {
          if (type === 'folder') onFolderClick(file.path);
        }}
      >
        <div className="">{getFileIcons(type, file.name.split('.').pop() as string, layout)}</div>
        <div className="min-w-0 flex-1">{renderFileName()}</div>
      </div>
      <div className="flex items-center space-x-4 text-xs text-muted-foreground">
        <p className="w-24">{file.file_type === FileType.File ? formatFileSize(file.size) : '-'}</p>
        <p className="w-20">{file.file_type}</p>
        <p className="w-32">{new Date(file.created_at).toLocaleDateString()}</p>
        <p className="w-32">{new Date(file.updated_at).toLocaleDateString()}</p>
      </div>
    </div>
  );

  return (
    <ResourceGuard 
      resource="file-manager" 
      action="read"
      loadingFallback={null}
    >
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
        <div onClick={() => onFolderClickActive(file.path)} className="cursor-pointer">
          {layout === 'grid' ? gridLayout : listLayout}
        </div>
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <FileInfo file={file} isLoading={isSizeLoading} fileSize={fileSize || null} />
        </Dialog>
        <ResourceGuard 
          resource="file-manager" 
          action="delete"
          loadingFallback={null}
        >
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
