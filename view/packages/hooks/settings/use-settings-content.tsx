'use client';

import { useState } from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useAppSidebar } from '@/packages/hooks/shared/use-app-sidebar';
import { useAppSelector } from '@/redux/hooks';
import { useGetAllDomainsQuery } from '@/redux/services/settings/domainsApi';
import useGeneralSettings from '@/packages/hooks/settings/use-general-settings';
import useNotificationSettings from '@/packages/hooks/settings/use-notification-settings';
import useTeamSettings from '@/packages/hooks/settings/use-team-settings';
import { useAdvancedSettings } from '@/packages/hooks/settings/use-advanced-settings';
import { SMTPFormData } from '@/redux/types/notification';
import { SelectOption } from '@/components/ui/select-wrapper';

export type SettingType = 'number' | 'switch' | 'select';

export interface BaseSettingConfig {
  id: string;
  titleKey: string;
  descriptionKey: string;
  type: SettingType;
}

export interface NumberSettingConfig extends BaseSettingConfig {
  type: 'number';
  min: number;
  max: number;
  step?: number;
  suffix?: string;
}

export interface SwitchSettingConfig extends BaseSettingConfig {
  type: 'switch';
}

export interface SelectSettingConfig extends BaseSettingConfig {
  type: 'select';
  options: SelectOption[];
  width?: string;
}

export type SettingConfig = NumberSettingConfig | SwitchSettingConfig | SelectSettingConfig;

export function useGeneralSettingsContent() {
  const settings = useGeneralSettings();
  const sidebar = useAppSidebar();

  return {
    settings,
    sidebar
  };
}

export function useNotificationsSettingsContent() {
  const settings = useNotificationSettings();

  const handleSave = (data: SMTPFormData) => settings.handleOnSave(data);

  const handleSaveSlack = (data: Record<string, string>) => {
    settings.slackConfig
      ? settings.handleUpdateWebhookConfig({
          type: 'slack',
          webhook_url: data.webhook_url,
          is_active: data.is_active === 'true'
        })
      : settings.handleCreateWebhookConfig({ type: 'slack', webhook_url: data.webhook_url });
  };

  const handleSaveDiscord = (data: Record<string, string>) => {
    settings.discordConfig
      ? settings.handleUpdateWebhookConfig({
          type: 'discord',
          webhook_url: data.webhook_url,
          is_active: data.is_active === 'true'
        })
      : settings.handleCreateWebhookConfig({ type: 'discord', webhook_url: data.webhook_url });
  };

  return {
    settings,
    handleSave,
    handleSaveSlack,
    handleSaveDiscord
  };
}

export function useTeamsSettingsContent() {
  const settings = useTeamSettings();
  return { settings };
}

export function useDomainsSettingsContent() {
  const { t } = useTranslation();
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);
  const { data: domains, isLoading } = useGetAllDomainsQuery();
  const [addDomainDialogOpen, setAddDomainDialogOpen] = useState(false);

  return {
    t,
    activeOrg,
    domains,
    isLoading,
    addDomainDialogOpen,
    setAddDomainDialogOpen
  };
}

export function useNetworkSettingsContent() {
  const { t } = useTranslation();
  const { settings, isLoading, updateSetting, resetToDefaults, DEFAULT_SETTINGS } =
    useAdvancedSettings();

  const configs: SettingConfig[] = [
    {
      id: 'websocketReconnectAttempts',
      titleKey: 'settings.network.reconnectAttempts.title',
      descriptionKey: 'settings.network.reconnectAttempts.description',
      type: 'number',
      min: 1,
      max: 20
    },
    {
      id: 'websocketReconnectInterval',
      titleKey: 'settings.network.reconnectInterval.title',
      descriptionKey: 'settings.network.reconnectInterval.description',
      type: 'number',
      min: 1000,
      max: 30000,
      step: 500,
      suffix: 'ms'
    },
    {
      id: 'apiRetryAttempts',
      titleKey: 'settings.network.apiRetryAttempts.title',
      descriptionKey: 'settings.network.apiRetryAttempts.description',
      type: 'number',
      min: 0,
      max: 5
    },
    {
      id: 'disableApiCache',
      titleKey: 'settings.network.disableApiCache.title',
      descriptionKey: 'settings.network.disableApiCache.description',
      type: 'switch'
    }
  ];

  const hasChanges = configs.some(
    (config) => (settings as any)[config.id] !== (DEFAULT_SETTINGS as any)[config.id]
  );

  return {
    t,
    settings,
    isLoading,
    updateSetting,
    resetToDefaults,
    configs,
    hasChanges
  };
}

