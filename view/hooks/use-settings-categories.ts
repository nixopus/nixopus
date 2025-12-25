'use client';

import { Settings, Bell, Globe, Users, Flag, Keyboard } from 'lucide-react';
import { useRBAC } from '@/lib/rbac';
import { useFeatureFlags } from '@/hooks/features_provider';
import { FeatureNames } from '@/types/feature-flags';
import { useAppSelector } from '@/redux/hooks';

export function useSettingsCategories() {
  const { canAccessResource } = useRBAC();
  const { isFeatureEnabled } = useFeatureFlags();
  const activeOrg = useAppSelector((state) => state.user.activeOrganization);

  return [
    { id: 'general', label: 'General', icon: Settings, visible: true },
    {
      id: 'notifications',
      label: 'Notifications',
      icon: Bell,
      visible: isFeatureEnabled(FeatureNames.FeatureNotifications)
    },
    {
      id: 'domains',
      label: 'Domains',
      icon: Globe,
      visible: isFeatureEnabled(FeatureNames.FeatureDomain) && !!activeOrg?.id
    },
    { id: 'teams', label: 'Teams', icon: Users, visible: !!activeOrg?.id },
    {
      id: 'feature-flags',
      label: 'Feature Flags',
      icon: Flag,
      visible: canAccessResource('feature-flags', 'read')
    },
    { id: 'keyboard-shortcuts', label: 'Keyboard Shortcuts', icon: Keyboard, visible: true }
  ];
}
