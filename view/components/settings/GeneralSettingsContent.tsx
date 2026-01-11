'use client';

import AvatarSection from '@/app/settings/general/components/AvatarSection';
import AccountSection from '@/app/settings/general/components/AccountSection';
import useGeneralSettings from '@/app/settings/hooks/use-general-settings';
import { useTranslation } from '@/hooks/use-translation';
import { useAppSidebar } from '@/hooks/use-app-sidebar';
import { LogoutDialog } from '@/components/ui/logout-dialog';
import { Button } from '@/components/ui/button';
import { LogOut } from 'lucide-react';

export function GeneralSettingsContent() {
  const { t } = useTranslation();
  const settings = useGeneralSettings();
  const { showLogoutDialog, handleLogoutClick, handleLogoutConfirm, handleLogoutCancel } =
    useAppSidebar();

  return (
    <div className="flex flex-col h-full">
      <div className="space-y-6 flex-1 overflow-y-auto">
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
              handleAutoUpdateChange={settings.handleAutoUpdateChange}
              handleFontUpdate={settings.handleFontUpdate}
            />
          </div>
        </div>
      </div>
      <Button variant="secondary" onClick={handleLogoutClick} className="w-full gap-2">
        <LogOut className="h-4 w-4" />
        {t('user.menu.logout')}
      </Button>
      <LogoutDialog
        open={showLogoutDialog}
        onConfirm={handleLogoutConfirm}
        onCancel={handleLogoutCancel}
      />
    </div>
  );
}
