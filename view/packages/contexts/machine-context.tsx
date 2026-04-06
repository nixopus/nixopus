'use client';
import { createContext, useContext, useEffect, useMemo, useRef } from 'react';
import { usePathname } from 'next/navigation';
import { useGetServersQuery } from '@/redux/services/servers/serversApi';
import { useAppDispatch } from '@/redux/hooks';
import { deployApi } from '@/redux/services/deploy/applicationsApi';
import { containerApi } from '@/redux/services/container/containerApi';
import { imagesApi } from '@/redux/services/container/imagesApi';
import { fileManagersApi } from '@/redux/services/file-manager/fileManagersApi';

const PLUGIN_MACHINE_APIS = ['machineLifecycleApi', 'machineBackupApi', 'machineBillingApi'];

interface MachineContextValue {
  machineId: string | null;
  isExplicit: boolean;
}

const MachineContext = createContext<MachineContextValue>({
  machineId: null,
  isExplicit: false
});

interface MachineProviderProps {
  machineId?: string;
  children: React.ReactNode;
}

export function MachineProvider({ machineId, children }: MachineProviderProps) {
  const pathname = usePathname();
  const dispatch = useAppDispatch();

  const urlMachineId = useMemo(() => {
    const match = pathname.match(/^\/machines\/([^/]+)/);
    return match ? match[1] : null;
  }, [pathname]);

  const explicitId = machineId ?? urlMachineId;
  const isExplicit = !!explicitId;
  const { data } = useGetServersQuery({ page: 1, page_size: 1 }, { skip: isExplicit });

  const resolvedId = isExplicit ? explicitId! : (data?.servers?.[0]?.id ?? null);

  const prevMachineIdRef = useRef<string | null>(resolvedId);
  useEffect(() => {
    if (resolvedId && resolvedId !== prevMachineIdRef.current) {
      prevMachineIdRef.current = resolvedId;
      dispatch(deployApi.util.resetApiState());
      dispatch(containerApi.util.resetApiState());
      dispatch(imagesApi.util.resetApiState());
      dispatch(fileManagersApi.util.resetApiState());
      PLUGIN_MACHINE_APIS.forEach((path) => dispatch({ type: `${path}/resetApiState` }));
    }
  }, [resolvedId, dispatch]);

  const value = useMemo(() => ({ machineId: resolvedId, isExplicit }), [resolvedId, isExplicit]);

  return <MachineContext.Provider value={value}>{children}</MachineContext.Provider>;
}

export function useMachineId(): string | null {
  return useContext(MachineContext).machineId;
}

export function useMachineContext(): MachineContextValue {
  return useContext(MachineContext);
}
