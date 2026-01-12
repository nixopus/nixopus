'use client';

import React from 'react';
import { Mail, User, CheckCircle, AlertCircle, CheckCircle2, XCircle, Search } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { LanguageSwitcher } from '@/packages/components/language-switcher';
import { SelectWrapper } from '@/components/ui/select-wrapper';
import { Switch } from '@/components/ui/switch';
import { RBACGuard } from '@/packages/components/rbac';
import { TypographySmall, TypographyMuted, TypographyH3 } from '@/components/ui/typography';
import UploadAvatar from '@/components/ui/upload_avatar';
import { Badge } from '@/components/ui/badge';
import { TabsContent } from '@/components/ui/tabs';
import { QRCodeSVG } from 'qrcode.react';
import { Shield } from 'lucide-react';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import {
  AccountSectionProps,
  AvatarSectionProps,
  SecuritySectionProps
} from '@/packages/types/settings';
import { useAccountSection } from '@/packages/hooks/settings/use-account-section';
import { useFeatureFlagsSettings } from '@/packages/hooks/settings/use-feature-flags-settings';
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
  isUpdatingFont,
  isUpdatingTheme,
  isUpdatingLanguage,
  isUpdatingAutoUpdate,
  handleThemeChange,
  handleLanguageChange,
  handleAutoUpdateChange,
  handleFontUpdate
}: AccountSectionProps) {
  const { t } = useTranslation();
  const {
    isSendingVerification,
    verificationSent,
    verificationError,
    handleFontChange,
    handleSendVerification
  } = useAccountSection({
    userSettings,
    handleFontUpdate
  });

  return (
    <div className="space-y-8">
      <div className="space-y-6">
        <div>
          <TypographySmall className="text-sm font-medium">
            {t('settings.account.title')}
          </TypographySmall>
          <TypographyMuted className="text-xs mt-1">
            {t('settings.account.description')}
          </TypographyMuted>
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
                    <TypographySmall className="text-red-500 text-xs">
                      {verificationError}
                    </TypographySmall>
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
        </div>
      </div>

      <div className="space-y-6">
        <div>
          <TypographySmall className="text-sm font-medium">
            {t('settings.account.preferences.title')}
          </TypographySmall>
          <TypographyMuted className="text-xs mt-1">
            {t('settings.account.preferences.description')}
          </TypographyMuted>
        </div>
        <div className="space-y-4">
          <div className="flex justify-between items-center">
            <TypographyMuted className="text-sm">{t('settings.preferences.font')}</TypographyMuted>
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
                  {
                    value: 'space-grotesk',
                    label: t('settings.preferences.fontOptions.spaceGrotesk')
                  },
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
        </div>
      </div>

      <div className="space-y-6">
        <div>
          <TypographySmall className="text-sm font-medium">
            {t('settings.preferences.language.title')}
          </TypographySmall>
          <TypographyMuted className="text-xs mt-1">
            {t('settings.preferences.language.description')}
          </TypographyMuted>
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
            <TypographyMuted className="text-xs mt-1">
              {t('settings.preferences.autoUpdate.description')}
            </TypographyMuted>
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
        <TypographyMuted className="text-xs mt-1">
          {t('settings.account.avatar.description')}
        </TypographyMuted>
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

export function FeatureFlagsSettings() {
  const { t } = useTranslation();
  const {
    isLoading,
    searchTerm,
    setSearchTerm,
    filterEnabled,
    setFilterEnabled,
    handleToggleFeature,
    getGroupIcon,
    groupedFeatures,
    enabledFeatures,
    disabledFeatures
  } = useFeatureFlagsSettings();

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <div className="space-y-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="animate-pulse">
                <div className="h-4 bg-muted rounded w-1/4 mb-2"></div>
                <div className="space-y-2">
                  {[1, 2].map((j) => (
                    <div
                      key={j}
                      className="flex items-center justify-between p-4 border rounded-lg"
                    >
                      <div className="space-y-2">
                        <div className="h-4 bg-muted rounded w-32"></div>
                        <div className="h-3 bg-muted rounded w-48"></div>
                      </div>
                      <div className="h-6 w-11 bg-muted rounded-full"></div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <RBACGuard resource="feature-flags" action="read">
      <div className="flex flex-col h-full">
        <div className="flex items-center justify-between mb-6">
          <div>
            <TypographyH3 className="text-lg font-semibold">
              {t('settings.featureFlags.title')}
            </TypographyH3>
            <TypographyMuted className="text-xs mt-1">
              {t('settings.featureFlags.description')}
            </TypographyMuted>
          </div>
          <div className="flex items-center gap-2">
            <Badge variant="secondary" className="flex items-center gap-1">
              <CheckCircle2 className="h-3 w-3" />
              {enabledFeatures}
            </Badge>
            <Badge variant="outline" className="flex items-center gap-1">
              <XCircle className="h-3 w-3" />
              {disabledFeatures}
            </Badge>
          </div>
        </div>
        <div className="flex-1 overflow-y-auto space-y-6">
          <div className="flex items-center gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder={t('settings.featureFlags.searchPlaceholder')}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>
            <div className="flex items-center gap-2">
              <div className="flex gap-1">
                {(['all', 'enabled', 'disabled'] as const).map((filter) => (
                  <Button
                    key={filter}
                    variant={filterEnabled === filter ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setFilterEnabled(filter)}
                  >
                    {t(`settings.featureFlags.filters.${filter}`)}
                  </Button>
                ))}
              </div>
            </div>
          </div>

          {groupedFeatures.size === 0 ? (
            <Alert>
              <Search className="h-4 w-4" />
              <AlertDescription>
                {searchTerm || filterEnabled !== 'all'
                  ? t('settings.featureFlags.noResults')
                  : t('settings.featureFlags.noFeatures')}
              </AlertDescription>
            </Alert>
          ) : (
            Array.from(groupedFeatures.entries())
              .filter(([group]) => group !== 'notifications')
              .map(([group, features], index) => {
                const GroupIcon = getGroupIcon(group);
                const enabledInGroup = features.filter((f) => f.is_enabled).length;

                return (
                  <div key={group} className="space-y-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <GroupIcon className="h-4 w-4 text-muted-foreground" />
                        <TypographySmall className="font-semibold">
                          {t(`settings.featureFlags.groups.${group}.title` as any)}
                        </TypographySmall>
                        <Badge variant="outline" className="text-xs">
                          {enabledInGroup}/{features.length}
                        </Badge>
                      </div>
                    </div>
                    <div className="space-y-3">
                      {features?.map((feature) => (
                        <div
                          key={feature.feature_name}
                          className="flex items-center justify-between p-4 rounded-md bg-muted/30 transition-colors hover:bg-muted/50"
                        >
                          <div className="space-y-1 flex-1">
                            <div className="flex items-center gap-2">
                              <TypographySmall className="font-medium">
                                {t(
                                  `settings.featureFlags.features.${feature.feature_name}.title` as any
                                )}
                              </TypographySmall>
                            </div>
                            <TypographyMuted className="text-sm">
                              {t(
                                `settings.featureFlags.features.${feature.feature_name}.description` as any
                              )}
                            </TypographyMuted>
                          </div>
                          <RBACGuard resource="feature-flags" action="update">
                            <Switch
                              checked={feature.is_enabled}
                              onCheckedChange={(checked) =>
                                handleToggleFeature(feature.feature_name, checked)
                              }
                            />
                          </RBACGuard>
                        </div>
                      ))}
                    </div>
                  </div>
                );
              })
          )}
        </div>
      </div>
    </RBACGuard>
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
