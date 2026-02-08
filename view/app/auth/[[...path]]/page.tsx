'use client';
import { LoginForm } from '@/packages/components/login-form';
import useAuth from '@/packages/hooks/auth/use-auth';

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
        <LoginForm
          email={email}
          password={password}
          handleEmailChange={handleEmailChange}
          handlePasswordChange={handlePasswordChange}
          handleLogin={handleLogin}
          isLoading={isLoading}
        />
      </div>
    </div>
  );
}
