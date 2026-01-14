'use client';

import { useEffect } from 'react';
import { useSettingsModal } from '@/packages/hooks/shared/use-settings-modal';
import { useRouter } from 'next/navigation';

function Page() {
  const { openSettings } = useSettingsModal();
  const router = useRouter();

  useEffect(() => {
    openSettings('general');
    router.replace('/dashboard');
  }, [openSettings, router]);

  return null;
}

export default Page;
