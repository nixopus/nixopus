'use client';
import { useState, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { FileData, FileType } from '@/redux/types/files';
import { useSearchable } from '@/hooks/use-searchable';
import { useCreateDirectoryMutation, useGetFilesInPathQuery } from '@/redux/services/file-manager/fileManagersApi';
import { useFileManagerActionsHook } from './use_file_actions';

function use_file_manager() {
    const [currentPath, setCurrentPath] = useState('/');
    const [layout, setLayout] = useState<'grid' | 'list'>('grid');
    const [showHidden, setShowHidden] = useState(false);
    const [selectedPath, setSelectedPath] = useState('');
    const [selectedFile, setSelectedFile] = useState<FileData | undefined>();
    const [fileToCopy, setFileToCopy] = useState<FileData | undefined>();
    const [fileToMove, setFileToMove] = useState<FileData | undefined>();
    const [createDirectory] =
        useCreateDirectoryMutation();
    const router = useRouter();
    const path = useSearchParams().get('path');
    const {
        data: files,
        isLoading,
        refetch,
    } = useGetFilesInPathQuery({ path: currentPath });
    const {
        handleFilePaste,
        handleFileMove,
        isCopyFileOrDirectoryLoading,
        isMoveOrRenameDirectoryLoading,
    } = useFileManagerActionsHook();
    const { filteredAndSortedData, searchTerm, handleSearchChange, handleSortChange } =
        useSearchable(
            files || [],
            ['name', 'created_at', 'updated_at', 'size', 'file_type'],
            { key: 'name', direction: 'asc' },
        );


    const visibleFiles = showHidden
        ? filteredAndSortedData
        : filteredAndSortedData?.filter((file) => !file.is_hidden);

    useEffect(() => {
        setCurrentPath(path || '/');
    }, [path]);

    useEffect(() => {
        refetch();
    }, [currentPath, refetch]);

    const fileClicked = (filePath: string | number | boolean) => {
        router.push(`/file-manager?path=${encodeURIComponent(filePath)}`);
    };

    const createNewFolder = async () => {
        if (!currentPath) return;

        const newFolders = files?.filter(
            (item) => item.file_type === FileType.Directory && /^New Folder \d+$/.test(item.name),
        );

        let highestNumber = 0;
        newFolders?.forEach((folder) => {
            const match = folder.name.match(/^New Folder (\d+)$/);
            if (match) {
                const number = parseInt(match[1], 10);
                if (number > highestNumber) {
                    highestNumber = number;
                }
            }
        });

        const newFolderName = `New Folder ${highestNumber + 1}`;

        await createDirectory({ path: currentPath, name: newFolderName });
        setSelectedPath(`${currentPath}/${newFolderName}`);
        refetch();
    };

    const handleFileSelect = (path: string) => {
        setSelectedPath(path);
        setSelectedFile(files?.find((file) => file.path === path));
    };



    return {
        currentPath,
        layout,
        showHidden,
        selectedPath,
        fileToCopy,
        fileToMove,
        isCopyFileOrDirectoryLoading,
        isMoveOrRenameDirectoryLoading,
        handleFileSelect,
        handleFilePaste,
        handleFileMove,
        handleSearchChange,
        handleSortChange,
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
        files
    };
}

export default use_file_manager