'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@nixopus/ui';
import { AlertTriangle, Home, RotateCcw } from 'lucide-react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';

interface ErrorProps {
  error: Error & { digest?: string };
  reset: () => void;
}

export default function Error({ error, reset }: ErrorProps) {
  const router = useRouter();
  const { t } = useTranslation();

  useEffect(() => {
    console.error('[Error Boundary]', error);
  }, [error]);

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <div className="flex flex-col items-center gap-6 text-center">
        <div className="text-destructive">
          <AlertTriangle className="h-12 w-12" strokeWidth={1.5} />
        </div>
        <h1 className="text-2xl font-bold">{t('common.error.title' as any)}</h1>
        <p className="text-muted-foreground text-balance max-w-md">
          {t('common.error.description' as any)}
        </p>
        {error.digest && (
          <p className="text-muted-foreground text-xs font-mono">
            {t('common.error.errorCode' as any)}: {error.digest}
          </p>
        )}
        <div className="flex flex-col gap-2 w-full max-w-xs">
          <Button onClick={reset} className="w-full">
            <RotateCcw className="mr-2 h-4 w-4" />
            {t('common.error.tryAgain' as any)}
          </Button>
          <Button variant="outline" onClick={() => router.push('/chats')} className="w-full">
            <Home className="mr-2 h-4 w-4" />
            {t('common.error.goHome' as any)}
          </Button>
        </div>
      </div>
    </div>
  );
}
