'use client';

import React, { useState } from 'react';
import { Mail, User, CheckCircle, AlertCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { TabsContent } from '@/components/ui/tabs';
import { Separator } from '@/components/ui/separator';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { UserSettings, User as UserType } from '@/redux/types/user';
// import { ModeToggler } from '@/components/ui/theme-toggler';
import { useSendVerificationEmailMutation } from '@/redux/services/users/authApi';
import { useTranslation } from '@/hooks/use-translation';
import { LanguageSwitcher } from '@/components/language-switcher';
import { SelectWrapper, SelectOption } from '@/components/ui/select-wrapper';
import { Switch } from '@/components/ui/switch';
import { RBACGuard } from '@/components/rbac/RBACGuard';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';

interface AccountSectionProps {
  username: string;
  setUsername: (username: string) => void;
  usernameError: string;
  usernameSuccess: boolean;
  setUsernameError: (error: string) => void;
  email: string;
  isLoading: boolean;
  handleUsernameChange: () => void;
  user: UserType;
  userSettings: UserSettings;
  isGettingUserSettings: boolean;
  isUpdatingFont: boolean;
  isUpdatingTheme: boolean;
  isUpdatingLanguage: boolean;
  isUpdatingAutoUpdate: boolean;
  handleThemeChange: (theme: string) => void;
  handleLanguageChange: (language: string) => void;
  // handleAutoUpdateChange: (autoUpdate: boolean) => void;
  handleFontUpdate: (fontFamily: string, fontSize: number) => Promise<void>;
}

function AccountSection({
  username,
  setUsername,
  usernameError,
  usernameSuccess,
  setUsernameError,
  email,
  isLoading,
  handleUsernameChange,
  user,
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
}: AccountSectionProps) {
  const { t } = useTranslation();
  const [sendVerificationEmail, { isLoading: isSendingVerification }] =
    useSendVerificationEmailMutation();
  const [verificationSent, setVerificationSent] = useState(false);
  const [verificationError, setVerificationError] = useState('');

  const handleFontChange = async (value: string) => {
    try {
      await handleFontUpdate(value, userSettings.font_size || 16);
      document.documentElement.style.setProperty('--font-sans', value);
      document.documentElement.style.setProperty(
        '--font-mono',
        value === 'geist' ? 'var(--font-geist-mono)' : value
      );
    } catch (error) {
      console.error('Failed to update font:', error);
    }
  };

  const handleSendVerification = async () => {
    try {
      await sendVerificationEmail().unwrap();
      setVerificationSent(true);
      setVerificationError('');
    } catch (error) {
      setVerificationError(t('settings.account.email.notVerified.error'));
    }
  };

  return (
    <TabsContent value="account" className="space-y-4 mt-4">
      <Card>
        <CardHeader>
          <TypographySmall>{t('settings.account.title')}</TypographySmall>
          <TypographyMuted>{t('settings.account.description')}</TypographyMuted>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-2">
            <Label htmlFor="username" className="flex items-center gap-2">
              <User size={16} />
              {t('settings.account.username.label')}
            </Label>
            <div className="flex gap-2">
              <Input
                id="username"
                value={username}
                onChange={(e) => {
                  setUsername(e.target.value);
                  setUsernameError('');
                }}
                placeholder={t('settings.account.username.placeholder')}
              />
              <RBACGuard resource="user" action="update">
                <Button
                  onClick={handleUsernameChange}
                  disabled={isLoading || username === user?.username}
                >
                  {t('settings.account.username.update')}
                </Button>
              </RBACGuard>
            </div>

            {usernameError && (
              <TypographySmall className="text-red-500">{usernameError}</TypographySmall>
            )}

            {usernameSuccess && (
              <Alert variant="default">
                <CheckCircle className="h-4 w-4" />
                <AlertTitle>Success</AlertTitle>
                <AlertDescription>{t('settings.account.username.success')}</AlertDescription>
              </Alert>
            )}
          </div>

          <Separator />

          <div className="space-y-2">
            <Label htmlFor="email" className="flex items-center gap-2">
              <Mail size={16} />
              {t('settings.account.email.label')}
            </Label>
            <div className="flex flex-col gap-2">
              <Input id="email" value={email} readOnly disabled className="bg-muted/50" />
              {user && !user.is_verified && (
                <div className="space-y-2">
                  <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>{t('settings.account.email.notVerified.title')}</AlertTitle>
                    <AlertDescription>
                      {t('settings.account.email.notVerified.description')}
                    </AlertDescription>
                  </Alert>
                  <RBACGuard resource="user" action="update">
                    <Button
                      onClick={handleSendVerification}
                      disabled={isSendingVerification || verificationSent}
                      variant="outline"
                      className="w-full"
                    >
                      {isSendingVerification
                        ? t('settings.account.email.notVerified.sending')
                        : verificationSent
                          ? t('settings.account.email.notVerified.sent')
                          : t('settings.account.email.notVerified.sendButton')}
                    </Button>
                  </RBACGuard>
                  {verificationError && (
                    <TypographySmall className="text-red-500">{verificationError}</TypographySmall>
                  )}
                  {verificationSent && (
                    <Alert variant="default">
                      <CheckCircle className="h-4 w-4" />
                      <AlertTitle>{t('settings.account.email.notVerified.sent')}</AlertTitle>
                      <AlertDescription>
                        {t('settings.account.email.notVerified.checkEmail')}
                      </AlertDescription>
                    </Alert>
                  )}
                </div>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <TypographySmall>{t('settings.account.preferences.title')}</TypographySmall>
          <TypographyMuted>{t('settings.account.preferences.description')}</TypographyMuted>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex justify-between items-center">
            <TypographyMuted>{t('settings.preferences.font')}</TypographyMuted>
            <RBACGuard resource="user" action="update">
              <SelectWrapper
                value={userSettings.font_family || 'outfit'}
                onValueChange={handleFontChange}
                options={[
                  { value: 'geist', label: t('settings.preferences.fontOptions.geist') },
                  { value: 'inter', label: t('settings.preferences.fontOptions.inter') },
                  { value: 'roboto', label: t('settings.preferences.fontOptions.roboto') },
                  { value: 'poppins', label: t('settings.preferences.fontOptions.poppins') },
                  { value: 'montserrat', label: t('settings.preferences.fontOptions.montserrat') },
                  { value: 'space-grotesk', label: t('settings.preferences.fontOptions.spaceGrotesk') },
                  { value: 'outfit', label: t('settings.preferences.fontOptions.outfit') },
                  { value: 'jakarta', label: t('settings.preferences.fontOptions.jakarta') },
                  { value: 'system', label: t('settings.preferences.fontOptions.system') }
                ]}
                placeholder={t('settings.preferences.font')}
                disabled={isUpdatingFont}
                className="w-[180px]"
              />
            </RBACGuard>
          </div>
        </CardContent>
      </Card>

      <div className="mt-6">
        <Card>
          <CardHeader>
            <TypographySmall>{t('settings.preferences.language.title')}</TypographySmall>
            <TypographyMuted>{t('settings.preferences.language.description')}</TypographyMuted>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <TypographyMuted>{t('settings.preferences.language.select')}</TypographyMuted>
              <RBACGuard resource="user" action="update">
                <LanguageSwitcher
                  handleLanguageChange={handleLanguageChange}
                  isUpdatingLanguage={isUpdatingLanguage}
                  userSettings={userSettings}
                />
              </RBACGuard>
            </div>
          </CardContent>
        </Card>
      </div>
      {/* <div className="mt-6">
        <Card>
          <CardHeader>
            <TypographySmall>{t('settings.preferences.autoUpdate.title')}</TypographySmall>
            <TypographyMuted>{t('settings.preferences.autoUpdate.description')}</TypographyMuted>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <TypographyMuted>{t('settings.preferences.autoUpdate.select')}</TypographyMuted>
              <RBACGuard resource="user" action="update">
                <Switch
                  checked={userSettings.auto_update}
                  onCheckedChange={handleAutoUpdateChange}
                  disabled={isUpdatingAutoUpdate}
                />
              </RBACGuard>
            </div>
          </CardContent>
        </Card>
      </div> */}
    </TabsContent>
  );
}

export default AccountSection;