export function useTerminalSettingsContent() {
  const { t } = useTranslation();
  const { settings, isLoading, updateSetting, resetToDefaults, DEFAULT_SETTINGS } =
    useAdvancedSettings();

  const configs: SettingConfig[] = [
    {
      id: 'terminalScrollback',
      titleKey: 'settings.terminal.scrollback.title',
      descriptionKey: 'settings.terminal.scrollback.description',
      type: 'number',
      min: 1000,
      max: 50000,
      step: 1000
    },
    {
      id: 'terminalFontSize',
      titleKey: 'settings.terminal.fontSize.title',
      descriptionKey: 'settings.terminal.fontSize.description',
      type: 'number',
      min: 8,
      max: 24,
      suffix: 'px'
    },
    {
      id: 'terminalCursorStyle',
      titleKey: 'settings.terminal.cursorStyle.title',
      descriptionKey: 'settings.terminal.cursorStyle.description',
      type: 'select',
      options: [
        { value: 'bar', label: t('settings.terminal.cursorStyle.options.bar') },
        { value: 'block', label: t('settings.terminal.cursorStyle.options.block') },
        { value: 'underline', label: t('settings.terminal.cursorStyle.options.underline') }
      ],
      width: 'w-32'
    },
    {
      id: 'terminalCursorBlink',
      titleKey: 'settings.terminal.cursorBlink.title',
      descriptionKey: 'settings.terminal.cursorBlink.description',
      type: 'switch'
    },
    {
      id: 'terminalLineHeight',
      titleKey: 'settings.terminal.lineHeight.title',
      descriptionKey: 'settings.terminal.lineHeight.description',
      type: 'number',
      min: 1.0,
      max: 2.5,
      step: 0.1
    },
    {
      id: 'terminalCursorWidth',
      titleKey: 'settings.terminal.cursorWidth.title',
      descriptionKey: 'settings.terminal.cursorWidth.description',
      type: 'number',
      min: 1,
      max: 5
    },
    {
      id: 'terminalTabStopWidth',
      titleKey: 'settings.terminal.tabStopWidth.title',
      descriptionKey: 'settings.terminal.tabStopWidth.description',
      type: 'number',
      min: 2,
      max: 8
    },
    {
      id: 'terminalFontFamily',
      titleKey: 'settings.terminal.fontFamily.title',
      descriptionKey: 'settings.terminal.fontFamily.description',
      type: 'select',
      options: [
        { value: 'JetBrains Mono', label: t('settings.terminal.fontFamily.options.jetbrainsMono') },
        { value: 'Fira Code', label: t('settings.terminal.fontFamily.options.firaCode') },
        { value: 'Cascadia Code', label: t('settings.terminal.fontFamily.options.cascadiaCode') },
        { value: 'SF Mono', label: t('settings.terminal.fontFamily.options.sfMono') },
        { value: 'Menlo', label: t('settings.terminal.fontFamily.options.menlo') },
        { value: 'Monaco', label: t('settings.terminal.fontFamily.options.monaco') },
        { value: 'Courier New', label: t('settings.terminal.fontFamily.options.courierNew') }
      ],
      width: 'w-48'
    },
    {
      id: 'terminalFontWeight',
      titleKey: 'settings.terminal.fontWeight.title',
      descriptionKey: 'settings.terminal.fontWeight.description',
      type: 'select',
      options: [
        { value: 'normal', label: t('settings.terminal.fontWeight.options.normal') },
        { value: 'bold', label: t('settings.terminal.fontWeight.options.bold') }
      ],
      width: 'w-32'
    },
    {
      id: 'terminalLetterSpacing',
      titleKey: 'settings.terminal.letterSpacing.title',
      descriptionKey: 'settings.terminal.letterSpacing.description',
      type: 'number',
      min: 0,
      max: 2,
      step: 0.1,
      suffix: 'px'
    }
  ];

  const hasChanges = configs.some(
    (config) => (settings as any)[config.id] !== (DEFAULT_SETTINGS as any)[config.id]
  );

  return {
    t,
    settings,
    isLoading,
    updateSetting,
    resetToDefaults,
    configs,
    hasChanges
  };
}

