'use client';

import {
  Settings,
  Bell,
  Globe,
  Users,
  Flag,
  Keyboard,
  Wifi,
  Bug,
  Terminal,
  Container
} from 'lucide-react';
import { useRBAC } from '@/packages/utils/rbac';
import { useFeatureFlags } from '@/packages/hooks/shared/features_provider';
import { FeatureNames } from '@/packages/types/feature-flags';
import { useAppSelector } from '@/redux/hooks';

export type SettingsScope = 'account' | 'organization';

export interface SettingsCategory {
  id: string;
  label: string;
  icon: typeof Settings;
  visible?: boolean;
  scope: SettingsScope;
}

export function useSettingsCategories(): SettingsCategory[] {
  const { canAccessResource } = useRBAC();
  const { isFeatureEnabled } = useFeatureFlags();
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);

  return [
    { id: 'general', label: 'General', icon: Settings, visible: true, scope: 'account' },
    {
      id: 'notifications',
      label: 'Notifications',
      icon: Bell,
      visible: isFeatureEnabled(FeatureNames.FeatureNotifications),
      scope: 'organization'
    },
    {
      id: 'domains',
      label: 'Domains',
      icon: Globe,
      visible: isFeatureEnabled(FeatureNames.FeatureDomain) && !!activeOrg?.id,
      scope: 'organization'
    },
    { id: 'teams', label: 'Teams', icon: Users, visible: !!activeOrg?.id, scope: 'organization' },
    {
      id: 'feature-flags',
      label: 'Feature Flags',
      icon: Flag,
      visible: canAccessResource('feature-flags', 'read'),
      scope: 'organization'
    },
    {
      id: 'keyboard-shortcuts',
      label: 'Keyboard Shortcuts',
      icon: Keyboard,
      visible: true,
      scope: 'account'
    },
    { id: 'network', label: 'Network', icon: Wifi, visible: true, scope: 'organization' },
    { id: 'terminal', label: 'Terminal', icon: Terminal, visible: true, scope: 'account' },
    { id: 'container', label: 'Container', icon: Container, visible: true, scope: 'organization' },
    { id: 'troubleshooting', label: 'Troubleshooting', icon: Bug, visible: true, scope: 'account' }
  ];
}
