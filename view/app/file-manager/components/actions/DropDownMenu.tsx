import React from 'react';
import { Button } from '@/components/ui/button';
import { Upload, FolderPlus, Eye, EyeOff, MoreVertical } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { DialogWrapper } from '@/components/ui/dialog-wrapper';
import FileUpload from './Upload';
import { FileData } from '@/redux/types/files';
import { useCreateDirectoryMutation } from '@/redux/services/file-manager/fileManagersApi';
import { useTranslation } from '@/hooks/use-translation';
import { toast } from 'sonner';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';

interface DropDownMenuProps {
  refetch: () => void;
  showHidden: boolean;
  setShowHidden: React.Dispatch<React.SetStateAction<boolean>>;
  currentPath: string;
  setSelectedPath: React.Dispatch<React.SetStateAction<string>>;
  selectedPath: string;
  files: FileData[];
}

const FileManagerDropDownMenu: React.FC<DropDownMenuProps> = ({
  refetch,
  showHidden,
  setShowHidden,
  currentPath,
  setSelectedPath,
  selectedPath,
  files
}) => {
  const { t } = useTranslation();
  const [isUploadOpen, setIsUploadOpen] = React.useState(false);
  const [createDirectory] = useCreateDirectoryMutation();

  const handleCreateDirectory = async () => {
    try {
      const newFolderNumber = files.filter((f) => f.name.startsWith('New Folder')).length + 1;
      const newFolderName = `New Folder ${newFolderNumber}`;
      await createDirectory({ path: currentPath, name: newFolderName });
      refetch();
    } catch (error) {
      console.error('Failed to create directory:', error);
      toast.error(t('fileManager.actions.errors.createDirectory'));
    }
  };

  return (
    <div className="flex items-center space-x-2">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size="icon">
            <MoreVertical className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <ResourceGuard resource="file-manager" action="create">
            <DialogWrapper
              open={isUploadOpen}
              onOpenChange={setIsUploadOpen}
              trigger={
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <Upload className="mr-2 h-4 w-4" />
                  {t('fileManager.actions.upload')}
                </DropdownMenuItem>
              }
              size="lg"
            >
              <FileUpload setIsDialogOpen={setIsUploadOpen} currentPath={currentPath} />
            </DialogWrapper>

            <DropdownMenuItem onSelect={handleCreateDirectory}>
              <FolderPlus className="mr-2 h-4 w-4" />
              {t('fileManager.actions.createDirectory')}
            </DropdownMenuItem>
          </ResourceGuard>

          <DropdownMenuItem onSelect={() => setShowHidden(!showHidden)}>
            {showHidden ? (
              <>
                <EyeOff className="mr-2 h-4 w-4" />
                {t('fileManager.actions.hideHidden')}
              </>
            ) : (
              <>
                <Eye className="mr-2 h-4 w-4" />
                {t('fileManager.actions.showHidden')}
              </>
            )}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
};

export default FileManagerDropDownMenu;
