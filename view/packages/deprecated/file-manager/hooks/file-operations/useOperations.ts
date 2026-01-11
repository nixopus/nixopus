import { useState } from 'react';
import { FileData } from '@/redux/types/files';
import {
  useCalculateDirectorySizeMutation,
  useDeleteDirectoryMutation,
  useMoveOrRenameDirectoryMutation,
  useUploadFileMutation,
  useCreateDirectoryMutation,
  useCopyFileOrDirectoryMutation
} from '@/redux/services/file-manager/fileManagersApi';
import { toast } from 'sonner';
import { useTranslation } from '@/packages/hooks/shared/use-translation';

export function useFileOperations(refetch: () => void) {
  const { t } = useTranslation();
  const [moveOrRenameDirectory] = useMoveOrRenameDirectoryMutation();
  const [deleteDirectory] = useDeleteDirectoryMutation();
  const [uploadFile] = useUploadFileMutation();
  const [createDirectory] = useCreateDirectoryMutation();
  const [copyFileOrDirectory] = useCopyFileOrDirectoryMutation();
  const [calculateDirectorySize, { isLoading: isSizeLoading, data: fileSize }] =
    useCalculateDirectorySizeMutation();

  const [isEditing, setIsEditing] = useState(false);
  const [editedFileName, setEditedFileName] = useState('');
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement> | File, path: string) => {
    const files = e instanceof File ? [e] : Array.from(e.target.files || []);
    const uploadResults = [];

    try {
      for (const file of files) {
        try {
          await uploadFile({ file, path });
          uploadResults.push({ success: true, file: file.name });
        } catch (error) {
          uploadResults.push({ success: false, file: file.name, error });
        }
      }

      const failedUploads = uploadResults.filter((result) => !result.success);
      if (failedUploads.length > 0) {
        const errorMessage = failedUploads.map((f) => `${f.file}: ${f.error}`).join('\n');
        toast.error(t('toasts.errors.uploadFile'), {
          description: errorMessage
        });
      } else {
        toast.success(t('toasts.success.uploadFile'));
        await refetch();
      }
    } catch (error) {
      toast.error(t('toasts.errors.uploadFile'), {
        description: error instanceof Error ? error.message : 'Unknown error'
      });
    }
  };

  const handleCreateDirectory = async (path: string, name: string) => {
    try {
      await createDirectory({ path, name });
      refetch();
    } catch (error) {
      toast.error(t('toasts.errors.createDirectory'), {
        description: error instanceof Error ? error.message : 'Unknown error'
      });
    }
  };

  const handleMove = async (from_path: string, to_path: string) => {
    try {
      await moveOrRenameDirectory({ from_path, to_path });
      refetch();
    } catch (error) {
      toast.error(t('toasts.errors.moveFile'), {
        description: error instanceof Error ? error.message : 'Unknown error'
      });
    }
  };

  const handleCopy = async (from_path: string, to_path: string) => {
    try {
      const result = await copyFileOrDirectory({ from_path, to_path });
      if (result) {
        await refetch();
        toast.success(t('toasts.success.copyFile'));
      }
    } catch (error) {
      toast.error(t('toasts.errors.copyFile'), {
        description: error instanceof Error ? error.message : 'Unknown error'
      });
    }
  };

  const handleDelete = async (path: string) => {
    try {
      await deleteDirectory({ path });
      refetch();
    } catch (error) {
      toast.error(t('toasts.errors.deleteFile'), {
        description: error instanceof Error ? error.message : 'Unknown error'
      });
    }
  };

  const calculateSize = async (path: string) => {
    try {
      await calculateDirectorySize({ path });
    } catch (error) {
      toast.error(t('toasts.errors.calculateSize'), {
        description: error instanceof Error ? error.message : 'Unknown error'
      });
    }
  };

  const startRenaming = (file: FileData) => {
    setIsEditing(true);
    setEditedFileName(file.name);
  };

  const handleRename = async (file: FileData) => {
    if (!editedFileName || editedFileName === file.name) {
      setIsEditing(false);
      return;
    }

    const newPath = file.path.replace(file.name, editedFileName);
    if (newPath !== file.path) {
      try {
        await moveOrRenameDirectory({ from_path: file.path, to_path: newPath });
        await refetch();
        setIsEditing(false);
      } catch (error) {
        toast.error(t('toasts.errors.renameFile'), {
          description: error instanceof Error ? error.message : 'Unknown error'
        });
      }
    }
  };

  const handleKeyDown = async (e: React.KeyboardEvent, file: FileData) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      await handleRename(file);
    } else if (e.key === 'Escape') {
      setIsEditing(false);
      setEditedFileName(file.name);
    }
  };

  const handleTextDoubleClick = () => {
    setIsEditing(true);
  };

  return {
    handleFileUpload,
    handleCreateDirectory,
    handleMove,
    handleCopy,
    handleDelete,
    calculateSize,
    isSizeLoading,
    fileSize,
    isEditing,
    editedFileName,
    setEditedFileName,
    isDialogOpen,
    setIsDialogOpen,
    handleRename,
    startRenaming,
    handleKeyDown,
    handleTextDoubleClick
  };
}
