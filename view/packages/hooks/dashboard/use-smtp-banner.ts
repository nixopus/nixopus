import React from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useRouter } from 'next/navigation';

const SMTP_BANNER_KEY = 'smtp_banner_dismissed';

export default function useSmtpBanner() {
  const { t } = useTranslation();
  const router = useRouter();
  const [isVisible, setIsVisible] = React.useState(false);

  React.useEffect(() => {
    const dismissed = localStorage.getItem(SMTP_BANNER_KEY);
    if (!dismissed) {
      setIsVisible(true);
    }
  }, []);

  const handleDismiss = () => {
    localStorage.setItem(SMTP_BANNER_KEY, 'true');
    setIsVisible(false);
  };

  const handleConfigure = () => {
    // TODO: Re-enable when notifications feature is working
    // router.push('/settings/notifications');
    // Temporarily disabled - redirect to general settings instead
    router.push('/settings/general');
  };

  return {
    t,
    isVisible,
    handleDismiss,
    handleConfigure
  };
}
