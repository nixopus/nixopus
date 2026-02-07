'use client';
import { LoginForm } from '@/packages/components/login-form';
import { OtpLoginForm } from '@/packages/components/otp-login-form';
import useAuth from '@/packages/hooks/auth/use-auth';
import useOtpAuth from '@/packages/hooks/auth/use-otp-auth';

const isSelfHosted = process.env.NEXT_PUBLIC_IS_SELF_HOSTED === 'true';

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

  const loaded = isSelfHosted ? loginLoaded : otpLoaded;

  if (!loaded) {
    return (
      <div className="flex h-screen flex-col items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-gray-300 border-t-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm md:max-w-3xl">
        {isSelfHosted ? (
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
