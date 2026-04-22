'use client';

import { useState } from 'react';
import { Button } from '@nixopus/ui';
import { Card, CardContent } from '@nixopus/ui';
import { Input } from '@nixopus/ui';
import { Alert, AlertDescription } from '@nixopus/ui';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import Link from 'next/link';

export function ForgotPasswordForm() {
  const { t } = useTranslation();
  const [email, setEmail] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    if (!email.trim()) {
      setError(t('auth.forgotPassword.errors.emailRequired'));
      return;
    }

    setIsLoading(true);
    try {
      const redirectTo = `${typeof window !== 'undefined' ? window.location.origin : ''}/auth/reset-password`;
      const res = await fetch('/api/auth/request-password-reset', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email: email.trim(), redirectTo }),
        credentials: 'include'
      });

      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        setError(data?.message || t('auth.forgotPassword.errors.requestFailed'));
        return;
      }

      setSuccess(true);
    } catch {
      setError(t('auth.forgotPassword.errors.requestFailed'));
    } finally {
      setIsLoading(false);
    }
  };

  if (success) {
    return (
      <Card className="overflow-hidden p-0">
        <CardContent className="p-4 md:p-6">
          <div className="flex flex-col gap-4 text-center">
            <h1 className="text-2xl font-bold">{t('auth.forgotPassword.successTitle')}</h1>
            <Button variant="outline" asChild>
              <Link href="/auth">{t('auth.forgotPassword.backToLogin')}</Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="overflow-hidden p-0">
      <CardContent className="p-4 md:p-6">
        <div className="flex flex-col gap-4">
          <div className="flex flex-col items-center text-center">
            <h1 className="text-2xl font-bold">{t('auth.forgotPassword.title')}</h1>
          </div>
          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <div className="grid gap-3">
              <Input
                id="email"
                type="email"
                placeholder={t('auth.login.emailPlaceholder')}
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
              {error && (
                <Alert variant="destructive">
                  <AlertDescription className="text-xs !text-red-600 font-medium">
                    {error}
                  </AlertDescription>
                </Alert>
              )}
            </div>
            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? t('auth.forgotPassword.sending') : t('auth.forgotPassword.submit')}
            </Button>
            <div className="text-center text-sm">
              <Link href="/auth" className="underline underline-offset-4">
                {t('auth.forgotPassword.backToLogin')}
              </Link>
            </div>
          </form>
        </div>
      </CardContent>
    </Card>
  );
}
