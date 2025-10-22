'use client';

import React from 'react';
import DashboardPageHeader from '@/components/layout/dashboard-page-header';
import useGeneralSettings from '../hooks/use-general-settings';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import AvatarSection from './components/AvatarSection';
import AccountSection from './components/AccountSection';
import SecuritySection from './components/SecuritySection';
import { useTranslation } from '@/hooks/use-translation';
import FeatureFlagsSettings from './components/FeatureFlagsSettings';
import { useRBAC } from '@/lib/rbac';
import PageLayout from '@/components/layout/page-layout';

function Page() {
  const { t, isLoading } = useTranslation();
  const { canAccessResource } = useRBAC();
  const {
    onImageChange,
    user,
    username,
    email,
    usernameError,
    usernameSuccess,
    isLoading: settingsLoading,
    handleUsernameChange,
    setUsername,
    setUsernameError,
    userSettings,
    isGettingUserSettings,
    isUpdatingFont,
    isUpdatingTheme,
    isUpdatingLanguage,
    isUpdatingAutoUpdate,
    handleThemeChange,
    handleLanguageChange,
    // handleAutoUpdateChange,
    handleFontUpdate
  } = useGeneralSettings();

  const hasFeatureFlagsReadPermission = canAccessResource('feature-flags', 'read');

  return (
    <PageLayout maxWidth="6xl" padding="md" spacing="lg">
      <DashboardPageHeader label={t('settings.title')} description={t('settings.description')} />
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <AvatarSection onImageChange={onImageChange} user={user} />
        <div className="col-span-1 lg:col-span-2">
          <Tabs defaultValue="account" className="w-full">
            <TabsList
              className={`grid w-full ${hasFeatureFlagsReadPermission ? 'grid-cols-3' : 'grid-cols-2'}`}
            >
              <TabsTrigger value="account">{t('settings.tabs.account')}</TabsTrigger>
              <TabsTrigger value="security">{t('settings.tabs.security')}</TabsTrigger>
              {hasFeatureFlagsReadPermission && (
                <TabsTrigger value="feature-flags">{t('settings.tabs.featureFlags')}</TabsTrigger>
              )}
            </TabsList>
            <AccountSection
              username={username}
              setUsername={setUsername}
              usernameError={usernameError}
              usernameSuccess={usernameSuccess}
              setUsernameError={setUsernameError}
              email={email}
              isLoading={settingsLoading}
              handleUsernameChange={handleUsernameChange}
              user={user}
              userSettings={
                userSettings || {
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
              isGettingUserSettings={isGettingUserSettings}
              isUpdatingFont={isUpdatingFont}
              isUpdatingTheme={isUpdatingTheme}
              isUpdatingLanguage={isUpdatingLanguage}
              isUpdatingAutoUpdate={isUpdatingAutoUpdate}
              handleThemeChange={handleThemeChange}
              handleLanguageChange={handleLanguageChange}
              // handleAutoUpdateChange={handleAutoUpdateChange}
              handleFontUpdate={handleFontUpdate}
            />
            <SecuritySection />
            {hasFeatureFlagsReadPermission && <FeatureFlagsSettings />}
          </Tabs>
        </div>
      </div>
    </PageLayout>
  );
}

export default Page;
