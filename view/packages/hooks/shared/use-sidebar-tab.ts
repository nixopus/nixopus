'use client';

import { useState, useEffect, useCallback } from 'react';
import { usePathname } from 'next/navigation';

export type SidebarTab = 'chat' | 'browse';

const STORAGE_KEY = 'sidebar_active_tab';

function getStoredTab(): SidebarTab {
  if (typeof window === 'undefined') return 'chat';
  const stored = localStorage.getItem(STORAGE_KEY);
  return stored === 'browse' ? 'browse' : 'chat';
}

export function useSidebarTab() {
  const pathname = usePathname();
  const [activeTab, setActiveTabState] = useState<SidebarTab>(getStoredTab);

  const setActiveTab = useCallback((tab: SidebarTab) => {
    setActiveTabState(tab);
    localStorage.setItem(STORAGE_KEY, tab);
  }, []);

  useEffect(() => {
    if (!pathname) return;
    const isChat = pathname === '/chats' || pathname.startsWith('/chats/');
    const routeTab: SidebarTab = isChat ? 'chat' : 'browse';
    if (routeTab !== activeTab) {
      setActiveTab(routeTab);
    }
  }, [pathname]);

  return { activeTab, setActiveTab };
}
