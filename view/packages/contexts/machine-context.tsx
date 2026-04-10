'use client';
import { createContext, useContext, useMemo } from 'react';
import { useGetServersQuery } from '@/redux/services/servers/serversApi';

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
  const isExplicit = !!machineId;
  const { data } = useGetServersQuery({ page: 1, page_size: 1 }, { skip: isExplicit });

  const resolvedId = isExplicit ? machineId! : (data?.servers?.[0]?.id ?? null);

  const value = useMemo(() => ({ machineId: resolvedId, isExplicit }), [resolvedId, isExplicit]);

  return <MachineContext.Provider value={value}>{children}</MachineContext.Provider>;
}

export function useMachineId(): string | null {
  return useContext(MachineContext).machineId;
}

export function useMachineContext(): MachineContextValue {
  return useContext(MachineContext);
}
