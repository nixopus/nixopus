'use client';

import React from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { FileData, FileType } from '@/redux/types/files';
import { cn } from '@/lib/utils';
import { Folder, File, ChevronUp, ChevronDown, Link2 } from 'lucide-react';
import { formatFileSize } from '@/packages/utils/util';
import { formatDistanceToNow, isValid } from 'date-fns';
import { FileContextMenu } from '../context-menu/FileContextMenu';
import { ResourceGuard } from '@/packages/components/rbac';

type SortField = 'name' | 'size' | 'file_type' | 'created_at' | 'updated_at';

function formatDate(dateString: string): string {
  if (!dateString) return '—';
  const date = new Date(dateString);
  if (!isValid(date)) return '—';
  return formatDistanceToNow(date, { addSuffix: true });
}

interface FileTableProps {
  files: FileData[];
  sortBy?: SortField;
  sortOrder?: 'asc' | 'desc';
  onSort?: (field: SortField) => void;
  activePath: string;
  onFileSelect: (path: string) => void;
  onFolderClick: (path: string) => void;
  isEditing: boolean;
  editedFileName: string;
  setEditedFileName: React.Dispatch<React.SetStateAction<string>>;
  handleRename: (file: FileData) => void;
  handleKeyDown: (e: React.KeyboardEvent<HTMLInputElement>, file: FileData) => void;
  handleTextDoubleClick: () => void;
  onInfo: (file: FileData) => void;
  onRename: (file: FileData) => void;
  onCopy: (file: FileData) => void;
  onMove: (file: FileData) => void;
  onDelete: (file: FileData) => void;
}

const FileTable: React.FC<FileTableProps> = ({
  files,
  sortBy = 'name',
  sortOrder = 'asc',
  onSort,
  activePath,
  onFileSelect,
  onFolderClick,
  isEditing,
  editedFileName,
  setEditedFileName,
  handleRename,
  handleKeyDown,
  handleTextDoubleClick,
  onInfo,
  onRename,
  onCopy,
  onMove,
  onDelete
}) => {
  const { t } = useTranslation();

  if (files.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
        <Folder className="h-12 w-12 mb-4 opacity-30" />
        <p className="text-sm">{t('fileManager.noFiles')}</p>
      </div>
    );
  }

  return (
    <div className="rounded-xl border overflow-hidden">
      <div className="grid grid-cols-[1fr_80px] sm:grid-cols-[1fr_80px_80px] lg:grid-cols-[1fr_100px_100px_140px_140px] gap-2 sm:gap-4 px-3 sm:px-4 py-3 bg-muted/30 text-xs font-medium text-muted-foreground uppercase tracking-wider">
        <SortableHeader
          label={t('fileManager.sort.methods.name')}
          field="name"
          currentSort={sortBy}
          currentOrder={sortOrder}
          onSort={onSort}
        />
        <SortableHeader
          label={t('fileManager.sort.methods.size')}
          field="size"
          currentSort={sortBy}
          currentOrder={sortOrder}
          onSort={onSort}
        />
        <div className="hidden sm:block">
          <SortableHeader
            label={t('fileManager.sort.methods.file_type')}
            field="file_type"
            currentSort={sortBy}
            currentOrder={sortOrder}
            onSort={onSort}
          />
        </div>
        <div className="hidden lg:block">
          <SortableHeader
            label={t('fileManager.sort.methods.created_at')}
            field="created_at"
            currentSort={sortBy}
            currentOrder={sortOrder}
            onSort={onSort}
          />
        </div>
        <div className="hidden lg:block">
          <SortableHeader
            label={t('fileManager.sort.methods.updated_at')}
            field="updated_at"
            currentSort={sortBy}
            currentOrder={sortOrder}
            onSort={onSort}
          />
        </div>
      </div>

      <div className="divide-y divide-border/50">
        {files.map((file) => (
          <FileRow
            key={file.path}
            file={file}
            isActive={activePath === file.path}
            isEditing={isEditing && activePath === file.path}
            editedFileName={editedFileName}
            setEditedFileName={setEditedFileName}
            handleRename={handleRename}
            handleKeyDown={handleKeyDown}
            handleTextDoubleClick={handleTextDoubleClick}
            onClick={() => onFileSelect(file.path)}
            onDoubleClick={() => {
              if (file.file_type === FileType.Directory) {
                onFolderClick(file.path);
              }
            }}
            onInfo={() => onInfo(file)}
            onRename={() => onRename(file)}
            onCopy={() => onCopy(file)}
            onMove={() => onMove(file)}
            onDelete={() => onDelete(file)}
          />
        ))}
      </div>
    </div>
  );
};

