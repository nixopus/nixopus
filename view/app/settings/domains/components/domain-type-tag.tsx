import React from 'react';
import { Badge } from '@/components/ui/badge';

interface DomainTypeTagProps {
  isWildcard: boolean;
}

export function DomainTypeTag({ isWildcard }: DomainTypeTagProps) {
  if (isWildcard) {
    return (
      <Badge variant="secondary" className="bg-secondary/20 text-secondary-foreground">
        Wildcard
      </Badge>
    );
  }

  return (
    <Badge variant="outline" className="bg-primary/10 text-primary">
      Standard
    </Badge>
  );
}
