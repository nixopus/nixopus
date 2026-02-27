'use client';

import { useState } from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useSendVerificationEmailMutation } from '@/redux/services/users/authApi';
import { UserSettings } from '@/redux/types/user';

interface UseAccountSectionProps {
  userSettings: UserSettings;
}

export function useAccountSection({ userSettings }: UseAccountSectionProps) {
  const { t } = useTranslation();
  const [sendVerificationEmail, { isLoading: isSendingVerification }] =
    useSendVerificationEmailMutation();
  const [verificationSent, setVerificationSent] = useState(false);
  const [verificationError, setVerificationError] = useState('');

  const handleSendVerification = async () => {
    try {
      await sendVerificationEmail().unwrap();
      setVerificationSent(true);
      setVerificationError('');
    } catch (error) {
      setVerificationError(t('settings.account.email.notVerified.error'));
    }
  };

  return {
    isSendingVerification,
    verificationSent,
    verificationError,
    handleSendVerification
  };
}
