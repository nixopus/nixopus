'use client';

import React from 'react';
import { Alert, AlertDescription } from '@nixopus/ui';
import { TypographyMuted } from '@nixopus/ui';
import { Button } from '@nixopus/ui';
import { X, Mail } from 'lucide-react';
import useSshBanner from '@/packages/hooks/dashboard/use-ssh-banner';
import type { translationKey } from '@/packages/hooks/shared/use-translation';

const SUPPORT_EMAIL = 'raghav@nixopus.com';
const SUPPORT_MAIL_SUBJECT = 'SSH Connection Issue - Unable to connect to server';
const SUPPORT_MAIL_BODY = `Hi Nixopus Support,

I'm unable to connect to my SSH server. Please ensure I have a machine connected and help me troubleshoot my SSH configuration.

Additional details:
- [Please describe your setup and any error messages you're seeing]

Thank you.`;

export function SSHBanner() {
  const { handleDismiss, t, isVisible, sshStatus, isLoading } = useSshBanner();

  // Don't render anything while loading or if there's no status
  // This ensures the banner never blocks page rendering
  if (isLoading) return null;
  if (!sshStatus) return null;

  // Don't render if not visible (dismissed or SSH is connected)
  if (!isVisible) return null;

  // Only show banner if SSH is not connected or not configured
  if (sshStatus.connected && sshStatus.is_configured) return null;

  const getMessage = () => {
    if (!sshStatus.is_configured) {
      return t('dashboard.sshBanner.notConfigured');
    }
    if (!sshStatus.connected) {
      return t('dashboard.sshBanner.notConnected');
    }
    return t('dashboard.sshBanner.message');
  };

  // Show Contact Support when SSH is configured but not connected (no machine connected)
  const showContactSupport = sshStatus.is_configured && !sshStatus.connected;
  const contactSupportUrl = `mailto:${SUPPORT_EMAIL}?subject=${encodeURIComponent(SUPPORT_MAIL_SUBJECT)}&body=${encodeURIComponent(SUPPORT_MAIL_BODY)}`;

  return (
    <Alert className="mb-4 border border-red-500/50 dark:border-red-800/50">
      <AlertDescription className="flex items-center justify-between">
        <TypographyMuted>{getMessage()}</TypographyMuted>
        <div className="flex items-center gap-2">
          {showContactSupport && (
            <Button variant="outline" size="sm" asChild>
              <a href={contactSupportUrl}>
                <Mail className="h-4 w-4 mr-1.5" />
                {t('dashboard.sshBanner.contactSupport' as translationKey)}
              </a>
            </Button>
          )}
          <Button variant="ghost" size="sm" onClick={handleDismiss}>
            <X className="h-4 w-4" />
          </Button>
        </div>
      </AlertDescription>
    </Alert>
  );
}
