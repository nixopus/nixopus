'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useTranslation } from '@/packages/hooks/shared/use-translation';

function useGithubCallback() {
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
  }, [router, t]);

  return {
    status,
    error
  };
}

export default useGithubCallback;
