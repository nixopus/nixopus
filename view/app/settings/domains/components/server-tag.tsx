import React from 'react';
import { Badge } from '@/components/ui/badge';

interface ServerTagProps {
  name: string;
  isPrimary: boolean;
}

export function ServerTag({ name, isPrimary }: ServerTagProps) {
  return (
    <div className="flex items-center">
      {name}
      {isPrimary && (
        <Badge variant="outline" className="ml-2 bg-primary/10 text-primary">
          Primary
        </Badge>
      )}
    </div>
  );
}
