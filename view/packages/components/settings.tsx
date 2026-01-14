'use client';

import React from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { LogoutDialog } from '@/components/ui/logout-dialog';
import { Button } from '@/components/ui/button';
import { LogOut } from 'lucide-react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ResourceGuard } from '@/packages/components/rbac';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { TypographyMuted, TypographyH1 } from '@/components/ui/typography';
import { RotateCcw } from 'lucide-react';
import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';
import { SelectWrapper } from '@/components/ui/select-wrapper';
import {
  AvatarSection,
  AccountSection,
  FeatureFlagsSettings
} from '@/packages/components/general-settings';
import { NotificationChannelsTab } from '@/packages/components/notification-settings';
import { NotificationPreferencesTab } from '@/packages/components/notification-settings';
import {
  AddMember,
  TeamMembers,
  EditTeam,
  TeamStats,
  RecentActivity
} from '@/packages/components/team-settings';
// import DomainsTable from '@/app/settings/domains/components/domainsTable';
// import UpdateDomainDialog from '@/app/settings/domains/components/update-domain';
import {
  useGeneralSettingsContent,
  useNotificationsSettingsContent,
  useTeamsSettingsContent,
  useNetworkSettingsContent,
  useTerminalSettingsContent,
  useContainerSettingsContent,
  useTroubleshootingSettingsContent,
  useKeyboardShortcutsSettingsContent,
  type SettingConfig
} from '@/packages/hooks/settings/use-settings-content';
import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog';
import { useSettingsModal } from '@/packages/hooks/shared/use-settings-modal';
import {
  SettingsCategory,
  useSettingsCategories
} from '@/packages/hooks/shared/use-settings-categories';
import { Heart, HelpCircle, AlertCircle, ArrowUpCircle } from 'lucide-react';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { useSettingsFooter } from '@/packages/hooks/settings/use-settings-footer';
import { SettingsSidebarProps } from '../types/settings';

interface SettingsContentProps {
  activeCategory: string;
}

function SettingRow({ children, className }: { children: React.ReactNode; className?: string }) {
  return <div className={cn('flex items-center justify-between py-3', className)}>{children}</div>;
}

function SettingLabel({
  htmlFor,
  title,
  description
}: {
  htmlFor: string;
  title: string;
  description: string;
}) {
  return (
    <div className="space-y-0.5">
      <Label htmlFor={htmlFor} className="text-sm font-medium">
        {title}
      </Label>
      <TypographyMuted className="text-xs">{description}</TypographyMuted>
    </div>
  );
}

function NumberInput({
  id,
  value,
  onChange,
  min,
  max,
  step = 1,
  suffix
}: {
  id: string;
  value: number;
  onChange: (value: number) => void;
  min: number;
  max: number;
  step?: number;
  suffix?: string;
}) {
  return (
    <div className="flex items-center gap-2">
      <Input
        id={id}
        type="number"
        value={value}
        onChange={(e) => {
          const val = parseFloat(e.target.value);
          if (!isNaN(val) && val >= min && val <= max) {
            onChange(val);
          }
        }}
        min={min}
        max={max}
        step={step}
        className="w-24 text-right"
      />
      {suffix && <span className="text-xs text-muted-foreground">{suffix}</span>}
    </div>
  );
}

function SettingField({
  config,
  value,
  onChange,
  t
}: {
  config: SettingConfig;
  value: any;
  onChange: (value: any) => void;
  t: (key: string) => string;
}) {
  const title = t(config.titleKey);
  const description = t(config.descriptionKey);

  const renderControl = () => {
    switch (config.type) {
      case 'number':
        return (
          <NumberInput
            id={config.id}
            value={value as number}
            onChange={onChange}
            min={config.min}
            max={config.max}
            step={config.step}
            suffix={config.suffix}
          />
        );
      case 'switch':
        return <Switch id={config.id} checked={value as boolean} onCheckedChange={onChange} />;
      case 'select':
        return (
          <SelectWrapper
            value={value as string}
            onValueChange={onChange}
            options={config.options}
            className={config.width || 'w-48'}
          />
        );
    }
  };

  return (
    <SettingRow>
      <SettingLabel htmlFor={config.id} title={title} description={description} />
      {renderControl()}
    </SettingRow>
  );
}

function SettingsRenderer({
  settings,
  configs,
  updateSetting,
  t
}: {
  settings: Record<string, any>;
  configs: SettingConfig[];
  updateSetting: (key: string, value: any) => void;
  t: (key: string, params?: Record<string, string>) => string;
}) {
  return (
    <div className="space-y-1">
      {configs.map((config) => (
        <SettingField
          key={config.id}
          config={config}
          value={(settings as any)[config.id]}
          onChange={(value) => updateSetting(config.id, value)}
          t={(key: string) => t(key as any)}
        />
      ))}
    </div>
  );
}

