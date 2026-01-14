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

export default function RegisterPage() {
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
