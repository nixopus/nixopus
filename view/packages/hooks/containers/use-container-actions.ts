import { useMemo } from 'react';
import { isNixopusContainer } from '@/lib/utils';

export const useContainerActions = (container: any) => {
  const containerName: string = typeof container?.name === 'string' ? container.name : '';
  const isProtected = useMemo(() => isNixopusContainer(containerName), [containerName]);
  const isRunning = useMemo(
    () =>
      (container.state || '').toLowerCase() === 'running' ||
      (container.status || '').toLowerCase() === 'running',
    [container.state, container.status]
  );
  const containerId = container.id;

  return {
    containerId,
    isProtected,
    isRunning
  };
};
