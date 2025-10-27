import React from 'react'
import { useTranslation } from '@/hooks/use-translation';
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
      router.push('/settings/notifications');
    };
  
    return {
        t,
        isVisible,
        handleDismiss,
        handleConfigure
    }
}