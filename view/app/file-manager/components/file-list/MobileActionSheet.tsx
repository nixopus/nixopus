import React from 'react';
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription
} from '@/components/ui/sheet';
import { Separator } from '@/components/ui/separator';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Button } from '@/components/ui/button';
import { FileData, FileType } from '@/redux/types/files';
import { getFileIcons } from '@/app/self-host/utils/getFileIcons';
import { formatFileSize } from '@/app/self-host/utils/formatFileSize';
import { useTranslation } from '@/hooks/use-translation';
import { TrashIcon, Download, Share, Edit3, Copy, Move, Info } from 'lucide-react';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';
import { cn } from '@/lib/utils';

interface MobileActionSheetProps {
  file: FileData;
  type: 'file' | 'folder';
  isOpen: boolean;
  onClose: () => void;
  onInfo: () => void;
  onRename: () => void;
  onCopy: () => void;
  onMove: () => void;
  onDelete: () => void;
}

export const MobileActionSheet: React.FC<MobileActionSheetProps> = ({
  file,
  type,
  isOpen,
  onClose,
  onInfo,
  onRename,
  onCopy,
  onMove,
  onDelete
}) => {
  const { t } = useTranslation();

  const actions = [
    { id: 'info', label: t('fileManager.contextMenu.info'), icon: Info, onClick: onInfo },
    { id: 'rename', label: t('fileManager.contextMenu.rename'), icon: Edit3, onClick: onRename },
    { id: 'copy', label: t('fileManager.contextMenu.copy'), icon: Copy, onClick: onCopy },
    { id: 'move', label: t('fileManager.contextMenu.move'), icon: Move, onClick: onMove },
    { id: 'share', label: t('fileManager.contextMenu.share'), icon: Share, onClick: () => {} },
    ...(type === 'file'
      ? [
          {
            id: 'download',
            label: t('fileManager.contextMenu.download'),
            icon: Download,
            onClick: () => {}
          }
        ]
      : []),
    {
      id: 'delete',
      label: t('fileManager.contextMenu.delete'),
      icon: TrashIcon,
      onClick: onDelete,
      destructive: true
    }
  ];

  return (
    <Sheet open={isOpen} onOpenChange={onClose}>
      <SheetContent side="bottom" className="rounded-t-2xl">
        <div className="mx-auto w-12 h-1.5 flex-shrink-0 rounded-full bg-muted mb-4" />

        <SheetHeader className="pb-4">
          <div className="flex items-center space-x-3">
            <div className="w-12 h-12 flex items-center justify-center bg-muted rounded-lg">
              {getFileIcons(type, file.name.split('.').pop() as string, 'list')}
            </div>
            <div className="flex-1 min-w-0 text-left">
              <SheetTitle className="truncate">{file.name}</SheetTitle>
              <SheetDescription>
                {file.file_type === FileType.File ? formatFileSize(file.size) : 'Folder'} •{' '}
                {new Date(file.updated_at).toLocaleDateString()}
              </SheetDescription>
            </div>
          </div>
        </SheetHeader>

        <Separator className="mb-4" />

        <ScrollArea className="max-h-[400px]">
          <div className="space-y-1">
            {actions.map((action) => (
              <ResourceGuard
                key={action.id}
                resource="file-manager"
                action={action.id === 'delete' ? 'delete' : 'read'}
                fallback={null}
              >
                <Button
                  variant="ghost"
                  onClick={() => {
                    action.onClick();
                    onClose();
                  }}
                  className={cn(
                    'w-full justify-start h-12 px-4',
                    action.destructive && 'text-destructive hover:text-destructive'
                  )}
                >
                  <action.icon className="w-5 h-5 mr-3" />
                  <span>{action.label}</span>
                </Button>
              </ResourceGuard>
            ))}
          </div>
        </ScrollArea>
        <Separator className="my-4" />
        <Button variant="outline" onClick={onClose} className="w-full">
          {t('fileManager.contextMenu.cancel')}
        </Button>
      </SheetContent>
    </Sheet>
  );
};
