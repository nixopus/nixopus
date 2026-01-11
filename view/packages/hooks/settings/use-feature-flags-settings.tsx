'use client';

import { useState, useMemo } from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useAppSelector } from '@/redux/hooks';
import {
  useGetAllFeatureFlagsQuery,
  useUpdateFeatureFlagMutation
} from '@/redux/services/feature-flags/featureFlagsApi';
import { FeatureFlag, FeatureName, featureGroups } from '@/packages/types/feature-flags';
import { toast } from 'sonner';
import { Server, Code, BarChart3, Bell, Settings } from 'lucide-react';

export function useFeatureFlagsSettings() {
  const { t } = useTranslation();
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);
  const { data: featureFlags, isLoading } = useGetAllFeatureFlagsQuery(undefined, {
    skip: !activeOrganization?.id
  });
  const [updateFeatureFlag] = useUpdateFeatureFlagMutation();
  const [searchTerm, setSearchTerm] = useState('');
  const [filterEnabled, setFilterEnabled] = useState<'all' | 'enabled' | 'disabled'>('all');

  const handleToggleFeature = async (featureName: string, isEnabled: boolean) => {
    try {
      await updateFeatureFlag({
        feature_name: featureName,
        is_enabled: isEnabled
      }).unwrap();
      toast.success(t('settings.featureFlags.messages.updated'));
    } catch (error) {
      toast.error(t('settings.featureFlags.messages.updateFailed'));
    }
  };

  const getGroupIcon = (group: string) => {
    const iconMap = {
      infrastructure: Server,
      development: Code,
      monitoring: BarChart3,
      notifications: Bell
    };
    return iconMap[group as keyof typeof iconMap] || Settings;
  };

  const filteredFeatures = useMemo(() => {
    if (!featureFlags) return [];

    return featureFlags.filter((feature) => {
      // Exclude domain and notifications features for now
      // TODO: Add them back later when we have them implemented
      if (feature.feature_name === 'domain' || feature.feature_name === 'notifications') {
        return false;
      }

      const matchesSearch =
        feature.feature_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        t(`settings.featureFlags.features.${feature.feature_name}.title` as any)
          .toLowerCase()
          .includes(searchTerm.toLowerCase());

      const matchesFilter =
        filterEnabled === 'all' ||
        (filterEnabled === 'enabled' && feature.is_enabled) ||
        (filterEnabled === 'disabled' && !feature.is_enabled);

      return matchesSearch && matchesFilter;
    });
  }, [featureFlags, searchTerm, filterEnabled, t]);

  const groupedFeatures = useMemo(() => {
    const grouped = new Map<string, FeatureFlag[]>();

    filteredFeatures.forEach((feature) => {
      for (const [group, features] of Object.entries(featureGroups)) {
        if (features.includes(feature.feature_name as FeatureName)) {
          if (!grouped.has(group)) {
            grouped.set(group, []);
          }
          grouped.get(group)?.push(feature as FeatureFlag);
          return;
        }
      }
    });
    return grouped;
  }, [filteredFeatures]);

  const totalFeatures = featureFlags?.length || 0;
  const enabledFeatures = featureFlags?.filter((f) => f.is_enabled).length || 0;
  const disabledFeatures = totalFeatures - enabledFeatures;

  return {
    featureFlags,
    isLoading,
    searchTerm,
    setSearchTerm,
    filterEnabled,
    setFilterEnabled,
    handleToggleFeature,
    getGroupIcon,
    groupedFeatures,
    totalFeatures,
    enabledFeatures,
    disabledFeatures
  };
}
