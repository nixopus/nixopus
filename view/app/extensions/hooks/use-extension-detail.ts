import { useEffect, useState } from 'react';
import { useParams, useRouter, useSearchParams } from 'next/navigation';
import { useTranslation } from '@/hooks/use-translation';
import {
  useGetExtensionQuery,
  useRunExtensionMutation
} from '@/redux/services/extensions/extensionsApi';

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

  useEffect(() => {
    const exec = search?.get('exec');
    const openLogs = search?.get('openLogs') === '1';
    if (exec && openLogs) {
      setTab('executions');
    }
  }, [search]);

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
