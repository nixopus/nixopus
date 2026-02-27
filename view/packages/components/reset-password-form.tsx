'use client';

import { useState } from 'react';
import { Button } from '@nixopus/ui';
import { Card, CardContent } from '@nixopus/ui';
import { PasswordInputField } from '@nixopus/ui';
import { Alert, AlertDescription } from '@nixopus/ui';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

export interface ResetPasswordFormProps {
  token: string;
}

export function ResetPasswordForm({ token }: ResetPasswordFormProps) {
  const { t } = useTranslation();
  const router = useRouter();
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (password.length < 8) {
      setError(t('auth.resetPassword.errors.passwordRequirements.minLength'));
      return;
    }
    if (password !== confirmPassword) {
      setError(t('auth.resetPassword.errors.passwordMismatch'));
      return;
    }

    setIsLoading(true);
    try {
      const res = await fetch('/api/auth/reset-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ newPassword: password, token }),
        credentials: 'include'
      });

      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        setError(data?.message || t('auth.resetPassword.errors.resetFailed'));
        return;
      }

      router.push('/auth');
      router.refresh();
    } catch {
      setError(t('auth.resetPassword.errors.resetFailed'));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Card className="overflow-hidden p-0">
      <CardContent className="p-4 md:p-6">
        <div className="flex flex-col gap-4">
          <div className="flex flex-col items-center text-center">
            <h1 className="text-2xl font-bold">{t('auth.resetPassword.title')}</h1>
          </div>
          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <div className="grid gap-3">
              <PasswordInputField
                id="newPassword"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder={t('auth.resetPassword.newPassword')}
                required
              />
            </div>
            <div className="grid gap-3">
              <PasswordInputField
                id="confirmPassword"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder={t('auth.resetPassword.confirmPassword')}
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
              {isLoading ? t('auth.resetPassword.submitting') : t('auth.resetPassword.submit')}
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
