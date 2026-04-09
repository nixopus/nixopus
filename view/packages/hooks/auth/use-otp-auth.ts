import React from 'react';
import { useEffect, useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { authClient } from '@/packages/lib/auth-client';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { toast } from 'sonner';
import { useAppDispatch } from '@/redux/hooks';
import { initializeAuth } from '@/redux/features/users/authSlice';

const OTP_EMAIL_STORAGE_KEY = 'otp_login_email';
const OTP_SENT_STORAGE_KEY = 'otp_login_sent';
const OTP_TIMER_STORAGE_KEY = 'otp_timer_end';

function formatTimer(seconds: number) {
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${mins}:${secs.toString().padStart(2, '0')}`;
}

function useOtpAuth() {
  const router = useRouter();
  const { t } = useTranslation();
  const dispatch = useAppDispatch();
  const [loaded, setLoaded] = useState(false);
  const [email, setEmail] = useState(() => {
    if (typeof window !== 'undefined') {
      return sessionStorage.getItem(OTP_EMAIL_STORAGE_KEY) || '';
    }
    return '';
  });
  const [otp, setOtp] = useState('');
  const [isSendingOtp, setIsSendingOtp] = useState(false);
  const [isVerifyingOtp, setIsVerifyingOtp] = useState(false);
  const [otpSent, setOtpSent] = useState(() => {
    if (typeof window !== 'undefined') {
      return sessionStorage.getItem(OTP_SENT_STORAGE_KEY) === 'true';
    }
    return false;
  });
  const [timer, setTimer] = useState(0);

  useEffect(() => {
    setLoaded(true);
  }, []);

  useEffect(() => {
    if (otpSent && timer > 0) {
      const interval = setInterval(() => {
        setTimer((prev) => {
          if (prev <= 1) {
            clearInterval(interval);
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
      return () => clearInterval(interval);
    }
  }, [otpSent, timer]);

  useEffect(() => {
    if (typeof window !== 'undefined' && otpSent) {
      const savedTimerEnd = sessionStorage.getItem(OTP_TIMER_STORAGE_KEY);
      if (savedTimerEnd) {
        const timeLeft = Math.max(0, Math.floor((parseInt(savedTimerEnd) - Date.now()) / 1000));
        if (timeLeft > 0) {
          setTimer(timeLeft);
        } else {
          sessionStorage.removeItem(OTP_TIMER_STORAGE_KEY);
        }
      }
    }
  }, [otpSent]);

  useEffect(() => {
    if (typeof window !== 'undefined') {
      if (email) {
        sessionStorage.setItem(OTP_EMAIL_STORAGE_KEY, email);
      } else {
        sessionStorage.removeItem(OTP_EMAIL_STORAGE_KEY);
      }
    }
  }, [email]);

  useEffect(() => {
    if (typeof window !== 'undefined') {
      if (otpSent) {
        sessionStorage.setItem(OTP_SENT_STORAGE_KEY, 'true');
      } else {
        sessionStorage.removeItem(OTP_SENT_STORAGE_KEY);
        sessionStorage.removeItem(OTP_TIMER_STORAGE_KEY);
      }
    }
  }, [otpSent]);

  const handleEmailChange = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(event.target.value);
  }, []);

  const handleOtpChange = useCallback((value: string) => {
    const numericValue = value.replace(/\D/g, '');
    if (numericValue.length <= 6) {
      setOtp(numericValue);
    }
  }, []);

  const handleSendOtp = useCallback(
    async (captchaToken?: string) => {
      if (!email) {
        toast.error(t('auth.otpLogin.errors.emailRequired'));
        return;
      }

      setIsSendingOtp(true);
      try {
        const result = await authClient.emailOtp.sendVerificationOtp({
          email,
          type: 'sign-in',
          fetchOptions: captchaToken
            ? { headers: { 'x-captcha-response': captchaToken } }
            : undefined
        });

        if (result.error) {
          toast.error(result.error.message || t('auth.otpLogin.errors.sendFailed'));
        } else {
          setOtpSent(true);
          const timerEnd = Date.now() + 300 * 1000;
          setTimer(300);
          if (typeof window !== 'undefined') {
            sessionStorage.setItem(OTP_TIMER_STORAGE_KEY, timerEnd.toString());
          }
          toast.success(t('auth.otpLogin.otpSentSuccess'));
        }
      } catch (error: any) {
        toast.error(error?.message || t('auth.otpLogin.errors.sendFailed'));
      } finally {
        setIsSendingOtp(false);
      }
    },
    [email, t]
  );

  const handleVerifyOtp = useCallback(
    async (captchaToken?: string) => {
      if (!email || !otp) {
        toast.error(t('auth.otpLogin.errors.otpRequired'));
        return;
      }

      if (isVerifyingOtp) return;

      setIsVerifyingOtp(true);
      try {
        const result = await authClient.signIn.emailOtp({
          email,
          otp,
          fetchOptions: captchaToken
            ? { headers: { 'x-captcha-response': captchaToken } }
            : undefined
        });

        if (result.error) {
          if (
            result.error.message?.includes('invalid') ||
            result.error.message?.includes('incorrect')
          ) {
            toast.error(t('auth.otpLogin.errors.verifyFailed'));
            setOtp('');
          } else if (result.error.message?.includes('expired')) {
            toast.error(t('auth.otpLogin.errors.verifyFailed'));
            setOtp('');
            setOtpSent(false);
            setTimer(0);
          } else {
            toast.error(result.error.message || t('auth.otpLogin.errors.verifyFailed'));
          }
        } else {
          setOtp('');
          setOtpSent(false);
          setTimer(0);
          if (typeof window !== 'undefined') {
            sessionStorage.removeItem(OTP_EMAIL_STORAGE_KEY);
            sessionStorage.removeItem(OTP_SENT_STORAGE_KEY);
            sessionStorage.removeItem(OTP_TIMER_STORAGE_KEY);
          }
          await dispatch(initializeAuth() as any);

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

          router.push('/chats');
        }
      } catch (error: any) {
        toast.error(error?.message || t('auth.otpLogin.errors.verifyFailed'));
      } finally {
        setIsVerifyingOtp(false);
      }
    },
    [email, otp, isVerifyingOtp, dispatch, router, t]
  );

  const handleChangeEmail = useCallback(() => {
    setOtpSent(false);
    setOtp('');
    setTimer(0);
  }, []);

  return {
    handleSendOtp,
    handleVerifyOtp,
    handleChangeEmail,
    handleEmailChange,
    handleOtpChange,
    isSendingOtp,
    isVerifyingOtp,
    loaded,
    email,
    otp,
    otpSent,
    timer,
    formatTimer
  };
}

export default useOtpAuth;
