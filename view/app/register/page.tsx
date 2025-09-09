'use client';

import { useTranslation } from '@/hooks/use-translation';
import { Card, CardContent, CardFooter } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { cn } from '@/lib/utils';
import nixopusLogo from '@/public/nixopus_logo_transparent.png';
import useRegister from './hooks/use-register';
import { Skeleton } from '@/components/ui/skeleton';
import { useRouter } from 'next/navigation';
import { LogIn } from 'lucide-react';
import Link from 'next/link';
import { PasswordInputField } from '@/components/ui/password-input-field';

export default function RegisterPage() {
  const { t } = useTranslation();
  const {
    form,
    onSubmit,
    isLoading,
    isAdminRegistered,
    isAdminRegisteredLoading,
    isAdminRegisteredError
  } = useRegister();

  if (isAdminRegisteredLoading) {
    return <AdminRegisteredSkeleton />;
  }

  if (isAdminRegisteredError) {
    return <AdminRegisteredError />;
  }

  if (isAdminRegistered) {
    return <AdminRegistered />;
  }

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm md:max-w-3xl">
        <div className={cn('flex flex-col gap-6')}>
          <Card className="overflow-hidden p-0">
            <CardContent className="grid p-0 md:grid-cols-2">
              <div className="p-6 md:p-8">
                <div className="flex flex-col gap-6">
                  <div className="flex flex-col items-center text-center">
                    <h1 className="text-2xl font-bold">{t('auth.register.title')}</h1>
                    <p className="text-muted-foreground text-balance">
                      {t('auth.register.description')}
                    </p>
                  </div>
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
                      <PasswordInputField id="password" type="password" {...form.register('password')} />
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
                      <Link href="/login" className="underline underline-offset-4">
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
          <div className="text-muted-foreground text-center text-xs text-balance">
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
          </div>
        </div>
      </div>
    </div>
  );
}

const AdminRegisteredError = () => {
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
                    <Button variant="outline" onClick={() => router.push('/login')}>
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

const AdminRegisteredSkeleton = () => {
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

const AdminRegistered = () => {
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
          <Button onClick={() => router.push('/login')}>
            <LogIn className="mr-2 h-4 w-4" />
            {t('auth.register.errors.loginButton')}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
};
