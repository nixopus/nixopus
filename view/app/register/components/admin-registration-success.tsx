'use client';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { useTranslation } from '@/hooks/use-translation';
import { CheckCircle2, ArrowRight } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { useSessionContext } from 'supertokens-auth-react/recipe/session';
import { useEffect, useState } from 'react';

export const AdminRegistrationSuccess = () => {
  const { t } = useTranslation();
  const router = useRouter();
  const session = useSessionContext();
  const [countdown, setCountdown] = useState(3);

  // User is already logged in after registration, so redirect to dashboard
  useEffect(() => {
    if (!session.loading) {
      const sessionExists = 'doesSessionExist' in session ? session.doesSessionExist : false;
      if (sessionExists) {
        const timer = setInterval(() => {
          setCountdown((prev) => {
            if (prev <= 1) {
              clearInterval(timer);
              router.push('/dashboard');
              return 0;
            }
            return prev - 1;
          });
        }, 1000);

        return () => clearInterval(timer);
      }
    }
  }, [session, router]);

  const handleGoToDashboard = () => {
    router.push('/dashboard');
  };

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-4">
            <CheckCircle2 className="h-16 w-16 text-green-500" />
          </div>
          <CardTitle className="text-2xl font-bold">
            {t('auth.register.successAdmin.title' as any)}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-4 text-center">
            <p className="text-muted-foreground text-balance">
              {t('auth.register.successAdmin.message' as any)}
            </p>
            <div className="bg-muted rounded-lg p-4 text-left space-y-2">
              <p className="text-sm font-medium">
                {t('auth.register.successAdmin.whatsNext' as any)}
              </p>
              <ul className="text-sm text-muted-foreground space-y-1 list-disc list-inside">
                <li>{t('auth.register.successAdmin.nextStep1' as any)}</li>
                <li>{t('auth.register.successAdmin.nextStep2' as any)}</li>
                <li>{t('auth.register.successAdmin.nextStep3' as any)}</li>
              </ul>
            </div>
            {countdown > 0 && (
              <p className="text-sm text-muted-foreground">
                {t('auth.register.successAdmin.redirecting' as any, { count: countdown.toString() })}
              </p>
            )}
          </div>
        </CardContent>
        <CardFooter className="flex justify-center">
          <Button onClick={handleGoToDashboard} className="w-full">
            <ArrowRight className="mr-2 h-4 w-4" />
            {t('auth.register.successAdmin.goToDashboard' as any)}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
};

