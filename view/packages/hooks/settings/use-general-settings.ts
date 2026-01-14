import { useAppSelector } from '@/redux/hooks';
import {
  useGetUserSettingsQuery,
  useUpdateFontMutation,
  useUpdateAutoUpdateMutation,
  useUpdateLanguageMutation,
  useUpdateThemeMutation,
  useUpdateUserNameMutation,
  useUpdateAvatarMutation
} from '@/redux/services/users/userApi';
import { useState } from 'react';
import { toast } from 'sonner';
import { useTranslation } from '@/packages/hooks/shared/use-translation';

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
  const [updateAvatar, { isLoading: isUpdatingAvatar }] = useUpdateAvatarMutation();

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

  const handlePasswordResetRequest = undefined as unknown as () => void;

  const onImageChange = async (imageUrl: string | null) => {
    if (!imageUrl) {
      toast.error(t('settings.account.errors.noImageSelected'));
      return;
    }

    try {
      setIsLoading(true);
      await updateAvatar({ avatarData: imageUrl }).unwrap();
      toast.success(t('settings.account.success.avatarUpdated'));
    } catch (error) {
      toast.error(t('settings.account.errors.avatarUpdateFailed'));
    } finally {
      setIsLoading(false);
    }
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
    isLoading: isLoading || isUpdatingUsername || isUpdatingAvatar,
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
