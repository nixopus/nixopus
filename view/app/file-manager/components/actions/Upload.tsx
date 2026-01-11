import React, { useRef, useState } from 'react';
import { Progress } from '@/components/ui/progress';
import { Button } from '@/components/ui/button';
import { Upload, Trash } from 'lucide-react';
import { DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { useFileOperations } from '../../hooks/file-operations/useOperations';
import { useTranslation } from '@/hooks/use-translation';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';

interface FileUploadProps {
  setIsDialogOpen: React.Dispatch<React.SetStateAction<boolean>>;
  currentPath: string;
}

interface UploadingFile {
  id: string;
  file: File;
  progress: number;
}

const FileUpload: React.FC<FileUploadProps> = ({ setIsDialogOpen, currentPath }) => {
  const { t } = useTranslation();
  const [isUploading, setIsUploading] = useState(false);
  const { handleFileUpload } = useFileOperations(() => {
    setFiles([]);
    setIsDialogOpen(false);
  });

  const [isDragging, setIsDragging] = useState(false);
  const [files, setFiles] = useState<UploadingFile[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleDragEnter = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
  };

  const processFiles = async (newFiles: File[]) => {
    if (isUploading) return;
    setIsUploading(true);

    const uploadingFiles = newFiles.map((file) => ({
      id: Math.random().toString(36).substring(7),
      file,
      progress: 0
    }));

    setFiles((prev) => [...prev, ...uploadingFiles]);

    try {
      for (const fileObj of uploadingFiles) {
        await handleFileUpload(fileObj.file, currentPath);
        setFiles((prev) => prev.map((f) => (f.id === fileObj.id ? { ...f, progress: 100 } : f)));
      }
    } catch (error) {
      console.error('Error processing files:', error);
    } finally {
      setIsUploading(false);
    }
  };

  const handleDrop = async (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    const droppedFiles = Array.from(e.dataTransfer.files);
    await processFiles(droppedFiles);
  };

  const handleFileInput = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFiles = Array.from(e.target.files || []);
    await processFiles(selectedFiles);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const deleteFile = (id: string) => {
    setFiles((prev) => prev.filter((f) => f.id !== id));
  };

  return (
    <ResourceGuard resource="file-manager" action="create" loadingFallback={null}>
      <DialogContent>
        <DialogHeader className="space-y-8">
          <DialogTitle>{t('fileManager.upload.title')}</DialogTitle>
          <div className="mx-auto w-full">
            <div
              className={`cursor-pointer rounded-lg border-2 border-dashed p-10 text-center ${
                isDragging ? 'border-blue-500 bg-muted' : 'border-gray-300'
              }`}
              onDragEnter={handleDragEnter}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              onClick={() => fileInputRef.current?.click()}
            >
              <Upload className="mx-auto h-12 w-12 text-muted-foreground" />
              <p className="text-md mt-2 text-muted-foreground">
                {t('fileManager.upload.dragDrop')}
              </p>
            </div>
            <input
              type="file"
              ref={fileInputRef}
              className="hidden"
              onChange={handleFileInput}
              multiple
            />
            <div className="mt-4 space-y-2">
              {files.map((fileObj) => (
                <div key={fileObj.id} className="rounded bg-muted p-4 shadow">
                  <div className="flex items-center justify-between">
                    <span className="truncate">{fileObj.file.name}</span>
                    <Button variant="ghost" size="icon" onClick={() => deleteFile(fileObj.id)}>
                      <Trash className="h-4 w-4" />
                    </Button>
                  </div>
                  <Progress value={fileObj.progress} className="mt-2 h-2 bg-muted-foreground" />
                </div>
              ))}
            </div>
          </div>
        </DialogHeader>
      </DialogContent>
    </ResourceGuard>
  );
};

export default FileUpload;
