'use client';
import { createContext, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { usePathname } from 'next/navigation';
import { useGetServersQuery, useLazyGetServersQuery } from '@/redux/services/servers/serversApi';
import { useAppDispatch, useAppSelector } from '@/redux/hooks';
import {
  getLastSelectedMachineId,
  setLastSelectedMachineId
} from '@/packages/utils/last-selected-machine-id';
import { deployApi } from '@/redux/services/deploy/applicationsApi';
import { containerApi } from '@/redux/services/container/containerApi';
import { imagesApi } from '@/redux/services/container/imagesApi';
import { fileManagersApi } from '@/redux/services/file-manager/fileManagersApi';
import { machineBackupApi } from '@/redux/services/machine/machineBackupApi';

const PLUGIN_MACHINE_APIS = ['machineLifecycleApi', 'machineBillingApi'];

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
  const orgId = useAppSelector((state) => state.user.activeOrganization?.id);

  const urlMachineId = useMemo(() => {
    const match = pathname.match(/^\/machines\/([^/]+)/);
    return match ? match[1] : null;
  }, [pathname]);

  const explicitId = machineId ?? urlMachineId;
  const isExplicit = !!explicitId;
  const { data } = useGetServersQuery({ page: 1, page_size: 100 }, { skip: isExplicit });
  const [fetchServers] = useLazyGetServersQuery();
  const [verifiedLastId, setVerifiedLastId] = useState<string | null>(null);

  // Verify persisted last-selected machine ID by checking additional pages if needed
  useEffect(() => {
    if (isExplicit || !orgId || !data) return;
    const last = getLastSelectedMachineId(orgId);
    if (!last) {
      setVerifiedLastId(null);
      return;
    }
    // Check if the persisted ID is in the first page
    const list = data.servers ?? [];
    if (list.some((s) => s.id === last)) {
      setVerifiedLastId(last);
      return;
    }
    // If not in first page and there are more results, verify by fetching with higher page_size
    if (data.total_count && data.total_count > 100) {
      fetchServers({ page: 1, page_size: data.total_count })
        .unwrap()
        .then((response) => {
          const found = response.servers?.some((s) => s.id === last);
          setVerifiedLastId(found ? last : null);
        })
        .catch(() => {
          // On error, fall back to null
          setVerifiedLastId(null);
        });
    } else {
      // No more pages to check, ID doesn't exist
      setVerifiedLastId(null);
    }
  }, [isExplicit, orgId, data, fetchServers]);

  const resolvedId = useMemo(() => {
    if (isExplicit) {
      return explicitId!;
    }
    const list = data?.servers ?? [];
    if (list.length === 0) {
      return null;
    }
    // Use verified last ID if available
    if (verifiedLastId) {
      return verifiedLastId;
    }
    return list[0]?.id ?? null;
  }, [isExplicit, explicitId, data?.servers, verifiedLastId]);

  // Track the last resolved machine ID so we can distinguish:
  //   - initial null -> first real value (just initialization, do NOT wipe caches)
  //   - real-A -> real-B (genuine machine switch, wipe machine-scoped caches)
  // Wiping on initialization would clear caches that other pages just populated,
  // causing skeletons to flash even though data was already fetched.
  const prevMachineIdRef = useRef<string | null>(null);
  useEffect(() => {
    if (!resolvedId) return;
    const prev = prevMachineIdRef.current;
    prevMachineIdRef.current = resolvedId;
    if (prev === null || prev === resolvedId) return;
    dispatch(deployApi.util.resetApiState());
    dispatch(containerApi.util.resetApiState());
    dispatch(imagesApi.util.resetApiState());
    dispatch(fileManagersApi.util.resetApiState());
    dispatch(machineBackupApi.util.resetApiState());
    PLUGIN_MACHINE_APIS.forEach((path) => dispatch({ type: `${path}/resetApiState` }));
  }, [resolvedId, dispatch]);

  useEffect(() => {
    if (urlMachineId && orgId) {
      setLastSelectedMachineId(orgId, urlMachineId);
    }
  }, [urlMachineId, orgId]);

  const value = useMemo(() => ({ machineId: resolvedId, isExplicit }), [resolvedId, isExplicit]);

  return <MachineContext.Provider value={value}>{children}</MachineContext.Provider>;
}

export function useMachineId(): string | null {
  return useContext(MachineContext).machineId;
}

export function useMachineContext(): MachineContextValue {
  return useContext(MachineContext);
}