import React from 'react';
import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { authClient } from '@/packages/lib/auth-client';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { toast } from 'sonner';
import { useAppDispatch } from '@/redux/hooks';
import { initializeAuth } from '@/redux/features/users/authSlice';

function useAuth() {
  const router = useRouter();
  const { t } = useTranslation();
  const dispatch = useAppDispatch();
  const [loaded, setLoaded] = useState(false);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    // Just mark as loaded - layout.tsx handles redirects for authenticated users
    setLoaded(true);
  }, []);

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
      const result = await authClient.signIn.email({
        email,
        password
      });

      if (result.error) {
        toast.error(result.error.message || t('auth.login.errors.loginFailed'));
      } else {
        await dispatch(initializeAuth() as any);
        
        // Check for pending organization invitation
        const pendingInvite = sessionStorage.getItem('pendingInvite');
        if (pendingInvite) {
          try {
            const inviteData = JSON.parse(pendingInvite);
            sessionStorage.removeItem('pendingInvite');
            // Redirect to organization invite page to complete the process
            router.push(`/auth/organization-invite?token=${inviteData.token}&org_id=${inviteData.orgId}&email=${inviteData.email || ''}&role=${inviteData.role || 'viewer'}`);
            return;
          } catch (error) {
            console.error('Error processing pending invite:', error);
          }
        }
        
        // Wait a moment for Redux state to update after initializeAuth
        // The layout.tsx will handle redirect automatically, but we can also redirect here
        // as a fallback after ensuring state is updated
        setTimeout(() => {
          router.push('/dashboard');
        }, 200);
      }
    } catch (error: any) {
      toast.error(error?.message || t('auth.login.errors.loginFailed'));
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
