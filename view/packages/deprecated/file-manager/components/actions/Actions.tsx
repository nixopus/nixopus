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
}

const Actions: React.FC<ActionsProps> = ({
  refetch,
  showHidden,
  setShowHidden,
  currentPath,
  setSelectedPath,
  selectedPath,
  files
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
    />
  );
};

export default Actions;
