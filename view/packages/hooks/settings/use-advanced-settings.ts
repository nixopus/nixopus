'use client';

import { useState, useEffect, useCallback, useMemo } from 'react';
import { toast } from 'sonner';
import { useTranslation, type translationKey } from '@/packages/hooks/shared/use-translation';
import {
  useGetUserPreferencesQuery,
  useUpdateUserPreferencesMutation,
  useGetOrganizationSettingsQuery,
  useUpdateOrganizationSettingsMutation
} from '@/redux/services/users/userApi';
import { UserPreferencesData, OrganizationSettingsData } from '@/redux/types/user';
import {
  AdvancedSettings,
  DEFAULT_SETTINGS,
  saveSettingsToStorage,
  getAdvancedSettings
} from '@/packages/utils/advanced-settings';

// Re-export for backward compatibility
export type { AdvancedSettings };
export { DEFAULT_SETTINGS, getAdvancedSettings };

export function useAdvancedSettings() {
  const { t } = useTranslation();

  const {
    data: userPreferences,
    isLoading: isLoadingUserPrefs,
    refetch: refetchUserPrefs
  } = useGetUserPreferencesQuery();
  const [updateUserPreferences] = useUpdateUserPreferencesMutation();

  const {
    data: orgSettings,
    isLoading: isLoadingOrgSettings,
    refetch: refetchOrgSettings
  } = useGetOrganizationSettingsQuery();
  const [updateOrgSettings] = useUpdateOrganizationSettingsMutation();

  const isLoading = isLoadingUserPrefs || isLoadingOrgSettings;

  const settings: AdvancedSettings = useMemo(() => {
    const merged: AdvancedSettings = {
      websocketReconnectAttempts:
        orgSettings?.settings?.websocket_reconnect_attempts ??
        DEFAULT_SETTINGS.websocketReconnectAttempts,
      websocketReconnectInterval:
        orgSettings?.settings?.websocket_reconnect_interval ??
        DEFAULT_SETTINGS.websocketReconnectInterval,
      apiRetryAttempts:
        orgSettings?.settings?.api_retry_attempts ?? DEFAULT_SETTINGS.apiRetryAttempts,
      disableApiCache: orgSettings?.settings?.disable_api_cache ?? DEFAULT_SETTINGS.disableApiCache,

      debugMode: userPreferences?.preferences?.debug_mode ?? DEFAULT_SETTINGS.debugMode,
      showApiErrorDetails:
        userPreferences?.preferences?.show_api_error_details ??
        DEFAULT_SETTINGS.showApiErrorDetails,

      terminalScrollback:
        userPreferences?.preferences?.terminal_scrollback ?? DEFAULT_SETTINGS.terminalScrollback,
      terminalFontSize:
        userPreferences?.preferences?.terminal_font_size ?? DEFAULT_SETTINGS.terminalFontSize,
      terminalCursorStyle:
        userPreferences?.preferences?.terminal_cursor_style ?? DEFAULT_SETTINGS.terminalCursorStyle,
      terminalCursorBlink:
        userPreferences?.preferences?.terminal_cursor_blink ?? DEFAULT_SETTINGS.terminalCursorBlink,
      terminalLineHeight:
        userPreferences?.preferences?.terminal_line_height ?? DEFAULT_SETTINGS.terminalLineHeight,
      terminalCursorWidth:
        userPreferences?.preferences?.terminal_cursor_width ?? DEFAULT_SETTINGS.terminalCursorWidth,
      terminalTabStopWidth:
        userPreferences?.preferences?.terminal_tab_stop_width ??
        DEFAULT_SETTINGS.terminalTabStopWidth,
      terminalFontFamily:
        userPreferences?.preferences?.terminal_font_family ?? DEFAULT_SETTINGS.terminalFontFamily,
      terminalFontWeight:
        userPreferences?.preferences?.terminal_font_weight ?? DEFAULT_SETTINGS.terminalFontWeight,
      terminalLetterSpacing:
        userPreferences?.preferences?.terminal_letter_spacing ??
        DEFAULT_SETTINGS.terminalLetterSpacing,

      containerLogTailLines:
        orgSettings?.settings?.container_log_tail_lines ?? DEFAULT_SETTINGS.containerLogTailLines,
      containerDefaultRestartPolicy:
        orgSettings?.settings?.container_default_restart_policy ??
        DEFAULT_SETTINGS.containerDefaultRestartPolicy,
      containerStopTimeout:
        orgSettings?.settings?.container_stop_timeout ?? DEFAULT_SETTINGS.containerStopTimeout,
      containerAutoPruneDanglingImages:
        orgSettings?.settings?.container_auto_prune_dangling_images ??
        DEFAULT_SETTINGS.containerAutoPruneDanglingImages,
      containerAutoPruneBuildCache:
        orgSettings?.settings?.container_auto_prune_build_cache ??
        DEFAULT_SETTINGS.containerAutoPruneBuildCache
    };

    return merged;
  }, [userPreferences, orgSettings]);

  // Persist settings to storage when they change
  useEffect(() => {
    saveSettingsToStorage(settings);
  }, [settings]);

  const updateSetting = useCallback(
    async <K extends keyof AdvancedSettings>(key: K, value: AdvancedSettings[K]) => {
      try {
        const userPrefKeys: (keyof AdvancedSettings)[] = [
          'debugMode',
          'showApiErrorDetails',
          'terminalScrollback',
          'terminalFontSize',
          'terminalCursorStyle',
          'terminalCursorBlink',
          'terminalLineHeight',
          'terminalCursorWidth',
          'terminalTabStopWidth',
          'terminalFontFamily',
          'terminalFontWeight',
          'terminalLetterSpacing'
        ];
        const orgSettingKeys: (keyof AdvancedSettings)[] = [
          'websocketReconnectAttempts',
          'websocketReconnectInterval',
          'apiRetryAttempts',
          'disableApiCache',
          'containerLogTailLines',
          'containerDefaultRestartPolicy',
          'containerStopTimeout',
          'containerAutoPruneDanglingImages',
          'containerAutoPruneBuildCache'
        ];

        if (userPrefKeys.includes(key)) {
          const newPrefs: UserPreferencesData = {
            debug_mode: key === 'debugMode' ? (value as boolean) : settings.debugMode,
            show_api_error_details:
              key === 'showApiErrorDetails' ? (value as boolean) : settings.showApiErrorDetails,
            terminal_scrollback:
              key === 'terminalScrollback' ? (value as number) : settings.terminalScrollback,
            terminal_font_size:
              key === 'terminalFontSize' ? (value as number) : settings.terminalFontSize,
            terminal_cursor_style:
              key === 'terminalCursorStyle'
                ? (value as 'bar' | 'block' | 'underline')
                : settings.terminalCursorStyle,
            terminal_cursor_blink:
              key === 'terminalCursorBlink' ? (value as boolean) : settings.terminalCursorBlink,
            terminal_line_height:
              key === 'terminalLineHeight' ? (value as number) : settings.terminalLineHeight,
            terminal_cursor_width:
              key === 'terminalCursorWidth' ? (value as number) : settings.terminalCursorWidth,
            terminal_tab_stop_width:
              key === 'terminalTabStopWidth' ? (value as number) : settings.terminalTabStopWidth,
            terminal_font_family:
              key === 'terminalFontFamily' ? (value as string) : settings.terminalFontFamily,
            terminal_font_weight:
              key === 'terminalFontWeight'
                ? (value as 'normal' | 'bold')
                : settings.terminalFontWeight,
            terminal_letter_spacing:
              key === 'terminalLetterSpacing' ? (value as number) : settings.terminalLetterSpacing
          };
          await updateUserPreferences(newPrefs).unwrap();
          await refetchUserPrefs();
        } else if (orgSettingKeys.includes(key)) {
          const newSettings: OrganizationSettingsData = {
            websocket_reconnect_attempts:
              key === 'websocketReconnectAttempts'
                ? (value as number)
                : settings.websocketReconnectAttempts,
            websocket_reconnect_interval:
              key === 'websocketReconnectInterval'
                ? (value as number)
                : settings.websocketReconnectInterval,
            api_retry_attempts:
              key === 'apiRetryAttempts' ? (value as number) : settings.apiRetryAttempts,
            disable_api_cache:
              key === 'disableApiCache' ? (value as boolean) : settings.disableApiCache,
            container_log_tail_lines:
              key === 'containerLogTailLines' ? (value as number) : settings.containerLogTailLines,
            container_default_restart_policy:
              key === 'containerDefaultRestartPolicy'
                ? (value as 'no' | 'always' | 'on-failure' | 'unless-stopped')
                : settings.containerDefaultRestartPolicy,
            container_stop_timeout:
              key === 'containerStopTimeout' ? (value as number) : settings.containerStopTimeout,
            container_auto_prune_dangling_images:
              key === 'containerAutoPruneDanglingImages'
                ? (value as boolean)
                : settings.containerAutoPruneDanglingImages,
            container_auto_prune_build_cache:
              key === 'containerAutoPruneBuildCache'
                ? (value as boolean)
                : settings.containerAutoPruneBuildCache
          };
          await updateOrgSettings(newSettings).unwrap();
          await refetchOrgSettings();
        }

        // Storage will be updated by useEffect when refetch completes
        toast.success(t('settings.network.messages.settingUpdated' as translationKey));
      } catch (error) {
        console.error('Failed to update setting:', error);
        toast.error(t('settings.network.messages.settingUpdateFailed' as translationKey));
      }
    },
    [settings, updateUserPreferences, updateOrgSettings, refetchUserPrefs, refetchOrgSettings, t]
  );

  const resetToDefaults = useCallback(async () => {
    try {
      await updateUserPreferences({
        debug_mode: DEFAULT_SETTINGS.debugMode,
        show_api_error_details: DEFAULT_SETTINGS.showApiErrorDetails,
        terminal_scrollback: DEFAULT_SETTINGS.terminalScrollback,
        terminal_font_size: DEFAULT_SETTINGS.terminalFontSize,
        terminal_cursor_style: DEFAULT_SETTINGS.terminalCursorStyle,
        terminal_cursor_blink: DEFAULT_SETTINGS.terminalCursorBlink,
        terminal_line_height: DEFAULT_SETTINGS.terminalLineHeight,
        terminal_cursor_width: DEFAULT_SETTINGS.terminalCursorWidth,
        terminal_tab_stop_width: DEFAULT_SETTINGS.terminalTabStopWidth,
        terminal_font_family: DEFAULT_SETTINGS.terminalFontFamily,
        terminal_font_weight: DEFAULT_SETTINGS.terminalFontWeight,
        terminal_letter_spacing: DEFAULT_SETTINGS.terminalLetterSpacing
      }).unwrap();

      await updateOrgSettings({
        websocket_reconnect_attempts: DEFAULT_SETTINGS.websocketReconnectAttempts,
        websocket_reconnect_interval: DEFAULT_SETTINGS.websocketReconnectInterval,
        api_retry_attempts: DEFAULT_SETTINGS.apiRetryAttempts,
        disable_api_cache: DEFAULT_SETTINGS.disableApiCache,
        container_log_tail_lines: DEFAULT_SETTINGS.containerLogTailLines,
        container_default_restart_policy: DEFAULT_SETTINGS.containerDefaultRestartPolicy,
        container_stop_timeout: DEFAULT_SETTINGS.containerStopTimeout,
        container_auto_prune_dangling_images: DEFAULT_SETTINGS.containerAutoPruneDanglingImages,
        container_auto_prune_build_cache: DEFAULT_SETTINGS.containerAutoPruneBuildCache
      }).unwrap();

      // Storage will be updated by useEffect when refetch completes
      await refetchUserPrefs();
      await refetchOrgSettings();

      toast.success(t('settings.network.messages.resetSuccess' as translationKey));
    } catch (error) {
      console.error('Failed to reset settings:', error);
      toast.error(t('settings.network.messages.resetFailed' as translationKey));
    }
  }, [updateUserPreferences, updateOrgSettings, refetchUserPrefs, refetchOrgSettings, t]);

  return {
    settings,
    isLoading,
    updateSetting,
    resetToDefaults,
    DEFAULT_SETTINGS
  };
}
