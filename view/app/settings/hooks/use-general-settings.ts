import { useAppSelector } from '@/redux/hooks';
import {
  useRequestPasswordResetLinkMutation,
  useUpdateUserNameMutation
} from '@/redux/services/users/userApi';
import { useState } from 'react';
import { toast } from 'sonner';

function useGeneralSettings() {
  const user = useAppSelector((state) => state.auth.user);

  const [username, setUsername] = useState(user?.username || '');
  const [usernameError, setUsernameError] = useState('');
  const [usernameSuccess, setUsernameSuccess] = useState(false);

  const [email, setEmail] = useState(user?.email || '');
  const [emailSent, setEmailSent] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [updateUserName, { isLoading: isUpdatingUsername }] = useUpdateUserNameMutation();
  const [requestPasswordResetLink, { isLoading: isRequestingPasswordReset }] =
    useRequestPasswordResetLinkMutation();

  const handleUsernameChange = async () => {
    if (username.trim() === '') {
      setUsernameError('Username cannot be empty');
      return;
    }

    if (username === user.username) {
      setUsernameError('Please enter a different username');
      return;
    }

    setIsLoading(true);
    try {
      await updateUserName(username);
      setUsernameSuccess(true);
      setUsernameError('');
    } catch (error) {
      setUsernameError('Failed to update username. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handlePasswordResetRequest = async () => {
    setIsLoading(true);
    try {
      await requestPasswordResetLink();
      setEmailSent(true);
    } catch (error) {
      console.error('Failed to send reset email', error);
    } finally {
      setIsLoading(false);
    }
  };

  const onImageChange = (imageUrl: string | null) => {
    console.log('Image URL:', imageUrl);
    toast.error('Not yet implemented');
  };

  return {
    onImageChange,
    username,
    usernameError,
    usernameSuccess,
    email,
    emailSent,
    isLoading: isLoading || isUpdatingUsername || isRequestingPasswordReset,
    handleUsernameChange,
    handlePasswordResetRequest,
    setUsername,
    setUsernameError,
    user
  };
}

export default useGeneralSettings;
