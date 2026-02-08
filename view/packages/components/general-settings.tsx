'use client';

import React from 'react';
import { Mail, User, CheckCircle, CheckCircle2, XCircle, Search, Shield } from 'lucide-react';
import { Button } from '@nixopus/ui';
import { Input } from '@nixopus/ui';
import { Label } from '@nixopus/ui';
import { Alert, AlertDescription, AlertTitle } from '@nixopus/ui';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { LanguageSwitcher } from '@/packages/components/language-switcher';
import { Switch } from '@nixopus/ui';
import { RBACGuard } from '@/packages/components/rbac';
import { TypographySmall, TypographyMuted, TypographyH3 } from '@nixopus/ui';
import UploadAvatar from '@/components/ui/upload_avatar';
import { Badge } from '@nixopus/ui';
import { TabsContent } from '@nixopus/ui';
import { QRCodeSVG } from 'qrcode.react';
import { Card, CardContent, CardHeader } from '@nixopus/ui';
import {
  AccountSectionProps,
  AvatarSectionProps,
  SecuritySectionProps
} from '@/packages/types/settings';
import { useAccountSection } from '@/packages/hooks/settings/use-account-section';
import { useTwoFactorSetup } from '@/packages/hooks/settings/use-two-factor-setup';

export function AccountSection({
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
  isUpdatingTheme,
  isUpdatingLanguage,
  isUpdatingAutoUpdate,
  handleThemeChange,
  handleLanguageChange,
  handleAutoUpdateChange
}: AccountSectionProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-8">
      <div className="space-y-6">
        <div>
          <TypographySmall className="text-sm font-medium">
            {t('settings.account.title')}
          </TypographySmall>
        </div>
        <div className="space-y-6">
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
              <TypographySmall className="text-red-500 text-xs">{usernameError}</TypographySmall>
            )}

            {usernameSuccess && (
              <Alert variant="default">
                <CheckCircle className="h-4 w-4" />
                <AlertTitle>Success</AlertTitle>
                <AlertDescription>{t('settings.account.username.success')}</AlertDescription>
              </Alert>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="email" className="flex items-center gap-2">
              <Mail size={16} />
              {t('settings.account.email.label')}
            </Label>
            <Input id="email" value={email} readOnly disabled className="bg-muted/50" />
          </div>
        </div>
      </div>

      <div className="space-y-6">
        <div>
          <TypographySmall className="text-sm font-medium">
            {t('settings.preferences.language.title')}
          </TypographySmall>
        </div>
        <div className="flex items-center justify-between">
          <TypographyMuted className="text-sm">
            {t('settings.preferences.language.select')}
          </TypographyMuted>
          <RBACGuard resource="user" action="update">
            <LanguageSwitcher
              handleLanguageChange={handleLanguageChange}
              isUpdatingLanguage={isUpdatingLanguage}
              userSettings={userSettings}
            />
          </RBACGuard>
        </div>
      </div>
      <div className="space-y-8">
        <div className="flex items-center justify-between">
          <div className="flex flex-col gap-2">
            <TypographySmall className="text-sm font-medium">
              {t('settings.preferences.autoUpdate.title')}
            </TypographySmall>
          </div>
          <RBACGuard resource="user" action="update">
            <Switch
              checked={userSettings.auto_update}
              onCheckedChange={handleAutoUpdateChange}
              disabled={isUpdatingAutoUpdate}
            />
          </RBACGuard>
        </div>
      </div>
    </div>
  );
}

export function AvatarSection({ onImageChange, user }: AvatarSectionProps) {
  const { t } = useTranslation();

  return (
    <div className="col-span-1 space-y-4 border border-border/50 rounded-lg p-6 h-fit">
      <div>
        <TypographySmall className="text-sm font-medium">
          {t('settings.account.avatar.title')}
        </TypographySmall>
      </div>
      <RBACGuard resource="user" action="update">
        <UploadAvatar
          onImageChange={onImageChange}
          username={user?.username}
          initialImage={user?.avatar}
        />
      </RBACGuard>
    </div>
  );
}

export function SecuritySection({}: SecuritySectionProps) {
  const { t } = useTranslation();

  return (
    <TabsContent value="security" className="space-y-4 mt-4">
      <TwoFactorSetup />
    </TabsContent>
  );
}

export default function TwoFactorSetup() {
  const { t } = useTranslation();
  const { code, setCode, setupData, user, handleSetup, handleVerify, handleDisable } =
    useTwoFactorSetup();

  return (
    <Card>
      <CardHeader>
        <TypographySmall className="flex items-center gap-2">
          <Shield className="h-5 w-5" />
          {t('settings.2fa.title')}
        </TypographySmall>
      </CardHeader>
      <CardContent>
        {user?.two_factor_enabled ? (
          <div className="space-y-6">
            <Alert>
              <CheckCircle2 className="h-4 w-4" />
              <AlertTitle>{t('settings.2fa.enabledTitle')}</AlertTitle>
              <AlertDescription>{t('settings.2fa.enabledDescription')}</AlertDescription>
            </Alert>

            <RBACGuard resource="user" action="update">
              <Button onClick={handleDisable} variant="destructive" className="w-full">
                {t('settings.2fa.disableButton')}
              </Button>
            </RBACGuard>
          </div>
        ) : !setupData ? (
          <div className="space-y-4">
            <RBACGuard resource="user" action="update">
              <Button onClick={handleSetup} className="w-full">
                {t('settings.2fa.setupButton')}
              </Button>
            </RBACGuard>
          </div>
        ) : (
          <div className="space-y-6">
            <div className="flex flex-col items-center space-y-4">
              <div className="rounded-lg border p-4">
                <QRCodeSVG value={setupData.qr_code} size={200} level="H" includeMargin={true} />
              </div>
              <TypographyMuted className="text-center">
                {t('settings.2fa.description')}
              </TypographyMuted>
            </div>

            <div className="space-y-4">
              <div className="space-y-2">
                <Label>{t('settings.2fa.enterCode')}</Label>
                <Input
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                  placeholder={t('settings.2fa.codePlaceholder')}
                  className="w-full"
                />
              </div>
              <RBACGuard resource="user" action="update">
                <Button
                  onClick={handleVerify}
                  className="w-full"
                  disabled={!code || code.length !== 6}
                >
                  {t('settings.2fa.verifyButton')}
                </Button>
              </RBACGuard>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
