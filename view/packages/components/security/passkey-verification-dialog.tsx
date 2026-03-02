'use client';

import * as React from 'react';
import { DialogWrapper } from '@nixopus/ui';
import { Fingerprint } from 'lucide-react';
import { authClient } from '@/packages/lib/auth-client';
import { toast } from 'sonner';

interface PasskeyVerificationDialogProps {
  open: boolean;
  onVerified: () => void;
  onCancel: () => void;
}

export function PasskeyVerificationDialog({
  open,
  onVerified,
  onCancel
}: PasskeyVerificationDialogProps) {
  const [isVerifying, setIsVerifying] = React.useState(false);

  const handleVerify = React.useCallback(async () => {
    setIsVerifying(true);
    try {
      const res = await authClient.signIn.passkey();
      if (res?.error) {
        toast.error(res.error.message || 'Verification failed');
        return;
      }
      onVerified();
    } catch (error: any) {
      const msg = error?.message || 'Verification failed';
      if (!msg.includes('cancelled') && !msg.includes('abort')) {
        toast.error(msg);
      }
    } finally {
      setIsVerifying(false);
    }
  }, [onVerified]);

  const actions = [
    {
      label: 'Cancel',
      onClick: onCancel,
      variant: 'outline' as const,
      disabled: isVerifying
    },
    {
      label: isVerifying ? 'Verifying...' : 'Verify with Passkey',
      onClick: handleVerify,
      disabled: isVerifying
    }
  ];

  return (
    <DialogWrapper
      open={open}
      onOpenChange={(val) => !val && onCancel()}
      title="Confirm your identity"
      description="This action requires additional verification. Use your passkey to continue."
      actions={actions}
      size="sm"
      contentClassName="sm:max-w-[450px]"
    >
      <div className="flex flex-col items-center gap-4 py-4">
        <div className="rounded-full bg-muted p-4">
          <Fingerprint className="h-8 w-8 text-primary" />
        </div>
        <p className="text-sm text-muted-foreground text-center max-w-[320px]">
          You'll be prompted to use your biometric, PIN, or security key to verify it's you.
        </p>
      </div>
    </DialogWrapper>
  );
}
