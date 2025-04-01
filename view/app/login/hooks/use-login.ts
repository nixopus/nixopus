import { useAppSelector } from '@/redux/hooks';
import { useLoginUserMutation } from '@/redux/services/users/authApi';
import { useRouter } from 'next/navigation';
import React, { useEffect } from 'react';
import { toast } from 'sonner';
import { useTranslation } from '@/hooks/use-translation';

function useLogin() {
  const { t } = useTranslation();
  const [email, setEmail] = React.useState('');
  const [password, setPassword] = React.useState('');
  const [loginUser, { isLoading, error }] = useLoginUserMutation();
  const router = useRouter();

  const authenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const user = useAppSelector((state) => state.auth.user);

  useEffect(() => {
    if (authenticated && user) {
      router.push('/dashboard');
    }
    return () => {};
  }, [authenticated, user, router]);

  const handleEmailChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    try {
      setEmail(event.target.value);
    } catch (error) {
      toast.error(t('toasts.errors.setEmail'), {
        description: error instanceof Error ? error.message : 'Unknown error'
      });
    }
  };

  const handlePasswordChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    try {
      setPassword(event.target.value);
    } catch (error) {
      toast.error(t('toasts.errors.setPassword'), {
        description: error instanceof Error ? error.message : 'Unknown error'
      });
    }
  };

  const handleLogin = async () => {
    try {
      if (!email || !password) {
        toast.error(t('auth.login.errors.requiredFields'));
        return;
      }
      await loginUser({ email, password }).unwrap();
      router.push('/dashboard');
    } catch (error) {
      toast.error(t('auth.login.errors.loginFailed'));
    }
  };

  return {
    email,
    password,
    handleEmailChange,
    handlePasswordChange,
    handleLogin,
    isLoading,
    error
  };
}

export default useLogin;
