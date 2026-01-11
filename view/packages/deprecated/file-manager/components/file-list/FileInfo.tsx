import React from 'react';
import { DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { FileData } from '@/redux/types/files';
import { formatFileSize } from '@/packages/utils/util';
import { useTranslation } from '@/packages/hooks/shared/use-translation';

interface FileInfoProps {
  file: FileData;
  isLoading: boolean;
  fileSize: number | null;
}

function FileInfo({ file, isLoading, fileSize }: FileInfoProps) {
  const { t } = useTranslation();

  return (
    <DialogContent className="sm:max-w-[425px]">
      <DialogHeader className="space-y-6">
        <DialogTitle className="text-2xl font-bold text-primary">{file.name}</DialogTitle>
        <div className="mx-auto w-full space-y-4 rounded-lg bg-secondary/10 p-4">
          <InfoItem label={t('fileManager.info.labels.path')} value={file.path} />
          <InfoItem
            label={t('fileManager.info.labels.size')}
            value={
              isLoading ? t('fileManager.info.calculating') : formatFileSize(fileSize || file.size)
            }
          />
          <InfoItem
            label={t('fileManager.info.labels.created')}
            value={new Date(file.created_at).toLocaleString()}
          />
          <InfoItem
            label={t('fileManager.info.labels.updated')}
            value={new Date(file.updated_at).toLocaleString()}
          />
          <InfoItem label={t('fileManager.info.labels.type')} value={file.file_type} />
          <InfoItem
            label={t('fileManager.info.labels.hidden')}
            value={file.is_hidden ? t('fileManager.info.yes') : t('fileManager.info.no')}
          />
          <InfoItem
            label={t('fileManager.info.labels.permissions')}
            value={
              (file.permissions & 0o400 ? 'r' : '-') +
              (file.permissions & 0o200 ? 'w' : '-') +
              (file.permissions & 0o100 ? 'x' : '-')
            }
          />
        </div>
      </DialogHeader>
    </DialogContent>
  );
}

function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center space-x-3 text-sm">
      <span className="font-medium text-secondary-foreground">{label}:</span>
      <span className="text-secondary-foreground/80">{value}</span>
    </div>
  );
}

export default FileInfo;
