import React, { createContext, useContext, useMemo, useEffect } from 'react';
import { useGetAllFeatureFlagsQuery } from '@/redux/services/feature-flags/featureFlagsApi';
import { useAppSelector } from '@/redux/hooks';
import type { FeatureFlag } from '@/packages/types/feature-flags';

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
  const { isAuthenticated, isInitialized } = useAppSelector((state) => state.auth);
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);

  const {
    data: features = [],
    isLoading,
    error,
    refetch
  } = useGetAllFeatureFlagsQuery(undefined, {
    skip: !isAuthenticated || !isInitialized
  });

  useEffect(() => {
    if (isAuthenticated && isInitialized) {
      refetch();
    }
  }, [isAuthenticated, isInitialized, activeOrganization?.id, refetch]);

  const isFeatureEnabled = useMemo(() => {
    return (featureName: string): boolean => {
      const feature = features.find((f) => f.feature_name === featureName);
      return feature?.is_enabled || false;
    };
  }, [features]);

  const value = useMemo(
    () => ({
      features: isAuthenticated ? features : [],
      isLoading: isAuthenticated ? isLoading : false,
      error: isAuthenticated ? error : null,
      isFeatureEnabled
    }),
    [features, isLoading, error, isFeatureEnabled, isAuthenticated]
  );

  return <FeatureFlagsContext.Provider value={value}>{children}</FeatureFlagsContext.Provider>;
};

export const useFeatureFlags = () => useContext(FeatureFlagsContext);