export function useContainerSettingsContent() {
  const { t } = useTranslation();
  const { settings, isLoading, updateSetting, resetToDefaults, DEFAULT_SETTINGS } =
    useAdvancedSettings();

  const configs: SettingConfig[] = [
    {
      id: 'containerLogTailLines',
      titleKey: 'settings.container.logTailLines.title',
      descriptionKey: 'settings.container.logTailLines.description',
      type: 'number',
      min: 50,
      max: 10000,
      step: 50
    },
    {
      id: 'containerDefaultRestartPolicy',
      titleKey: 'settings.container.defaultRestartPolicy.title',
      descriptionKey: 'settings.container.defaultRestartPolicy.description',
      type: 'select',
      options: [
        { value: 'no', label: t('settings.container.defaultRestartPolicy.options.no') },
        { value: 'always', label: t('settings.container.defaultRestartPolicy.options.always') },
        {
          value: 'on-failure',
          label: t('settings.container.defaultRestartPolicy.options.onFailure')
        },
        {
          value: 'unless-stopped',
          label: t('settings.container.defaultRestartPolicy.options.unlessStopped')
        }
      ],
      width: 'w-48'
    },
    {
      id: 'containerStopTimeout',
      titleKey: 'settings.container.stopTimeout.title',
      descriptionKey: 'settings.container.stopTimeout.description',
      type: 'number',
      min: 1,
      max: 300,
      suffix: 's'
    },
    {
      id: 'containerAutoPruneDanglingImages',
      titleKey: 'settings.container.autoPruneDanglingImages.title',
      descriptionKey: 'settings.container.autoPruneDanglingImages.description',
      type: 'switch'
    },
    {
      id: 'containerAutoPruneBuildCache',
      titleKey: 'settings.container.autoPruneBuildCache.title',
      descriptionKey: 'settings.container.autoPruneBuildCache.description',
      type: 'switch'
    }
  ];

  const hasChanges = configs.some(
    (config) => (settings as any)[config.id] !== (DEFAULT_SETTINGS as any)[config.id]
  );

  return {
    t,
    settings,
    isLoading,
    updateSetting,
    resetToDefaults,
    configs,
    hasChanges
  };
}

export function useTroubleshootingSettingsContent() {
  const { t } = useTranslation();
  const { settings, isLoading, updateSetting, resetToDefaults, DEFAULT_SETTINGS } =
    useAdvancedSettings();

  const configs: SettingConfig[] = [
    {
      id: 'debugMode',
      titleKey: 'settings.troubleshooting.debugMode.title',
      descriptionKey: 'settings.troubleshooting.debugMode.description',
      type: 'switch'
    },
    {
      id: 'showApiErrorDetails',
      titleKey: 'settings.troubleshooting.showApiErrorDetails.title',
      descriptionKey: 'settings.troubleshooting.showApiErrorDetails.description',
      type: 'switch'
    }
  ];

  const hasChanges = configs.some(
    (config) => (settings as any)[config.id] !== (DEFAULT_SETTINGS as any)[config.id]
  );

  return {
    t,
    settings,
    isLoading,
    updateSetting,
    resetToDefaults,
    configs,
    hasChanges
  };
}

export function useKeyboardShortcutsSettingsContent() {
  interface Shortcut {
    keys: string[];
    description: string;
  }

  const shortcuts: Shortcut[] = [
    { keys: ['Ctrl', 'J'], description: 'Toggle terminal' },
    { keys: ['Ctrl', 'T'], description: 'Change terminal position' },
    { keys: ['Ctrl', 'B'], description: 'Toggle sidebar' },
    { keys: ['Ctrl', 'C'], description: 'Copy file' },
    { keys: ['Ctrl', 'X'], description: 'Cut file' },
    { keys: ['Ctrl', 'V'], description: 'Paste file' },
    { keys: ['Ctrl', 'H'], description: 'Toggle hidden files' },
    { keys: ['Ctrl', 'L'], description: 'Toggle layout (grid/list)' },
    { keys: ['Ctrl', 'Shift', 'N'], description: 'Create new folder' },
    { keys: ['F2'], description: 'Rename file' }
  ];

  return { shortcuts };
}
