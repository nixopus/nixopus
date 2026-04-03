'use client';
import { cn } from '@/lib/utils';
import { LoginForm } from '@/packages/components/login-form';
import { OtpLoginForm } from '@/packages/components/otp-login-form';
import { ForgotPasswordForm } from '@/packages/components/forgot-password-form';
import { ResetPasswordForm } from '@/packages/components/reset-password-form';
import useAuth from '@/packages/hooks/auth/use-auth';
import useOtpAuth from '@/packages/hooks/auth/use-otp-auth';
import { usePasswordLoginEnabled } from '@/packages/hooks/shared/use-config';
import { useIsAdminRegisteredQuery } from '@/redux/services/users/authApi';
import { usePathname, useSearchParams, useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import { getPluginCaptcha } from '@/plugins/registry';
import { getTurnstileSiteKey } from '@/redux/conf';

const CaptchaComponent = getPluginCaptcha();

export default function Auth() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const router = useRouter();
  const passwordLoginEnabled = usePasswordLoginEnabled();
  const { data: isAdminRegistered } = useIsAdminRegisteredQuery();
  const [captchaSiteKey, setCaptchaSiteKey] = useState('');

  useEffect(() => {
    getTurnstileSiteKey().then((key) => {
      if (key) setCaptchaSiteKey(key);
    });
  }, []);

  const isResetPasswordPage = pathname === '/auth/reset-password';
  const resetToken = searchParams.get('token');
  const showForgotPassword = isResetPasswordPage && !resetToken;
  const showResetPassword = isResetPasswordPage && !!resetToken;

  // Redirect password-related pages to main auth when password login is disabled
  const isPasswordRelatedPage = showForgotPassword || showResetPassword;
  useEffect(() => {
    if (passwordLoginEnabled === false && isPasswordRelatedPage) {
      router.replace('/auth');
    }
  }, [passwordLoginEnabled, isPasswordRelatedPage, router]);

  const {
    isLoading,
    handleEmailChange: handleEmailLoginChange,
    handleLogin,
    email: loginEmail,
    password,
    handlePasswordChange,
    loaded: loginLoaded
  } = useAuth();

  const {
    isSendingOtp,
    isVerifyingOtp,
    handleEmailChange: handleOtpEmailChange,
    handleOtpChange,
    handleSendOtp,
    handleVerifyOtp,
    handleChangeEmail,
    email: otpEmail,
    otp,
    otpSent,
    loaded: otpLoaded,
    timer,
    formatTimer
  } = useOtpAuth();

  const loaded =
    passwordLoginEnabled === null ? false : passwordLoginEnabled ? loginLoaded : otpLoaded;

  if (!loaded) {
    return (
      <div className="flex h-screen flex-col items-center justify-center bg-background">
        <div className="flex items-center gap-1.5">
          <div
            className="app-loading-dot h-1.5 w-1.5 rounded-full bg-primary/60"
            style={{ animationDelay: '0ms' }}
          />
          <div
            className="app-loading-dot h-1.5 w-1.5 rounded-full bg-primary/60"
            style={{ animationDelay: '150ms' }}
          />
          <div
            className="app-loading-dot h-1.5 w-1.5 rounded-full bg-primary/60"
            style={{ animationDelay: '300ms' }}
          />
        </div>
      </div>
    );
  }

  if (showForgotPassword && passwordLoginEnabled) {
    return (
      <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
        <div className="w-full max-w-sm md:max-w-md">
          <ForgotPasswordForm />
        </div>
      </div>
    );
  }

  if (showResetPassword && resetToken && passwordLoginEnabled) {
    return (
      <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
        <div className="w-full max-w-sm md:max-w-md">
          <ResetPasswordForm token={resetToken} />
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <div
        className={cn(
          'w-full',
          passwordLoginEnabled ? 'max-w-sm md:max-w-3xl' : 'max-w-sm md:max-w-md'
        )}
      >
        {passwordLoginEnabled ? (
          <LoginForm
            email={loginEmail}
            password={password}
            handleEmailChange={handleEmailLoginChange}
            handlePasswordChange={handlePasswordChange}
            handleLogin={handleLogin}
            isLoading={isLoading}
            hideRegistration={isAdminRegistered}
            CaptchaComponent={CaptchaComponent ?? undefined}
            captchaSiteKey={captchaSiteKey}
          />
        ) : (
          <OtpLoginForm
            email={otpEmail}
            otp={otp}
            handleEmailChange={handleOtpEmailChange}
            handleOtpChange={handleOtpChange}
            handleSendOtp={handleSendOtp}
            handleVerifyOtp={handleVerifyOtp}
            handleChangeEmail={handleChangeEmail}
            isSendingOtp={isSendingOtp}
            isVerifyingOtp={isVerifyingOtp}
            otpSent={otpSent}
            timer={timer}
            formatTimer={formatTimer}
            CaptchaComponent={CaptchaComponent ?? undefined}
            captchaSiteKey={captchaSiteKey}
          />
        )}
      </div>
    </div>
  );
}
