'use client';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter } from '@/components/ui/card';
import { useTranslation } from '@/hooks/use-translation';
import { LogIn } from 'lucide-react';
import { useRouter } from 'next/router';

export const AdminRegistered = () => {
  const { t } = useTranslation();
  const router = useRouter();

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <Card>
        <CardContent>
          <div className="flex flex-col items-center text-center">
            <h1 className="text-2xl font-bold">
              {t('auth.register.errors.adminAlreadyRegistered')}
            </h1>
            <p className="text-muted-foreground text-balance mt-4">
              {t('auth.register.errors.adminAlreadyRegisteredDescription')}
            </p>
          </div>
        </CardContent>
        <CardFooter className="flex justify-center">
          <Button onClick={() => router.push('/auth')}>
            <LogIn className="mr-2 h-4 w-4" />
            {t('auth.register.errors.loginButton')}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
};
