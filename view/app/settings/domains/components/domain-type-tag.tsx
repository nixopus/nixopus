import React from 'react';
import { Badge } from '@/components/ui/badge';
import { useTranslation } from '@/hooks/use-translation';

interface DomainTypeTagProps {
  isWildcard: boolean;
}

export function DomainTypeTag({ isWildcard }: DomainTypeTagProps) {
  const { t } = useTranslation();

  if (isWildcard) {
    return (
      <Badge variant="secondary" className="bg-secondary/20 text-secondary-foreground rounded-full">
        {t('settings.domains.types.wildcard')}
      </Badge>
    );
  }

  return (
    <Badge variant="outline" className="bg-primary/10 text-primary rounded-full">
      {t('settings.domains.types.standard')}
    </Badge>
  );
}
