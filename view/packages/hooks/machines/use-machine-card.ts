'use client';

import type { Machine } from './use-machines';

export function useMachineCard(machine: Machine) {
  const isProvisioning = machine.status === 'provisioning';
  const isFailed = machine.status === 'failed';

  return {
    isProvisioning,
    isFailed,
    isInteractive: !isProvisioning && !isFailed
  };
}
