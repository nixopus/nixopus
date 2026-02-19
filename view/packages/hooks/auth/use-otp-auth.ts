import React from 'react';
import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { authClient } from '@/packages/lib/auth-client';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { toast } from 'sonner';
import { useAppDispatch } from '@/redux/hooks';
import { initializeAuth } from '@/redux/features/users/authSlice';

function useOtpAuth() {
  const router = useRouter();
  const { t } = useTranslation();
  const dispatch = useAppDispatch();
  const [loaded, setLoaded] = useState(false);
  const [email, setEmail] = useState('');
  const [otp, setOtp] = useState('');
  const [isSendingOtp, setIsSendingOtp] = useState(false);
  const [isVerifyingOtp, setIsVerifyingOtp] = useState(false);
  const [otpSent, setOtpSent] = useState(false);

  useEffect(() => {
    setLoaded(true);
  }, []);

  const handleEmailChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(event.target.value);
  };

  const handleOtpChange = (value: string) => {
    // Only allow numeric input
    const numericValue = value.replace(/\D/g, '');
    setOtp(numericValue);
  };

  const handleSendOtp = async () => {
    if (!email) {
      toast.error(t('auth.otpLogin.errors.emailRequired'));
      return;
    }

    setIsSendingOtp(true);
    try {
      const result = await authClient.emailOtp.sendVerificationOtp({
        email,
        type: 'sign-in'
      });

      if (result.error) {
        toast.error(result.error.message || t('auth.otpLogin.errors.sendFailed'));
      } else {
        setOtpSent(true);
        toast.success(t('auth.otpLogin.otpSentSuccess'));
      }
    } catch (error: any) {
      toast.error(error?.message || t('auth.otpLogin.errors.sendFailed'));
    } finally {
      setIsSendingOtp(false);
    }
  };

  const handleVerifyOtp = async () => {
    if (!email || !otp) {
      toast.error(t('auth.otpLogin.errors.otpRequired'));
      return;
    }

    setIsVerifyingOtp(true);
    try {
      const result = await authClient.signIn.emailOtp({
        email,
        otp
      });

      if (result.error) {
        toast.error(result.error.message || t('auth.otpLogin.errors.verifyFailed'));
      } else {
        await dispatch(initializeAuth() as any);

        // Check for pending organization invitation
        const pendingInvite = sessionStorage.getItem('pendingInvite');
        if (pendingInvite) {
          try {
            const inviteData = JSON.parse(pendingInvite);
            sessionStorage.removeItem('pendingInvite');
            router.push(
              `/auth/organization-invite?token=${inviteData.token}&org_id=${inviteData.orgId}&email=${inviteData.email || ''}&role=${inviteData.role || 'viewer'}`
            );
            return;
          } catch (error) {
            console.error('Error processing pending invite:', error);
          }
        }

        setTimeout(() => {
          router.push('/charts');
        }, 200);
      }
    } catch (error: any) {
      toast.error(error?.message || t('auth.otpLogin.errors.verifyFailed'));
    } finally {
      setIsVerifyingOtp(false);
    }
  };

  return {
    handleSendOtp,
    handleVerifyOtp,
    handleEmailChange,
    handleOtpChange,
    isSendingOtp,
    isVerifyingOtp,
    loaded,
    email,
    otp,
    otpSent
  };
}

export default useOtpAuth;
