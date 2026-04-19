'use client';

import { useMachines } from './use-machines';

export function useMachinesView(isLoadingProp?: boolean) {
  const machinesData = useMachines();
  const isLoading = isLoadingProp !== undefined ? isLoadingProp : machinesData.isLoading;

  return {
    ...machinesData,
    isLoading
  };
}
