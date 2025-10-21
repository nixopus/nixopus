'use client';

import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { useTranslation } from '@/hooks/use-translation';
import { cn } from '@/lib/utils';
import { useRouter } from 'next/router';

export const AdminRegisteredError = () => {
  const { t } = useTranslation();
  const router = useRouter();

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm md:max-w-3xl">
        <div className={cn('flex flex-col gap-6')}>
          <Card className="overflow-hidden p-0">
            <CardContent className="p-0">
              <div className="p-6 md:p-8">
                <div className="flex flex-col gap-6">
                  <div className="flex flex-col items-center text-center">
                    <h1 className="text-2xl font-bold">
                      {t('auth.register.errors.somethingWentWrong')}
                    </h1>
                    <p className="text-muted-foreground text-balance mt-4">
                      {t('auth.register.errors.loadingAdminRegistration')}
                    </p>
                  </div>
                  <div className="flex justify-center gap-4 mt-4 mb-4">
                    <Button variant="outline" onClick={() => window.location.reload()}>
                      {t('auth.register.errors.tryAgain')}
                    </Button>
                    <Button variant="outline" onClick={() => router.push('/auth')}>
                      {t('auth.register.errors.loginButton')}
                    </Button>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
};