function SettingsSkeleton() {
  return (
    <div className="space-y-6">
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-4 w-72" />
      <Skeleton className="h-48 w-full" />
    </div>
  );
}

function GeneralSettingsContent() {
  const { t } = useTranslation();
  const { settings, sidebar } = useGeneralSettingsContent();

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
      <Button variant="secondary" onClick={sidebar.handleLogoutClick} className="w-full gap-2">
        <LogOut className="h-4 w-4" />
        {t('user.menu.logout')}
      </Button>
      <LogoutDialog
        open={sidebar.showLogoutDialog}
        onConfirm={sidebar.handleLogoutConfirm}
        onCancel={sidebar.handleLogoutCancel}
      />
    </div>
  );
}

function NotificationsSettingsContent() {
  const { t } = useTranslation();
  const { settings, handleSave, handleSaveSlack, handleSaveDiscord } =
    useNotificationsSettingsContent();

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-semibold">{t('settings.notifications.page.title')}</h2>
      <Tabs defaultValue="channels" className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <ResourceGuard resource="notification" action="create">
            <TabsTrigger value="channels">
              {t('settings.notifications.page.tabs.channels')}
            </TabsTrigger>
          </ResourceGuard>
          <TabsTrigger value="preferences">
            {t('settings.notifications.page.tabs.preferences')}
          </TabsTrigger>
        </TabsList>
        <ResourceGuard resource="notification" action="create">
          <TabsContent value="channels">
            <NotificationChannelsTab
              smtpConfigs={settings.smtpConfigs || undefined}
              slackConfig={settings.slackConfig}
              discordConfig={settings.discordConfig}
              isLoading={settings.isLoading}
              handleOnSave={handleSave}
              handleOnSaveSlack={handleSaveSlack}
              handleOnSaveDiscord={handleSaveDiscord}
            />
          </TabsContent>
        </ResourceGuard>
        <TabsContent value="preferences">
          <NotificationPreferencesTab
            activityPreferences={settings.preferences?.activity}
            securityPreferences={settings.preferences?.security}
            onUpdatePreference={settings.handleUpdatePreference}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function TeamsSettingsContent() {
  const { t } = useTranslation();
  const { settings } = useTeamsSettingsContent();

  return (
    <ResourceGuard resource="organization" action="read">
      <div className="space-y-6">
        <h2 className="text-2xl font-semibold">Teams</h2>
        <div className="flex items-center justify-between">
          <div>
            <TypographyH1>{settings.teamName}</TypographyH1>
            <TypographyMuted>{settings.teamDescription}</TypographyMuted>
          </div>
          <div className="flex gap-2">
            <ResourceGuard resource="organization" action="update">
              <EditTeam
                teamName={settings.teamName || ''}
                teamDescription={settings.teamDescription || ''}
                setEditTeamDialogOpen={settings.setEditTeamDialogOpen}
                handleUpdateTeam={settings.handleUpdateTeam}
                setTeamName={settings.setTeamName}
                setTeamDescription={settings.setTeamDescription}
                isEditTeamDialogOpen={settings.isEditTeamDialogOpen}
                isUpdating={settings.isUpdating}
              />
            </ResourceGuard>
            <ResourceGuard resource="user" action="create">
              <AddMember
                isAddUserDialogOpen={settings.isAddUserDialogOpen}
                setIsAddUserDialogOpen={settings.setIsAddUserDialogOpen}
                newUser={settings.newUser}
                setNewUser={settings.setNewUser}
                handleSendInvite={settings.handleSendInvite}
                isInviteLoading={settings.isInviteLoading}
              />
            </ResourceGuard>
          </div>
        </div>
        {settings.users.length > 0 ? (
          <TeamMembers
            users={settings.users}
            handleRemoveUser={settings.handleRemoveUser}
            getRoleBadgeVariant={settings.getRoleBadgeVariant}
            onUpdateUser={settings.handleUpdateUser}
          />
        ) : (
          <div className="text-center text-muted-foreground">{t('settings.teams.noMembers')}</div>
        )}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <TeamStats users={settings.users} />
          <RecentActivity />
        </div>
      </div>
    </ResourceGuard>
  );
}

function FeatureFlagsSettingsContent() {
  return <FeatureFlagsSettings />;
}

function KeyboardShortcutsSettingsContent() {
  const { t } = useTranslation();
  const { shortcuts } = useKeyboardShortcutsSettingsContent();

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-semibold">Keyboard Shortcuts</h2>
      <div className="space-y-4">
        {shortcuts.map((shortcut, index) => (
          <div key={index} className="flex items-center justify-between">
            <div className="text-sm text-muted-foreground">{shortcut.description}</div>
            <div className="flex items-center gap-1">
              {shortcut.keys.map((key, keyIndex) => (
                <React.Fragment key={keyIndex}>
                  <kbd className="pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground">
                    {key}
                  </kbd>
                  {keyIndex < shortcut.keys.length - 1 && (
                    <span className="text-muted-foreground">+</span>
                  )}
                </React.Fragment>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function NetworkSettingsContent() {
  const { t, settings, isLoading, updateSetting, resetToDefaults, configs, hasChanges } =
    useNetworkSettingsContent();

  if (isLoading) {
    return <SettingsSkeleton />;
  }

  return (
    <div className="flex flex-col h-full">
      <div className="space-y-6 flex-1 overflow-y-auto">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-semibold">{t('settings.network.title')}</h2>
            <TypographyMuted className="text-sm mt-1">
              {t('settings.network.description')}
            </TypographyMuted>
          </div>
          {hasChanges && (
            <Button variant="outline" size="sm" onClick={resetToDefaults} className="gap-2">
              <RotateCcw className="h-4 w-4" />
              {t('settings.advanced.resetDefaults')}
            </Button>
          )}
        </div>
        <SettingsRenderer
          settings={settings as any}
          configs={configs}
          updateSetting={(key: string, value: any) => updateSetting(key as any, value)}
          t={(key: string) => t(key as any)}
        />
      </div>
    </div>
  );
}

function TerminalSettingsContent() {
  const { t, settings, isLoading, updateSetting, resetToDefaults, configs, hasChanges } =
    useTerminalSettingsContent();

  if (isLoading) {
    return <SettingsSkeleton />;
  }

  return (
    <div className="flex flex-col h-full">
      <div className="space-y-6 flex-1 overflow-y-auto">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-semibold">{t('settings.terminal.title')}</h2>
            <TypographyMuted className="text-sm mt-1">
              {t('settings.terminal.description')}
            </TypographyMuted>
          </div>
          {hasChanges && (
            <Button variant="outline" size="sm" onClick={resetToDefaults} className="gap-2">
              <RotateCcw className="h-4 w-4" />
              {t('settings.advanced.resetDefaults')}
            </Button>
          )}
        </div>
        <SettingsRenderer
          settings={settings as any}
          configs={configs}
          updateSetting={(key: string, value: any) => updateSetting(key as any, value)}
          t={(key: string) => t(key as any)}
        />
      </div>
    </div>
  );
}

function ContainerSettingsContent() {
  const { t, settings, isLoading, updateSetting, resetToDefaults, configs, hasChanges } =
    useContainerSettingsContent();

  if (isLoading) {
    return <SettingsSkeleton />;
  }

  return (
    <div className="flex flex-col h-full">
      <div className="space-y-6 flex-1 overflow-y-auto">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-semibold">{t('settings.container.title')}</h2>
            <TypographyMuted className="text-sm mt-1">
              {t('settings.container.description')}
            </TypographyMuted>
          </div>
          {hasChanges && (
            <Button variant="outline" size="sm" onClick={resetToDefaults} className="gap-2">
              <RotateCcw className="h-4 w-4" />
              {t('settings.advanced.resetDefaults')}
            </Button>
          )}
        </div>
        <SettingsRenderer
          settings={settings as any}
          configs={configs}
          updateSetting={(key: string, value: any) => updateSetting(key as any, value)}
          t={(key: string) => t(key as any)}
        />
      </div>
    </div>
  );
}

function TroubleshootingSettingsContent() {
  const { t, settings, isLoading, updateSetting, resetToDefaults, configs, hasChanges } =
    useTroubleshootingSettingsContent();

  if (isLoading) {
    return <SettingsSkeleton />;
  }

  return (
    <div className="flex flex-col h-full">
      <div className="space-y-6 flex-1 overflow-y-auto">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-semibold">{t('settings.troubleshooting.title')}</h2>
            <TypographyMuted className="text-sm mt-1">
              {t('settings.troubleshooting.description')}
            </TypographyMuted>
          </div>
          {hasChanges && (
            <Button variant="outline" size="sm" onClick={resetToDefaults} className="gap-2">
              <RotateCcw className="h-4 w-4" />
              {t('settings.advanced.resetDefaults')}
            </Button>
          )}
        </div>
        <SettingsRenderer
          settings={settings as any}
          configs={configs}
          updateSetting={(key: string, value: any) => updateSetting(key as any, value)}
          t={(key: string) => t(key as any)}
        />
      </div>
    </div>
  );
}

export function SettingsContent({ activeCategory }: SettingsContentProps) {
  return (
    <div className="flex-1 flex flex-col overflow-hidden p-6">
      {activeCategory === 'general' && <GeneralSettingsContent />}
      {activeCategory === 'notifications' && <NotificationsSettingsContent />}
      {activeCategory === 'teams' && <TeamsSettingsContent />}
      {/* {activeCategory === 'domains' && <DomainsSettingsContent />} */}
      {activeCategory === 'feature-flags' && <FeatureFlagsSettingsContent />}
      {activeCategory === 'keyboard-shortcuts' && <KeyboardShortcutsSettingsContent />}
      {activeCategory === 'network' && <NetworkSettingsContent />}
      {activeCategory === 'terminal' && <TerminalSettingsContent />}
      {activeCategory === 'container' && <ContainerSettingsContent />}
      {activeCategory === 'troubleshooting' && <TroubleshootingSettingsContent />}
    </div>
  );
}

export function SettingsModal() {
  const { open, closeSettings, activeCategory, setActiveCategory } = useSettingsModal();
  const categories = useSettingsCategories();
  return (
    <Dialog open={open} onOpenChange={closeSettings}>
      <DialogContent className="!max-w-[1200px] w-[90vw] max-h-[90vh] h-[90vh] p-0 flex overflow-hidden">
        <DialogTitle className="sr-only">Settings</DialogTitle>
        <SettingsSidebar
          categories={categories}
          activeCategory={activeCategory}
          onCategoryChange={setActiveCategory}
        />
        <SettingsContent activeCategory={activeCategory} />
      </DialogContent>
    </Dialog>
  );
}

export function SettingsFooter() {
  const { t } = useTranslation();
  const { updateInfo, isCheckingUpdates, handleSponsor, handleReportIssue, handleHelp } =
    useSettingsFooter();

  return (
    <div className="border-t p-2 space-y-2">
      <div className="flex items-center justify-center gap-1">
        <Button
          variant="ghost"
          size="icon"
          onClick={handleSponsor}
          className="h-8 w-8"
          title={t('user.menu.sponsor')}
        >
          <Heart className="h-4 w-4 text-red-500" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          onClick={handleHelp}
          className="h-8 w-8"
          title={t('user.menu.help')}
        >
          <HelpCircle className="h-4 w-4" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          onClick={handleReportIssue}
          className="h-8 w-8"
          title={t('user.menu.reportIssue')}
        >
          <AlertCircle className="h-4 w-4" />
        </Button>
      </div>
      <div className="flex items-center justify-center">
        {isCheckingUpdates ? (
          <span className="text-xs text-muted-foreground">Checking version...</span>
        ) : updateInfo?.current_version ? (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <div
                  className={cn(
                    'flex items-center gap-1.5 text-xs px-2 py-1 rounded-md cursor-default',
                    updateInfo.update_available
                      ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400'
                      : 'text-muted-foreground'
                  )}
                >
                  {updateInfo.update_available && <ArrowUpCircle className="h-3 w-3" />}
                  <span>{updateInfo.current_version}</span>
                </div>
              </TooltipTrigger>
              <TooltipContent side="top">
                {updateInfo.update_available ? (
                  <p>Update available: {updateInfo.latest_version}</p>
                ) : (
                  <p>You&apos;re on the latest version</p>
                )}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        ) : null}
      </div>
    </div>
  );
}

export function SettingsSidebar({
  categories,
  activeCategory,
  onCategoryChange
}: SettingsSidebarProps) {
  const { t } = useTranslation();
  const visibleCategories = categories.filter((cat) => cat.visible !== false);
  const accountCategories = visibleCategories.filter((cat) => cat.scope === 'account');
  const orgCategories = visibleCategories.filter((cat) => cat.scope === 'organization');

  const renderCategoryButton = (cat: SettingsCategory) => {
    const Icon = cat.icon;
    return (
      <button
        key={cat.id}
        onClick={() => onCategoryChange(cat.id)}
        className={cn(
          'w-full flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors',
          activeCategory === cat.id ? 'bg-muted font-medium' : 'hover:bg-muted/50'
        )}
      >
        <Icon className="h-4 w-4" />
        <span>{cat.label}</span>
      </button>
    );
  };

  return (
    <div className="w-[240px] flex-shrink-0 bg-muted/50 border-r flex flex-col">
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {accountCategories.length > 0 && (
          <div className="space-y-1">
            <div className="px-3 py-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              {t('settings.sidebar.account')}
            </div>
            {accountCategories.map(renderCategoryButton)}
          </div>
        )}
        {orgCategories.length > 0 && (
          <div className="space-y-1">
            <div className="px-3 py-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              {t('settings.sidebar.organization')}
            </div>
            {orgCategories.map(renderCategoryButton)}
          </div>
        )}
      </div>
      <SettingsFooter />
    </div>
  );
}
