import React from 'react';
import { Button } from '@/components/ui/button';
import { FileData, FileType } from '@/redux/types/files';
import { formatFileSize } from '@/app/self-host/utils/formatFileSize';
import { MoreVertical, File, Folder } from 'lucide-react';
import { cn } from '@/lib/utils';

interface MobileLayoutProps {
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
  onFolderClickActive: (path: string) => void;
  onMoreClick: () => void;
}

export const MobileLayout: React.FC<MobileLayoutProps> = ({
  file,
  type,
  activePath,
  isEditing,
  editedFileName,
  setEditedFileName,
  handleRename,
  handleKeyDown,
  handleTextDoubleClick,
  onFolderClick,
  onFolderClickActive,
  onMoreClick
}) => {
  const getFileTypeIcon = (size: 'sm' | 'md' | 'lg' = 'md') => {
    const iconSize = size === 'sm' ? 'w-5 h-5' : size === 'md' ? 'w-6 h-6' : 'w-8 h-8';

    if (file.file_type === FileType.Directory) {
      return <Folder className={cn(iconSize, 'text-primary')} />;
    }
    return <File className={cn(iconSize, 'text-primary')} />;
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
        className={cn(
          'w-full px-2 py-1 rounded-md border-0 bg-transparent focus:ring-2 focus:ring-primary',
          activePath === file.path ? 'bg-secondary text-secondary-foreground' : 'bg-background'
        )}
        onClick={(e) => e.stopPropagation()}
      />
    ) : (
      <span
        className={cn(
          'px-2 py-1 rounded-md truncate leading-normal text-left',
          activePath === file.path ? 'bg-secondary text-secondary-foreground' : 'text-foreground'
        )}
        title={file.name}
        onDoubleClick={handleTextDoubleClick}
      >
        {file.name}
      </span>
    );

  return (
    <div
      className="flex items-center space-x-4 p-4 active:bg-muted/50 transition-colors touch-manipulation"
      onClick={() => onFolderClickActive(file.path)}
      onTouchStart={() => {
        const timer = setTimeout(() => onMoreClick(), 500);
        const cleanup = () => clearTimeout(timer);

        const handleTouchEnd = () => {
          cleanup();
          document.removeEventListener('touchend', handleTouchEnd);
          document.removeEventListener('touchmove', handleTouchMove);
        };

        const handleTouchMove = () => {
          cleanup();
          document.removeEventListener('touchend', handleTouchEnd);
          document.removeEventListener('touchmove', handleTouchMove);
        };

        document.addEventListener('touchend', handleTouchEnd);
        document.addEventListener('touchmove', handleTouchMove);
      }}
    >
      <div className="flex-shrink-0">{getFileTypeIcon('lg')}</div>

      <div
        className="flex-1 min-w-0"
        onDoubleClick={() => {
          if (type === 'folder') onFolderClick(file.path);
        }}
      >
        <div className="flex items-center space-x-2 h-6">
          <div className="font-medium text-foreground truncate flex-1">{renderFileName()}</div>
          <div className="flex items-center space-x-2 text-sm text-muted-foreground flex-shrink-0">
            <span className="truncate">
              {file.file_type === FileType.File ? formatFileSize(file.size) : 'Folder'}
            </span>
            <span>•</span>
            <span className="truncate">{new Date(file.updated_at).toLocaleDateString()}</span>
          </div>
        </div>
      </div>
      <Button
        variant="ghost"
        size="sm"
        onClick={(e) => {
          e.stopPropagation();
          onMoreClick();
        }}
        className="h-10 w-10 p-0 text-muted-foreground hover:text-foreground"
      >
        <MoreVertical className="w-5 h-5" />
      </Button>
    </div>
  );
};
