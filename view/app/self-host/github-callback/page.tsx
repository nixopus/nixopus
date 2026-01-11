'use client';
import React, { useEffect, useState } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Loader2 } from 'lucide-react';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { useRouter } from 'next/navigation';
import { useTranslation } from '@/hooks/use-translation';

const Page = () => {
  const [status, setStatus] = useState<'processing' | 'success' | 'error'>('processing');
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();
  const { t } = useTranslation();

  useEffect(() => {
    const handleCallback = async () => {
      const params = new URLSearchParams(window.location.search);
      const installationId = params.get('installation_id');
      const setupAction = params.get('setup_action');

      if (!installationId) {
        setError(t('selfHost.githubCallback.error.invalidParams'));
        setStatus('error');
        return;
      }

      if (installationId && setupAction === 'install') {
        try {
          setStatus('success');
          console.log('installationId', installationId);
          window.history.replaceState({}, document.title, window.location.pathname);
          router.push('/self-host/create/');
        } catch (err) {
          setError(t('selfHost.githubCallback.error.installationFailed'));
          setStatus('error');
        }
      }
    };

    handleCallback();
  }, []);

  if (status === 'processing') {
    return (
      <Card className="w-[350px]">
        <CardContent className="flex flex-col items-center justify-center py-8">
          <Loader2 className="mb-4 h-8 w-8 animate-spin" />
          <p>{t('selfHost.githubCallback.processing')}</p>
        </CardContent>
      </Card>
    );
  }

  if (status === 'success') {
    return (
      <Card className="w-[350px]">
        <CardContent className="py-6">
          <Alert>
            <AlertDescription className="text-green-600">
              {t('selfHost.githubCallback.success')}
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-[350px]">
      <CardContent className="py-6">
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      </CardContent>
    </Card>
  );
};

export default Page;