function SortableHeader({
  label,
  field,
  currentSort,
  currentOrder,
  onSort
}: {
  label: string;
  field: SortField;
  currentSort: SortField;
  currentOrder: 'asc' | 'desc';
  onSort?: (field: SortField) => void;
}) {
  const isActive = currentSort === field;

  return (
    <button
      onClick={() => onSort?.(field)}
      className="flex items-center gap-1 hover:text-foreground transition-colors text-left"
    >
      {label}
      <span className="flex flex-col">
        <ChevronUp
          className={cn(
            'h-3 w-3 -mb-1',
            isActive && currentOrder === 'asc' ? 'text-foreground' : 'opacity-30'
          )}
        />
        <ChevronDown
          className={cn(
            'h-3 w-3',
            isActive && currentOrder === 'desc' ? 'text-foreground' : 'opacity-30'
          )}
        />
      </span>
    </button>
  );
}

function FileRow({
  file,
  isActive,
  isEditing,
  editedFileName,
  setEditedFileName,
  handleRename,
  handleKeyDown,
  handleTextDoubleClick,
  onClick,
  onDoubleClick,
  onInfo,
  onRename,
  onCopy,
  onMove,
  onDelete
}: {
  file: FileData;
  isActive: boolean;
  isEditing: boolean;
  editedFileName: string;
  setEditedFileName: React.Dispatch<React.SetStateAction<string>>;
  handleRename: (file: FileData) => void;
  handleKeyDown: (e: React.KeyboardEvent<HTMLInputElement>, file: FileData) => void;
  handleTextDoubleClick: () => void;
  onClick: () => void;
  onDoubleClick: () => void;
  onInfo: () => void;
  onRename: () => void;
  onCopy: () => void;
  onMove: () => void;
  onDelete: () => void;
}) {
  const { t } = useTranslation();
  const isDirectory = file.file_type === FileType.Directory;
  const isSymlink = file.file_type === FileType.Symlink;

  const getFileIcon = () => {
    if (isDirectory) {
      return <Folder className="h-5 w-5 flex-shrink-0 fill-[#58b7e9] text-[#58b7e9]" />;
    }
    if (isSymlink) {
      return <Link2 className="h-5 w-5 flex-shrink-0 text-muted-foreground" />;
    }
    return <File className="h-5 w-5 flex-shrink-0 text-muted-foreground" />;
  };

  const renderFileName = () => {
    if (isEditing) {
      return (
        <input
          type="text"
          value={editedFileName}
          onChange={(e) => setEditedFileName(e.target.value)}
          onBlur={() => handleRename(file)}
          onKeyDown={(e) => handleKeyDown(e, file)}
          autoFocus
          className={cn(
            'w-full px-2 py-1 rounded-md border-0 bg-transparent focus:ring-2 focus:ring-primary text-left text-sm',
            isActive ? 'bg-secondary text-secondary-foreground' : 'bg-background'
          )}
          onClick={(e) => e.stopPropagation()}
        />
      );
    }
    return (
      <span
        className="text-sm truncate"
        title={file.name}
        onDoubleClick={(e) => {
          e.stopPropagation();
          handleTextDoubleClick();
        }}
      >
        {file.name}
      </span>
    );
  };

  return (
    <ResourceGuard resource="file-manager" action="read" loadingFallback={null}>
      <FileContextMenu
        isItem
        onInfo={onInfo}
        onRename={onRename}
        onCopy={onCopy}
        onMoveItem={onMove}
        onDelete={onDelete}
      >
        <div
          onClick={onClick}
          onDoubleClick={onDoubleClick}
          className={cn(
            'grid grid-cols-[1fr_80px] sm:grid-cols-[1fr_80px_80px] lg:grid-cols-[1fr_100px_100px_140px_140px] gap-2 sm:gap-4 px-3 sm:px-4 py-3 items-center cursor-pointer hover:bg-muted/30 transition-colors group',
            isActive && 'bg-muted/50'
          )}
        >
          <div className="flex items-center gap-2 sm:gap-3 min-w-0">
            {getFileIcon()}
            <div className="min-w-0 flex-1">
              <p className="font-medium truncate text-sm">{renderFileName()}</p>
              {file.is_hidden && (
                <p className="text-xs text-muted-foreground/60">
                  {t('fileManager.info.labels.hidden')}
                </p>
              )}
            </div>
          </div>

          <div className="text-xs sm:text-sm text-muted-foreground">
            {file.file_type === FileType.File ? formatFileSize(file.size) : '—'}
          </div>

          <div className="hidden sm:block text-sm text-muted-foreground">
            {file.extension ? `.${file.extension}` : file.file_type}
          </div>

          <div className="hidden lg:block text-sm text-muted-foreground">
            {formatDate(file.created_at)}
          </div>

          <div className="hidden lg:block text-sm text-muted-foreground">
            {formatDate(file.updated_at)}
          </div>
        </div>
      </FileContextMenu>
    </ResourceGuard>
  );
}

export default FileTable;
