'use client';

import { useEffect, useState, useRef } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useVerifyEmailMutation } from '@/redux/services/users/authApi';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Loader2 } from 'lucide-react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { toast } from 'sonner';

export default function VerifyEmailPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { t } = useTranslation();
  const [verifyEmail] = useVerifyEmailMutation();
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [message, setMessage] = useState(t('auth.verifyEmail.loading'));
  const isVerifying = useRef(false);
  const token = searchParams.get('token');

  useEffect(() => {
    if (!token) {
      setStatus('error');
      setMessage(t('auth.verifyEmail.error.invalidLink'));
      return;
    }

    if (isVerifying.current) return;

    const verify = async () => {
      isVerifying.current = true;
      try {
        await verifyEmail({ token }).unwrap();
        setStatus('success');
        setMessage(t('auth.verifyEmail.success.message'));
      } catch (error) {
        setStatus('error');
        setMessage(t('auth.verifyEmail.error.message'));
      } finally {
        isVerifying.current = false;
      }
    };

    verify();

    return () => {
      isVerifying.current = false;
    };
  }, [token]);

  useEffect(() => {
    if (status === 'success') {
      router.push('/dashboard');
      toast.success(t('auth.verifyEmail.success.message'));
    }
  }, [status]);

  return (
    <div className="flex min-h-screen items-center justify-center">
      <Card className="w-[350px]">
        <CardHeader>
          <CardTitle>{t('auth.verifyEmail.title')}</CardTitle>
          <CardDescription>{t('auth.verifyEmail.description')}</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col items-center gap-4">
          {status === 'loading' && <Loader2 className="h-8 w-8 animate-spin" />}
          <p className="text-center">{message}</p>
          {status === 'error' && (
            <Button variant="outline" onClick={() => router.push('/login')}>
              {t('auth.verifyEmail.buttons.backToLogin')}
            </Button>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
