'use client';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter } from '@/components/ui/card';
import { useTranslation } from '@/hooks/use-translation';
import { cn } from '@/lib/utils';
import { useRouter } from 'next/navigation';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { AlertCircle, WifiOff, ServerOff, HelpCircle } from 'lucide-react';
import { useState } from 'react';

interface AdminRegisteredErrorProps {
  error?: {
    type?: 'network' | 'server' | 'unknown';
    message?: string;
    code?: string;
  };
}

export const AdminRegisteredError = ({ error }: AdminRegisteredErrorProps = {}) => {
  const { t } = useTranslation();
  const router = useRouter();
  const [showDetails, setShowDetails] = useState(false);

  const errorType = error?.type || 'unknown';
  const errorMessage = error?.message;
  const errorCode = error?.code;

  const getErrorContent = () => {
    switch (errorType) {
      case 'network':
        return {
          title: t('auth.register.errors.networkError.title' as any),
          description: t('auth.register.errors.networkError.description' as any),
          icon: <WifiOff className="h-5 w-5" />
        };
      case 'server':
        return {
          title: t('auth.register.errors.serverError.title' as any),
          description: t('auth.register.errors.serverError.description' as any),
          icon: <ServerOff className="h-5 w-5" />
        };
      default:
        return {
          title: t('auth.register.errors.somethingWentWrong'),
          description: t('auth.register.errors.loadingAdminRegistration'),
          icon: <AlertCircle className="h-5 w-5" />
        };
    }
  };

  const errorContent = getErrorContent();

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-md">
        <Card>
          <CardContent className="pt-6">
            <div className="flex flex-col gap-6">
              <div className="flex flex-col items-center text-center">
                <div className="mb-4 text-destructive">{errorContent.icon}</div>
                <h1 className="text-2xl font-bold">{errorContent.title}</h1>
                <p className="text-muted-foreground text-balance mt-4">
                  {errorContent.description}
                </p>
              </div>
              {(errorMessage || errorCode) && (
                <Alert className="border-0 bg-muted/30">
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>
                    <div className="space-y-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setShowDetails(!showDetails)}
                        className="h-auto p-0 text-xs"
                      >
                        {showDetails
                          ? t('auth.register.errors.hideDetails' as any)
                          : t('auth.register.errors.showDetails' as any)}
                      </Button>
                      {showDetails && (
                        <div className="text-xs space-y-1 font-mono bg-muted p-2 rounded">
                          {errorCode && (
                            <div>
                              <span className="font-semibold">Code:</span> {errorCode}
                            </div>
                          )}
                          {errorMessage && (
                            <div>
                              <span className="font-semibold">Message:</span> {errorMessage}
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  </AlertDescription>
                </Alert>
              )}
            </div>
          </CardContent>
          <CardFooter className="flex flex-col gap-2">
            <Button variant="outline" onClick={() => window.location.reload()} className="w-full">
              {t('auth.register.errors.tryAgain')}
            </Button>
            <Button variant="outline" onClick={() => router.push('/auth')} className="w-full">
              {t('auth.register.errors.loginButton')}
            </Button>
            <Button
              variant="ghost"
              onClick={() => window.open('https://discord.gg/nixopus', '_blank')}
              className="w-full"
            >
              <HelpCircle className="mr-2 h-4 w-4" />
              {t('auth.register.errors.contactSupport' as any)}
            </Button>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
};
