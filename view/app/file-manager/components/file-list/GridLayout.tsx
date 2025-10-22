import React from 'react';
import { FileData } from '@/redux/types/files';
import { getFileIcons } from '@/app/self-host/utils/getFileIcons';
import { cn } from '@/lib/utils';

interface GridLayoutProps {
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

export const GridLayout: React.FC<GridLayoutProps> = ({
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
          'w-full px-2 py-1 rounded-md border-0 bg-transparent focus:ring-2 focus:ring-primary text-center',
          activePath === file.path ? 'bg-secondary text-secondary-foreground' : 'bg-background'
        )}
        onClick={(e) => e.stopPropagation()}
      />
    ) : (
      <span
        className={cn(
          'px-2 py-1 rounded-md break-words leading-normal text-center',
          activePath === file.path ? 'bg-secondary text-secondary-foreground' : 'text-foreground'
        )}
        title={file.name}
        onDoubleClick={handleTextDoubleClick}
      >
        {file.name}
      </span>
    );

  return (
    <div className="flex flex-col items-center w-full min-h-[120px] sm:min-h-[130px] md:min-h-[140px] p-2 sm:p-3">
      <div
        className="mb-3 flex items-center justify-center flex-1 cursor-pointer"
        onDoubleClick={() => {
          if (type === 'folder') onFolderClick(file.path);
        }}
      >
        <div className="h-12 w-12 sm:h-16 sm:w-16 md:h-20 md:w-20 flex items-center justify-center">
          {getFileIcons(type, file.name.split('.').pop() as string, 'grid')}
        </div>
      </div>
      <div className="w-full px-1 sm:px-2 text-center text-xs sm:text-sm truncate">
        {renderFileName()}
      </div>
    </div>
  );
};
