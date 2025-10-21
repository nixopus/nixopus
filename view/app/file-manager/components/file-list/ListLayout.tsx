import React from 'react';
import { FileData, FileType } from '@/redux/types/files';
import { getFileIcons } from '@/app/self-host/utils/getFileIcons';
import { formatFileSize } from '@/app/self-host/utils/formatFileSize';
import { cn } from '@/lib/utils';

interface ListLayoutProps {
  file: FileData;
  type: 'file' | 'folder';
  activePath: string;
  isEditing: boolean;
  editedFileName: string;
  setEditedFileName: React.Dispatch<React.SetStateAction<string>>;
  handleRename: (file: FileData) => void;
  handleKeyDown: (e: React.KeyboardEvent<HTMLInputElement>, file: FileData) => void;
  handleTextDoubleClick: () => void;
  onFolderClick: (filePath: string) => void;
}

export const ListLayout: React.FC<ListLayoutProps> = ({
  file,
  type,
  activePath,
  isEditing,
  editedFileName,
  setEditedFileName,
  handleRename,
  handleKeyDown,
  handleTextDoubleClick,
  onFolderClick
}) => {
  const renderFileName = () =>
    isEditing ? (
      <input
        type="text"
        value={editedFileName}
        onChange={(e) => setEditedFileName(e.target.value)}
        onBlur={() => handleRename(file)}
        onKeyDown={(e) => handleKeyDown(e, file)}
        autoFocus
        className={cn(
          'w-full px-2 py-1 rounded-md border-0 bg-transparent focus:ring-2 focus:ring-primary text-left',
          activePath === file.path ? 'bg-secondary text-secondary-foreground' : 'bg-background'
        )}
        onClick={(e) => e.stopPropagation()}
      />
    ) : (
      <span
        className={cn(
          'px-2 py-1 rounded-md break-words leading-normal text-left',
          activePath === file.path ? 'bg-secondary text-secondary-foreground' : 'text-foreground'
        )}
        title={file.name}
        onDoubleClick={handleTextDoubleClick}
      >
        {file.name}
      </span>
    );

  return (
    <div className="flex items-center p-2">
      <div
        className="flex flex-1 items-center cursor-pointer"
        onDoubleClick={() => {
          if (type === 'folder') onFolderClick(file.path);
        }}
      >
        <div className="mr-3">
          {getFileIcons(type, file.name.split('.').pop() as string, 'list')}
        </div>
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
};
