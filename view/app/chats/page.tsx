'use client';

import React from 'react';
import dynamic from 'next/dynamic';
import { Skeleton } from '@nixopus/ui';

const ChatPage = dynamic(
  () => import('@/packages/components/ai-chat').then((m) => ({ default: m.ChatPage })),
  {
    ssr: false,
    loading: () => <Skeleton className="h-full w-full" />
  }
);

export default function ChatsPage() {
  return (
    <div className="h-[calc(100vh-4rem)] w-full">
      <ChatPage />
    </div>
  );
}
