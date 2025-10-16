'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useSessionContext } from 'supertokens-auth-react/recipe/session';
import { signIn } from 'supertokens-auth-react/recipe/emailpassword';
import { ResetPasswordUsingToken } from 'supertokens-auth-react/recipe/emailpassword/prebuiltui';
import { LoginForm } from '@/components/features/login-form';
import { useTranslation } from '@/hooks/use-translation';
import { toast } from 'sonner';
import { useAppDispatch } from '@/redux/hooks';
import { initializeAuth } from '@/redux/features/users/authSlice';

export default function Auth() {
  const router = useRouter();
  const session = useSessionContext();
  const { t } = useTranslation();
  const dispatch = useAppDispatch();
  const [loaded, setLoaded] = useState(false);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (session.loading === false) {
      if (session.doesSessionExist) {
        router.push('/dashboard');
      } else {
        setLoaded(true);
      }
    }
  }, [session, router]);

  const handleEmailChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(event.target.value);
  };

  const handlePasswordChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setPassword(event.target.value);
  };

  const handleLogin = async () => {
    if (!email || !password) {
      toast.error(t('auth.login.errors.requiredFields'));
      return;
    }

    setIsLoading(true);
    try {
      const response = await signIn({
        formFields: [
          { id: 'email', value: email },
          { id: 'password', value: password }
        ]
      });

      if (response.status === 'FIELD_ERROR') {
        response.formFields.forEach(field => {
          toast.error(field.error);
        });
      } else if (response.status === 'WRONG_CREDENTIALS_ERROR') {
        toast.error(t('auth.login.errors.loginFailed'));
      } else {
        dispatch(initializeAuth() as any);
        router.push('/dashboard');
      }
    } catch (error) {
      toast.error(t('auth.login.errors.loginFailed'));
    } finally {
      setIsLoading(false);
    }
  };

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
