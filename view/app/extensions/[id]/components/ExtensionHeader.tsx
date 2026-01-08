'use client';

import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import type { Extension } from '@/redux/types/extension';

interface ExtensionHeaderProps {
  extension?: Extension;
  isLoading: boolean;
}

export function ExtensionHeader({ extension, isLoading }: ExtensionHeaderProps) {
  if (isLoading) {
    return <Skeleton className="h-6 w-48" />;
  }

  return (
    <div className="flex items-center gap-3">
      <div className="h-10 w-10 rounded bg-accent flex items-center justify-center text-lg">
        {extension?.icon}
      </div>
      <div>
        <div className="flex items-center gap-2">
          <div className="text-xl font-semibold">{extension?.name}</div>
          {extension?.featured && (
            <Badge
              variant="outline"
              className="text-xs bg-yellow-500/10 border-yellow-500/50 text-yellow-600 dark:text-yellow-500 font-semibold"
            >
              Featured
            </Badge>
          )}
        </div>
        <div className="text-sm text-muted-foreground">{extension?.author}</div>
      </div>
    </div>
  );
}
