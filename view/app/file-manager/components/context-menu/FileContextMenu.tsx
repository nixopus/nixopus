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
  Info,
  MoveIcon,
  Pencil,
  TrashIcon,
  UploadCloudIcon
} from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import { FileData } from '@/redux/types/files';
import { toast } from 'sonner';
import { DialogWrapper } from '@/components/ui/dialog-wrapper';
import FileUpload from '../actions/Upload';
import { useCreateDirectoryMutation } from '@/redux/services/file-manager/fileManagersApi';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';

interface FileContextMenuProps {
  children: React.ReactNode;
  showHidden?: boolean;
  setShowHidden?: (show: boolean) => void;
  fileToMove?: FileData;
  fileToCopy?: FileData;
  currentPath?: string;
  files?: FileData[];
  onFileUpload?: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onCreateFolder?: () => void;
  onMove?: () => void;
  onPaste?: () => void;
  onInfo?: () => void;
  onRename?: () => void;
  onCopy?: () => void;
  onMoveItem?: () => void;
  onDelete?: () => void;
  isItem?: boolean;
  refetch?: () => void;
}

export function FileContextMenu({
  children,
  showHidden,
  setShowHidden,
  fileToMove,
  fileToCopy,
  currentPath,
  files = [],
  onFileUpload,
  onCreateFolder,
  onMove,
  onPaste,
  onInfo,
  onRename,
  onCopy,
  onMoveItem,
  onDelete,
  isItem = false,
  refetch
}: FileContextMenuProps) {
  const { t } = useTranslation();
  const [isUploadOpen, setIsUploadOpen] = React.useState(false);
  const [createDirectory] = useCreateDirectoryMutation();

  const handleCreateDirectory = async () => {
    try {
      const newFolderNumber = files.filter((f) => f.name.startsWith('New Folder')).length + 1;
      const newFolderName = `New Folder ${newFolderNumber}`;
      await createDirectory({ path: currentPath || '/', name: newFolderName });
      refetch?.();
    } catch (error) {
      toast.error(t('fileManager.actions.errors.createDirectory'), {
        description: error instanceof Error ? error.message : 'Unknown error'
      });
    }
  };

  const handleMove = () => {
    onMove?.();
  };

  const handlePaste = () => {
    onPaste?.();
  };

  return (
    <ContextMenu>
      <ContextMenuTrigger className="h-full w-full">{children}</ContextMenuTrigger>
      <ContextMenuContent className="w-64">
        {isItem ? (
          <>
            <ContextMenuItem onSelect={onInfo}>
              <Info className="mr-2 h-4 w-4" /> {t('fileManager.item.actions.info')}
            </ContextMenuItem>
            <ResourceGuard resource="file-manager" action="update">
              <ContextMenuItem onSelect={onRename}>
                <Pencil className="mr-2 h-4 w-4" /> {t('fileManager.item.actions.rename')}
              </ContextMenuItem>
              <ContextMenuItem onSelect={onCopy}>
                <Copy className="mr-2 h-4 w-4" /> {t('fileManager.item.actions.copy')}
              </ContextMenuItem>
              <ContextMenuItem onSelect={onMoveItem}>
                <MoveIcon className="mr-2 h-4 w-4" /> {t('fileManager.item.actions.move')}
              </ContextMenuItem>
            </ResourceGuard>
            <ResourceGuard resource="file-manager" action="delete">
              <ContextMenuItem onSelect={onDelete}>
                <TrashIcon className="mr-2 h-4 w-4" /> {t('fileManager.item.actions.delete')}
              </ContextMenuItem>
            </ResourceGuard>
          </>
        ) : (
          <>
            <ResourceGuard resource="file-manager" action="create">
              <DialogWrapper
                open={isUploadOpen}
                onOpenChange={setIsUploadOpen}
                trigger={
                  <ContextMenuItem onSelect={(e) => e.preventDefault()}>
                    <UploadCloudIcon className="mr-2 h-5 w-5" />
                    <span>{t('fileManager.actions.upload')}</span>
                  </ContextMenuItem>
                }
                size="lg"
              >
                <FileUpload setIsDialogOpen={setIsUploadOpen} currentPath={currentPath || ''} />
              </DialogWrapper>

              <ContextMenuItem onSelect={handleCreateDirectory}>
                <FolderPlusIcon className="mr-2 h-5 w-5" />
                <span>{t('fileManager.actions.createDirectory')}</span>
              </ContextMenuItem>
            </ResourceGuard>
            <ResourceGuard resource="file-manager" action="update">
              {fileToMove && (
                <ContextMenuItem onSelect={handleMove}>
                  <MoveIcon className="mr-2 h-5 w-5" />
                  <span>{t('fileManager.item.actions.moveHere')}</span>
                </ContextMenuItem>
              )}
              {fileToCopy && (
                <ContextMenuItem onSelect={handlePaste}>
                  <Copy className="mr-2 h-5 w-5" />
                  <span>{t('fileManager.item.actions.pasteHere')}</span>
                </ContextMenuItem>
              )}
            </ResourceGuard>
            <ContextMenuItem onSelect={() => setShowHidden?.(!showHidden)}>
              {showHidden ? (
                <EyeOffIcon className="mr-2 h-5 w-5" />
              ) : (
                <EyeIcon className="mr-2 h-5 w-5" />
              )}
              <span>
                {showHidden
                  ? t('fileManager.actions.hideHidden')
                  : t('fileManager.actions.showHidden')}
              </span>
            </ContextMenuItem>
          </>
        )}
      </ContextMenuContent>
    </ContextMenu>
  );
}
