import React from 'react';
import { Grid, List } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface FileManagerLayoutsProps {
  layout: 'grid' | 'list';
  setLayout: React.Dispatch<React.SetStateAction<'grid' | 'list'>>;
}

function FileManagerLayouts({ layout, setLayout }: FileManagerLayoutsProps) {
  return (
    <div>
      <Button
        variant={'outline'}
        size="icon"
        onClick={() => setLayout(layout === 'grid' ? 'list' : 'grid')}
        className="mr-2"
      >
        {layout === 'list' ? <Grid className="h-4 w-4" /> : <List className="h-4 w-4" />}
      </Button>
    </div>
  );
}

export default FileManagerLayouts;
