import React from 'react';
import { FileData } from '@/redux/types/files';
import FileContextMenu from './DropDownMenu';

interface ActionsProps {
  refetch: () => void;
  showHidden: boolean;
  setShowHidden: React.Dispatch<React.SetStateAction<boolean>>;
  currentPath: string;
  setSelectedPath: React.Dispatch<React.SetStateAction<string>>;
  selectedPath: string;
  files: FileData[];
  canCreate: boolean;
  canUpdate: boolean;
  canDelete: boolean;
}

const Actions: React.FC<ActionsProps> = ({
  refetch,
  showHidden,
  setShowHidden,
  currentPath,
  setSelectedPath,
  selectedPath,
  files,
  canCreate,
  canUpdate,
  canDelete
}) => {
  return (
    <FileContextMenu
      refetch={refetch}
      showHidden={showHidden}
      setShowHidden={setShowHidden}
      currentPath={currentPath}
      setSelectedPath={setSelectedPath}
      selectedPath={selectedPath}
      files={files}
      canCreate={canCreate}
      canUpdate={canUpdate}
      canDelete={canDelete}
    />
  );
};

export default Actions;
