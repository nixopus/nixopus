import React, { createContext, useContext, useMemo } from 'react';
import { useGetAllFeatureFlagsQuery } from '@/redux/services/feature-flags/featureFlagsApi';
import type { FeatureFlag } from '@/types/feature-flags';

interface FeatureFlagsContextType {
  features: FeatureFlag[];
  isLoading: boolean;
  error: any;
  isFeatureEnabled: (feature: string) => boolean;
}

const FeatureFlagsContext = createContext<FeatureFlagsContextType>({
  features: [],
  isLoading: false,
  error: null,
  isFeatureEnabled: () => false
});

export const FeatureFlagsProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { data: features = [], isLoading, error } = useGetAllFeatureFlagsQuery();

  const isFeatureEnabled = useMemo(() => {
    return (featureName: string): boolean => {
      const feature = features.find((f) => f.feature_name === featureName);
      return feature?.is_enabled || false;
    };
  }, [features]);

  const value = useMemo(
    () => ({
      features,
      isLoading,
      error,
      isFeatureEnabled
    }),
    [features, isLoading, error, isFeatureEnabled]
  );

  return <FeatureFlagsContext.Provider value={value}>{children}</FeatureFlagsContext.Provider>;
};

export const useFeatureFlags = () => useContext(FeatureFlagsContext);
