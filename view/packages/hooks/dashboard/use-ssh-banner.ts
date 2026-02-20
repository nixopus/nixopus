import React from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useRouter } from 'next/navigation';

const SSH_BANNER_KEY = 'ssh_banner_dismissed';

interface SSHConnectionStatus {
  status: 'connected' | 'disconnected' | 'not_configured' | 'error';
  connected: boolean;
  message: string;
  is_configured: boolean;
}

export default function useSshBanner() {
  const { t } = useTranslation();
  const router = useRouter();
  const [isVisible, setIsVisible] = React.useState(false);
  const [isLoading, setIsLoading] = React.useState(true);
  const [sshStatus, setSshStatus] = React.useState<SSHConnectionStatus | null>(null);

  React.useEffect(() => {
    let isMounted = true;

    const checkSSHStatus = async () => {
      try {
        // Get API base URL from config with timeout
        const configController = new AbortController();
        const configTimeout = setTimeout(() => configController.abort(), 5000);

        const configResponse = await fetch('/api/config', {
          signal: configController.signal
        });
        clearTimeout(configTimeout);

        if (!configResponse.ok || !isMounted) {
          if (isMounted) setIsLoading(false);
          return;
        }

        const config = await configResponse.json();
        const apiBaseUrl = config.baseUrl || 'http://localhost:8080/api';

        // Check SSH status with timeout
        const sshController = new AbortController();
        const sshTimeout = setTimeout(() => sshController.abort(), 10000);

        const response = await fetch(`${apiBaseUrl}/v1/servers/ssh/status`, {
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json'
          },
          signal: sshController.signal
        });
        clearTimeout(sshTimeout);

        if (!isMounted) return;

        if (response.ok) {
          const data: SSHConnectionStatus = await response.json();
          if (!isMounted) return;

          setSshStatus(data);

          // Show banner if SSH is not connected or not configured
          const dismissed = localStorage.getItem(SSH_BANNER_KEY);
          if (!dismissed && (!data.connected || !data.is_configured)) {
            setIsVisible(true);
          }
        } else {
          // If API returns error, don't show banner - might be auth issue or endpoint doesn't exist
          // Just silently fail - don't block the UI
          if (isMounted) {
            setSshStatus(null);
          }
        }
      } catch (error: any) {
        // On error (including timeout), don't show banner - silently fail
        if (error.name === 'AbortError') {
          console.warn('SSH status check timed out');
        } else {
          console.error('Failed to check SSH status:', error);
        }
        if (isMounted) {
          setSshStatus(null);
        }
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    };

    checkSSHStatus();

    return () => {
      isMounted = false;
    };
  }, []);

  const handleDismiss = () => {
    localStorage.setItem(SSH_BANNER_KEY, 'true');
    setIsVisible(false);
  };

  const handleTroubleshoot = () => {
    router.push('/settings/general');
  };

  return {
    t,
    isVisible,
    isLoading,
    sshStatus,
    handleDismiss,
    handleTroubleshoot
  };
}
