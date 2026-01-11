import React from 'react';
import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useSessionContext } from 'supertokens-auth-react/recipe/session';
import { signIn } from 'supertokens-auth-react/recipe/emailpassword';
import { useTranslation } from '@/hooks/use-translation';
import { toast } from 'sonner';
import { useAppDispatch } from '@/redux/hooks';
import { initializeAuth } from '@/redux/features/users/authSlice';

function useAuth() {
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
        response.formFields.forEach((field) => {
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
  return {
    handleLogin,
    handleEmailChange,
    handlePasswordChange,
    isLoading,
    loaded,
    email,
    password
  };
}

export default useAuth;
