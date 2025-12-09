'use client';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter } from '@/components/ui/card';
import { useTranslation } from '@/hooks/use-translation';
import { LogIn, Info, HelpCircle } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { Alert, AlertDescription } from '@/components/ui/alert';

export const AdminRegistered = () => {
  const { t } = useTranslation();
  const router = useRouter();

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <Card className="w-full max-w-md">
        <CardContent className="pt-6">
          <div className="flex flex-col gap-6">
            <div className="flex flex-col items-center text-center">
              <h1 className="text-2xl font-bold">
                {t('auth.register.adminAlreadyRegistered.title' as any)}
              </h1>
              <p className="text-muted-foreground text-balance mt-4">
                {t('auth.register.adminAlreadyRegistered.description' as any)}
              </p>
            </div>
            <Alert className="border-0 bg-muted/30">
              <Info className="h-4 w-4" />
              <AlertDescription className="text-left">
                <p className="font-medium mb-2">
                  {t('auth.register.adminAlreadyRegistered.policyTitle' as any)}
                </p>
                <p className="text-sm">{t('auth.register.adminAlreadyRegistered.policy' as any)}</p>
              </AlertDescription>
            </Alert>
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground text-center">
                {t('auth.register.adminAlreadyRegistered.nextSteps' as any)}
              </p>
            </div>
          </div>
        </CardContent>
        <CardFooter className="flex flex-col gap-2">
          <Button onClick={() => router.push('/auth')} className="w-full">
            <LogIn className="mr-2 h-4 w-4" />
            {t('auth.register.adminAlreadyRegistered.goToLogin' as any)}
          </Button>
          <Button
            variant="outline"
            onClick={() => window.open('https://invite.nixopus.com', '_blank')}
            className="w-full"
          >
            <HelpCircle className="mr-2 h-4 w-4" />
            {t('auth.register.adminAlreadyRegistered.contactSupport' as any)}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
};
