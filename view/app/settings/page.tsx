'use client';

import { useEffect } from 'react';
import { useSettingsModal } from '@/hooks/use-settings-modal';
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
