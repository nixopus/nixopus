import { useMemo, useState, useCallback } from 'react';
import { Application } from '@/redux/types/applications';

export interface LabelFilterState {
  selectedLabels: string[];
  availableLabels: string[];
  filteredApplications: Application[];
}

export interface LabelFilterActions {
  toggleLabel: (label: string) => void;
  clearFilters: () => void;
  hasActiveFilters: boolean;
}

export type UseLabelFilterReturn = LabelFilterState & LabelFilterActions;

export function useLabelFilter(applications: Application[]): UseLabelFilterReturn {
  const [selectedLabels, setSelectedLabels] = useState<string[]>([]);

  const availableLabels = useMemo(() => extractUniqueLabels(applications), [applications]);

  const toggleLabel = useCallback((label: string) => {
    setSelectedLabels((prev) => toggleLabelInArray(prev, label));
  }, []);

  const clearFilters = useCallback(() => {
    setSelectedLabels([]);
  }, []);

  const filteredApplications = useMemo(
    () => filterByLabels(applications, selectedLabels),
    [applications, selectedLabels]
  );

  const hasActiveFilters = selectedLabels.length > 0;

  return {
    selectedLabels,
    availableLabels,
    filteredApplications,
    toggleLabel,
    clearFilters,
    hasActiveFilters
  };
}

function extractUniqueLabels(apps: Application[]): string[] {
  const labelsSet = new Set<string>();

  apps.forEach((app) => {
    app.labels?.forEach((label) => labelsSet.add(label));
  });

  return Array.from(labelsSet).sort();
}

function toggleLabelInArray(labels: string[], label: string): string[] {
  return labels.includes(label) ? labels.filter((l) => l !== label) : [...labels, label];
}

function filterByLabels(apps: Application[], selectedLabels: string[]): Application[] {
  if (selectedLabels.length === 0) return apps;

  return apps.filter((app) => hasMatchingLabels(app, selectedLabels));
}

function hasMatchingLabels(app: Application, selectedLabels: string[]): boolean {
  if (!app.labels?.length) return false;

  return selectedLabels.some((label) => app.labels?.includes(label));
}
