'use client';

import { useState } from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useSendVerificationEmailMutation } from '@/redux/services/users/authApi';
import { UserSettings } from '@/redux/types/user';

interface UseAccountSectionProps {
  userSettings: UserSettings;
  handleFontUpdate: (fontFamily: string, fontSize: number) => Promise<void>;
}

export function useAccountSection({ userSettings, handleFontUpdate }: UseAccountSectionProps) {
  const { t } = useTranslation();
  const [sendVerificationEmail, { isLoading: isSendingVerification }] =
    useSendVerificationEmailMutation();
  const [verificationSent, setVerificationSent] = useState(false);
  const [verificationError, setVerificationError] = useState('');

  const handleFontChange = async (value: string) => {
    try {
      await handleFontUpdate(value, userSettings.font_size || 16);
      document.documentElement.style.setProperty('--font-sans', value);
      document.documentElement.style.setProperty(
        '--font-mono',
        value === 'geist' ? 'var(--font-geist-mono)' : value
      );
    } catch (error) {
      console.error('Failed to update font:', error);
    }
  };

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
    handleFontChange,
    handleSendVerification
  };
}
