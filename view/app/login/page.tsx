'use client';
import { LoginForm } from '@/components/features/login-form';
import { useTranslation } from '@/hooks/use-translation';
import { useLoginUserMutation, useTwoFactorLoginMutation } from '@/redux/services/users/authApi';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import { toast } from 'sonner';
import { AuthResponse } from '@/redux/types/user';
import { useDispatch, useSelector } from 'react-redux';
import { setCredentials } from '@/redux/features/users/authSlice';
import { RootState } from '@/redux/store';

export default function LoginPage() {
  const { t } = useTranslation();
  const router = useRouter();
  const dispatch = useDispatch();
  const { twoFactor } = useSelector((state: RootState) => state.auth);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [code, setCode] = useState('');
  const [loginUser, { isLoading: isLoginLoading }] = useLoginUserMutation();
  const [twoFactorLogin, { isLoading: isTwoFactorLoading }] = useTwoFactorLoginMutation();

  const handleLogin = async () => {
    try {
      const response = (await loginUser({ email, password }).unwrap()) as AuthResponse;

      if (response.temp_token) {
        dispatch(
          setCredentials({
            user: null,
            token: response.temp_token,
            tempToken: response.temp_token,
            expiresIn: response.expires_in
          })
        );
      } else if (response.access_token) {
        dispatch(
          setCredentials({
            user: response.user,
            token: response.access_token,
            refreshToken: response.refresh_token,
            expiresIn: response.expires_in
          })
        );
        router.push('/dashboard');
      }
    } catch (error) {
      toast.error(t('auth.login.errors.loginFailed'));
    }
  };

  const handleTwoFactorLogin = async () => {
    try {
      const response = (await twoFactorLogin({ email, password, code }).unwrap()) as AuthResponse;

      if (response.access_token) {
        dispatch(
          setCredentials({
            user: response.user,
            token: response.access_token,
            refreshToken: response.refresh_token,
            expiresIn: response.expires_in
          })
        );
        router.push('/dashboard');
      }
    } catch (error) {
      toast.error(t('auth.login.errors.2faFailed'));
    }
  };

  return (
    <div className="flex min-h-svh flex-col items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm md:max-w-3xl">
        <LoginForm
          email={email}
          password={password}
          handleEmailChange={(e) => setEmail(e.target.value)}
          handlePasswordChange={(e) => setPassword(e.target.value)}
          handleLogin={handleLogin}
          isLoading={isLoginLoading}
          twoFactorCode={code}
          handleTwoFactorCodeChange={(e) => setCode(e.target.value)}
          showTwoFactor={twoFactor.isRequired}
          handleTwoFactorLogin={handleTwoFactorLogin}
          isTwoFactorLoading={isTwoFactorLoading}
        />
      </div>
    </div>
  );
}
