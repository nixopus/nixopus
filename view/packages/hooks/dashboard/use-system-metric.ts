import { useMemo } from 'react';
import { SystemStatsType } from '@/redux/types/monitor';
import { useTranslation, translationKey } from '@/packages/hooks/shared/use-translation';

interface UseSystemMetricOptions<T> {
  systemStats: SystemStatsType | null;
  extractData: (stats: SystemStatsType) => T;
  defaultData: T;
}

interface UseSystemMetricResult<T> {
  data: T;
  isLoading: boolean;
  t: (key: translationKey, params?: Record<string, string>) => string;
}

/**
 * Custom hook for handling system metric components
 * Provides common logic for loading state, data extraction, and translations
 */
export function useSystemMetric<T>({
  systemStats,
  extractData,
  defaultData
}: UseSystemMetricOptions<T>): UseSystemMetricResult<T> {
  const { t } = useTranslation();
  const isLoading = !systemStats;

  const data = useMemo(() => {
    if (!systemStats) {
      return defaultData;
    }
    return extractData(systemStats);
  }, [systemStats, extractData, defaultData]);

  return {
    data,
    isLoading,
    t
  };
}
