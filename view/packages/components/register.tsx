'use client';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import {
  LogIn,
  Info,
  HelpCircle,
  AlertCircle,
  WifiOff,
  ServerOff,
  CheckCircle2,
  ArrowRight
} from 'lucide-react';
import { useRouter } from 'next/navigation';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Skeleton } from '@/components/ui/skeleton';
import { useState, useEffect } from 'react';
import { useSessionContext } from 'supertokens-auth-react/recipe/session';
import { TypographyH1, TypographyMuted } from '@/components/ui/typography';
import { cn } from '@/lib/utils';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { PasswordInputField } from '@/components/ui/password-input-field';
import Link from 'next/link';
import nixopusLogo from '@/public/nixopus_logo_transparent.png';
import { UseFormReturn } from 'react-hook-form';
import { RegisterForm } from '../hooks/auth/use-register';

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

export const AdminRegisteredSkeleton = () => {
  const { t } = useTranslation();

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm md:max-w-3xl">
        <div className="flex flex-col gap-6">
          <Card className="overflow-hidden p-0">
            <CardContent className="grid p-0 md:grid-cols-2">
              <div className="p-6 md:p-8">
                <div className="flex flex-col gap-6">
                  <div className="flex flex-col items-center text-center">
                    <Skeleton className="h-8 w-48" />
                    <Skeleton className="mt-4 h-4 w-64" />
                  </div>
                  <div className="text-center">
                    <p className="text-sm text-muted-foreground">
                      {t('auth.register.loadingStatus' as any)}
                    </p>
                  </div>
                  <div className="space-y-4">
                    <div className="grid gap-3">
                      <Skeleton className="h-4 w-16" />
                      <Skeleton className="h-10 w-full" />
                    </div>
                    <div className="grid gap-3">
                      <Skeleton className="h-4 w-20" />
                      <Skeleton className="h-10 w-full" />
                    </div>
                    <div className="grid gap-3">
                      <Skeleton className="h-4 w-32" />
                      <Skeleton className="h-10 w-full" />
                    </div>
                    <Skeleton className="h-10 w-full" />
                    <Skeleton className="mx-auto h-4 w-48" />
                  </div>
                </div>
              </div>
              <div className="bg-muted relative hidden md:block">
                <Skeleton className="absolute inset-0 h-full w-full" />
              </div>
            </CardContent>
          </Card>
          <Skeleton className="mx-auto h-4 w-64" />
        </div>
      </div>
    </div>
  );
};

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
                {t('auth.register.successAdmin.redirecting' as any, {
                  count: countdown.toString()
                })}
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

export const RegisterFormComponent = ({
  form,
  onSubmit,
  isLoading
}: {
  form: UseFormReturn<RegisterForm>;
  onSubmit: (data: RegisterForm) => void;
  isLoading: boolean;
}) => {
  const { t } = useTranslation();
  return (
    <div className="w-full max-w-sm md:max-w-3xl">
      <div className={cn('flex flex-col gap-6')}>
        <Card className="overflow-hidden p-0">
          <CardContent className="grid p-0 md:grid-cols-2">
            <div className="p-6 md:p-8">
              <div className="flex flex-col gap-6">
                <div className="flex flex-col items-center text-center">
                  <TypographyH1 className="text-2xl">{t('auth.register.title')}</TypographyH1>
                  <TypographyMuted className="text-balance">
                    {t('auth.register.description')}
                  </TypographyMuted>
                </div>
                <Alert className="border-0 bg-muted/30">
                  <Info className="h-4 w-4" />
                  <AlertDescription>{t('auth.register.adminInfoBanner' as any)}</AlertDescription>
                </Alert>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                  <div className="grid gap-3">
                    <Label htmlFor="email">{t('auth.email')}</Label>
                    <Input
                      id="email"
                      type="email"
                      placeholder={t('auth.register.emailPlaceholder')}
                      {...form.register('email')}
                    />
                    {form.formState.errors.email && (
                      <p className="text-sm text-destructive">
                        {form.formState.errors.email.message}
                      </p>
                    )}
                  </div>
                  <div className="grid gap-3">
                    <Label htmlFor="password">{t('auth.password')}</Label>
                    <PasswordInputField
                      id="password"
                      type="password"
                      {...form.register('password')}
                    />
                    {form.formState.errors.password && (
                      <p className="text-sm text-destructive">
                        {form.formState.errors.password.message}
                      </p>
                    )}
                  </div>
                  <div className="grid gap-3">
                    <Label htmlFor="confirmPassword">{t('auth.register.confirmPassword')}</Label>
                    <PasswordInputField
                      id="confirmPassword"
                      type="password"
                      {...form.register('confirmPassword')}
                    />

                    {form.formState.errors.confirmPassword && (
                      <p className="text-sm text-destructive">
                        {form.formState.errors.confirmPassword.message}
                      </p>
                    )}
                  </div>
                  <Button type="submit" className="w-full" disabled={isLoading}>
                    {isLoading ? t('auth.register.loading') : t('auth.register.submit')}
                  </Button>
                  <div className="text-center text-sm">
                    {t('auth.register.alreadyHaveAccount')}{' '}
                    <Link href="/auth" className="underline underline-offset-4">
                      {t('auth.login.title')}
                    </Link>
                  </div>
                </form>
              </div>
            </div>
            <div className="bg-muted relative hidden md:block">
              <img
                src={nixopusLogo.src}
                alt="Nixopus Logo"
                className="absolute inset-0 h-full w-full object-cover"
              />
            </div>
          </CardContent>
        </Card>
        <TypographyMuted className="text-center text-xs text-balance">
          {t('auth.register.terms')}{' '}
          <a
            href="https://docs.nixopus.com/license"
            className="underline underline-offset-4 hover:text-primary"
            target="_blank"
            rel="noopener noreferrer"
          >
            {t('auth.register.termsOfService')}
          </a>{' '}
          {t('auth.register.and')}{' '}
          <a
            href="https://docs.nixopus.com/privacy-policy"
            className="underline underline-offset-4 hover:text-primary"
            target="_blank"
            rel="noopener noreferrer"
          >
            {t('auth.register.privacyPolicy')}
          </a>
          .
        </TypographyMuted>
      </div>
    </div>
  );
};
