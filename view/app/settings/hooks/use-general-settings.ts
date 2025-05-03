import { useAppSelector } from '@/redux/hooks';
import {
  useGetUserSettingsQuery,
  useUpdateFontMutation,
  useRequestPasswordResetLinkMutation,
  useUpdateAutoUpdateMutation,
  useUpdateLanguageMutation,
  useUpdateThemeMutation,
  useUpdateUserNameMutation
} from '@/redux/services/users/userApi';
import { useState } from 'react';
import { toast } from 'sonner';
import { useTranslation } from '@/hooks/use-translation';

function useGeneralSettings() {
  const { t } = useTranslation();
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

  const {
    data: userSettings,
    isLoading: isGettingUserSettings,
    refetch: refetchUserSettings
  } = useGetUserSettingsQuery();
  const [updateFont, { isLoading: isUpdatingFont }] = useUpdateFontMutation();
  const [updateTheme, { isLoading: isUpdatingTheme }] = useUpdateThemeMutation();
  const [updateLanguage, { isLoading: isUpdatingLanguage }] = useUpdateLanguageMutation();
  const [updateAutoUpdate, { isLoading: isUpdatingAutoUpdate }] = useUpdateAutoUpdateMutation();

  const handleUsernameChange = async () => {
    if (username.trim() === '') {
      setUsernameError(t('settings.account.errors.emptyUsername'));
      return;
    }

    if (username === user.username) {
      setUsernameError(t('settings.account.errors.sameUsername'));
      return;
    }

    setIsLoading(true);
    try {
      await updateUserName(username);
      setUsernameSuccess(true);
      setUsernameError('');
    } catch (error) {
      setUsernameError(t('settings.account.errors.updateFailed'));
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
      console.error(t('settings.account.errors.resetEmailFailed'), error);
    } finally {
      setIsLoading(false);
    }
  };

  const onImageChange = (imageUrl: string | null) => {
    console.log('Image URL:', imageUrl);
    toast.error(t('settings.account.errors.imageNotImplemented'));
  };

  const handleFontChange = async (fontFamily: string, fontSize: number) => {
    try {
      await updateFont({ font_family: fontFamily, font_size: fontSize });
      refetchUserSettings();
    } catch (error) {
      console.error(t('settings.account.errors.fontUpdateFailed'), error);
    }
  };

  const handleThemeChange = async (theme: string) => {
    try {
      await updateTheme({ theme });
      refetchUserSettings();
    } catch (error) {
      console.error(t('settings.account.errors.themeUpdateFailed'), error);
    }
  };

  const handleLanguageChange = async (language: string) => {
    try {
      await updateLanguage({ language });
      refetchUserSettings();
    } catch (error) {
      console.error(t('settings.account.errors.languageUpdateFailed'), error);
    }
  };

  const handleAutoUpdateChange = async (autoUpdate: boolean) => {
    try {
      await updateAutoUpdate({ auto_update: autoUpdate });
      refetchUserSettings();
    } catch (error) {
      console.error(t('settings.account.errors.autoUpdateUpdateFailed'), error);
    }
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
    user,
    userSettings,
    isGettingUserSettings,
    isUpdatingFont,
    isUpdatingTheme,
    isUpdatingLanguage,
    isUpdatingAutoUpdate,
    handleThemeChange,
    handleLanguageChange,
    handleAutoUpdateChange,
    handleFontUpdate: handleFontChange
  };
}

export default useGeneralSettings;
