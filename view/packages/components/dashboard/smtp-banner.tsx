'use client';

import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { X } from 'lucide-react';
import { TypographyMuted } from '@/components/ui/typography';
import useSmtpBanner from '@/packages/hooks/dashboard/use-smtp-banner';

export function SMTPBanner() {
  const { handleDismiss, handleConfigure, t, isVisible } = useSmtpBanner();

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
