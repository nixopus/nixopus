'use client';
import { ResetPasswordUsingToken } from 'supertokens-auth-react/recipe/emailpassword/prebuiltui';
import { LoginForm } from '@/components/features/login-form';
import useAuth from '../hooks/use-auth';

export default function Auth() {
  const {
    isLoading,
    handleEmailChange,
    handleLogin,
    email,
    password,
    handlePasswordChange,
    loaded
  } = useAuth();

  if (!loaded) {
    return (
      <div className="flex h-screen flex-col items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-gray-300 border-t-blue-600"></div>
      </div>
    );
  }

  const path = typeof window !== 'undefined' ? window.location.pathname : '';
  const isResetPath = path === '/auth/reset-password';

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm md:max-w-3xl">
        {isResetPath ? (
          <ResetPasswordUsingToken />
        ) : (
          <LoginForm
            email={email}
            password={password}
            handleEmailChange={handleEmailChange}
            handlePasswordChange={handlePasswordChange}
            handleLogin={handleLogin}
            isLoading={isLoading}
          />
        )}
      </div>
    </div>
  );
}
