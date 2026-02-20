'use client';

import React from 'react';
import { Alert, AlertDescription } from '@nixopus/ui';
import { TypographyMuted } from '@nixopus/ui';
import { Button } from '@nixopus/ui';
import { X } from 'lucide-react';
import useSshBanner from '@/packages/hooks/dashboard/use-ssh-banner';

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

  return (
    <Alert className="mb-4 border-red-500/50 bg-red-50 dark:bg-red-950/20">
      <AlertDescription className="flex items-center justify-between">
        <TypographyMuted className="text-red-800 dark:text-red-200">{getMessage()}</TypographyMuted>
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleDismiss}
            className="text-red-800 dark:text-red-200 hover:bg-red-100 dark:hover:bg-red-900/30"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
      </AlertDescription>
    </Alert>
  );
}
