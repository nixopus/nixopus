'use client';

import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { cn } from '@/lib/utils';
import nixopusLogo from '@/public/nixopus_logo_transparent.png';
import Link from 'next/link';
import { PasswordInputField } from '@/components/ui/password-input-field';
import { AdminRegisteredSkeleton } from './components/admin-registered-skeleton';
import { AdminRegisteredError } from './components/admin-registerd-error';
import { AdminRegistered } from './components/admin-registered';
import useRegister from './hooks/use-register';

export default function RegisterPage() {
  const {
    form,
    onSubmit,
    isLoading,
    isAdminRegistered,
    isAdminRegisteredLoading,
    isAdminRegisteredError,
    t
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
