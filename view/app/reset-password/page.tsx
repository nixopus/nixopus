'use client';

import { useSearchParams } from 'next/navigation';
import { ResetPasswordForm } from './components/ResetPasswordForm';

export default function ResetPasswordPage() {
  const searchParams = useSearchParams();
  const token = searchParams.get('token');

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <ResetPasswordForm token={token} />
    </div>
  );
}
