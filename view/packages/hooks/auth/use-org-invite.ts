import { useEffect, useState } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { authClient } from '@/packages/lib/auth-client';
import { useAppDispatch } from '@/redux/hooks';
import { initializeAuth } from '@/redux/features/users/authSlice';
import { useAcceptInviteMutation } from '@/redux/services/users/userApi';

function useOrganizationInvite() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const dispatch = useAppDispatch();
  const [acceptInvite, { isLoading: isAcceptingInvite }] = useAcceptInviteMutation();
  const orgId = searchParams.get('org_id');
  const token = searchParams.get('token'); // Invitation token from URL
  const email = searchParams.get('email');
  const role = searchParams.get('role') || 'viewer';
  const [status, setStatus] = useState<'loading' | 'success' | 'error' | 'needs-auth'>('loading');
  const [message, setMessage] = useState('');

  useEffect(() => {
    handleInvitation();
  }, []);

  const handleInvitation = async () => {
    setStatus('loading');
    setMessage('Processing your invitation...');

    try {
      // Check if user is authenticated
      const session = await authClient.getSession();
      const isAuthenticated = !!session?.data?.session;

      if (!isAuthenticated) {
        // User needs to authenticate first
        setStatus('needs-auth');
        setMessage('Please sign in to accept the invitation.');
        return;
      }

      if (!token) {
        setStatus('error');
        setMessage('Invalid invitation link. Missing token.');
        return;
      }

      // Accept the invitation via RTK Query
      await acceptInvitation();
    } catch (error: any) {
      console.error('Error processing invitation:', error);
      setStatus('error');
      setMessage(error?.message || 'An error occurred while processing your invitation.');
    }
  };

  const acceptInvitation = async () => {
    try {
      // Use RTK Query mutation to accept invitation
      await acceptInvite({
        token: token!,
        organization_id: orgId || undefined,
        role: role || undefined,
        email: email || undefined
      }).unwrap();

      // Refresh auth state to get updated organization list
      await dispatch(initializeAuth() as any);

      setStatus('success');
      setMessage('Welcome! You have successfully joined the organization.');

      // Redirect to dashboard after a short delay
      setTimeout(() => {
        router.push('/dashboard');
      }, 2000);
    } catch (error: any) {
      console.error('Error accepting invitation:', error);
      setStatus('error');
      setMessage(error?.data?.message || error?.message || 'Failed to accept invitation. Please try again.');
      throw error;
    }
  };

  const handleLoginAndAccept = () => {
    // Store invitation data in sessionStorage to retrieve after login
    if (token && orgId) {
      sessionStorage.setItem('pendingInvite', JSON.stringify({ token, orgId, email, role }));
    }
    router.push('/auth');
  };

  return {
    handleLoginAndAccept,
    isLoading: isAcceptingInvite,
    status,
    message,
    router,
    orgId
  };
}

export default useOrganizationInvite;
