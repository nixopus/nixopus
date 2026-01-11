import React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { AlertTriangle } from 'lucide-react';

interface AccessDeniedProps {
  title?: string;
  description?: string;
  className?: string;
}

export const AccessDenied: React.FC<AccessDeniedProps> = ({
  title,
  description,
  className = ''
}) => {
  const { t } = useTranslation();

  return (
    <div className={`flex h-full items-center justify-center ${className}`}>
      <div className="text-center">
        <AlertTriangle className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
        <h2 className="text-2xl font-bold">{title || t('common.accessDenied')}</h2>
        <p className="text-muted-foreground">{description || t('common.noPermissionView')}</p>
      </div>
    </div>
  );
};
