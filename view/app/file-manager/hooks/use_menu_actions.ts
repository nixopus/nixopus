import { useCreateDirectoryMutation } from '@/redux/services/file-manager/fileManagersApi';
import { FileData, FileType } from '@/redux/types/files';
import React, { useState } from 'react'

type UseMenuActionsProps = {
    refetch: () => void;
    currentPath: string;
    setSelectedPath: React.Dispatch<React.SetStateAction<string>>;
    files: FileData[] | [];
}

function use_menu_actions({ refetch, currentPath, setSelectedPath, files }: UseMenuActionsProps) {
    const [createDirectory, { isLoading, data, error }] = useCreateDirectoryMutation();
    const [isDialogOpen, setIsDialogOpen] = useState(false);

    const createNewFolder = async () => {
        if (!currentPath) return;

        const newFolders = files?.filter(
            (item) => item.file_type === FileType.Directory && /^New Folder \d+$/.test(item.name),
        );

        let highestNumber = 0;
        newFolders.forEach((folder) => {
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

    const handleFileUpload = () => {
        setIsDialogOpen(true);
    };

    return {
        createNewFolder,
        handleFileUpload,
        isDialogOpen,
        setIsDialogOpen,
    }
}

export default use_menu_actions