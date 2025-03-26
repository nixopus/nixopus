import { useCopyFileOrDirectoryMutation, useMoveOrRenameDirectoryMutation } from "@/redux/services/file-manager/fileManagersApi";
import { FileData } from "@/redux/types/files";

export const useFileManagerActionsHook = () => {
    const [
        moveOrRenameDirectory,
        { isLoading: isMoveOrRenameDirectoryLoading, error: moveOrRenameDirectoryError },
    ] = useMoveOrRenameDirectoryMutation();
    const [
        copyFileOrDirectory,
        { isLoading: isCopyFileOrDirectoryLoading, error: copyFileOrDirectoryError },
    ] = useCopyFileOrDirectoryMutation();
    const handleCopyFile = (
        data: FileData,
        setFileToCopy: React.Dispatch<React.SetStateAction<FileData | undefined>>,
    ) => {
        setFileToCopy(data);
    };

    const handleFileMove = async (from_path: string, to_path: string) => {
        await moveOrRenameDirectory({ from_path, to_path });
    };

    const handleFilePaste = async (from_path: string, to_path: string) => {
        await copyFileOrDirectory({ from_path, to_path });
    };

    return {
        handleCopyFile,
        handleFileMove,
        handleFilePaste,
        copyFileOrDirectoryError,
        moveOrRenameDirectoryError,
        isCopyFileOrDirectoryLoading,
        isMoveOrRenameDirectoryLoading,
    };
};
