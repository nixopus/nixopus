import { CHART_COLORS } from '@/packages/utils/dashboard';

export interface LoadAverageItem {
  label: string;
  value: number;
  color: string;
}

export interface LoadAverageData {
  oneMin: number;
  fiveMin: number;
  fifteenMin: number;
}

export function useLoadAverageItems(load: LoadAverageData): LoadAverageItem[] {
  return [
    {
      label: '1 min',
      value: load.oneMin,
      color: CHART_COLORS.blue
    },
    {
      label: '5 min',
      value: load.fiveMin,
      color: CHART_COLORS.green
    },
    {
      label: '15 min',
      value: load.fifteenMin,
      color: CHART_COLORS.orange
    }
  ];
}
