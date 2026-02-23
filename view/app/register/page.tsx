'use client';
import {
  AdminRegisteredSkeleton,
  AdminRegisteredError,
  AdminRegistrationSuccess,
  RegisterFormComponent
} from '../../packages/components/register';
import { AdminRegistered } from '../../packages/components/register';
import useRegister from '../../packages/hooks/auth/use-register';
import { useIsAdminRegisteredQuery } from '@/redux/services/users/authApi';
import { usePasswordLoginEnabled } from '@/packages/hooks/shared/use-config';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';

export default function RegisterPage() {
  const router = useRouter();
  const passwordLoginEnabled = usePasswordLoginEnabled();

  // Redirect to auth when password login is disabled (registration is password-based)
  useEffect(() => {
    if (passwordLoginEnabled === false) {
      router.replace('/auth');
    }
  }, [passwordLoginEnabled, router]);
  const {
    form,
    onSubmit,
    isLoading,
    isAdminRegistered,
    isAdminRegisteredLoading,
    isAdminRegisteredError,
    registrationSuccess
  } = useRegister();

  const { error: adminRegisteredQueryError } = useIsAdminRegisteredQuery();

  // Show loading while checking if password login is enabled
  if (passwordLoginEnabled === null) {
    return (
      <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
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

  if (passwordLoginEnabled === false) {
    return null;
  }

  if (isAdminRegisteredLoading) {
    return <AdminRegisteredSkeleton />;
  }

  if (isAdminRegisteredError) {
    const errorDetails = adminRegisteredQueryError
      ? {
          type: 'network' as const,
          message: (adminRegisteredQueryError as any)?.message,
          code: (adminRegisteredQueryError as any)?.code
        }
      : undefined;
    return <AdminRegisteredError error={errorDetails} />;
  }

  if (isAdminRegistered) {
    return <AdminRegistered />;
  }

  if (registrationSuccess) {
    return <AdminRegistrationSuccess />;
  }

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <RegisterFormComponent form={form} onSubmit={onSubmit} isLoading={isLoading} />
    </div>
  );
}
