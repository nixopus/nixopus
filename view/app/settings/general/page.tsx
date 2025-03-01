'use client';

import React from 'react';
import DashboardPageHeader from '@/components/dashboard-page-header';
import useGeneralSettings from '../hooks/use-general-settings';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import AvatarSection from './components/AvatarSection';
import AccountSection from './components/AccountSection';
import SecuritySection from './components/SecuritySection';

function Page() {
  const {
    onImageChange,
    user,
    username,
    email,
    usernameError,
    usernameSuccess,
    emailSent,
    isLoading,
    handleUsernameChange,
    handlePasswordResetRequest,
    setUsername,
    setUsernameError
  } = useGeneralSettings();

  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl">
      <DashboardPageHeader label="Settings" description="Manage your account" />
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <AvatarSection onImageChange={onImageChange} user={user} />
        <div className="col-span-1 lg:col-span-2">
          <Tabs defaultValue="account" className="w-full">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="account">Account</TabsTrigger>
              <TabsTrigger value="security">Security</TabsTrigger>
            </TabsList>
            <AccountSection
              username={username}
              setUsername={setUsername}
              usernameError={usernameError}
              usernameSuccess={usernameSuccess}
              setUsernameError={setUsernameError}
              email={email}
              isLoading={isLoading}
              handleUsernameChange={handleUsernameChange}
              user={user}
            />
            <SecuritySection
              emailSent={emailSent}
              isLoading={isLoading}
              handlePasswordResetRequest={handlePasswordResetRequest}
            />
          </Tabs>
        </div>
      </div>
    </div>
  );
}

export default Page;
