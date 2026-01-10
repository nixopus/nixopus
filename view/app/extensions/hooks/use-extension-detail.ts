import { useEffect, useState } from 'react';
import { useParams, useRouter, useSearchParams } from 'next/navigation';
import { useTranslation } from '@/hooks/use-translation';
import {
  useGetExtensionQuery,
  useRunExtensionMutation
} from '@/redux/services/extensions/extensionsApi';
import { useListExecutionsQuery } from '@/redux/services/extensions/extensionsApi';
import { useRef } from 'react';

function useExtensionDetails() {
  const { t } = useTranslation();
  const params = useParams();
  const search = useSearchParams();
  const router = useRouter();
  const id = (params?.id as string) || '';

  const { data: extension, isLoading } = useGetExtensionQuery({ id });
  const [tab, setTab] = useState<string>('overview');
  const [runModalOpen, setRunModalOpen] = useState(false);
  const [runExtension, { isLoading: isRunning }] = useRunExtensionMutation();
  const { data: executions, isLoading: isExecsLoading } = useListExecutionsQuery(
    { extensionId: id },
    { skip: !id }
  );
  const initializedDefaultTab = useRef(false);

  useEffect(() => {
    const exec = search?.get('exec');
    const openLogs = search?.get('openLogs') === '1';
    if (exec && openLogs) {
      setTab('executions');
      initializedDefaultTab.current = true;
    }
  }, [search]);

  useEffect(() => {
    if (initializedDefaultTab.current) return;
    if (isExecsLoading) return;
    if (!executions) return;
    setTab(executions.length > 0 ? 'executions' : 'overview');
    initializedDefaultTab.current = true;
  }, [executions, isExecsLoading]);

  return {
    runModalOpen,
    runExtension,
    isRunning,
    isLoading,
    tab,
    extension,
    router,
    setRunModalOpen,
    t,
    setTab
  };
}

export default useExtensionDetails;
