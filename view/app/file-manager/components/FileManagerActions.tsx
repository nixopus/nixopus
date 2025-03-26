import React from 'react';
import {
    UploadCloudIcon,
    FolderPlusIcon,
    Trash,
    EyeIcon,
    EyeOffIcon,
} from 'lucide-react';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { DotsVerticalIcon } from '@radix-ui/react-icons';
import { Button } from '@/components/ui/button';
import { Dialog } from '@/components/ui/dialog';
import FileUpload from './FileUpload';
import { FileData } from '@/redux/types/files';
import use_menu_actions from '../hooks/use_menu_actions';

interface FileManagerActionsProps {
    refetch: () => void;
    showHidden: boolean;
    setShowHidden: React.Dispatch<React.SetStateAction<boolean>>;
    currentPath: string;
    setSelectedPath: React.Dispatch<React.SetStateAction<string>>;
    selectedPath: string;
    files: FileData[] | [];
}

function FileManagerActions({
    refetch,
    showHidden,
    setShowHidden,
    currentPath,
    setSelectedPath,
    selectedPath,
    files,
}: FileManagerActionsProps) {
    const { createNewFolder, handleFileUpload, isDialogOpen,setIsDialogOpen } = use_menu_actions({ refetch, currentPath, setSelectedPath, files })
    return (
        <div>
            <DropdownMenu>
                <DropdownMenuTrigger>
                    <Button variant={'outline'} size="icon">
                        <DotsVerticalIcon className="h-4 w-4" />
                    </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                    <DropdownMenuItem
                        className="flex cursor-pointer items-center gap-3"
                        onClick={handleFileUpload}
                    >
                        <UploadCloudIcon className="h-5 w-5 flex-shrink-0" />
                        <span>New File</span>
                    </DropdownMenuItem>
                    <DropdownMenuItem
                        className="flex cursor-pointer items-center gap-3"
                        onClick={createNewFolder}
                    >
                        <FolderPlusIcon className="h-5 w-5 flex-shrink-0" />
                        <span>New Folder</span>
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                        className="flex cursor-pointer items-center gap-3"
                        onClick={() => setShowHidden(!showHidden)}
                    >
                        {!showHidden ? (
                            <EyeIcon className="h-5 w-5 flex-shrink-0" />
                        ) : (
                            <EyeOffIcon className="h-5 w-5 flex-shrink-0" />
                        )}
                        <span className="truncate">
                            {!showHidden ? 'Show Hidden Files' : 'Hide Hidden Files'}
                        </span>
                    </DropdownMenuItem>
                    <DropdownMenuItem className="flex cursor-pointer items-center gap-3">
                        <Trash className="h-5 w-5 flex-shrink-0" />
                        <span>Trash</span>
                    </DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>

            <Dialog
                open={isDialogOpen}
                onOpenChange={(open) => {
                    setIsDialogOpen(open);
                }}
            >
                <FileUpload setIsDialogOpen={setIsDialogOpen} />
            </Dialog>
        </div>
    );
}

export default FileManagerActions;
