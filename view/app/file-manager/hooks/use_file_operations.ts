import { useState, useEffect } from 'react';
import { FileData } from '@/redux/types/files';
import { useCalculateDirectorySizeMutation, useDeleteDirectoryMutation, useMoveOrRenameDirectoryMutation } from '@/redux/services/file-manager/fileManagersApi';

export const useFileOperations = (file: FileData, refetch: () => void) => {
    const [moveOrRenameDirectory] = useMoveOrRenameDirectoryMutation();
    const [deleteDirectory] = useDeleteDirectoryMutation();
    const [calculateDirectorySize, { isLoading: isSizeLoading, data: fileSize }] =
        useCalculateDirectorySizeMutation();

    const [isEditing, setIsEditing] = useState(false);
    const [editedFileName, setEditedFileName] = useState(file.name);
    const [isDialogOpen, setIsDialogOpen] = useState(false);

    useEffect(() => {
        if (isDialogOpen) {
            calculateDirectorySize({ path: file.path });
        }
    }, [isDialogOpen, file.path, calculateDirectorySize]);

    const handleRename = async () => {
        if (editedFileName !== file.name && editedFileName.trim() !== '') {
            const from_path = file.path;
            const to_path = from_path.replace(file.name, editedFileName);
            try {
                await moveOrRenameDirectory({ from_path, to_path });
                refetch();
            } catch (error) {
                console.error('Error renaming file:', error);
                setEditedFileName(file.name);
            }
        } else {
            setEditedFileName(file.name);
        }
        setIsEditing(false);
    };

    const onDeleteFolder = async () => {
        try {
            await deleteDirectory({ path: file.path });
            refetch();
        } catch (error) {
            console.error('Error deleting file:', error);
        }
    };

    const startRenaming = () => {
        setIsEditing(true);
        setEditedFileName(file.name);
    };


    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            handleRename();
        } else if (e.key === 'Escape') {
            setIsEditing(false);
            setEditedFileName(file.name);
        }
    };

    const handleTextDoubleClick = (e: React.MouseEvent) => {
        e.stopPropagation();
        setIsEditing(true);
    };

    return {
        isEditing,
        setIsEditing,
        editedFileName,
        setEditedFileName,
        isDialogOpen,
        setIsDialogOpen,
        isSizeLoading,
        fileSize,
        handleRename,
        onDeleteFolder,
        startRenaming,
        handleKeyDown,
        handleTextDoubleClick,
    };
};
