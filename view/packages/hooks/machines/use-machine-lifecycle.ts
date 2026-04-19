'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { toast } from 'sonner';
import {
  useGetMachineStatusQuery,
  useRestartMachineMutation,
  usePauseMachineMutation,
  useResumeMachineMutation
} from '@/redux/services/servers/serversApi';

export type LiveState = 'Running' | 'Paused' | 'Stopped' | 'unknown';

const BASE_POLL_MS = 30_000;
const MAX_POLL_MS = 5 * 60_000;

export function useMachineLifecycle(skip: boolean, serverId?: string) {
  const [consecutiveErrors, setConsecutiveErrors] = useState(0);
  const prevFetchingRef = useRef(false);

  const pollingInterval = useMemo(() => {
    if (consecutiveErrors === 0) return BASE_POLL_MS;
    return Math.min(BASE_POLL_MS * Math.pow(2, consecutiveErrors), MAX_POLL_MS);
  }, [consecutiveErrors]);

  const queryArg = serverId ? { server_id: serverId } : undefined;

  const {
    data: machineStatus,
    isLoading: isStatusLoading,
    isFetching: isStatusFetching,
    isError: isStatusError
  } = useGetMachineStatusQuery(queryArg, {
    skip,
    pollingInterval,
    refetchOnMountOrArgChange: 30
  });

  useEffect(() => {
    const wasFetching = prevFetchingRef.current;
    prevFetchingRef.current = isStatusFetching;

    if (wasFetching && !isStatusFetching && !skip) {
      if (isStatusError) {
        setConsecutiveErrors((c: number) => c + 1);
      } else {
        setConsecutiveErrors(0);
      }
    }
  }, [isStatusFetching, isStatusError, skip]);

  const [restartMachine, restartState] = useRestartMachineMutation();
  const [pauseMachine, pauseState] = usePauseMachineMutation();
  const [resumeMachine, resumeState] = useResumeMachineMutation();

  const isActionInFlight = restartState.isLoading || pauseState.isLoading || resumeState.isLoading;

  const activeAction: string | null = restartState.isLoading
    ? 'restart'
    : pauseState.isLoading
      ? 'pause'
      : resumeState.isLoading
        ? 'resume'
        : null;

  const liveState: LiveState = useMemo(() => {
    if (!machineStatus) return 'unknown';
    if (!machineStatus.active) return 'Stopped';
    return (machineStatus.state as LiveState) || 'unknown';
  }, [machineStatus]);

  const handleAction = useCallback(
    async (action: 'restart' | 'pause' | 'resume') => {
      const actionLabels = {
        restart: { ing: 'Restarting', ed: 'restarted' },
        pause: { ing: 'Pausing', ed: 'paused' },
        resume: { ing: 'Resuming', ed: 'resumed' }
      };
      const label = actionLabels[action];
      const params = serverId ? { server_id: serverId } : undefined;
      const mutationFn =
        action === 'restart'
          ? () => restartMachine(params)
          : action === 'pause'
            ? () => pauseMachine(params)
            : () => resumeMachine(params);

      try {
        await mutationFn().unwrap();
        toast.success(`Machine ${label.ed} successfully`);
      } catch (err: unknown) {
        const e = err as { data?: { detail?: string; message?: string } };
        const detail = e?.data?.detail || e?.data?.message || `Failed to ${action} machine`;
        toast.error(detail);
      }
    },
    [restartMachine, pauseMachine, resumeMachine, serverId]
  );

  return {
    liveState,
    isStatusLoading,
    isStatusFetching,
    isActionInFlight,
    activeAction,
    machineStatus,
    handleAction
  };
}
