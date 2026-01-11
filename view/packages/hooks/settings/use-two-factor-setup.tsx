'use client';

import { useState } from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import {
  useSetupTwoFactorMutation,
  useVerifyTwoFactorMutation,
  useDisableTwoFactorMutation
} from '@/redux/services/users/authApi';
import { useAppSelector, useAppDispatch } from '@/redux/hooks';
import { setTwoFactorEnabled } from '@/redux/features/users/authSlice';
import { toast } from 'sonner';

export function useTwoFactorSetup() {
  const { t } = useTranslation();
  const [code, setCode] = useState('');
  const [setupTwoFactor, { data: setupData }] = useSetupTwoFactorMutation();
  const [verifyTwoFactor] = useVerifyTwoFactorMutation();
  const [disableTwoFactor] = useDisableTwoFactorMutation();
  const user = useAppSelector((state) => state.auth.user);
  const dispatch = useAppDispatch();

  const handleSetup = async () => {
    try {
      const response = await setupTwoFactor().unwrap();
      dispatch(setTwoFactorEnabled(true));
      toast.success(t('settings.2fa.setupSuccess'));
    } catch (error) {
      toast.error(t('settings.2fa.setupError'));
    }
  };

  const handleVerify = async () => {
    try {
      await verifyTwoFactor({ code }).unwrap();
      toast.success(t('settings.2fa.verifySuccess'));
    } catch (error) {
      toast.error(t('settings.2fa.verifyError'));
    }
  };

  const handleDisable = async () => {
    try {
      await disableTwoFactor().unwrap();
      toast.success(t('settings.2fa.disableSuccess'));
    } catch (error) {
      toast.error(t('settings.2fa.disableError'));
    }
  };

  return {
    code,
    setCode,
    setupData,
    user,
    handleSetup,
    handleVerify,
    handleDisable
  };
}
