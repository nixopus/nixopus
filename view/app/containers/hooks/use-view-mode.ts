import { useState } from 'react';
import { ViewMode, CONTAINERS_VIEW_STORAGE_KEY } from '@/packages/types/containers';

export function useViewMode() {
  const [viewMode, setViewModeState] = useState<ViewMode>(() => {
    if (typeof window !== 'undefined') {
      const existing = window.localStorage.getItem(CONTAINERS_VIEW_STORAGE_KEY);
      return (existing as ViewMode) || 'table';
    }
    return 'table';
  });

  const setViewMode = (mode: ViewMode) => {
    setViewModeState(mode);
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(CONTAINERS_VIEW_STORAGE_KEY, mode);
    }
  };

  return {
    viewMode,
    setViewMode
  };
}
