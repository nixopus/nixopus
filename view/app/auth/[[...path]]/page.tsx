'use client';
import { LoginForm } from '@/packages/components/login-form';
import { OtpLoginForm } from '@/packages/components/otp-login-form';
import useAuth from '@/packages/hooks/auth/use-auth';
import useOtpAuth from '@/packages/hooks/auth/use-otp-auth';

const passwordLoginEnabled = process.env.NEXT_PUBLIC_PASSWORD_LOGIN_ENABLED !== 'false';

export default function Auth() {
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
    email: otpEmail,
    otp,
    otpSent,
    loaded: otpLoaded
  } = useOtpAuth();

  const loaded = passwordLoginEnabled ? loginLoaded : otpLoaded;

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

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm md:max-w-3xl">
        {passwordLoginEnabled ? (
          <LoginForm
            email={loginEmail}
            password={password}
            handleEmailChange={handleEmailLoginChange}
            handlePasswordChange={handlePasswordChange}
            handleLogin={handleLogin}
            isLoading={isLoading}
          />
        ) : (
          <OtpLoginForm
            email={otpEmail}
            otp={otp}
            handleEmailChange={handleOtpEmailChange}
            handleOtpChange={handleOtpChange}
            handleSendOtp={handleSendOtp}
            handleVerifyOtp={handleVerifyOtp}
            isSendingOtp={isSendingOtp}
            isVerifyingOtp={isVerifyingOtp}
            otpSent={otpSent}
          />
        )}
      </div>
    </div>
  );
}
