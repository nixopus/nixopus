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
import { useTranslation } from '@/packages/hooks/shared/use-translation';

interface FileSortMethodsProps {
  files: FileData[] | [];
  onSortChange: (method: keyof FileData) => void;
}

function FileSortMethods({ files, onSortChange }: FileSortMethodsProps) {
  const { t } = useTranslation();
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
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size="icon" onClick={() => console.log(files)}>
            <ArrowUpDown className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuRadioGroup value={sortMethod} onValueChange={handleSortChange}>
            <DropdownMenuRadioItem value="name" className="flex cursor-pointer items-center gap-3">
              {t('fileManager.sort.methods.name')}
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem
              value="updated_at"
              className="flex cursor-pointer items-center gap-3"
            >
              {t('fileManager.sort.methods.updated_at')}
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem
              value="created_at"
              className="flex cursor-pointer items-center gap-3"
            >
              {t('fileManager.sort.methods.created_at')}
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="size" className="flex cursor-pointer items-center gap-3">
              {t('fileManager.sort.methods.size')}
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem
              value="file_type"
              className="flex cursor-pointer items-center gap-3"
            >
              {t('fileManager.sort.methods.file_type')}
            </DropdownMenuRadioItem>
          </DropdownMenuRadioGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}

export default FileSortMethods;
