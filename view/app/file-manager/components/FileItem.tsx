import React from 'react';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger
} from '@/components/ui/context-menu';
import { Code, Copy, Info, MoveIcon, Pencil, TrashIcon } from 'lucide-react';
import { Dialog } from '@/components/ui/dialog';
import { FileData, FileType } from '@/redux/types/files';
import { useFileOperations } from '../hooks/use_file_operations';
import { useFileManagerActionsHook } from '../hooks/use_file_actions';
import FileInfo from './FileInfo';
import { getFileIcons } from '@/app/self-host/utils/getFileIcons';
import { formatFileSize } from '@/app/self-host/utils/formatFileSize';

export interface FileItemProps {
  onFolderClick: (filePath: string) => void;
  type: 'folder' | 'file';
  layout: 'grid' | 'list';
  activePath: string;
  onFolderClickActive: (filePath: string) => void;
  refetch: () => void;
  file: FileData;
  isHeader?: boolean;
  setFileToCopy: React.Dispatch<React.SetStateAction<FileData | undefined>>;
  setFileToMove: React.Dispatch<React.SetStateAction<FileData | undefined>>;
  index: number;
}

export const FileItem = ({
  file,
  onFolderClick,
  type,
  layout,
  activePath,
  onFolderClickActive,
  refetch,
  isHeader = false,
  setFileToCopy,
  setFileToMove,
  index
}: FileItemProps) => {
  const {
    isEditing,
    setIsEditing,
    editedFileName,
    setEditedFileName,
    isDialogOpen,
    setIsDialogOpen,
    isSizeLoading,
    fileSize,
    handleRename,
    onDeleteFolder,
    startRenaming,
    handleKeyDown,
    handleTextDoubleClick
  } = useFileOperations(file, refetch);

  const { handleCopyFile, handleFileMove } = useFileManagerActionsHook();

  const renderFileName = () =>
    isEditing ? (
      <input
        type="text"
        value={editedFileName}
        onChange={(e) => setEditedFileName(e.target.value)}
        onBlur={handleRename}
        onKeyDown={handleKeyDown}
        autoFocus
        className={`w-full px-2 py-1 ${activePath === file.path ? 'bg-secondary text-white' : 'bg-transparent'} rounded-md`}
        onClick={(e) => e.stopPropagation()}
      />
    ) : (
      <span
        className={`px-2 py-1 ${layout === 'grid' ? 'text-center' : 'text-left'} 
                ${activePath === file.path ? 'bg-secondary text-white' : ''} break-words rounded-md leading-normal`}
        title={editedFileName}
        onDoubleClick={handleTextDoubleClick}
      >
        {editedFileName}
      </span>
    );

  const gridLayout = (
    <div
      className={`flex flex-col items-center ${activePath === file.path ? '' : ''} w-24 rounded-lg lg:w-32`}
    >
      <div
        className="mb-2"
        onDoubleClick={() => {
          if (type === 'folder') onFolderClick(file.path);
        }}
      >
        {getFileIcons(type, file.name.split('.').pop() as string, layout)}
      </div>
      <div className="w-full break-all text-center">{renderFileName()}</div>
    </div>
  );

  const listLayout = (
    <div
      className={`flex items-center p-2 ${activePath === file.path ? 'bg-muted' : ''} rounded-lg`}
    >
      <div
        className="flex flex-1 items-center"
        onDoubleClick={() => {
          if (type === 'folder') onFolderClick(file.path);
        }}
      >
        {!isHeader && (
          <div className="">{getFileIcons(type, file.name.split('.').pop() as string, layout)}</div>
        )}
        <div className="min-w-0 flex-1">
          {isHeader ? <span className="text-muted-foreground">Name</span> : renderFileName()}
        </div>
      </div>
      <div className="flex items-center space-x-4 text-xs text-muted-foreground">
        <p className="w-24">
          {isHeader ? 'Size' : file.file_type === FileType.File ? formatFileSize(file.size) : '-'}
        </p>
        <p className="w-20">{isHeader ? 'Type' : file.file_type}</p>
        <p className="w-32">
          {isHeader ? 'Created On' : new Date(file.created_at).toLocaleDateString()}
        </p>
        <p className="w-32">
          {isHeader ? 'Modified On' : new Date(file.updated_at).toLocaleDateString()}
        </p>
      </div>
    </div>
  );

  if (isHeader && layout === 'list') {
    return listLayout;
  }

  return (
    <ContextMenu>
      <ContextMenuTrigger>
        <div onClick={() => onFolderClickActive(file.path)} className="cursor-pointer">
          {layout === 'grid' ? gridLayout : listLayout}
        </div>
      </ContextMenuTrigger>
      <ContextMenuContent>
        <ContextMenuItem onSelect={() => setIsDialogOpen(true)}>
          <Info className="mr-2 h-4 w-4" /> Info
        </ContextMenuItem>
        <ContextMenuItem
          onSelect={() => {
            window.open('http://localhost:8087?folder=' + file.path, '_blank');
          }}
        >
          <Code className="mr-2 h-4 w-4" /> Open in Code Editor
        </ContextMenuItem>
        <ContextMenuItem onSelect={startRenaming}>
          <Pencil className="mr-2 h-4 w-4" /> Rename
        </ContextMenuItem>
        <ContextMenuItem onSelect={() => handleCopyFile(file, setFileToCopy)}>
          <Copy className="mr-2 h-4 w-4" /> Copy
        </ContextMenuItem>
        <ContextMenuItem onSelect={() => setFileToMove(file)}>
          <MoveIcon className="mr-2 h-4 w-4" /> Move
        </ContextMenuItem>
        <ContextMenuItem onSelect={onDeleteFolder}>
          <TrashIcon className="mr-2 h-4 w-4" /> Move to Trash
        </ContextMenuItem>
      </ContextMenuContent>
      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <FileInfo file={file} isLoading={isSizeLoading} fileSize={fileSize || null} />
      </Dialog>
    </ContextMenu>
  );
};
