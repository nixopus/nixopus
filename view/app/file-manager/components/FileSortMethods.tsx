import React, { useState } from 'react';
import { Button } from '@/components/ui/button';
import { ArrowUpDown } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
  DropdownMenuRadioItem,
  DropdownMenuRadioGroup
} from '@/components/ui/dropdown-menu';
import { FileData } from '@/redux/types/files';

interface FileSortMethodsProps {
  files: FileData[] | [];
  onSortChange: (method: keyof FileData) => void;
}

function FileSortMethods({ files, onSortChange }: FileSortMethodsProps) {
  const [sortMethod, setSortMethod] = useState<keyof FileData | ''>('');

  const handleSortChange = (value: string) => {
    if (value === sortMethod) {
      setSortMethod('');
      onSortChange('name');
    } else {
      setSortMethod(value as keyof FileData);
      onSortChange(value as keyof FileData);
    }
  };

  return (
    <div>
      <DropdownMenu>
        <DropdownMenuTrigger>
          <Button variant="outline" size="icon" onClick={() => console.log(files)}>
            <ArrowUpDown className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuRadioGroup value={sortMethod} onValueChange={handleSortChange}>
            <DropdownMenuRadioItem value="name" className="flex cursor-pointer items-center gap-3">
              Name
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem
              value="updated_at"
              className="flex cursor-pointer items-center gap-3"
            >
              Date Modified
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem
              value="created_at"
              className="flex cursor-pointer items-center gap-3"
            >
              Date Created
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="size" className="flex cursor-pointer items-center gap-3">
              Size
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem
              value="file_type"
              className="flex cursor-pointer items-center gap-3"
            >
              Type
            </DropdownMenuRadioItem>
          </DropdownMenuRadioGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}

export default FileSortMethods;
