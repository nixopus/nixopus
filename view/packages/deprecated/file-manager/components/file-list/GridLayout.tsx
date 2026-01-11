import React from 'react';
import { FileData, FileType } from '@/redux/types/files';
import { cn } from '@/lib/utils';
import { Folder, File, Link2 } from 'lucide-react';
import { formatFileSize } from '@/packages/utils/util';

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
  const isDirectory = file.file_type === FileType.Directory;
  const isSymlink = file.file_type === FileType.Symlink;

  const getIcon = () => {
    const iconClass = 'h-14 w-14 sm:h-16 sm:w-16 md:h-20 md:w-20';
    if (isDirectory) {
      return <Folder className={cn(iconClass, ' fill-[#58b7e9] text-[#58b7e9]')} />;
    }
    if (isSymlink) {
      return <Link2 className={cn(iconClass, 'text-muted-foreground')} />;
    }
    return <File className={cn(iconClass, 'text-muted-foreground')} />;
  };

  const renderFileName = () =>
    isEditing ? (
      <input
        type="text"
        value={editedFileName}
        onChange={(e) => setEditedFileName(e.target.value)}
        onBlur={() => handleRename(file)}
        onKeyDown={(e) => handleKeyDown(e, file)}
        autoFocus
        className="w-full px-2 py-1 rounded-md border-0 bg-background focus:ring-2 focus:ring-primary text-center text-xs sm:text-sm"
        onClick={(e) => e.stopPropagation()}
      />
    ) : (
      <span
        className="line-clamp-2 text-center text-xs sm:text-sm leading-tight"
        title={file.name}
        onDoubleClick={handleTextDoubleClick}
      >
        {file.name}
      </span>
    );

  return (
    <div
      className="flex flex-col items-center w-full p-4 sm:p-5 cursor-pointer"
      onDoubleClick={() => {
        if (type === 'folder') onFolderClick(file.path);
      }}
    >
      <div className="mb-3 sm:mb-4 flex items-center justify-center">{getIcon()}</div>
      <div className="w-full text-center min-h-[2.5rem]">{renderFileName()}</div>
      {!isDirectory && (
        <p className="text-[10px] sm:text-xs text-muted-foreground mt-1">
          {formatFileSize(file.size)}
        </p>
      )}
    </div>
  );
};
