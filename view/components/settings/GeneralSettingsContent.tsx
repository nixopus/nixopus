'use client';

import AvatarSection from '@/app/settings/general/components/AvatarSection';
import AccountSection from '@/app/settings/general/components/AccountSection';
import useGeneralSettings from '@/app/settings/hooks/use-general-settings';
import { useTranslation } from '@/hooks/use-translation';

export function GeneralSettingsContent() {
  const { t } = useTranslation();
  const settings = useGeneralSettings();

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-semibold">{t('settings.title')}</h2>
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <AvatarSection onImageChange={settings.onImageChange} user={settings.user!} />
        <div className="col-span-1 lg:col-span-2">
          <AccountSection
            username={settings.username}
            setUsername={settings.setUsername}
            usernameError={settings.usernameError}
            usernameSuccess={settings.usernameSuccess}
            setUsernameError={settings.setUsernameError}
            email={settings.email}
            isLoading={settings.isLoading}
            handleUsernameChange={settings.handleUsernameChange}
            user={settings.user!}
            userSettings={
              settings.userSettings || {
                id: '0',
                user_id: '0',
                font_family: 'outfit',
                font_size: 16,
                language: 'en',
                theme: 'light',
                auto_update: true,
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString()
              }
            }
            isGettingUserSettings={settings.isGettingUserSettings}
            isUpdatingFont={settings.isUpdatingFont}
            isUpdatingTheme={settings.isUpdatingTheme}
            isUpdatingLanguage={settings.isUpdatingLanguage}
            isUpdatingAutoUpdate={settings.isUpdatingAutoUpdate}
            handleThemeChange={settings.handleThemeChange}
            handleLanguageChange={settings.handleLanguageChange}
            handleFontUpdate={settings.handleFontUpdate}
          />
        </div>
      </div>
    </div>
  );
}
