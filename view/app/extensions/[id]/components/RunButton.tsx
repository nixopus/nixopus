'use client';

import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useTranslation } from '@/hooks/use-translation';
import type { Extension } from '@/redux/types/extension';

interface RunButtonProps {
  extension?: Extension;
  isLoading: boolean;
  isRunning: boolean;
  onClick: () => void;
}

export function RunButton({ extension, isLoading, isRunning, onClick }: RunButtonProps) {
  const { t } = useTranslation();

  if (isLoading) {
    return <Skeleton className="h-9 w-28" />;
  }

  const buttonText =
    extension?.extension_type === 'install'
      ? t('extensions.install') || 'Install'
      : t('extensions.run') || 'Run';

  return (
    <Button className="min-w-[112px]" onClick={onClick} disabled={!extension || isRunning}>
      {buttonText}
    </Button>
  );
}
