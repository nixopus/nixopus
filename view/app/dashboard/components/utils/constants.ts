// Color constants for charts and visualizations
export const CHART_COLORS = {
  blue: '#3b82f6',
  green: '#10b981',
  orange: '#f59e0b',
  red: '#ef4444',
  purple: '#a855f7',
  yellow: '#eab308'
} as const;

// Default values for system metrics
export const DEFAULT_METRICS = {
  load: {
    oneMin: 0 as number,
    fiveMin: 0 as number,
    fifteenMin: 0 as number
  },
  cpu: {
    overall: 0 as number,
    per_core: [] as Array<{ core_id: number; usage: number }>
  },
  memory: {
    total: 0 as number,
    used: 0 as number,
    percentage: 0 as number
  },
  disk: {
    percentage: 0 as number,
    used: 0 as number,
    total: 0 as number,
    allMounts: [] as any[]
  }
};
