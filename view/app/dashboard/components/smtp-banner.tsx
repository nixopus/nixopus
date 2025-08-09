'use client';

import React from 'react';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { X } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import { useRouter } from 'next/navigation';
import { TypographyMuted } from '@/components/ui/typography';

const SMTP_BANNER_KEY = 'smtp_banner_dismissed';

export function SMTPBanner() {
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

  if (!isVisible) return null;

  return (
    <Alert className="mb-4">
      <AlertDescription className="flex items-center justify-between">
        <TypographyMuted>{t('dashboard.smtpBanner.message')}</TypographyMuted>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={handleConfigure}>
            {t('dashboard.smtpBanner.configure')}
          </Button>
          <Button variant="ghost" size="sm" onClick={handleDismiss}>
            <X className="h-4 w-4" />
          </Button>
        </div>
      </AlertDescription>
    </Alert>
  );
}
