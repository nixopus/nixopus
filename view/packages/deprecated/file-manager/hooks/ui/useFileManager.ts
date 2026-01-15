'use client';
import { useState, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { FileData } from '@/redux/types/files';
import { useSearchable } from '@/packages/hooks/shared/use-searchable';
import { useGetFilesInPathQuery } from '@/redux/services/file-manager/fileManagersApi';
import { useFileOperations } from '../file-operations/useOperations';
import { toast } from 'sonner';
import { useTranslation } from '@/packages/hooks/shared/use-translation';

function use_file_manager() {
  const { t } = useTranslation();
  const [currentPath, setCurrentPath] = useState('/');
  const [layout, setLayout] = useState<'grid' | 'list'>('grid');
  const [showHidden, setShowHidden] = useState(false);
  const [selectedPath, setSelectedPath] = useState('');
  const [selectedFile, setSelectedFile] = useState<FileData | undefined>();
  const [fileToCopy, setFileToCopy] = useState<FileData | undefined>();
  const [fileToMove, setFileToMove] = useState<FileData | undefined>();
  const [showCopyFeedback, setShowCopyFeedback] = useState(false);
  const [copyFeedbackMessage, setCopyFeedbackMessage] = useState('');
  const router = useRouter();
  const path = useSearchParams().get('path');

  const {
    data: files,
    isLoading,
    refetch
  } = useGetFilesInPathQuery(
    { path: currentPath },
    {
      refetchOnMountOrArgChange: true,
      refetchOnFocus: true,
      refetchOnReconnect: true
    }
  );

  const {
    handleFileUpload,
    handleCreateDirectory,
    handleMove: moveFile,
    handleCopy: copyFile,
    handleDelete,
    calculateSize,
    isSizeLoading: isCopyFileOrDirectoryLoading,
    fileSize,
    handleKeyDown,
    handleTextDoubleClick,
    handleRename,
    startRenaming,
    isEditing,
    editedFileName,
    setEditedFileName,
    isDialogOpen,
    setIsDialogOpen
  } = useFileOperations(() => {
    refetch();
  });

  const { filteredAndSortedData, searchTerm, handleSearchChange, handleSortChange, sortConfig } =
    useSearchable(files || [], ['name', 'created_at', 'updated_at', 'size', 'file_type'], {
      key: 'name',
      direction: 'asc'
    });

  const visibleFiles = showHidden
    ? filteredAndSortedData
    : filteredAndSortedData?.filter((file) => !file.is_hidden);

  useEffect(() => {
    setCurrentPath(path || '/');
  }, [path]);

  useEffect(() => {
    refetch();
  }, [currentPath, refetch]);

  const handleCopy = async (fromPath: string, toPath: string) => {
    try {
      await copyFile(fromPath, toPath);
      await refetch();
    } catch (error) {
      toast.error(t('toasts.errors.copyFile'), {
        description: error instanceof Error ? error.message : 'Unknown error'
      });
    }
  };

  const handleMove = async (fromPath: string, toPath: string) => {
    try {
      await moveFile(fromPath, toPath);
      refetch();
    } catch (error) {
      toast.error(t('toasts.errors.moveFile'), {
        description: error instanceof Error ? error.message : 'Unknown error'
      });
    }
  };

  const fileClicked = (filePath: string | number | boolean) => {
    try {
      router.push(`/file-manager?path=${encodeURIComponent(filePath)}`);
    } catch (error) {
      toast.error(t('toasts.errors.navigateToFile'), {
        description: error instanceof Error ? error.message : 'Unknown error'
      });
    }
  };

  const createNewFolder = async () => {
    const existingFolders = files?.filter((file: FileData) => file.name.startsWith('New Folder'));
    const numbers = existingFolders?.map((folder: FileData) => {
      const match = folder.name.match(/New Folder (\d+)/);
      return match ? parseInt(match[1]) : 0;
    });
    const highestNumber = numbers?.length ? Math.max(...numbers) : 0;

    const newFolderName = `New Folder ${highestNumber + 1}`;
    await handleCreateDirectory(currentPath, newFolderName);
    setSelectedPath(`${currentPath}/${newFolderName}`);
  };

  const handleKeyboardShortcuts = async (e: KeyboardEvent) => {
    if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
      return;
    }

    const isModifierKey = e.ctrlKey || e.metaKey;

    if (isModifierKey && e.key === 'c') {
      e.preventDefault();
      if (selectedFile) {
        setFileToCopy(selectedFile);
        setCopyFeedbackMessage(t('fileManager.copiedToClipboard', { name: selectedFile.name }));
        setShowCopyFeedback(true);
        setTimeout(() => {
          setShowCopyFeedback(false);
          setCopyFeedbackMessage('');
        }, 2000);
      }
    }

    if (isModifierKey && e.key === 'x') {
      e.preventDefault();
      if (selectedFile) {
        setFileToMove(selectedFile);
      }
    }

    if (isModifierKey && e.key === 'v') {
      e.preventDefault();
      if (fileToCopy) {
        const fileName = fileToCopy.name;
        const basePath = currentPath;
        let newPath = `${basePath}/${fileName}`;
        let counter = 1;

        while (files?.some((f) => f.path === newPath)) {
          const extension = fileName.includes('.')
            ? fileName.substring(fileName.lastIndexOf('.'))
            : '';
          const baseName = extension ? fileName.substring(0, fileName.lastIndexOf('.')) : fileName;
          newPath = `${basePath}/${baseName} (${counter})${extension}`;
          counter++;
        }

        try {
          await handleCopy(fileToCopy.path, newPath);
          setCopyFeedbackMessage(t('fileManager.copySuccess', { name: fileName }));
          setShowCopyFeedback(true);
          setTimeout(() => {
            setShowCopyFeedback(false);
            setCopyFeedbackMessage('');
          }, 2000);
          setFileToCopy(undefined);
        } catch (error) {
          toast.error(t('toasts.errors.copyFile'), {
            description: error instanceof Error ? error.message : 'Unknown error'
          });
        }
      } else if (fileToMove) {
        try {
          await handleMove(fileToMove.path, currentPath + '/' + fileToMove.name);
          setFileToMove(undefined);
        } catch (error) {
          toast.error(t('toasts.errors.moveFile'), {
            description: error instanceof Error ? error.message : 'Unknown error'
          });
        }
      }
    }

    if (isModifierKey && e.key === 'h') {
      e.preventDefault();
      setShowHidden(!showHidden);
    }

    if (isModifierKey && e.key === 'l') {
      e.preventDefault();
      setLayout(layout === 'grid' ? 'list' : 'grid');
    }

    if (isModifierKey && e.shiftKey && e.key === 'n') {
      e.preventDefault();
      createNewFolder();
    }

    if (e.key === 'F2' && selectedPath) {
      const file = files?.find((f: FileData) => f.path === selectedPath);
      if (file) {
        handleMove(file.path, `${file.path}/renamed`);
      }
    }
  };

  useEffect(() => {
    window.addEventListener('keydown', handleKeyboardShortcuts);
    return () => window.removeEventListener('keydown', handleKeyboardShortcuts);
  }, [
    selectedFile,
    fileToCopy,
    fileToMove,
    currentPath,
    layout,
    showHidden,
    copyFile,
    handleMove,
    refetch,
    setFileToCopy,
    setFileToMove,
    setShowHidden,
    setLayout,
    createNewFolder
  ]);

  const handleFileSelect = (path: string) => {
    setSelectedPath(path);
    setSelectedFile(files?.find((file) => file.path === path));
  };

  return {
    currentPath,
    layout,
    showHidden,
    selectedPath,
    selectedFile,
    fileToCopy,
    fileToMove,
    isCopyFileOrDirectoryLoading,
    showCopyFeedback,
    copyFeedbackMessage,
    handleFileSelect,
    handleCopy,
    handleMove,
    handleSearchChange,
    handleSortChange,
    sortConfig,
    visibleFiles,
    createNewFolder,
    fileClicked,
    isLoading,
    refetch,
    searchTerm,
    setLayout,
    setShowHidden,
    setFileToCopy,
    setFileToMove,
    setSelectedPath,
    setSelectedFile,
    files,
    handleFileUpload,
    handleDelete,
    handleKeyDown,
    handleTextDoubleClick,
    handleRename,
    startRenaming,
    isEditing,
    editedFileName,
    setEditedFileName,
    isDialogOpen,
    setIsDialogOpen
  };
}

export default use_file_manager;
