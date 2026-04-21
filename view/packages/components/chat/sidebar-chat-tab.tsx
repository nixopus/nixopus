'use client';

import React from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { Plus } from 'lucide-react';
import { useChatThreads } from '@/packages/hooks/ai/use-chat-threads';
import { ThreadSidebar } from '@/packages/components/chat/thread-sidebar';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useSidebar, SidebarMenuButton } from '@/components/ui/sidebar';

export function SidebarChatTab() {
  const { state } = useSidebar();
  const isCollapsed = state === 'collapsed';
  const router = useRouter();
  const pathname = usePathname();
  const { t } = useTranslation();
  const threads = useChatThreads();

  const navigateToChats = () => {
    if (!pathname.startsWith('/chats')) {
      router.push('/chats');
    }
  };

  const handleSelectThread = (id: string) => {
    threads.setActiveThreadId(id);
    navigateToChats();
  };

  const handleNewChat = () => {
    threads.createThread(t('ai.threads.untitledChat'));
    navigateToChats();
  };

  if (isCollapsed) {
    return (
      <div className="flex flex-col items-center py-2 gap-1">
        <SidebarMenuButton tooltip="New Chat" onClick={handleNewChat} className="cursor-pointer">
          <Plus className="size-4" />
        </SidebarMenuButton>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full overflow-hidden px-1">
      <ThreadSidebar
        threads={threads.threads}
        activeThreadId={threads.activeThreadId}
        resourceId={threads.resourceId}
        isLoading={!threads.isInitialized}
        isCollapsed={false}
        onToggleCollapse={() => {}}
        onSelectThread={handleSelectThread}
        onNewChat={handleNewChat}
        onDeleteThread={threads.deleteThread}
        onRenameThread={threads.updateThreadTitle}
        onRefresh={threads.refreshThreads}
        isRefreshing={threads.isRefreshing}
        hideCollapseToggle
      />
    </div>
  );
}
