'use client';

import { useRouter } from 'next/navigation';
import { Button } from '@nixopus/ui';
import { ArrowLeft, Home, FileQuestion } from 'lucide-react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';

export default function NotFound() {
  const router = useRouter();
  const { t } = useTranslation();

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <div className="flex flex-col items-center gap-6 text-center">
        <div className="text-muted-foreground">
          <FileQuestion className="h-12 w-12" strokeWidth={1.5} />
        </div>
        <p className="text-7xl font-bold tracking-tighter text-foreground">
          {t('common.notFound.code' as any)}
        </p>
        <h1 className="text-2xl font-bold">{t('common.notFound.title' as any)}</h1>
        <p className="text-muted-foreground text-balance">
          {t('common.notFound.description' as any)}
        </p>
        <div className="flex flex-col gap-2 w-full max-w-xs">
          <Button onClick={() => router.push('/chats')} className="w-full">
            <Home className="mr-2 h-4 w-4" />
            {t('common.notFound.goHome' as any)}
          </Button>
          <Button variant="outline" onClick={() => router.back()} className="w-full">
            <ArrowLeft className="mr-2 h-4 w-4" />
            {t('common.notFound.goBack' as any)}
          </Button>
        </div>
      </div>
    </div>
  );
}
